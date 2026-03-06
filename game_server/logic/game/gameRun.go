package game

import (
	"time"
	"xfx/game_server/logic/base"
	"xfx/game_server/logic/player"
	"xfx/game_server/pkg/network"
	"xfx/game_server/pkg/packet/pb_packet"
	"xfx/pkg/log"
	proto_id "xfx/proto"
	"xfx/proto/proto_room"

	l4g "github.com/alecthomas/log4go"
	"xfx/proto/proto_game"
)

// GameState 游戏状态
type GameState int

const (
	k_Ready  GameState = 0 // 准备阶段
	k_Gaming           = 1 // 战斗中阶段
	k_Over             = 2 // 结束阶段
	k_Stop             = 3 // 停止
)

const (
	MaxGameFrame          uint32 = 30 * 5000 // 每局最大帧数
	BroadcastOffsetFrames        = 10        // 每隔多少帧广播一次
	kMaxFrameDataPerMsg          = 60        // 每个消息包最多包含多少个帧数据
	kBadNetworkThreshold         = 15        // 这个时间段没有收到心跳包认为他网络很差，不再持续给发包(网络层的读写时间设置的比较长，客户端要求的方案)
)

type gameListener interface {
	OnJoinGame(uint64, uint64)
	OnGameStart(uint64)
	OnLeaveGame(uint64, uint64)
	OnGameOver(uint64)
}

type packet struct {
	id  uint64
	msg pb_packet.IPacket
}

// GameRun 一局游戏
type GameRun struct {
	base.IGameRun
	id               uint64
	startTime        int64
	State            GameState
	players          map[uint64]base.IPlayer //玩家
	logic            *lockstep               //帧逻辑
	clientFrameCount uint32

	dirty       bool
	enterLoding bool

	listener  gameListener
	logicConn *network.Conn //逻辑服

	sync base.ISync //同步
}

// NewGame 构造游戏
func NewGameRun(id uint64, info *proto_room.RoomInfo, listener gameListener, logicConn *network.Conn) *GameRun {
	g := &GameRun{
		id:          id,
		players:     make(map[uint64]base.IPlayer),
		logic:       newLockstep(),
		startTime:   time.Now().Unix(),
		listener:    listener,
		enterLoding: false,
		logicConn:   logicConn,
	}

	//同步
	g.sync = NewGameSync(g.logic)

	//玩家
	for k := 0; k < len(info.GetPlayers()); k++ {
		g.players[uint64(info.GetPlayers()[k].GetCommonPlayerInfo().GetPlayerId())] = player.NewPlayer(info.GetPlayers()[k], k)
	}

	log.Info("构建一局游戏,游戏ID == %v, 玩家有: %v", id, info.GetPlayers())
	return g
}

// JoinGame 加入游戏
func (g *GameRun) OnJoinGame(id uint64, conn *network.Conn) bool {

	msg := &proto_game.S2CConnectMsg{
		State: int32(proto_room.ERRORCODE_ERR_Ok),
	}

	p, ok := g.players[id]
	if !ok {
		log.Error("[game(%d)] player[%d] join room failed", g.id, id)
		return false
	}

	if k_Ready != g.State && k_Gaming != g.State {
		msg.State = int32(proto_room.ERRORCODE_ERR_RoomState)
		sid, _ := proto_id.MessageID(&proto_game.S2CConnectMsg{})
		p.SendMessage(pb_packet.NewPacket(sid, msg))
		log.Error("[game(%d)] player[%d] game is over", g.id, id)
		return true
	}

	// 把现有的玩家顶掉
	if nil != p.GetClient() {
		// TODO 这里有多线程操作的危险 如果调 p.client.Close() 会把现有刚进来的玩家提调
		p.GetClient().PutExtraData(nil)
		log.Error("[game(%d)] player[%d] replace", g.id, id)
	}

	p.Connect(conn)

	sid, _ := proto_id.MessageID(&proto_game.S2CConnectMsg{})
	p.SendMessage(pb_packet.NewPacket(sid, msg))

	g.listener.OnJoinGame(g.id, id)

	log.Info("加入一名玩家: %v", id)
	return true
}

