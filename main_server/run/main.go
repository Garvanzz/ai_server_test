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
	"xfx/main_server/logic/paradise"
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
		paradise.Module(),
	)
}

// 启动
func startup(app module.App) {
	id.Init(uint32(app.GetEnv().ID))

	// 设置时间偏移
	utils.SetTimeOffsetEnabled(app.GetEnv().Debug)

	db.Start(app)

	// 从 Redis 加载之前保存的时间偏移
	utils.LoadTimeOffsetFromRedis()

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
