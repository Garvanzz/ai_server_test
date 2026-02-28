package wsgate

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"xfx/pkg/gate"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/pkg/net/ws"
)

// var S gate.Gate = new(Gate)

type Gate struct {
	modules.BaseModule
	ws.SessionHandler
	server      *ws.Server
	sessions    sync.Map
	createAgent func(gate.Gate, gate.Session) (gate.Agent, error)
}

func (gt *Gate) OnInit(app module.App) {
	gt.BaseModule.OnInit(app)

	e := app.GetEnv().Gate
	if e == nil {
		panic("no config gate")
	}

	conf := ws.DefaultConfig
	if e.WriteWait != 0 {
		conf.WriteWait = e.WriteWait
	}
	if e.PongWait != 0 {
		conf.PongWait = e.PongWait
	}
	if e.PingPeriod != 0 {
		conf.PingPeriod = e.PingPeriod
	}
	if e.MaxMessageSize != 0 {
		conf.MaxMessageSize = int64(e.MaxMessageSize)
	}
	if e.MessageBufferSize != 0 {
		conf.MessageBufferSize = e.MessageBufferSize
	}

	server := ws.NewServer(conf)
	server.HandleProduceSessionHandler(func(*http.Request) ws.ISessionHandler { return gt })
	err := server.Serve(e.Host)
	if err != nil {
		panic("Gate start error")
	}

	gt.server = server
}

func (gt *Gate) SetCreateAgent(fn func(gate.Gate, gate.Session) (gate.Agent, error)) {
	gt.createAgent = fn
}

func (gt *Gate) NewSession(conn interface{}) (gate.Session, error) {
	s, ok := conn.(*ws.Session)
	if !ok {
		return nil, fmt.Errorf("ws gate create sessin error")
	}
	return NewSession(s), nil
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

func (gt *Gate) Close(sessionID string) error {
	if value, ok := gt.sessions.Load(sessionID); ok {
		session := value.(gate.Session)
		return session.Close()
	} else {
		return nil
	}
}

func (gt *Gate) GetSession(sessionID string) gate.Session {
	value, ok := gt.sessions.Load(sessionID)
	if ok {
		return value.(gate.Session)
	}
	return nil
}

// -------------------------------------------------------------------------
func (gt *Gate) HandleTextMessage(*ws.Session, []byte) {
	log.Error("text message is not supported")
}

func (gt *Gate) HandleBinaryMessage(s *ws.Session, data []byte) {
	value, ok := s.Get("#session")
	if !ok {
		log.Error("No session for session %v", s.ID())
		return
	}
	session, _ := value.(*Session)
	session.doRecv(data)
}

func (gt *Gate) HandleError(s *ws.Session, err error) {
	log.Error("* session %d error: %v", s.ID(), err)
}

func (gt *Gate) HandleClose(s *ws.Session, code int, text string) error {
	log.Debug("* session %d closed", s.ID())
	return nil
}

// 建立了一个连接
func (gt *Gate) HandleConnect(s *ws.Session) error {
	session, err := gt.NewSession(s)
	if err != nil {
		s.Close()
		return err
	}
	gt.sessions.Store(session.ID(), session)

	a, err := gt.NewAgent(session)
	if err != nil {
		s.Close()
		return err
	}
	s.Set("#id", session.ID())
	s.Set("#session", session)
	s.Set("#agent", a)
	log.Debug("* session %d connected", s.ID())
	return nil
}

func (gt *Gate) HandleDisconnect(s *ws.Session) {
	log.Debug("* session %d disconnected", s.ID())
	if value, ok := s.Get("#agent"); ok {
		agent := value.(gate.Agent)
		agent.Close()
	}
	sessionID := strconv.FormatUint(s.ID(), 10)
	gt.sessions.Delete(sessionID)
	s.Delete("#id")
	s.Delete("#session")
	s.Delete("#agent")
}
