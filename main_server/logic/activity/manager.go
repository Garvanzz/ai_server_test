package activity

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"xfx/core/cache"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/fsm"
	"xfx/core/model"
	"xfx/main_server/logic/activity/data"
	"xfx/main_server/logic/activity/impl"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/gogo/protobuf/proto"
)

var Mgr *Manager

var Module = func() module.Module {
	return Mgr
}

func init() {
	Mgr = new(Manager)
}

type Manager struct {
	modules.BaseModule
	entities sync.Map
	sm       *fsm.StateMachine
	lastTick int64
}

func (m *Manager) OnInit(app module.App) {
	m.BaseModule.OnInit(app)
	m.sm = fsm.NewStateMachine(&fsm.DefaultDelegate{P: m}, transitions...)
	data.Cache = cache.New[int64, any](cache.Options[int64, any]{
		Capacity:      10000,
		DefaultTTL:    5 * time.Minute,
		FlushInterval: 30 * time.Second,
		SaveFunc:      data.SavePlayerData,
	})
	data.ServerId = m.App.GetEnv().ID

	activities, err := data.LoadAllActivityData()
	if err != nil {
		panic("activity load activity data error")
		return
	}

	existIds := make(map[int64]struct{})

	for _, activity := range activities {
		ent := new(entity)
		ent.Id = activity.Id
		ent.CfgId = activity.CfgId
		ent.Type = activity.Type
		ent.State = activity.State
		ent.StartTime = activity.StartTime
		ent.EndTime = activity.EndTime
		ent.CloseTime = activity.CloseTime
		ent.TimeType = activity.TimeType
		ent.TimeValue = activity.TimeValue
		ent.mod = Mgr

		desc := impl.GetActivityDesc(ent.Type)
		if desc == nil || desc.NewHandler == nil {
			panic(fmt.Sprintf("missing activity handler: %v", ent.Type))
		}

		ent.handler = desc.NewHandler()
		ent.handler.SetBaseInfo(ent)

		if desc.InjectFunc != nil {
			desc.InjectFunc(ent.handler, impl.UnmarshalActivityData(activity))
		}

		ent.handler.OnInit()

		existIds[ent.CfgId] = struct{}{}
		log.Info("初始加载活动:%v", activity.CfgId)
		m.entities.Store(ent.Id, ent)
	}

	// 根据配置加载新活动
	activityConfs := impl.GetAllCommonConf()
	for _, activityConf := range activityConfs {
		if _, ok := existIds[activityConf.Id]; !ok {
			log.Info("注册新活动:%v", activityConf.Id)
			ent := m.register(activityConf.Id)
			m.entities.Store(ent.Id, ent)
		}
	}

	// register func
	m.Register("GetActivityStatus", m.OnGetActivityStatus)
	m.Register("GetActivityStatusByType", m.OnGetActivityStatusByType)
	m.Register("GetActivityData", m.OnGetActivityData)
	m.Register("GetActivityDataList", m.OnGetActivityDataList)
	m.Register("OnRouterMsg", m.OnRouterMsg)
}

func (m *Manager) OnStart(ctx module.Context) {
	m.BaseModule.OnStart(ctx)
	event.AddEventListener(define.EventTypeActivity, m.Self())
	event.AddEventListener(define.EventTypePlayerOnline, m.Self())
	event.AddEventListener(define.EventTypePlayerOffline, m.Self())
}

func (m *Manager) GetType() string { return define.ModuleActivity }

func (m *Manager) OnTick(delta time.Duration) {
	now := time.Now()
	if m.lastTick == 0 {
		m.lastTick = now.Unix()
	} else {
		if now.Unix()-m.lastTick >= 60*5 {
			m.saveData()
			m.lastTick = now.Unix()
		}
	}

	m.entities.Range(func(key, value any) bool {
		ent := value.(*entity)

		// 检查配置是否变动
		if !ent.checked {
			eventStr := ent.determineStateFromConfig(m.App.GetEnv().ID)
			if eventStr != EventNone {
				err := m.sm.Trigger(ent.State, eventStr, ent)
				if err != nil {
					log.Error("%v", err)
				}
			}
			ent.checked = true
			return true
		}

		if ent.State == StateStopped || ent.State == StateClosed {
			return true
		}

		// 状态转换处理
		triggerEvent := ent.checkState()
		if triggerEvent != EventNone {
			err := m.sm.Trigger(ent.State, triggerEvent, ent)
			if err != nil {
				log.Error("sm trigger error:%v", err)
				return true
			}
		}

		// 活动业务更新
		if ent.State == StateRunning {
			ent.handler.Update(now)
		}

		return true
	})
}

