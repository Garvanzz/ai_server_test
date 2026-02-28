package wsgate

import (
	"xfx/pkg/gate"
	"xfx/pkg/log"
)

type Agent struct {
	gate.Agent
	session gate.Session
	Gate    gate.Gate
}

func NewAgent() *Agent {
	a := &Agent{}
	return a
}

func (a *Agent) OnInit(gate gate.Gate, session gate.Session) {
	a.Gate = gate
	a.session = session
}

func (a *Agent) GetGate() gate.Gate { return a.Gate }

// Must be threadsafe
func (a *Agent) Send(msg interface{}) {
	err := a.session.Send(msg)
	if err != nil {
		log.Error("ws_agent.Send: %v", err)
	}
}

// Called from I/O
func (a *Agent) Recv(msg interface{})     {}
func (a *Agent) GetSession() gate.Session { return a.session }
func (a *Agent) Close() error {
	return nil
}
