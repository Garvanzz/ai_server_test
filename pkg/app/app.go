package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"xfx/pkg/agent"
	"xfx/pkg/env"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
)

func NewApp(opts ...module.Option) module.App {
	opt := module.Options{}
	for _, o := range opts {
		o(&opt)
	}
	app := new(DefaultApp)
	app.opts = opt
	return app
}

type DefaultApp struct {
	opts     module.Options
	env      *env.Env
	startup  func(app module.App)
	shutdown func(app module.App)
	system   *agent.System
	manager  *modules.ModuleManager
}

func (app *DefaultApp) OnInit() {}

func (app *DefaultApp) OnDestroy() {}

func (app *DefaultApp) Run(mods ...module.Module) {
	e, err := env.LoadEnv()
	if err != nil {
		panic(fmt.Sprintf("load env error %v", err))
	}
	app.env = e
	log.Init(app.env.Log)

	var tick int64
	if app.opts.Fps > 0 {
		tick = int64(time.Second) / int64(app.opts.Fps)
	}

	app.system = agent.NewSystem(
		agent.WithName(app.env.RemoteName),
		agent.WithHost(app.env.RemoteHost),
		agent.WithPort(app.env.RemotePort),
		agent.WithTick(time.Duration(tick)),
	)
	app.system.Start()

	app.system.Root()

	app.OnInit()
	if app.startup != nil {
		app.startup(app)
	}

	// 初始化模块
	manager := modules.NewModuleManager()
	for i := 0; i < len(mods); i++ {
		manager.Register(mods[i])
	}
	manager.Init(app, app.system)
	app.manager = manager

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
	sig := <-c

	// 一定时间内关不了则强制关闭
	timeout := time.NewTimer(app.opts.KillWaitTTL)
	wait := make(chan struct{})
	go func() {
		manager.Destroy(app.system)

		if app.shutdown != nil {
			app.shutdown(app)
		}

		app.system.Stop()
		app.OnDestroy()
		wait <- struct{}{}
	}()
	select {
	case <-timeout.C:
		panic(fmt.Sprintf("app close timeout (signal: %v)", sig))
	case <-wait:
		fmt.Printf("app closing down (signal: %v)\n", sig)
	}
}

func (app *DefaultApp) GetModule(mod string) module.Module {
	return app.manager.Get(mod)
}

func (app *DefaultApp) GetEnv() *env.Env {
	return app.env
}

func (app *DefaultApp) OnStartup(_func func(app module.App)) {
	app.startup = _func
}

func (app *DefaultApp) OnShutdown(_func func(app module.App)) {
	app.shutdown = _func
}

func (app *DefaultApp) System() *agent.System {
	return app.system
}
