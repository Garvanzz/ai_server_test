package activity

import (
	"errors"
	"fmt"
	"time"
	"xfx/core/cache"
	"xfx/core/config"
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
	"xfx/pkg/utils"
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
	entities entityStore
	sm       *fsm.StateMachine
	lastTick int64
}

func (m *Manager) OnInit(app module.App) {
	m.BaseModule.OnInit(app)
	m.entities = newEntityStore()
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
		m.entities.store(ent)
	}

	// 根据配置加载新活动
	activityConfs := impl.GetAllCommonConf()
	for _, activityConf := range activityConfs {
		if _, ok := existIds[activityConf.Id]; !ok {
			log.Info("注册新活动:%v", activityConf.Id)
			ent := m.register(activityConf.Id)
			if ent != nil {
				m.entities.store(ent)
			}
		}
	}

	// register func
	m.Register("GetActivityStatus", m.OnGetActivityStatus)
	m.Register("GetActivityStatusByType", m.OnGetActivityStatusByType)
	m.Register("GetActivityData", m.OnGetActivityData)
	m.Register("GetActivityDataList", m.OnGetActivityDataList)
	m.Register("OnRouterMsg", m.OnRouterMsg)

	// GM 后台接口
	m.Register("ListAllActivities", m.OnListAllActivities)
	m.Register("GetActivityByActId", m.OnGetActivityByActId)
	m.Register("GetActivityByCfgId", m.OnGetActivityByCfgId)
	m.Register("StopActivity", m.OnStopActivity)
	m.Register("RecoverActivity", m.OnRecoverActivity)
	m.Register("CloseActivity", m.OnCloseActivity)
	m.Register("RestartActivity", m.OnRestartActivity)
	m.Register("RemoveActivity", m.OnRemoveActivity)
	m.Register("CloseActivityByCfgId", m.OnCloseActivityByCfgId)
	m.Register("StopActivityByType", m.OnStopActivityByType)
}

func (m *Manager) OnStart(ctx module.Context) {
	m.BaseModule.OnStart(ctx)
	event.AddEventListener(define.EventTypeActivity, m.Self())
	event.AddEventListener(define.EventTypePlayerOnline, m.Self())
	event.AddEventListener(define.EventTypePlayerOffline, m.Self())
	event.AddEventListener(define.EventTypeConfigReload, m.Self())
}

func (m *Manager) GetType() string { return define.ModuleActivity }

func (m *Manager) OnTick(delta time.Duration) {
	now := utils.Now()
	if m.lastTick == 0 {
		m.lastTick = now.Unix()
	} else {
		if now.Unix()-m.lastTick >= 60*5 {
			m.saveData()
			m.lastTick = now.Unix()
		}
	}

	for _, ent := range m.entities.snapshot() {

		// 检查配置是否变动
		if !ent.checked {
			eventStr := ent.determineStateFromConfig(m.App.GetEnv().ID)
			if eventStr != EventNone {
				err := m.sm.Trigger(ent.State, eventStr, ent)
				if err != nil {
					log.Error("%v", err)
				}
			}
			ent.handler.OnReloadConfig()
			ent.checked = true
			continue
		}

		if ent.State == StateClosed {
			continue
		}

		// 状态转换处理
		triggerEvent := ent.checkState()
		if triggerEvent != EventNone {
			err := m.sm.Trigger(ent.State, triggerEvent, ent)
			if err != nil {
				log.Error("sm trigger error:%v", err)
				continue
			}
		}

		// 活动业务更新
		if ent.State == StateRunning {
			// 跨天检测
			currentDay := int32(now.Year()*10000 + int(now.Month())*100 + now.Day())
			if ent.lastUpdateDay == 0 {
				ent.lastUpdateDay = currentDay
			}
			if ent.lastUpdateDay != currentDay {
				// 跨天了，触发跨天重置
				ent.handler.OnDayReset(now)
				ent.lastUpdateDay = currentDay
				log.Debug("activity day reset triggered: actId=%v, cfgId=%v", ent.Id, ent.CfgId)
			}
			ent.handler.Update(now)
		}
	}
}