// LeaveGame 离开游戏
func (g *GameRun) OnLeaveGame(id uint64) bool {

	p, ok := g.players[id]
	if !ok {
		return false
	}

	p.Cleanup()

	g.listener.OnLeaveGame(g.id, id)

	return true
}

// ProcessMsg 处理消息
func (g *GameRun) OnProcessMsg(id uint64, msg *pb_packet.Packet) {
	l4g.Info("[game(%d)] processMsg msg=[%d]", g.id, msg.GetMessageID())
	player, ok := g.players[id]
	if !ok {
		l4g.Error("[game(%d)] processMsg player[%d] msg=[%d]", g.id, player.GetID(), msg.GetMessageID())
		return
	}

	msgID, _ := proto_id.NewMessage(msg.GetMessageID())

	switch msgID.(type) {
	case *proto_game.C2SReadyMsg:
	case *proto_game.C2SPing:
		ProId, _ := proto_id.MessageID(&proto_game.S2CPong{})
		player.SendMessage(pb_packet.NewPacket(uint32(ProId), nil))
		player.RefreshHeartbeatTime()

	case *proto_game.C2SProgressMsg:
		if g.State > k_Ready {
			break
		}
		m := &proto_game.C2SProgressMsg{}
		if err := msg.Unmarshal(m); nil != err {
			log.Error("[game(%d)] processMsg player[%d] msg=[%d] UnmarshalPB error:[%s]", g.id, player.GetID(), msg.GetMessageID(), err.Error())
			return
		}
		player.SetloadProgress(m.GetPro())
		ProId, _ := proto_id.MessageID(&proto_game.S2CProgressMsg{})
		playerPro := make(map[int64]int32)
		for _, v := range g.GetAllPlayer() {
			playerPro[int64(v.GetID())] = v.GetloadProgress()
		}
		msg := pb_packet.NewPacket(ProId, &proto_game.S2CProgressMsg{
			Pro: playerPro,
		})
		log.Debug("加载进度:%v", m.GetPro())
		g.broadcast(msg)

		if m.GetPro() >= 100 {
			if k_Ready == g.State {
				g.OnGameReady(player)
			} else if k_Gaming == g.State {
				g.OnGameReady(player)
				// 重连进来 TODO 对重连进行检查，重连比较耗费
				g.doReconnect(player)
				log.Info("[game(%d)] doReconnect [%d]", g.id, player.GetID())
			} else {
				log.Error("[game(%d)] ID_MSG_Ready player[%d] state error:[%d]", g.id, player.GetID(), g.State)
			}
		}
	case *proto_game.C2SGameOver:
		m := &proto_game.C2SGameOver{}
		if err := msg.Unmarshal(m); nil != err {
			log.Error("[game(%d)] processMsg player[%d] msg=[%d] UnmarshalPB error:[%s]", g.id, player.GetID(), msg.GetMessageID(), err.Error())
			return
		}
		g.State = k_Over
	default:
		log.Debug("[game(%d)] processMsg unknown message id[%d]", msgID)
	}

}

// 广播给其他玩家
func (g *GameRun) BroadcastExclude(msg pb_packet.IPacket, id uint64) {
	g.broadcastExclude(msg, id)
}

// 广播给指定玩家
func (g *GameRun) BroadcastPlayers(msg pb_packet.IPacket, ids []uint64) {
	g.broadcastPlayers(msg, ids)
}

// 广播给所有玩家
func (g *GameRun) BroadcastAll(msg pb_packet.IPacket) {
	g.broadcast(msg)
}

