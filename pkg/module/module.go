package module

import (
	"time"
	"xfx/pkg/agent"
	"xfx/pkg/env"
)

type App interface {
	OnInit()
	OnDestroy()
	Run(mods ...Module)
	GetEnv() *env.Env
	OnStartup(func(app App))
	OnShutdown(func(app App))
	GetModule(mod string) Module
	System() *agent.System
}

// Module 基本模块定义
type Module interface {
	Agent
	GetType() string
	GetApp() App
	OnInit(app App)
	OnDestroy()
}

type (
	PID     = agent.PID
	Context = agent.Context
)

type Agent interface {
	GetApp() App
	Self() PID
	OnStart(ctx Context)
	OnStop()
	OnTerminated(pid PID, reason int)
	OnMessage(msg interface{}) interface{}
	OnTick(delta time.Duration)
	Cast(mod string, msg interface{})
	Call(mod string, msg interface{}) (interface{}, error)
	CallNR(mod string, msg interface{}) error
	Invoke(mod, fn string, args ...interface{}) (interface{}, error)
	InvokeP(pid PID, fn string, args ...interface{}) (interface{}, error)
}
