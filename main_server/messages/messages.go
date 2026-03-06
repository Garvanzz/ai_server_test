package messages

import (
	"xfx/core/model"
	"xfx/pkg/agent"
	"xfx/proto/proto_player"
)

type CreatePlayer struct {
	Session agent.PID
	Model   model.Player
}

type Login struct {
	Session agent.PID
	Request *proto_player.C2SLogin
}

// LoginResult 玩家登录返回
type LoginResult struct {
	PlayerId  int64
	PlayerPid agent.PID
	Result    int
}

type LoginSuccess struct {
}

type LoginReplace struct {
	Session agent.PID
}

type Logout struct{}

type Disconnect struct{}

type SysMessage struct {
	Content interface{}
}

// SysKick 踢线指令（作为 SysMessage.Content）
type SysKick struct {
	Reason string
}

// SysRefreshActivity 刷新活动指令（作为 SysMessage.Content）
type SysRefreshActivity struct{}

// SysGrantItemEntry 发放道具条目（与 conf.ItemE 一致，避免 messages 依赖 conf）
type SysGrantItemEntry struct {
	ItemId   int32
	ItemType int32
	ItemNum  int32
}

// SysGrantItems GM 发放道具指令（作为 SysMessage.Content）
type SysGrantItems struct {
	Items []SysGrantItemEntry
}

type DispatchMessage struct {
	Content interface{}
}
type GetPlayerDataMessage struct {
}

// 关卡结算
type StageSettle struct {
	IsWin           bool
	StageId         int32
	Damage          int64
	WithStandDamage int64
	Heal            int64
	KillNum         int32
}
