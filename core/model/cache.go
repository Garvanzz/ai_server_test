package model

import (
	"xfx/pkg/agent"
	"xfx/pkg/module"
)

type Cache struct {
	App        module.App
	Session    agent.PID // 网关session pid
	Self       agent.PID
	Disconnect bool
	GameRun    GameRun // 游戏相关缓存数据
	SaveTick   int
	RoomId     int32 //房间Id,标记是否有房间
	RoomType   int32
}

type GameRun struct {
	PID        agent.PID
	GameId     int64
	LastGameId int64
}
