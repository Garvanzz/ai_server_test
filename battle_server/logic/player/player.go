package player

import (
	"time"
	"xfx/game_server/logic/base"
	"xfx/game_server/pkg/network"
	"xfx/game_server/pkg/packet/pb_packet"
	"xfx/pkg/log"
	"xfx/proto/proto_public"
	proto_room "xfx/proto/proto_room"
)

// 人物
type Player struct {
	base.IPlayer
	Id                uint64
	IsOnline          bool
	IsReady           bool
	LoadingProgress   int32
	LastHeartbeatTime int64
	SendFrameCount    uint32
	Client            *network.Conn
	IsMonster         bool                           //是否人机
	IsLeader          bool                           //是否房主
	Group             int32                          //阵营
	PlayerInfo        *proto_public.CommonPlayerInfo //玩家信息
}

func NewPlayer(player *proto_room.RoomPlayers, index int) base.IPlayer {
	p := new(Player)
	p.Id = uint64(player.GetCommonPlayerInfo().GetPlayerId())
	p.IsMonster = player.GetIsMonster()
	p.IsLeader = player.GetIsleader()
	p.Group = player.Group
	p.PlayerInfo = player.GetCommonPlayerInfo()

	if p.IsMonster {
		p.SetIsReady(true)
		p.SetloadProgress(100)
	}

	return p
}

func (p *Player) Connect(conn *network.Conn) {
	p.Client = conn
	p.IsOnline = true
	p.IsReady = false
	p.LastHeartbeatTime = time.Now().Unix()
}

func (p *Player) PlayerIsOnline() bool {
	return nil != p.Client && p.IsOnline
}

func (p *Player) RefreshHeartbeatTime() {
	p.LastHeartbeatTime = time.Now().Unix()
}

func (p *Player) GetLastHeartbeatTime() int64 {
	return p.LastHeartbeatTime
}

func (p *Player) GetID() uint64 {
	return p.Id
}

func (p *Player) GetIsMonster() bool {
	return p.IsMonster
}

func (p *Player) SetSendFrameCount(c uint32) {
	p.SendFrameCount = c
}

func (p *Player) GetSendFrameCount() uint32 {
	return p.SendFrameCount
}

func (p *Player) SetloadProgress(c int32) {
	p.LoadingProgress = c
}

func (p *Player) GetloadProgress() int32 {
	return p.LoadingProgress
}

func (p *Player) SetIsReady(r bool) {
	p.IsReady = r
}

func (p *Player) GetIsReady() bool {
	return p.IsReady
}

func (p *Player) GetClient() *network.Conn {
	return p.Client
}

func (p *Player) SendMessage(msg pb_packet.IPacket) {
	if !p.PlayerIsOnline() {
		return
	}

	if err := p.Client.AsyncWritePacket(msg, 0); err != nil {
		log.Error("write package error: %v", err)
		p.Client.Close()
	}
}

func (p *Player) Cleanup() {

	if nil != p.Client {
		p.Client.Close()
	}
	p.Client = nil
	p.IsReady = false
	p.IsOnline = false
}