// OnEvent 事件回调
func (m *Manager) OnEvent(ev *event.Event) {
	if ev == nil {
		return
	}
	if ev.Type == define.EventTypeConfigReload {
		m.resetAllConfigChecked()
		return
	}
	if ev.M == nil {
		return
	}

	// 玩家基础信息
	ctx, ok := ev.M["player"].(*proto_player.Context)
	if !ok {
		log.Error("activity event find no player data")
		return
	}

	switch ev.Type {
	case define.EventTypeActivity:
		m.notify(ctx, ev.M)
	case define.EventTypePlayerOffline:
	case define.EventTypePlayerOnline:
		ev.M["key"] = "player_online"
		m.notify(ctx, ev.M)
	default:
	}
}

// 重置所有配置检查标记
func (m *Manager) resetAllConfigChecked() {
	m.entities.resetChecked()
	log.Info("activity: config reload notified, all entity checked reset")
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
	event.DelEventListener(define.EventTypeConfigReload, m.Self())

	m.saveData()
	data.Cache.Close()
}

func (m *Manager) saveData() bool {
	for _, ent := range m.entities.snapshot() {

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
			continue
		}
		if desc.ExtractFunc != nil {
			actData.Data = desc.ExtractFunc(ent.handler)
		}

		err := data.SaveActivityData(actData)
		if err != nil {
			log.Error("save activity data error:%v", err)
		}
	}

	return true
}

// =========================================FSM PROCESS============================================

func (m *Manager) OnExit(fromState string, args []interface{}) {
	if len(args) == 0 {
		return
	}
	e := args[0].(*entity)
	if e.State != fromState {
		log.Error("OnExit state error:%v,currentState:%v", fromState, e.State)
		return
	}
}

func (m *Manager) Action(action string, fromState string, toState string, args []interface{}) error {
	if len(args) == 0 {
		return errors.New("activity action missing entity argument")
	}
	ent := args[0].(*entity)

	switch action {
	case ActionStart:
		return m.handleActionStart(ent)
	case ActionClose:
		return m.handleActionClose(ent, fromState)
	case ActionStop:
		return m.handleActionStop(ent, fromState)
	case ActionRecover:
		return m.handleActionRecover(ent)
	case ActionRestart:
		return m.handleActionRestart(ent)
	default:
		log.Error("unprocessed action:%v", action)
	}

	return nil
}

func (m *Manager) OnActionFailure(action string, fromState string, toState string, args []interface{}, err error) {
}

func (m *Manager) handleActionStart(ent *entity) error {
	ent.handler.OnStart()
	log.Debug("activity start:%v,%v", ent.Id, ent.CfgId)
	return nil
}

func (m *Manager) handleActionClose(ent *entity, fromState string) error {
	if fromState == StateRunning {
		ent.handler.OnClose()
		data.PurgeActivityPlayerData(ent.Id)
	}
	log.Debug("activity close:%v,%v", ent.Id, ent.CfgId)
	return nil
}

func (m *Manager) handleActionStop(ent *entity, fromState string) error {
	if fromState == StateRunning {
		ent.handler.OnStop()
	}
	log.Debug("activity stop:%v,%v", ent.Id, ent.CfgId)
	return nil
}

func (m *Manager) handleActionRecover(ent *entity) error {
	ent.handler.OnRecover()
	log.Debug("activity recover:%v,%v", ent.Id, ent.CfgId)
	return nil
}

func (m *Manager) handleActionRestart(ent *entity) error {
	previousActID := ent.Id
	data.PurgeActivityPlayerData(previousActID)
	m.entities.delete(previousActID)
	data.DelActivityData(previousActID)

	actID, err := db.GetActivityId()
	if err != nil {
		log.Error("get activity id from redis error:%v", err)
		return err
	}
	ent.Id = int64(actID)
	m.entities.store(ent)
	ent.handler.OnRestart(previousActID)
	log.Debug("activity restart:%v,%v", ent.Id, ent.CfgId)
	return nil
}

func (m *Manager) OnEnter(fromState string, toState string, args []interface{}) {
	if len(args) == 0 {
		return
	}
	ent := args[0].(*entity)
	ent.State = toState
	m.entities.updateState(ent, fromState, toState)
}

