package agent

import "time"

type Producer func() Agent

var (
	todoAgent = &emptyAgent{}
)

type Agent interface {
	OnStart(ctx Context)
	OnStop()
	OnTerminated(pid PID, reason int)
	OnMessage(msg interface{}) interface{}
	OnTick(delta time.Duration)
}

type emptyAgent struct{}

func (ag *emptyAgent) OnStart(ctx Context)                   {}
func (ag *emptyAgent) OnStop()                               {}
func (ag *emptyAgent) OnTerminated(pid PID, reason int)      {}
func (ag *emptyAgent) OnMessage(msg interface{}) interface{} { return nil }
func (ag *emptyAgent) OnTick(delta time.Duration)            {}
