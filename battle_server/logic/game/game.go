package game

import (
	"sync"
	"sync/atomic"
	"time"
	"xfx/game_server/pkg/network"
	"xfx/pkg/log"
	"xfx/proto/proto_room"

	"xfx/game_server/pkg/packet/pb_packet"
)

const (
	Frequency   = 60                      // 每分钟心跳频率
	TickTimer   = time.Second / Frequency // 心跳Timer
	TimeoutTime = time.Minute * 60        // 超时时间
)

// 战斗
type Game struct {
	wg sync.WaitGroup

	roomID          uint64
	players         []*proto_room.RoomPlayers
	typeID          int32
	closeFlag       int32
	timeStamp       int64
	secretKey       string
	logicServerPort string

	exitChan chan struct{}
	msgQ     chan *packet
	inChan   chan *network.Conn
	outChan  chan *network.Conn
	game     *GameRun
}

// NewGame 构造
func NewGame(id uint64, info *proto_room.RoomInfo, logicServer string, logicConn *network.Conn) *Game {
	r := &Game{
		roomID:          id,
		players:         info.GetPlayers(),
		typeID:          info.GetType(),
		exitChan:        make(chan struct{}),
		msgQ:            make(chan *packet, 2048),
		outChan:         make(chan *network.Conn, 8),
		inChan:          make(chan *network.Conn, 8),
		timeStamp:       time.Now().Unix(),
		logicServerPort: logicServer,
		secretKey:       "test_room",
	}

	r.game = NewGameRun(id, info, r, logicConn)
	return r
}

// ID room ID
func (r *Game) ID() uint64 {
	return r.roomID
}

// Type room Type
func (r *Game) Type() int32 {
	return r.typeID
}

func (r *Game) GetGame() *GameRun {
	return r.game
}

func (r *Game) Players() []*proto_room.RoomPlayers {
	return r.players
}

// SecretKey secret key
func (r *Game) SecretKey() string {
	return r.secretKey
}

// TimeStamp time stamp
func (r *Game) TimeStamp() int64 {
	return r.timeStamp
}

// IsOver 是否已经结束
func (r *Game) IsOver() bool {
	return atomic.LoadInt32(&r.closeFlag) != 0
}

// HasPlayer 是否有这个player
func (r *Game) HasPlayer(id uint64) bool {
	for _, v := range r.players {
		if uint64(v.CommonPlayerInfo.PlayerId) == id {
			return true
		}
	}

	return false
}

func (r *Game) OnJoinGame(id, pid uint64) {
	log.Info("[room(%d)] onJoinGame %d", id, pid)
}
func (r *Game) OnGameStart(id uint64) {
	log.Info("[room(%d)] onGameStart", id)
}

func (r *Game) OnLeaveGame(id, pid uint64) {
	log.Info("[room(%d)] onLeaveGame %d", id, pid)
}
func (r *Game) OnGameOver(id uint64) {
	atomic.StoreInt32(&r.closeFlag, 1)

	log.Info("[room(%d)] onGameOver", id)

	r.wg.Add(1)

	go func() {
		defer r.wg.Done()
		// TODO
		// http result
	}()

}

// OnConnect network.Conn callback
func (r *Game) OnConnect(conn *network.Conn) bool {

	if conn != nil {
		conn.SetCallback(r) // SetCallback只能在OnConnect里调
		r.inChan <- conn
	}

	log.Info("[room(%d)] OnConnect %d", r.roomID, conn.GetExtraData().(uint64))

	return true
}

// OnMessage network.Conn callback
func (r *Game) OnMessage(conn *network.Conn, msg pb_packet.IPacket) bool {

	id, ok := conn.GetExtraData().(uint64)
	if !ok {
		log.Info("[room] OnMessage error conn don't have id")
		return false
	}

	p := &packet{
		id:  id,
		msg: msg,
	}
	r.msgQ <- p

	return true
}

// OnClose network.Conn callback
func (r *Game) OnClose(conn *network.Conn) {
	r.outChan <- conn
	if id, ok := conn.GetExtraData().(uint64); ok {
		log.Info("[room(%d)] OnClose %d", r.roomID, id)
	} else {
		log.Info("[room(%d)] OnClose no id", r.roomID)
	}

}

// Run 主循环
func (r *Game) Run() {
	r.wg.Add(1)
	defer r.wg.Done()
	defer func() {
		/*
			err := recover()
			if nil != err {
				l4g.Error("[room(%d)] Run error:%+v", r.roomID, err)
			}*/
		r.game.OnCleanup()
		log.Info("[room(%d)] quit! total time=[%d]", r.roomID, time.Now().Unix()-r.timeStamp)
	}()

	// 心跳
	tickerTick := time.NewTicker(TickTimer)
	defer tickerTick.Stop()

	// 超时timer
	timeoutTimer := time.NewTimer(TimeoutTime)

	log.Info("[room(%d)] running...", r.roomID)
LOOP:
	for {
		select {
		case <-r.exitChan:
			log.Error("[room(%d)] force exit", r.roomID)
			return
		case <-timeoutTimer.C:
			log.Error("[room(%d)] time out", r.roomID)
			break LOOP
		case msg := <-r.msgQ:
			r.game.OnProcessMsg(msg.id, msg.msg.(*pb_packet.Packet))
		case <-tickerTick.C:
			if !r.game.Tick(time.Now().Unix()) {
				log.Info("[room(%d)] tick over", r.roomID)
				break LOOP
			}
		case c := <-r.inChan:
			id, ok := c.GetExtraData().(uint64)
			if ok {
				if r.game.OnJoinGame(id, c) {
					log.Info("[room(%d)] player[%d] join room ok", r.roomID, id)
				} else {
					log.Error("[room(%d)] player[%d] join room failed", r.roomID, id)
					c.Close()
				}
			} else {
				c.Close()
				log.Error("[room(%d)] inChan don't have id", r.roomID)
			}

		case c := <-r.outChan:
			if id, ok := c.GetExtraData().(uint64); ok {
				r.game.OnLeaveGame(id)
			} else {
				c.Close()
				log.Error("[room(%d)] outChan don't have id", r.roomID)
			}
		}
	}

	r.game.OnGameClose()

	for i := 3; i > 0; i-- {
		<-time.After(time.Second)
		log.Info("[room(%d)] quiting %d...", r.roomID, i)
	}
}

// Stop 强制关闭
func (r *Game) Stop() {
	close(r.exitChan)
	r.wg.Wait()
}
