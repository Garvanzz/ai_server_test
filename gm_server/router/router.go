package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xfx/gm_server/logic"
	"xfx/gm_server/middleware"
)

// Register 注册 GM 服务所有路由：鉴权组（无 token）、需鉴权的 GM 业务组
func Register(r *gin.Engine) {
	// CORS 中间件 - 处理跨域请求
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	r.Use(gin.Recovery())
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	// ========== 鉴权（无需 token）==========
	auth := r.Group("/gm/auth")
	{
		auth.POST("/login", logic.GmLogin)
		auth.POST("/logout", logic.GmLogout)
		auth.POST("/user-info", logic.GmAdminUserInfo)
	}

	// ========== 以下所有 GM 接口均需请求头携带 xiaoxiaoxiyou(token) 鉴权 ==========
	gm := r.Group("/gm")
	gm.Use(middleware.GmAuth)

	// ========== 区服与进程 ==========
	servers := gm.Group("/servers")
	{
		servers.POST("/list", logic.GmGetServerList)
		servers.POST("/create", logic.GmCreateManagedServer)
		servers.POST("/update", logic.GmUpdateManagedServer)
		servers.POST("/delete", logic.GmDeleteManagedServer)
		servers.POST("/batch-update", logic.GmBatchUpdateManagedServer)
		servers.POST("/start", logic.GmStartServer)
		servers.POST("/stop", logic.GmStopServer)
		servers.POST("/restart", logic.GmReStartServer)
		servers.POST("/game-list", logic.GmGetGameServerList)
		servers.POST("/game-start", logic.GmStartGameServer)
		servers.POST("/game-stop", logic.GmStopGameServer)
		servers.POST("/game-restart", logic.GmReStartGameServer)
		servers.POST("/game-list-all", logic.GmGetGameServerProcessList)
		servers.GET("/time", logic.GmGetServerTime)
		servers.POST("/time", middleware.RequirePermission(1, 2), logic.GmSetServerTime)
	}

	serverGroups := gm.Group("/server-groups")
	serverGroups.Use(middleware.RequirePermission(1, 2))
	{
		serverGroups.POST("/list", logic.GmGetServerGroupManageList)
		serverGroups.POST("/create", logic.GmCreateServerGroup)
		serverGroups.POST("/update", logic.GmUpdateServerGroup)
		serverGroups.POST("/delete", logic.GmDeleteServerGroup)
	}

	// ========== 玩家数据查询 ==========
	players := gm.Group("/players")
	{
		players.POST("/info", logic.GmGetPlayerInfo)
		players.POST("/game-info", logic.GmGetPlayerGameInfo)
		players.POST("/bag", logic.GmItem)
		players.POST("/equip", logic.GmEquip)
		players.POST("/equip-delete", logic.GmDeleteEquip)
		players.POST("/hero", logic.GmHero)
		players.POST("/hero-edit", logic.GmEditHero)
		players.POST("/stage-info", logic.GmGetStageInfo)
		players.POST("/stage-set", logic.GmSetStageInfo)
		players.POST("/stage-add", logic.GmAddStageInfo)
		players.POST("/stage-delete", logic.GmDeleteStageInfo)
		players.POST("/grant-item", logic.GmAddItem)
		players.POST("/grant-item-all", logic.GmOneKeyAddItem)
		players.POST("/item-delete", logic.GmDeleteItem)
		players.POST("/kick", logic.GmKickPlayer)
	}

	// ========== 需转发 main_server 的接口 ==========
	gm.POST("/mail/send", logic.GmCreateAdminMail)
	gm.POST("/notice/horse", logic.GmSendHorse)
	gm.POST("/notice/send", logic.GmSendNotice)
	gm.POST("/notices/list", logic.GmGetNoticeList)

	// ========== 订单、帮会 ==========
	gm.POST("/orders/list", logic.GmGetOrderList)
	gm.POST("/orders/cache-list", logic.GmGetCacheOrderList)
	gm.POST("/guild/list", logic.GmGetGuildList)

	// ========== 合服管理（仅管理员以上 permission=1,2）==========
	merge := gm.Group("/merge")
	merge.Use(middleware.RequirePermission(1, 2))
	{
		merge.POST("/plan/create", logic.GmCreateMergePlan)
		merge.POST("/plan/list", logic.GmListMergePlans)
		merge.POST("/precheck", logic.GmPrecheckMerge)
		merge.POST("/redis-check", logic.GmRedisMergeCheck)
		merge.POST("/redis-script", logic.GmExportRedisMergeScript)
		merge.POST("/execute", logic.GmExecuteMergePlan)
		merge.POST("/rollback", logic.GmRollbackMergePlan)
		merge.POST("/conflicts", logic.GmListMergeConflicts)
	}

	// ========== 进程管理（仅管理员以上 permission=1,2）==========
	processes := gm.Group("/processes")
	processes.Use(middleware.RequirePermission(1, 2))
	{
		processes.POST("/list", logic.GmListProcesses)
		processes.POST("/create", logic.GmCreateProcess)
		processes.POST("/update", logic.GmUpdateProcess)
		processes.POST("/delete", logic.GmDeleteProcess)
		processes.POST("/start", logic.GmStartProcess)
		processes.POST("/stop", logic.GmStopProcess)
		processes.POST("/restart", logic.GmRestartProcess)
		processes.POST("/build", logic.GmBuildProcess)
	}

	// ========== 热更 ==========
	hotfix := gm.Group("/hotfix")
	{
		hotfix.POST("/list", logic.GmGetHotUpdate)
		hotfix.POST("/version-edit", logic.GmEditHotUpdateVersion)
		hotfix.POST("/version-create", logic.GmCreateHotUpdateVersion)
		hotfix.POST("/version-delete", logic.GmDeleteHotUpdateVersion)
		hotfix.POST("/path-create", logic.GmCreateHotUpdatePath)
	}

	// ========== 上传、配置、编译 ==========
	gm.POST("/upload", logic.GmUpload)
	gm.POST("/config/update", logic.GmUpdateConfig)
	gm.POST("/config/game-update", logic.GmGameUpdateConfig)
	gm.POST("/build", logic.GmBuildServer)
	gm.POST("/build/game", logic.GmGameBuildServer)

	// ========== GM 账号管理（仅超级管理员 permission=1）==========
	accounts := gm.Group("/accounts")
	accounts.Use(middleware.RequirePermission(1))
	{
		accounts.POST("/list", logic.GmListAdminAccounts)
		accounts.POST("/create", logic.GmCreateAdminAccount)
		accounts.POST("/update", logic.GmUpdateAdminAccount)
		accounts.POST("/delete", logic.GmDeleteAdminAccount)
	}

	// ========== 活动管理 ==========
	activity := gm.Group("/activity")
	{
		activity.POST("/list", logic.GmActivityList)
		activity.POST("/get-by-act-id", logic.GmActivityGetByActId)
		activity.POST("/get-by-cfg-id", logic.GmActivityGetByCfgId)
		activity.POST("/stop", logic.GmActivityStop)
		activity.POST("/recover", logic.GmActivityRecover)
		activity.POST("/close", logic.GmActivityClose)
		activity.POST("/restart", logic.GmActivityRestart)
		activity.POST("/remove", logic.GmActivityRemove)
		activity.POST("/close-by-cfg-id", logic.GmActivityCloseByCfgId)
		activity.POST("/stop-by-type", logic.GmActivityStopByType)
		activity.POST("/adjust-time", logic.GmActivityAdjustTime)
		activity.POST("/force-start", logic.GmActivityForceStart)
		activity.POST("/player-count", logic.GmActivityPlayerCount)
	}
}
