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

// activity_daily_acc_recharge.go - 缺少跨天重置
// 每日累充金额 pd.Money 只会累加，但没有跨天重置逻辑，Update 方法是空实现，rank.go 中也有对应 TODO 未完成：

// activity_month_card.go - 月卡每日邮件未实现
// 月卡玩家每天都应该领取奖励，但 Update 完全没有实现，仅在 rank.go 中留了 TODO：

// activity_passport.go - getAward 已领取判断逻辑有误

// 对于未购买高级通行证的玩家，应该只检查 pd.NormalIds 是否已领取就跳过，但现在未购买玩家的普通奖励已领也不会 return（因为 pd.IsBuy 为 false 使整个条件为 false），会进入下面继续领奖导致重复发放。

// activity_arena.go - BattleReport 胜利后未更新排行榜积分
// 竞技场战报只刷新了对手列表，但没有根据胜负调整积分更新排行榜，导致排行榜分数永远停留在初始值：

// 配置热更：config 层 Reload 成功后会发 EventTypeConfigReload，activity 监听并 resetAllConfigChecked，下一 tick 会重新执行 determineStateFromConfig。
