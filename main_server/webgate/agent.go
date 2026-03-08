package wgate

import (
	"sync/atomic"
	"time"
	"xfx/main_server/invoke"
	"xfx/pkg/utils"
	"xfx/main_server/messages"
	"xfx/pkg/agent"
	"xfx/pkg/gate"
	"xfx/pkg/gate/tcpgate"
	"xfx/pkg/log"
	"xfx/pkg/module/modules"
	Proto_Player "xfx/proto/proto_player"
	Proto_Public "xfx/proto/proto_public"
)

const PingTime = 30 * time.Second

type Agent struct {
	tcpgate.Agent // 基础agent实现
	modules.BaseAgent
	gate *Gate

	state     int32
	playerId  int64         // 断开连接时给login的回调
	playerPid agent.PID     // 用于转发网关消息到玩家进程
	pingTime  time.Duration // 及剩余超时时间
}

func NewAgent(gate *Gate) *Agent {
	a := &Agent{gate: gate}
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
	log.Debug("* gate agent recv %v msg: %v", a.GetSession().ID(), msg)

	// Disable wrap msg in agent
	if a.GetSession().IsClosed() == true {
		log.Debug("* gate agent recv %v msg: %v is closed", a.GetSession().ID(), msg)
		return
	}

	for {
		if atomic.LoadInt32(&a.state) == 1 {
			if a.Context == nil {
				log.Error("* gate agent[%v] context is nil,pid:%v", a.GetSession().ID(), a.Self())
			} else {
				a.Context.Cast(a.Context.Self(), &sessionmsg{msg: msg})
			}
			break
		}
	}
}

func (a *Agent) OnStart(ctx agent.Context) {
	log.Debug("* gate agent context started:%v", ctx)
	a.BaseAgent.OnStart(ctx)
	a.pingTime = PingTime
	atomic.AddInt32(&a.state, 1)
}

func (a *Agent) OnStop() { // actor stop call
	log.Debug("* agent %s actor stopped", a.GetSession().ID())
	a.GetSession().Close()

	if a.playerId != 0 {
		invoke.LoginClient(a).Disconnect(a.playerId)
	}
}

func (a *Agent) Close() error {
	log.Debug("game_agent close:%v", a.GetSession().ID())
	a.Context.Stop()
	return nil
}

func (a *Agent) OnTick(delta time.Duration) {
	a.pingTime -= delta
	if a.pingTime <= 0 {
		// 超时断开连接
		a.GetSession().Close()
		invoke.LoginClient(a).Disconnect(a.playerId)
	}
}

func (a *Agent) OnMessage(msg any) any {
	// Receive msg from I/O
	switch m := msg.(type) {
	case *sessionmsg:
		a.OnSessionMessage(m.msg)
	case *Proto_Player.S2CKick:
		a.playerId = 0
		a.playerPid = nil
		a.Send(m)
		a.Close()
	default:
		// 发给客户端
		log.Debug("gate agent sent to client: %v", m)
		a.Send(m)
	}
	return nil
}

// OnSessionMessage 转发网关消息
func (a *Agent) OnSessionMessage(msg any) {
	switch m := msg.(type) {
	case *Proto_Player.C2SLogout: // 登出
		invoke.LoginClient(a).Logout(a.playerId)
	case *Proto_Player.C2SPing:
		a.pingTime = PingTime
		a.Send(&Proto_Player.S2CPong{
			ZoneOffset: utils.Now().Unix(),
		})
	case *Proto_Player.C2SLogin: // 登录
		loginResult, err := invoke.LoginClient(a).Login(&messages.Login{
			Session: a.Context.Self(),
			Request: m,
		})

		log.Debug("agent login result : %v", loginResult)
		if err == nil {
			a.playerId = loginResult.PlayerId
			a.playerPid = loginResult.PlayerPid
			state := Proto_Public.CommonState(loginResult.Result)
			if state != Proto_Public.CommonState_Success {
				a.Send(&Proto_Player.S2CLogin{State: state})
				return
			}

			a.Context.Cast(loginResult.PlayerPid, &messages.LoginSuccess{})
		} else {
			a.Send(&Proto_Player.S2CLogin{State: Proto_Public.CommonState_Faild})
		}
	default:
		// 转给玩家进程
		if a.playerPid != nil {
			a.Context.Cast(a.playerPid, msg)
		} else {
			log.Error("session message error:%v", msg)
		}
	}
}
