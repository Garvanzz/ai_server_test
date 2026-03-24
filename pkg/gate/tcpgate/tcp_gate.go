package tcpgate

import (
	"errors"
	"fmt"
	"io"
	"net"
	"xfx/pkg/gate"
	"xfx/pkg/gate/tcpgate/codec"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/pkg/net/tcp"
)

var _ gate.Gate = (*Gate)(nil)

type Gate struct {
	modules.BaseModule
	server      *tcp.Server
	createAgent func(gate.Gate, gate.Session) (gate.Agent, error)
}

func (gt *Gate) OnInit(app module.App) {
	gt.BaseModule.OnInit(app)

	if app.GetEnv().TcpGate == nil {
		panic("tcp gate config is nil")
	}

	cfg := app.GetEnv().TcpGate
	host := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	log.Debug("tcp连接:%s", host)

	sendChanSize := cfg.MessageBufferSize
	if sendChanSize <= 0 {
		sendChanSize = 2000
	}

	server, err := tcp.Listen("tcp", host, tcp.ProtocolFunc(codec.NewParser), sendChanSize, tcp.HandlerFunc(gt.handleConnect))
	if err != nil {
		panic(err)
	}

	gt.server = server

	go server.Serve()
}

func (gt *Gate) handleConnect(sess *tcp.Session) {
	agent, err := gt.NewAgent(sess)
	if err != nil {
		sess.Close()
		return
	}

	//sess.Codec().(*codec.Parser).SetAgent(agent)

	// 不再通过 codec 持有 agent，改用 CloseCallback
	// 当 session 关闭时，通知上层 agent 做清理
	sess.AddCloseCallback(gt, sess.ID(), func() {
		agent.Close()
	})

	log.Debug("* session %d connected", sess.ID())

	for {
		msg, err := sess.Receive()
		if err != nil {
			// 区分正常断开（EOF）和真正的错误
			if err == io.EOF {
				log.Debug("session %d disconnected (EOF)", sess.ID())
			} else if errors.Is(err, net.ErrClosed) {
				log.Debug("session %d disconnected (connection closed)", sess.ID())
			} else {
				log.Warn("gate handler msg error: %v", err)
			}
			return
		}
		agent.OnRecv(msg)
	}
}

func (gt *Gate) NewAgent(session gate.Session) (gate.Agent, error) {
	var (
		agent gate.Agent
		err   error
	)

	if gt.createAgent == nil {
		agent = NewAgent()
	} else {
		agent, err = gt.createAgent(gt, session)
	}
	if err != nil {
		return nil, err
	}
	return agent, nil
}

func (gt *Gate) SetCreateAgent(fn func(gate.Gate, gate.Session) (gate.Agent, error)) {
	gt.createAgent = fn
}

func (gt *Gate) OnDestroy() {
	gt.server.Stop()
}