// Tick 主逻辑
func (g *GameRun) Tick(now int64) bool {

	switch g.State {
	case k_Ready:
		if g.checkReady() && g.enterLoding == false && g.checkLoaded() == false {
			g.enterLoding = true
			id, _ := proto_id.MessageID(&proto_game.S2CEnterGameMsg{})
			ret := pb_packet.NewPacket(uint32(id), &proto_game.S2CEnterGameMsg{})
			g.broadcast(ret)
		}

		if g.checkReady() && g.checkLoaded() == true {
			if g.getOnlinePlayerCount() > 0 {
				// 大于最大准备时间，只要有在线的，就强制开始
				g.State = k_Gaming
				g.OnGameStart()
				log.Debug("[game(%d)] force start game because ready state is timeout ", g.id)
			} else {
				// 全都没连进来，直接结束
				g.State = k_Over
				log.Debug("[game(%d)] game over!! nobody ready", g.id)
			}
		}

		return true
	case k_Gaming:
		if g.checkOver() {
			g.State = k_Over
			log.Info("[game(%d)] game over successfully!!", g.id)
			return true
		}

		if g.isTimeout() {
			g.State = k_Over
			log.Debug("[game(%d)] game timeout", g.id)
			return true
		}
		g.logic.tick()
		g.broadcastFrameData()

		return true
	case k_Over:
		g.OnGameOver()
		g.State = k_Stop
		log.Info("[game(%d)] do game over", g.id)
		return true
	case k_Stop:
		return false
	}

	return false
}

// Close 关闭游戏
func (g *GameRun) OnGameClose() {
	id, _ := proto_id.MessageID(&proto_game.S2CGameCloseMsg{})
	msg := pb_packet.NewPacket(id, &proto_game.S2CGameCloseMsg{})
	g.broadcast(msg)
}

// Cleanup 清理游戏
func (g *GameRun) OnCleanup() {
	for _, v := range g.players {
		v.Cleanup()
	}
	g.players = make(map[uint64]base.IPlayer)

}

func (g *GameRun) OnGameReady(p base.IPlayer) {
	if p.GetIsReady() == true {
		return
	}

	p.SetIsReady(true)
}

func (g *GameRun) OnGameStart() {
	g.clientFrameCount = 0
	g.logic.reset()

	for _, v := range g.players {
		v.SetIsReady(true)
		v.SetloadProgress(100)
	}
	g.startTime = time.Now().Unix()
	g.enterLoding = true
	msg := &proto_game.S2CStartMsg{
		TimeStamp: g.startTime,
	}
	id, _ := proto_id.MessageID(&proto_game.S2CStartMsg{})
	ret := pb_packet.NewPacket(uint32(id), msg)

	g.broadcast(ret)
	g.listener.OnGameStart(g.id)
}

func (g *GameRun) OnGameOver() {
	g.listener.OnGameOver(g.id)

	//通知客户端游戏结束
	id, _ := proto_id.MessageID(&proto_game.S2CGameOver{})
	ret := pb_packet.NewPacket(id, &proto_game.S2CGameOver{})
	g.broadcast(ret)
}

func (g *GameRun) doReconnect(p base.IPlayer) {

	g.enterLoding = true
	msg := &proto_game.S2CStartMsg{
		TimeStamp: g.startTime,
	}
	id, _ := proto_id.MessageID(&proto_game.S2CStartMsg{})
	ret := pb_packet.NewPacket(uint32(id), msg)
	p.SendMessage(ret)

	framesCount := g.clientFrameCount
	var i uint32 = 0
	c := 0
	frameMsg := &proto_game.S2CFrameMsg{}

	for ; i < framesCount; i++ {

		frameData := g.logic.getFrame(i)
		if nil == frameData && i != (framesCount-1) {
			continue
		}

		f := &proto_game.FrameData{
			FrameID: i,
		}

		if nil != frameData {
			f.Input = frameData.cmds
		}
		frameMsg.Frames = append(frameMsg.Frames, f)
		c++

		if c >= kMaxFrameDataPerMsg || i == (framesCount-1) {
			id, _ := proto_id.MessageID(&proto_game.S2CFrameMsg{})
			p.SendMessage(pb_packet.NewPacket(uint32(id), frameMsg))
			c = 0
			frameMsg = &proto_game.S2CFrameMsg{}
		}
	}

	p.SetSendFrameCount(g.clientFrameCount)

}

