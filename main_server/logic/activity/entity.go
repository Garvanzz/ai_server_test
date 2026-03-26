package activity

import (
	"time"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/logic/activity/impl"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/utils"
)

type entity struct {
	Id            int64
	CfgId         int64
	Type          string
	State         string
	StartTime     int64
	EndTime       int64
	CloseTime     int64
	TimeType      int32
	TimeValue     int32
	handler       impl.IActivity
	mod           module.Module
	checked       bool
	lastUpdateDay int32 // 上次更新的日期，用于跨天检测 (格式：YYYYMMDD)
	manualClosed  bool  // GM 主动关闭标记，防止 Tick 自动重启
}

func (e *entity) GetId() int64          { return e.Id }
func (e *entity) GetCfgId() int64       { return e.CfgId }
func (e *entity) GetType() string       { return e.Type }
func (e *entity) GetStartTime() int64   { return e.StartTime }
func (e *entity) GetCloseTime() int64   { return e.CloseTime }
func (e *entity) GetEndTime() int64     { return e.EndTime }
func (e *entity) Module() module.Module { return e.mod }

// 检查活动状态
func (e *entity) checkState() (event string) {
	event = EventNone

	now := utils.Now().Unix()
	switch e.State {
	case StateWaiting:
		if (now >= e.StartTime && now < e.EndTime) || e.TimeType == define.ActTimeAlwaysOpen {
			event = EventStart
		} else if now >= e.EndTime {
			event = EventClose
		}
	case StateRunning:
		if e.TimeType == define.ActTimeClose {
			event = EventClose
		} else if e.TimeType == define.ActTimeConfigured || e.TimeType == define.ActTimeServerConfigured || e.TimeType == define.ActTimeSeason {
			if now < e.StartTime || now > e.EndTime {
				event = EventClose
			}
		}
	case StateStopped:
		// 暂停状态下若配置的活动结束时间已过，自动转为 Closed
		if e.TimeType != define.ActTimeAlwaysOpen && e.EndTime > 0 && now >= e.EndTime {
			event = EventClose
		}
	case StateClosed:
		// manualClosed 为 true 表示 GM 主动关闭，不自动重启
		if !e.manualClosed {
			if (now >= e.StartTime && now < e.EndTime) || e.TimeType == define.ActTimeAlwaysOpen {
				event = EventRestart
			}
		}
	}

	return
}

func (e *entity) determineStateFromConfig(Sid int) (event string) {
	event = EventNone

	conf, ok := impl.GetCommonConf(e.CfgId)
	if !ok {
		log.Error("activity config error:%v", e.CfgId)
		return
	}

	// set time type
	e.TimeType = conf.ActTime

	switch e.TimeType {
	case define.ActTimeAlwaysOpen: // 常驻活动
		e.StartTime = 0
		e.EndTime = 0
		e.CloseTime = 0

		if e.State == StateWaiting {
			event = EventStart
		} else if e.State == StateClosed {
			event = EventRestart
		}
	case define.ActTimeConfigured, define.ActTimeSeason: // 检查活动配置表
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(conf.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			return
		}

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(conf.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			return
		}

		closeTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(conf.CloseTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			return
		}

		if startTime.Unix() >= endTime.Unix() {
			log.Error("checkCfg startTime>=endTime err")
			return
		}

		// 配置更新带来新的时间窗口时，清除 manualClosed 标记，允许自动重启
		if e.manualClosed && (startTime.Unix() != e.StartTime || endTime.Unix() != e.EndTime) {
			e.manualClosed = false
			log.Info("activity manualClosed cleared due to new schedule: cfgId=%v", e.CfgId)
		}

		e.StartTime = startTime.Unix()
		e.EndTime = endTime.Unix()
		e.CloseTime = closeTime.Unix()
		e.TimeValue = conf.Param1
	case define.ActTimeClose: // 关闭活动
		e.StartTime = 0
		e.EndTime = 0
		e.CloseTime = 0

		if e.State == StateRunning || e.State == StateWaiting {
			event = EventClose
		}
	case define.ActTimeServerConfigured: //按照服务器开启时间
		serverItem := new(model.ServerItem)
		ok, err := db.Engine.Mysql.Table(define.GameServerTable).Cols("open_server_time").Where("id = ?", Sid).Get(serverItem)
		if !ok || err != nil {
			log.Error("activity determineStateFromConfig mysql error, cfgId:%v, err:%v", e.CfgId, err)
			return
		}

		startTime := serverItem.OpenServerTime
		endTime := serverItem.OpenServerTime + int64(conf.LastTime*86400)
		if startTime >= endTime {
			log.Error("checkCfg startTime>=endTime err")
			return
		}

		e.StartTime = startTime
		e.EndTime = endTime
		e.CloseTime = endTime
	default:
		log.Error("checkCfg ActTime error:%v", conf.ActTime)
	}

	return
}