// OnEvent 事件回调
func (m *Manager) OnEvent(event *event.Event) {
	if event == nil {
		return
	}

	if event.M == nil {
		return
	}

	// 玩家基础信息
	ctx, ok := event.M["player"].(*proto_player.Context)
	if !ok {
		log.Error("activity event find no player data")
		return
	}

	switch event.Type {
	case define.EventTypeActivity:
		m.notify(ctx, event.M)
	case define.EventTypePlayerOffline:
	case define.EventTypePlayerOnline:
		event.M["key"] = "player_online"
		m.notify(ctx, event.M)
	}
}

func (m *Manager) OnMessage(msg any) any {
	switch v := msg.(type) {
	case *event.Event:
		m.OnEvent(v)
	default:
		return nil
	}
	return nil
}

func (m *Manager) OnStop() {
	event.DelEventListener(define.EventTypeActivity, m.Self())
	event.DelEventListener(define.EventTypePlayerOnline, m.Self())
	event.DelEventListener(define.EventTypePlayerOffline, m.Self())

	m.saveData()
	data.Cache.Close()
}

func (m *Manager) saveData() bool {
	m.entities.Range(func(key, value any) bool {
		ent := value.(*entity)

		actData := new(model.ActivityData)
		actData.Id = ent.Id
		actData.CfgId = ent.CfgId
		actData.Type = ent.Type
		actData.State = ent.State
		actData.StartTime = ent.StartTime
		actData.EndTime = ent.EndTime
		actData.CloseTime = ent.CloseTime
		actData.TimeType = ent.TimeType
		actData.TimeValue = ent.TimeValue

		desc := impl.GetActivityDesc(ent.Type)
		if desc == nil {
			log.Error("activity saveData: no activity factory for type: %v", actData.Type)
			return true
		}
		if desc.ExtractFunc != nil {
			actData.Data = desc.ExtractFunc(ent.handler)
		}

		err := data.SaveActivityData(actData)
		if err != nil {
			log.Error("save activity data error:%v", err)
		}
		return true
	})

	return true
}

// =========================================FSM PROCESS============================================

func (m *Manager) OnExit(fromState string, args []interface{}) {
	e := args[0].(*entity)
	if e.State != fromState {
		log.Error("OnExit state error:%v,currentState:%v", fromState, e.State)
		return
	}
}

func (m *Manager) Action(action string, fromState string, toState string, args []interface{}) error {
	ent := args[0].(*entity)

	switch action {
	case ActionStart:
		// waiting - running

		ent.handler.OnStart()
		log.Debug("activity start:%v,%v", ent.Id, ent.CfgId)
	case ActionClose:
		// waiting - closed
		// running - closed
		// stopped - closed

		if fromState == StateRunning {
			ent.handler.OnClose()
			data.PurgeActivityPlayerData(ent.Id) // 清除活动对应玩家数据
		}

		log.Debug("activity close:%v,%v", ent.Id, ent.CfgId)
	case ActionStop:
		// running - stopped
		if fromState == StateRunning {
			ent.handler.OnStop()
		}
		log.Debug("activity stop:%v,%v", ent.Id, ent.CfgId)
	case ActionRecover:
		// stopped - running
		log.Debug("activity recover:%v,%v", ent.Id, ent.CfgId)
	case ActionRestart:
		// closed - waiting
		// stopped - waiting
		m.entities.Delete(ent.Id)
		data.DelActivityData(ent.Id) // 清空活动数据

		// 分配新的id
		rdb, _ := db.GetEngine(m.App.GetEnv().ID)
		actId, err := rdb.GetActivityId()
		if err != nil {
			log.Error("get activity id from redis error:%v", err)
			return err
		}
		ent.Id = int64(actId)
		m.entities.Store(ent.Id, ent)
		log.Debug("activity restart:%v,%v", ent.Id, ent.CfgId)
	default:
		log.Error("unprocessed action:%v", action)
	}

	return nil
}

func (m *Manager) OnActionFailure(action string, fromState string, toState string, args []interface{}, err error) {
}

func (m *Manager) OnEnter(toState string, args []interface{}) {
	ent := args[0].(*entity)
	ent.State = toState
}

// ==================================================FSM PROCESS END======================================================================

// 事件分发
func (m *Manager) notify(obj *proto_player.Context, content map[string]any) {
	if key, ok := content["key"]; !ok {
		return
	} else {
		eventKey, ok := key.(string)
		if ok && eventKey != "" {
			m.entities.Range(func(k, v interface{}) bool {
				ent := v.(*entity)
				if ent.State == StateRunning {
					ent.handler.OnEvent(eventKey, obj, content)
				}
				return true
			})
		}
	}
}

