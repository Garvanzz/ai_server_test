package activity

import "xfx/core/fsm"

const (
	StateWaiting = "waiting"
	StateRunning = "running"
	StateStopped = "stopped"
	StateClosed  = "closed"

	EventNone    = ""
	EventStart   = "event_start"
	EventStop    = "event_stop"
	EventClose   = "event_close"
	EventRecover = "event_recover"
	EventRestart = "event_restart"

	ActionStart   = "action_start"
	ActionClose   = "action_close"
	ActionStop    = "action_stop"
	ActionRecover = "action_recover"
	ActionRestart = "action_restart"
)

var (
	transitions = []fsm.Transition{
		{StateWaiting, EventStart, StateRunning, ActionStart},
		{StateWaiting, EventClose, StateClosed, ActionClose},

		{StateRunning, EventStop, StateStopped, ActionStop},
		{StateRunning, EventClose, StateClosed, ActionClose},

		{StateStopped, EventRecover, StateRunning, ActionRecover},
		{StateStopped, EventClose, StateClosed, ActionClose},
		{StateStopped, EventRestart, StateWaiting, ActionRestart},

		{StateClosed, EventRestart, StateWaiting, ActionRestart},
	}
)

// activity_competition.go - chooseGroupId 重复判断 pd.IsChoose

// activity_daily_acc_recharge.go - 跨天重置已添加 OnDayReset 接口

// activity_month_card.go - 跨天重置已添加 OnDayReset 接口

// activity_passport.go - 已修复：getAward 已领取判断逻辑

// activity_arena.go - BattleReport 胜利后未更新排行榜积分