// ==================================================FSM PROCESS END======================================================================

// 事件分发
func (m *Manager) notify(obj *proto_player.Context, content map[string]any) {
	if key, ok := content["key"]; !ok {
		return
	} else {
		eventKey, ok := key.(string)
		if ok && eventKey != "" {
			for _, ent := range m.entities.snapshot() {
				if ent.State == StateRunning {
					ent.handler.OnEvent(eventKey, obj, content)
				}
			}
		}
	}
}

// redis 回调
//func (m *Manager) OnRet(ret *dbengine.CDBRet) {}

// register 注册新的活动
func (m *Manager) register(cfgId int64) *entity {
	id, err := db.GetActivityId()
	if err != nil {
		log.Error("register new activity get id error:%v", err)
		return nil
	}

	activityConf, ok := config.Activity.Find(int64(cfgId))
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
		rdb, _ := db.GetEngine()
		serverItem := new(model.ServerItem)
		ok, err := rdb.Mysql.Table(define.GameServerTable).Cols("open_server_time").Where("id = ?", m.App.GetEnv().ID).Get(serverItem)
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
	ent, ok := m.entities.load(id)
	if !ok {
		log.Error("GetActivityData id:%v", id)
		return nil, errors.New("GetActivityData id is null")
	}
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
	for _, ent := range m.entities.snapshot() {
		if info := statusInfoForEntity(ent); info != nil {
			result = append(result, info)
		}
	}
	return result, nil
}

// OnGetActivityStatusByType 获取活动根据类型
func (m *Manager) OnGetActivityStatusByType(typ string) (*proto_activity.ActivityStatusInfo, error) {
	ent := m.getEntityByType(typ)
	if ent == nil {
		return new(proto_activity.ActivityStatusInfo), nil
	}
	return statusInfoForEntity(ent), nil
}