// redis 回调
//func (m *Manager) OnRet(ret *dbengine.CDBRet) {}

// register 注册新的活动
func (m *Manager) register(cfgId int64) *entity {
	rdb, _ := db.GetEngine(m.App.GetEnv().ID)
	id, err := rdb.GetActivityId()
	if err != nil {
		log.Error("register new activity get id error:%v", err)
		return nil
	}

	activityConf, ok := config.CfgMgr.AllJson()["Activity"].(map[int64]conf.Activity)[cfgId]
	if !ok {
		log.Error("register new activity get config id error:%v", cfgId)
		return nil
	}

	var startTime, endTime, closeTime int64
	if activityConf.ActTime == define.ActTimeConfigured || activityConf.ActTime == define.ActTimeSeason {
		if activityConf.StartTime == "" || activityConf.EndTime == "" {
			log.Error("register timer error")
			return nil
		}
		parseTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(activityConf.StartTime), time.Local)
		if err != nil {
			log.Error("parse start time error:%v", err)
			return nil
		}
		startTime = parseTime.Unix()

		parseTime, err = time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(activityConf.EndTime), time.Local)
		if err != nil {
			log.Error("parse end time error")
			return nil
		}
		endTime = parseTime.Unix()

		parseTime, err = time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(activityConf.CloseTime), time.Local)
		if err != nil {
			log.Error("parse end time error")
			return nil
		}
		closeTime = parseTime.Unix()

		if startTime >= endTime {
			log.Error("register timer error1")
			return nil
		}
	} else if activityConf.ActTime == define.ActTimeServerConfigured {
		serverItem := new(model.ServerItem)
		ok, err := rdb.Mysql.Table(define.ServerGroup).Where("id = ?", m.App.GetEnv().ID).Get(serverItem)
		if !ok || err != nil {
			panic("mysql数据库连接失败")
		}

		startTime = serverItem.OpenServerTime
		endTime = serverItem.OpenServerTime + int64(activityConf.LastTime*86400)
		closeTime = endTime
		if startTime >= endTime {
			log.Error("checkCfg startTime>=endTime err")
			return nil
		}
	}

	ent := new(entity)
	ent.Id = int64(id)
	ent.Type = activityConf.Type
	ent.CfgId = activityConf.Id
	ent.StartTime = startTime
	ent.EndTime = endTime
	ent.CloseTime = closeTime
	ent.TimeType = activityConf.ActTime
	ent.mod = Mgr

	desc := impl.GetActivityDesc(ent.Type)
	if desc == nil || desc.NewHandler == nil {
		panic(fmt.Sprintf("missing activity handler: %v", ent.Type))
	}

	ent.handler = desc.NewHandler()
	ent.handler.SetBaseInfo(ent)

	if desc.InjectFunc != nil {
		desc.InjectFunc(ent.handler, desc.NewActivityData())
	}

	ent.handler.OnInit()

	switch ent.TimeType {
	case define.ActTimeClose:
		ent.State = StateClosed
	case define.ActTimeConfigured, define.ActTimeAlwaysOpen, define.ActTimeServerConfigured:
		ent.State = StateWaiting
	case define.ActTimeSeason:
		ent.State = StateWaiting
		ent.TimeValue = activityConf.Param1
	}

	return ent
}

// OnGetActivityData 获取活动数据
func (m *Manager) OnGetActivityData(ctx *proto_player.Context, id int64) (*proto_activity.ActivityData, error) {
	v, ok := m.entities.Load(id)
	if !ok {
		log.Error("GetActivityData id:%v", id)
		return nil, errors.New("GetActivityData id is null")
	}

	ent := v.(*entity)
	if ent.State != StateRunning {
		log.Error("GetActivityData state is not running:%v", id)
		return nil, errors.New("GetActivityData id is not run")
	}

	result := new(proto_activity.ActivityData)
	result.ActivityId = ent.Id
	result.ConfigId = ent.CfgId

	formatData := ent.handler.Format(ctx)
	log.Debug("加载活动数据:%v", ent.CfgId)
	impl.SetProtoByType(ent.Type, result, formatData)
	return result, nil
}

// OnGetActivityStatus 获取活动状态列表
func (m *Manager) OnGetActivityStatus() ([]*proto_activity.ActivityStatusInfo, error) {
	result := make([]*proto_activity.ActivityStatusInfo, 0)
	m.entities.Range(func(key, value interface{}) bool {
		ent := value.(*entity)
		if ent.State == StateRunning { // 开启中的活动

			var endTime, closeTime int64
			if ent.TimeType == define.ActTimeAlwaysOpen {
				endTime = 0
				closeTime = 0
			} else {
				closeTime = ent.CloseTime
				endTime = ent.EndTime
			}

			result = append(result, &proto_activity.ActivityStatusInfo{
				ActivityId: ent.Id,
				ConfigId:   ent.CfgId,
				StartTime:  ent.StartTime,
				EndTime:    endTime,
				CloseTime:  closeTime,
				IsOpen:     true,
				Season:     ent.TimeValue,
			})
		}

		return true
	})
	return result, nil
}

