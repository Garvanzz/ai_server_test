package tcpgate

import (
	"xfx/pkg/gate"
	"xfx/pkg/log"
)

var _ gate.Agent = (*Agent)(nil)

// Agent 上层网关agent 默认实现
type Agent struct {
	sess gate.Session
}

func NewAgent() *Agent {
	a := &Agent{}
	return a
}

func (a *Agent) OnInit(gate gate.Gate, session gate.Session) {
	a.sess = session
}

func (a *Agent) Send(msg any) {
	err := a.sess.Send(msg)
	if err != nil {
		log.Error("tcp_agent.Send: %v", err)
	}
}

func (a *Agent) OnRecv(msg interface{}) {}

func (a *Agent) GetSession() gate.Session { return a.sess }

func (a *Agent) Close() error {
	return nil
}

func (a *Agent) Get(key string) (any, bool) {
	return a.sess.Get(key)
}

func (a *Agent) Set(key string, value any) {
	a.sess.Set(key, value)
}

func (a *Agent) ID() uint64 {
	return a.sess.ID()
}

func (a *Agent) IsClosed() bool {
	return a.sess.IsClosed()
}
