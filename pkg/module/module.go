package module

import (
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

type (
	PID     = agent.PID
	Context = agent.Context
)

// Agent 在 agent 生命周期之上，增加 App、按模块名通信与 Invoke 能力。
// 实现 Agent 的类型同时满足 agent.Agent，可作为 actor 被 System 调度。
type Agent interface {
	agent.Agent
	GetApp() App
	Self() PID
	Cast(mod string, msg interface{})
	Call(mod string, msg interface{}) (interface{}, error)
	CallNR(mod string, msg interface{}) error
	Invoke(mod, fn string, args ...interface{}) (interface{}, error)
	InvokeP(pid PID, fn string, args ...interface{}) (interface{}, error)
}

// Module 是注册在 App 中、按 GetType() 区分的 Agent，具备初始化/销毁钩子。
type Module interface {
	Agent
	GetType() string
	GetApp() App
	OnInit(app App)
	OnDestroy()
}