// OnGetActivityStatusByType 获取活动根据类型
func (m *Manager) OnGetActivityStatusByType(typ string) (*proto_activity.ActivityStatusInfo, error) {
	result := new(proto_activity.ActivityStatusInfo)
	m.entities.Range(func(key, value interface{}) bool {
		ent := value.(*entity)
		if ent.State == StateRunning && ent.Type == typ { // 开启中的活动
			var endTime, closeTime int64
			if ent.TimeType == define.ActTimeAlwaysOpen {
				endTime = 0
				closeTime = 0
			} else {
				closeTime = ent.CloseTime
				endTime = ent.EndTime
			}

			result = &proto_activity.ActivityStatusInfo{
				ActivityId: ent.Id,
				ConfigId:   ent.CfgId,
				StartTime:  ent.StartTime,
				EndTime:    endTime,
				CloseTime:  closeTime,
				IsOpen:     true,
				Season:     ent.TimeValue,
			}
		}

		return true
	})
	return result, nil
}

func (m *Manager) OnGetActivityDataList(ctx *proto_player.Context, ids []int64) []*proto_activity.ActivityData {
	result := make([]*proto_activity.ActivityData, 0)

	for _, id := range ids {
		v, ok := m.entities.Load(id)
		if !ok {
			log.Error("get activity data list id error:%v", id)
			continue
		}

		ent := v.(*entity)
		if ent.State != StateRunning {
			continue
		}

		actData := new(proto_activity.ActivityData)
		actData.ActivityId = ent.Id
		actData.ConfigId = ent.CfgId

		formatData := ent.handler.Format(ctx)
		impl.SetProtoByType(ent.Type, actData, formatData)
		result = append(result, actData)
	}
	return result
}

// OnRouterMsg 直接转发proto到活动内部
func (m *Manager) OnRouterMsg(ctx *proto_player.Context, actId int64, req proto.Message) (any, error) {
	v, ok := m.entities.Load(actId)
	if !ok {
		return nil, errors.New("router msg activity id error")
	}

	ent := v.(*entity)
	if ent.State != StateRunning {
		return nil, errors.New("router msg activity is not running")
	}

	return ent.handler.Router(ctx, req)
}

// TODO:后台接口
//func (m *Manager) StopActivity(id int32) bool {
//	v, ok := m.entities.Load(id)
//	if !ok {
//		log.Debug("stop activity id error:%v", id)
//		return false
//	}
//
//	entity := v.(*entity)
//
//	if !entity.isActive() {
//		return false
//	}
//
//	if err := m.sm.Trigger(entity.State, EventStop, entity); err != nil {
//		log.Error("%v", err)
//		return false
//	}
//
//	return true
//}

// TODO:后台接口
//func (m *Manager) RecoverActivity(id int32) bool {
//	v, ok := m.entities.Load(id)
//	if !ok {
//		log.Debug("recover activity id error:%v", id)
//		return false
//	}
//
//	entity := v.(*entity)
//
//	if entity.State != StateStopped {
//		log.Error("recover activity state error:%v", entity.State)
//		return false
//	}
//func (m *ActivityManager) RemoveActivity(id int32) {
//	if entity, ok := m.GetActivity(id); ok {
//		ctx := m.createActivityContext(id)
//		if err := entity.handler.OnClose(ctx); err != nil {
//			log.Warn().Err(err).Int32("activityID", id).Msg("Activity close error during removal")
//		}
//		m.cleanupActivityResources(id)
//		m.entities.Delete(id)
//	}
//}
//
//	err := m.sm.Trigger(entity.State, EventRecover, entity)
//	if err != nil {
//		log.Error("%v", err)
//		return false
//	}
//
//	return true
//}

// TODO:  删除活动 后台接口
//func (m *Manager) DelActivity(id int32) bool {
//	v, ok := m.entities.Load(id)
//	if !ok {
//		log.Debug("del activity id error:%v", id)
//		return false
//	}
//
//	entity := v.(*entity)
//
//	if entity.State == StateRunning {
//		err := m.sm.Trigger(entity.State, EventStop, entity)
//		if err != nil {
//			log.Error("%v", err)
//			return false
//		}
//	}
//
//	data.DelData(id)
//	m.entities.Delete(id)
//	return true
//}