func (m *Manager) OnGetActivityDataList(ctx *proto_player.Context, ids []int64) []*proto_activity.ActivityData {
	result := make([]*proto_activity.ActivityData, 0)

	for _, id := range ids {
		ent, ok := m.entities.load(id)
		if !ok {
			log.Error("get activity data list id error:%v", id)
			continue
		}
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
	ent, ok := m.entities.load(actId)
	if !ok {
		return nil, errors.New("router msg activity id error")
	}
	if ent.State != StateRunning {
		return nil, errors.New("router msg activity is not running")
	}

	return ent.handler.Router(ctx, req)
}

func entityToInfo(ent *entity) *model.ActivityInfo {
	if ent == nil {
		return nil
	}
	return &model.ActivityInfo{
		ActId:     ent.Id,
		CfgId:     ent.CfgId,
		Type:      ent.Type,
		State:     ent.State,
		StartTime: ent.StartTime,
		EndTime:   ent.EndTime,
		CloseTime: ent.CloseTime,
		TimeType:  ent.TimeType,
		Season:    ent.TimeValue,
	}
}

func statusInfoForEntity(ent *entity) *proto_activity.ActivityStatusInfo {
	if ent == nil || ent.State != StateRunning {
		return nil
	}

	endTime := ent.EndTime
	closeTime := ent.CloseTime
	if ent.TimeType == define.ActTimeAlwaysOpen {
		endTime = 0
		closeTime = 0
	}

	return &proto_activity.ActivityStatusInfo{
		ActivityId: ent.Id,
		ConfigId:   ent.CfgId,
		StartTime:  ent.StartTime,
		EndTime:    endTime,
		CloseTime:  closeTime,
		IsOpen:     true,
		Season:     ent.TimeValue,
	}
}

func (m *Manager) getEntityByActId(actId int64) *entity {
	ent, ok := m.entities.load(actId)
	if !ok {
		return nil
	}
	return ent
}

func (m *Manager) getEntityByCfgId(cfgId int64) *entity {
	return m.entities.getByCfgID(cfgId)
}

func (m *Manager) getEntityByType(typ string) *entity {
	return m.entities.getRunningByType(typ)
}

// ==================== GM 后台接口 ====================

// OnListAllActivities 列出所有活动（含状态），供 GM 后台展示
func (m *Manager) OnListAllActivities() ([]*model.ActivityInfo, error) {
	list := make([]*model.ActivityInfo, 0)
	for _, ent := range m.entities.snapshot() {
		list = append(list, entityToInfo(ent))
	}
	return list, nil
}

// OnGetActivityByActId 按活动实例 ID 查询
func (m *Manager) OnGetActivityByActId(actId int64) (*model.ActivityInfo, error) {
	ent := m.getEntityByActId(actId)
	if ent == nil {
		return nil, errors.New("activity not found")
	}
	return entityToInfo(ent), nil
}

// OnGetActivityByCfgId 按配置 ID 查询，优先返回 Running 的实例
func (m *Manager) OnGetActivityByCfgId(cfgId int64) (*model.ActivityInfo, error) {
	ent := m.getEntityByCfgId(cfgId)
	if ent == nil {
		return nil, errors.New("activity not found")
	}
	return entityToInfo(ent), nil
}

// OnStopActivity 暂停活动（Running -> Stopped）
func (m *Manager) OnStopActivity(actId int64) error {
	ent := m.getEntityByActId(actId)
	if ent == nil {
		return errors.New("activity not found")
	}
	if ent.State != StateRunning {
		return fmt.Errorf("invalid state for stop: want running, got %s", ent.State)
	}
	if err := m.sm.Trigger(ent.State, EventStop, ent); err != nil {
		return err
	}
	return nil
}

// OnRecoverActivity 恢复活动（Stopped -> Running）
func (m *Manager) OnRecoverActivity(actId int64) error {
	ent := m.getEntityByActId(actId)
	if ent == nil {
		return errors.New("activity not found")
	}
	if ent.State != StateStopped {
		return fmt.Errorf("invalid state for recover: want stopped, got %s", ent.State)
	}
	if err := m.sm.Trigger(ent.State, EventRecover, ent); err != nil {
		return err
	}
	return nil
}

// OnCloseActivity 强制结束活动（Waiting/Running/Stopped -> Closed）
func (m *Manager) OnCloseActivity(actId int64) error {
	ent := m.getEntityByActId(actId)
	if ent == nil {
		return errors.New("activity not found")
	}
	if ent.State == StateClosed {
		return errors.New("activity already closed")
	}
	if err := m.sm.Trigger(ent.State, EventClose, ent); err != nil {
		return err
	}
	return nil
}

// OnRestartActivity 重启活动（Stopped/Closed -> Waiting，分配新 actId）
func (m *Manager) OnRestartActivity(actId int64) error {
	ent := m.getEntityByActId(actId)
	if ent == nil {
		return errors.New("activity not found")
	}
	if ent.State != StateStopped && ent.State != StateClosed {
		return fmt.Errorf("invalid state for restart: want stopped or closed, got %s", ent.State)
	}
	if err := m.sm.Trigger(ent.State, EventRestart, ent); err != nil {
		return err
	}
	return nil
}

// OnRemoveActivity 彻底移除活动
func (m *Manager) OnRemoveActivity(actId int64) error {
	ent := m.getEntityByActId(actId)
	if ent == nil {
		return errors.New("activity not found")
	}
	if ent.State == StateRunning || ent.State == StateStopped {
		if err := m.sm.Trigger(ent.State, EventClose, ent); err != nil {
			return err
		}
	}
	m.entities.delete(actId)
	data.DelActivityData(actId)
	return nil
}

// OnCloseActivityByCfgId 按配置 ID 强制结束活动
func (m *Manager) OnCloseActivityByCfgId(cfgId int64) error {
	ent := m.getEntityByCfgId(cfgId)
	if ent == nil {
		return errors.New("activity not found")
	}
	return m.OnCloseActivity(ent.Id)
}

// OnStopActivityByType 按类型暂停活动（该类型当前 Running 的实例）
func (m *Manager) OnStopActivityByType(typ string) error {
	ent := m.getEntityByType(typ)
	if ent == nil {
		return errors.New("activity not found or not running")
	}
	return m.OnStopActivity(ent.Id)
}