func (g *GameRun) broadcastFrameData() {

	framesCount := g.logic.getFrameCount()

	if !g.dirty && framesCount-g.clientFrameCount < BroadcastOffsetFrames {
		return
	}

	defer func() {
		g.dirty = false
		g.clientFrameCount = framesCount
	}()

	//msg := &proto_game.S2CFrameMsg{}
	//
	//for i := g.clientFrameCount; i < g.logic.getFrameCount(); i++ {
	//	frameData := g.logic.getFrame(i)
	//
	//	if nil == frameData && i != (g.logic.getFrameCount()-1) {
	//		continue
	//	}
	//
	//	f := &proto_game.FrameData{}
	//	f.FrameID = proto.Uint32(i)
	//	msg.Frames = append(msg.Frames, f)
	//
	//	if nil != frameData {
	//		f.Input = frameData.cmds
	//	}
	//
	//}
	//if len(msg.Frames) > 0 {
	//	id := proto_id.MessageID(&proto_game.S2CFrameMsg{})
	//	g.broadcast(pb_packet.NewPacket(uint16(id), msg))
	//}

	now := time.Now().Unix()

	for _, p := range g.players {

		// 掉线的
		if !p.PlayerIsOnline() {
			continue
		}

		if !p.GetIsReady() {
			continue
		}

		// 网络不好的
		if now-p.GetLastHeartbeatTime() >= kBadNetworkThreshold {
			continue
		}

		// 获得这个玩家已经发到哪一帧
		i := p.GetSendFrameCount()
		c := 0
		msg := &proto_game.S2CFrameMsg{}
		for ; i < framesCount; i++ {
			frameData := g.logic.getFrame(i)
			if nil == frameData {
				continue
			}

			f := &proto_game.FrameData{
				FrameID: i,
			}

			f.Input = frameData.cmds
			msg.Frames = append(msg.Frames, f)
			c++
			// 如果是最后一帧或者达到这个消息包能装下的最大帧数，就发送
			if i >= (framesCount-1) || c >= kMaxFrameDataPerMsg {
				id, _ := proto_id.MessageID(&proto_game.S2CFrameMsg{})
				p.SendMessage(pb_packet.NewPacket(id, msg))
				c = 0
				msg = &proto_game.S2CFrameMsg{}
			}

		}

		p.SetSendFrameCount(framesCount)

	}

}

func (g *GameRun) broadcast(msg pb_packet.IPacket) {
	for _, v := range g.players {
		v.SendMessage(msg)
	}
}

func (g *GameRun) broadcastExclude(msg pb_packet.IPacket, id uint64) {
	for _, v := range g.players {
		if v.GetID() == id {
			continue
		}
		v.SendMessage(msg)
	}
}

func (g *GameRun) broadcastPlayers(msg pb_packet.IPacket, ids []uint64) {
	for _, b := range ids {
		g.GetPlayer(b).SendMessage(msg)
	}
}

func (g *GameRun) GetAllPlayer() []base.IPlayer {
	var arr []base.IPlayer
	for _, value := range g.players {
		arr = append(arr, value)
	}

	return arr
}

func (g *GameRun) GetPlayer(id uint64) base.IPlayer {

	return g.players[id]
}

func (g *GameRun) getPlayerCount() int {

	return len(g.players)
}

func (g *GameRun) getOnlinePlayerCount() int {

	i := 0
	for _, v := range g.players {
		if v.PlayerIsOnline() {
			i++
		}
	}

	return i
}

// 检查是否全部加载完毕
func (g *GameRun) checkLoaded() bool {
	for _, v := range g.players {
		if v.GetloadProgress() < 100 {
			return false
		}
	}

	return true
}

func (g *GameRun) checkReady() bool {
	for _, v := range g.players {
		if !v.GetIsReady() {
			return false
		}
	}

	return true
}

func (g *GameRun) checkOver() bool {
	// 只要有人没发结果并且还在线，就不结束
	for _, v := range g.players {
		if v.PlayerIsOnline() {
			return false
		}
	}

	return true
}

func (g *GameRun) isTimeout() bool {
	return g.logic.getFrameCount() > MaxGameFrame
}

func (g *GameRun) SetState(state int) {
	g.State = GameState(state)
}

func (g *GameRun) GetSync() base.ISync {
	return g.sync
}

func (g *GameRun) GetLogicServer() *network.Conn {
	return g.logicConn
}

func (g *GameRun) GetRoomId() int32 {
	return int32(g.id)
}
