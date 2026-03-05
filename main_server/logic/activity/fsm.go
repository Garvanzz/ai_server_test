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


// PurgeActivityPlayerData 调用后需要同步遍历 Cache，删除所有 key / ActivityPlayerDataBase == actId 的条目。

// ActionRestart 分配了新的 actId，但玩家数据 key 是 actId * Base + playerId，旧活动的玩家数据永远留在 Redis/Cache 里无法被清除。
// 在 DelActivityData 之前应先调用 data.PurgeActivityPlayerData(ent.Id) 清理旧玩家数据。

// checkState() 对 StateStopped 状态缺少时间检查
// StateStopped 状态下如果配置的活动时间已经结束，理应转为 Closed，但 checkState() 里没有覆盖这个场景，导致活动停留在 Stopped 状态永远不自动关闭。

// determineStateFromConfig 在 DB 失败时 panic
// ActTimeServerConfigured 类型在连接 MySQL 失败时直接 panic，在生产环境不应该这样处理，应该返回错误并跳过该活动。

// checked 标志导致配置热更无法生效
// ent.checked 被设为 true 后永不重置，意味着配置热更新（修改活动开始/结束时间）后，服务不重启则不会生效。

// activity_competition.go - chooseGroupId 重复判断 pd.IsChoose

// activity_daily_acc_recharge.go - 缺少跨天重置
// 每日累充金额 pd.Money 只会累加，但没有跨天重置逻辑，Update 方法是空实现，rank.go 中也有对应 TODO 未完成：


// activity_month_card.go - 月卡每日邮件未实现
// 月卡玩家每天都应该领取奖励，但 Update 完全没有实现，仅在 rank.go 中留了 TODO：

// activity_passport.go - getAward 已领取判断逻辑有误


// 对于未购买高级通行证的玩家，应该只检查 pd.NormalIds 是否已领取就跳过，但现在未购买玩家的普通奖励已领也不会 return（因为 pd.IsBuy 为 false 使整个条件为 false），会进入下面继续领奖导致重复发放。


// activity_arena.go - BattleReport 胜利后未更新排行榜积分
// 竞技场战报只刷新了对手列表，但没有根据胜负调整积分更新排行榜，导致排行榜分数永远停留在初始值：