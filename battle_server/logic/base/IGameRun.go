package base

import (
	"xfx/game_server/pkg/network"
)

type IGameRun interface {
	ISync
	GetAllPlayer() []IPlayer
	GetPlayer(id uint64) IPlayer
	GetSync() ISync
	SetState(state int)
	GetLogicServer() *network.Conn
	GetRoomId() int32
}
