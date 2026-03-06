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
	"xfx/main_server/logic/activity"
	"xfx/main_server/logic/battle"
	"xfx/main_server/logic/common"
	"xfx/main_server/logic/guild"
	"xfx/main_server/logic/huaguoshan"
	"xfx/main_server/logic/login"
	"xfx/main_server/logic/mail"
	"xfx/main_server/logic/match"
	"xfx/main_server/logic/recruit"
	"xfx/main_server/logic/room"
	"xfx/main_server/logic/transaction"
	"xfx/pkg/app"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/utils"
	"xfx/pkg/utils/id"
	"xfx/pkg/utils/sensitive"
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
		room.Module(),
		match.Module(),
		recruit.Module(),
		launcher.Module(),
		mail.Module(),
		common.Module(),
		mgate.Module(),
		guild.Module(),
		battle.Module(),
		activity.Module(),
		transaction.Module(),
		huaguoshan.Module(),
	)
}

// 启动
func startup(app module.App) {
	id.Init(uint32(app.GetEnv().ID))

	// 仅在 Debug 模式下启用时间偏移，线上 Debug=false 时始终使用服务器真实时间
	if app.GetEnv().Debug && app.GetEnv().TimeOffsetDays != 0 {
		d := app.GetEnv().TimeOffsetDays
		utils.SetTimeOffset(time.Duration(d) * 24 * time.Hour)
		log.Info("game time offset: +%d days (debug only)", d)
	}

	db.Start(app)

	//加载敏感词库
	sensitive.Init()

	//加载配置
	coreconfig.InitConfig(app.GetEnv().ConfPath)

	event.Init(app.System())

	log.Info("* startup")
}

func shutdown(app module.App) {
	db.Close()

	log.Info("* shutdown")
}
