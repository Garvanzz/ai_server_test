package server

import (
	"sync/atomic"
	"time"
	"xfx/game_server/pkg/packet/pb_packet"
	"xfx/pkg/log"
	"xfx/proto"
	"xfx/proto/proto_game"
	"xfx/proto/proto_room"

	"xfx/game_server/pkg/network"
)

// TODO
func verifyToken(secret string) string {
	return secret
}

// OnConnect 链接进来
func (r *LockStepServer) OnConnect(conn *network.Conn) bool {
	count := atomic.AddInt64(&r.totalConn, 1)
	log.Debug("[router] OnConnect [%s] totalConn=%d", conn.GetRawConn().RemoteAddr().String(), count)
	// TODO 可以做一些check，不合法返回false
	return true
}

// OnMessage 消息处理
func (r *LockStepServer) OnMessage(conn *network.Conn, p pb_packet.IPacket) bool {
	msg := p.(*pb_packet.Packet)

	log.Info("[router] OnMessage [%s] msg=[%d] len=[%d]", conn.GetRawConn().RemoteAddr().String(), msg.GetMessageID(), len(msg.GetData()))

	protoMsg, err := proto_id.NewMessage(msg.GetMessageID())
	if err != nil {
		log.Error("[router] OnMessage.err:%v", err)
		return false
	}
	log.Info("[router] msg.Unmarshal ***************")
	switch protoMsg.(type) {
	case *proto_room.S2SMSGStartGame:
		rec := &proto_room.S2SMSGStartGame{}
		if err := msg.Unmarshal(rec); nil != err {
			log.Error("[router] msg.Unmarshal error=[%s]", err.Error())
			return false
		}

		//构建一局游戏
		re := &proto_room.S2CStartGame{}
		_, err := r.mgr.CreateGame(rec.GetRoomInfo(), rec.GetLogicServer(), conn)
		if err != nil {
			re.State = int32(proto_room.ERRORCODE_ERR_RoomState)
			log.Error("CreateRoom err:", err)
		} else {
			re.State = int32(proto_room.ERRORCODE_ERR_Ok)
			re.RoomInfo = rec.GetRoomInfo()
		}

		proId, _ := proto_id.MessageID(&proto_room.S2CStartGame{})
		conn.AsyncWritePacket(pb_packet.NewPacket(proId, re), time.Millisecond)
		log.Info("[router] creategame Addr=[%v]: %v", conn.GetRawConn().RemoteAddr(), re.RoomInfo)
		return true
	case *proto_game.C2SConnectMsg:
		//客户端连接
		rec := &proto_game.C2SConnectMsg{}
		if err := msg.Unmarshal(rec); nil != err {
			log.Error("[router] msg.Unmarshal error=[%s]", err.Error())
			return false
		}

		// player id
		playerID := rec.GetPlayerId()
		// room id
		roomID := rec.GetRoomId()
		// token
		token := rec.GetToken()

		ret := &proto_game.S2CConnectMsg{
			State: int32(proto_room.ERRORCODE_ERR_Ok),
		}

		room := r.mgr.GetGame(uint64(roomID))
		proId, _ := proto_id.MessageID(&proto_game.S2CConnectMsg{})
		if nil == room {
			ret.State = int32(proto_room.ERRORCODE_ERR_NoRoom)
			conn.AsyncWritePacket(pb_packet.NewPacket(proId, ret), time.Millisecond)
			log.Error("[router] no room player=[%d] room=[%d] token=[%s]", playerID, roomID, token)
			return true
		}

		if room.IsOver() {
			ret.State = int32(proto_room.ERRORCODE_ERR_RoomState)
			conn.AsyncWritePacket(pb_packet.NewPacket(proId, ret), time.Millisecond)
			log.Error("[router] room is over player=[%d] room==[%d] token=[%s]", playerID, roomID, token)
			return true
		}

		if !room.HasPlayer(uint64(playerID)) {
			ret.State = int32(proto_room.ERRORCODE_ERR_NoPlayer)
			conn.AsyncWritePacket(pb_packet.NewPacket(proId, ret), time.Millisecond)
			log.Error("[router] !room.HasPlayer(playerID) player=[%d] room==[%d] token=[%s]", playerID, roomID, token)
			return true
		}

		//验证token
		if token != verifyToken(token) {
			ret.State = int32(proto_room.ERRORCODE_ERR_Token)
			conn.AsyncWritePacket(pb_packet.NewPacket(proId, ret), time.Millisecond)
			log.Error("[router] verifyToken failed player=[%d] room==[%d] token=[%s]", playerID, roomID, token)
			return true
		}

		conn.PutExtraData(uint64(playerID))

		// 这里只是先给加上身份标识，不能直接返回Connect成功，又后面Game返回
		//conn.AsyncWritePacket(pb_packet.NewPacket(uint8(pb.ID_MSG_Connect), ret), time.Millisecond)
		return room.OnConnect(conn)
	case *proto_game.C2SPing:
		proId, _ := proto_id.MessageID(&proto_game.S2CPong{})
		log.Info("ping !!!")
		conn.AsyncWritePacket(pb_packet.NewPacket(proId, nil), 0)
		return true
	}

	return false

}

// OnClose 链接断开
func (r *LockStepServer) OnClose(conn *network.Conn) {
	count := atomic.AddInt64(&r.totalConn, -1)

	log.Info("[router] OnClose: total=%d, addr = %d", count, conn.GetRawConn().RemoteAddr().String())
}
