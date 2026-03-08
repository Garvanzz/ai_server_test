package mgate

import (
	"sync"
	"time"
	"xfx/main_server/invoke"
	"xfx/main_server/messages"
	"xfx/pkg/agent"
	"xfx/pkg/gate"
	"xfx/pkg/gate/tcpgate"
	"xfx/pkg/log"
	"xfx/pkg/module/modules"
	"xfx/pkg/utils"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"
)

const PingTime = 30 * time.Second

type Agent struct {
	tcpgate.Agent // 基础agent实现
	modules.BaseAgent
	gate *Gate

	// state     int32
	closeOnce sync.Once
	startedCh chan struct{} // 替换 state int32，用 channel 通知 actor 已就绪

	playerId  int64         // 断开连接时给login的回调
	playerPid agent.PID     // 用于转发网关消息到玩家进程
	pingTime  time.Duration // 及剩余超时时间
}

func NewAgent(gate *Gate) *Agent {
	a := &Agent{
		gate:      gate,
		startedCh: make(chan struct{}),
	}
	return a
}

func (a *Agent) OnInit(gate gate.Gate, session gate.Session) {
	a.BaseAgent.OnInit(a.gate.App)
	a.Agent.OnInit(gate, session)
}

// Message from session
type sessionmsg struct {
	msg any
}

// OnRecv Called from session
func (a *Agent) OnRecv(msg any) {
	//log.Debug("* gate agent recv %v msg: %v", a.GetSession().ID(), msg)

	// Disable wrap msg in agent
	// if a.GetSession().IsClosed() == true {
	// 	log.Debug("* gate agent recv %v msg: %v is closed", a.GetSession().ID(), msg)
	// 	return
	// }

	// if atomic.LoadInt32(&a.state) == 1 {
	// 	a.Context.Cast(a.Context.Self(), &sessionmsg{msg: msg})
	// } else {
	// 	for {
	// 		if atomic.LoadInt32(&a.state) == 1 {
	// 			if a.Context == nil {
	// 				log.Error("* gate agent[%v] context is nil,pid:%v", a.GetSession().ID(), a.Self())
	// 			} else {
	// 				a.Context.Cast(a.Context.Self(), &sessionmsg{msg: msg})
	// 			}
	// 			break
	// 		}
	// 	}
	// }

	// if a.GetSession().IsClosed() {
	//     return
	// }

	// 等待 actor 启动，最多等 3 秒，防止死等
	select {
	case <-a.startedCh:
		// actor 已就绪
	case <-time.After(3 * time.Second):
		log.Error("* gate agent[%v] wait start timeout, drop msg", a.GetSession().ID())
		a.GetSession().Close()
		return
	}

	if a.Context == nil {
		log.Error("* gate agent[%v] context is nil", a.GetSession().ID())
		return
	}
	a.Context.Cast(a.Context.Self(), &sessionmsg{msg: msg})
}

func (a *Agent) OnStart(ctx agent.Context) {
	log.Debug("* gate agent context started:%v", ctx)
	a.BaseAgent.OnStart(ctx)
	a.pingTime = PingTime
	close(a.startedCh) // 通知所有等待方
}

func (a *Agent) OnStop() { // actor stop call
	log.Debug("* agent %d actor stopped", a.GetSession().ID())
	a.GetSession().Close()

	if a.playerId != 0 {
		invoke.LoginClient(a).Disconnect(a.playerId)
	}
}

// kickAfterLoginFailure 登录失败后延迟断开连接，让客户端先收到 S2CLogin 再关闭
func (a *Agent) kickAfterLoginFailure() {
	go func() {
		time.Sleep(500 * time.Millisecond)
		a.GetSession().Close()
	}()
}

func (a *Agent) Close() error {
	a.closeOnce.Do(func() {
		log.Debug("game_agent close:%v", a.GetSession().ID())
		if a.Context == nil {
			log.Error("game_agent close: context is nil, session:%v", a.GetSession().ID())
			// Context 为 nil 说明 actor 还未启动，直接关闭 session
			a.GetSession().Close()
			return
		}
		a.Context.Stop()
	})
	return nil
}

func (a *Agent) OnTick(delta time.Duration) {
	a.pingTime -= delta
	if a.pingTime <= 0 {
		// Session.Close → CloseCallback → agent.Close → Context.Stop → OnStop → Disconnect
		a.GetSession().Close()
	}
}

func (a *Agent) OnMessage(msg any) any {
	// Receive msg from I/O
	switch m := msg.(type) {
	case *sessionmsg:
		a.OnSessionMessage(m.msg)
	case *proto_player.S2CKick:
		a.playerId = 0
		a.playerPid = nil
		a.Send(m)

		// 等待 Kick 包发出后再关闭，最多等 500ms
		go func() {
			a.GetSession().CloseWithFlush(500 * time.Millisecond)
		}()
	default:
		// 发给客户端
		//log.Debug("gate agent sent to client: %v", m)
		a.Send(m)
	}
	return nil
}

// OnSessionMessage 转发网关消息
func (a *Agent) OnSessionMessage(msg any) {
	switch m := msg.(type) {
	case *proto_player.C2SLogout: // 登出
		invoke.LoginClient(a).Logout(a.playerId)
	case *proto_player.C2SPing:
		a.pingTime = PingTime
		//log.Info("ping!!!")
		a.Send(&proto_player.S2CPong{
			ZoneOffset: utils.Now().Unix(),
		})
	case *proto_player.C2SLogin: // 登录
		loginResult, err := invoke.LoginClient(a).Login(&messages.Login{
			Session: a.Context.Self(),
			Request: m,
		})

		log.Debug("agent login result : %v", loginResult)
		if err != nil {
			a.Send(&proto_player.S2CLogin{State: proto_public.CommonState_Faild})
			a.kickAfterLoginFailure()
			return
		}

		state := proto_public.CommonState(loginResult.Result)
		if state != proto_public.CommonState_Success {
			a.Send(&proto_player.S2CLogin{
				State:      state,
				ZoneOffset: utils.Now().Unix(),
			})
			a.kickAfterLoginFailure()
			return
		}

		a.playerId = loginResult.PlayerId
		a.playerPid = loginResult.PlayerPid
		a.Context.Cast(loginResult.PlayerPid, &messages.LoginSuccess{})
	default:
		// 转给玩家进程
		if a.playerPid != nil {
			a.Context.Cast(a.playerPid, msg)
		} else {
			log.Error("session message error:%v", msg)
		}
	}
}
