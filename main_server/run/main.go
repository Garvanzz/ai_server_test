package main

import (
	_ "embed"
	"time"
	coreconfig "xfx/core/config"
	"xfx/core/db"
	"xfx/core/event"
	mgate "xfx/main_server/gate"
	"xfx/main_server/http"
	"xfx/main_server/launcher"
	"xfx/main_server/logic/common"
	"xfx/main_server/logic/login"
	"xfx/pkg/app"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/utils/id"
)

func main() {
	app := app.NewApp(
		module.WithVersion("1.0.0"),
		module.WithKillWaitTTL(3*time.Second),
		module.WithFps(1),
	)
	app.OnStartup(startup)
	app.OnShutdown(shutdown)
	app.Run(
		http.Module(),
		login.Module(),
		launcher.Module(),
		common.Module(),
		mgate.Module(),
		//activity.Module(),
	)
}

// 启动
func startup(app module.App) {
	id.Init(uint32(app.GetEnv().ID))

	db.Start(app)

	//加载配置
	coreconfig.InitConfig(app.GetEnv().ConfPath)

	event.Init(app.System())

	log.Info("* startup")
}

func shutdown(app module.App) {
	db.Close()

	log.Info("* shutdown")
}
