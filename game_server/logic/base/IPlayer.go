package base

import (
	"xfx/game_server/pkg/network"
	"xfx/game_server/pkg/packet/pb_packet"
)

// 玩家接口
type IPlayer interface {
	SendMessage(msg pb_packet.IPacket)
	GetClient() *network.Conn
	Connect(conn *network.Conn)
	Cleanup()
	GetID() uint64
	RefreshHeartbeatTime()
	SetloadProgress(pro int32)
	SetIsReady(r bool)
	GetIsReady() bool
	SetSendFrameCount(c uint32)
	PlayerIsOnline() bool
	GetLastHeartbeatTime() int64
	GetSendFrameCount() uint32
	GetloadProgress() int32
	GetIsMonster() bool
}
