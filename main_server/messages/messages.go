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
