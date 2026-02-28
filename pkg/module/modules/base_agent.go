package modules

import (
	"fmt"
	"time"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/invoke"
)

var _ module.Agent = (*BaseAgent)(nil)

type BaseAgent struct {
	Context module.Context
	App     module.App
	invoker *invoke.Invoker
}

func (m *BaseAgent) OnInit(app module.App) { m.App = app }
func (m *BaseAgent) GetApp() module.App    { return m.App }

func (m *BaseAgent) Cast(mod string, msg interface{}) {
	pid := m.GetApp().GetModule(mod).Self()
	if pid != nil {
		m.Context.Cast(pid, msg)
	} else {
		log.Error("Cast:%v", mod)
	}
}

func (m *BaseAgent) Call(mod string, msg interface{}) (interface{}, error) {
	pid := m.GetApp().GetModule(mod).Self()
	if pid != nil {
		return m.Context.Call(pid, msg)
	} else {
		return nil, fmt.Errorf("can't find module %s", mod)
	}
}

func (m *BaseAgent) CallNR(mod string, msg interface{}) error {
	pid := m.GetApp().GetModule(mod).Self()
	if pid != nil {
		return m.Context.CallNR(pid, msg)
	} else {
		return fmt.Errorf("can't find module %s", mod)
	}
}

func (m *BaseAgent) OnStart(ctx module.Context) {
	m.Context = ctx
	m.Context.SetMessageFilter(m._message)
}

func (m *BaseAgent) Self() module.PID {
	if m.Context == nil {
		log.Error("BaseAgent Self Context is nil")
		return nil
	}
	return m.Context.Self()
}
func (m *BaseAgent) OnStop()                                 {}
func (m *BaseAgent) OnTerminated(pid module.PID, reason int) {}
func (m *BaseAgent) OnTick(delta time.Duration)              {}
func (m *BaseAgent) OnMessage(msg interface{}) interface{}   { return nil }

func (m *BaseAgent) Invoke(mod, fn string, args ...interface{}) (interface{}, error) {
	pid := m.GetApp().GetModule(mod).Self()
	if pid == nil {
		return nil, fmt.Errorf("can't find module(%s) actor", mod)
	}
	return m.InvokeP(pid, fn, args...)
}

func (m *BaseAgent) InvokeP(pid module.PID, fn string, args ...interface{}) (interface{}, error) {
	result, err := m.Context.Call(pid, &invokeRequest{
		Fn:   fn,
		Args: args,
	})
	if err != nil {
		return nil, err
	}

	if m, ok := result.(*invokeResponse); ok {
		return m.res, m.err
	} else {
		return nil, fmt.Errorf("invoke unexpect response message type error")
	}
}

func (m *BaseModule) Register(fn string, f interface{}) {
	if m.invoker == nil {
		m.invoker = invoke.NewInvoker()
	}
	m.invoker.Register(fn, f)
}

type invokeRequest struct {
	Fn   string
	Args []interface{}
}

type invokeResponse struct {
	res interface{}
	err error
}

func (m *BaseAgent) _message(msg interface{}) (bool, interface{}) {
	switch e := msg.(type) {
	case *invokeRequest:
		// process invoke
		res, err := m.invoker.Invoke(e.Fn, e.Args...)
		return true, &invokeResponse{
			res: res,
			err: err,
		}
	default:
		return false, nil
	}
}
