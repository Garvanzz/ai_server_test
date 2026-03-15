package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xfx/gm_server/logic"
	"xfx/gm_server/middleware"
)

// Register 注册 GM 服务所有路由：鉴权组（无 token）、需鉴权的 GM 业务组
func Register(r *gin.Engine) {
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
		servers.POST("/start", logic.GmStartServer)
		servers.POST("/stop", logic.GmStopServer)
		servers.POST("/restart", logic.GmReStartServer)
		servers.POST("/game-list", logic.GmGetGameServerList)
		servers.POST("/game-start", logic.GmStartGameServer)
		servers.POST("/game-stop", logic.GmStopGameServer)
		servers.POST("/game-restart", logic.GmReStartGameServer)
		servers.POST("/game-list-all", logic.GmGetGameServerProcessList)
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
	}

	// ========== 需转发 main_server 的接口 ==========
	gm.POST("/mail/send", logic.GmCreateAdminMail)
	gm.POST("/notice/horse", logic.GmSendHorse)
	gm.POST("/notice/send", logic.GmSendNotice)

	// ========== 订单、帮会 ==========
	gm.POST("/orders/list", logic.GmGetOrderList)
	gm.POST("/orders/cache-list", logic.GmGetCacheOrderList)
	gm.POST("/guild/list", logic.GmGetGuildList)

	// ========== 合服管理 ==========
	merge := gm.Group("/merge")
	{
		merge.POST("/plan/create", logic.GmCreateMergePlan)
		merge.POST("/plan/list", logic.GmListMergePlans)
		merge.POST("/precheck", logic.GmPrecheckMerge)
		merge.POST("/execute", logic.GmExecuteMergePlan)
		merge.POST("/rollback", logic.GmRollbackMergePlan)
		merge.POST("/conflicts", logic.GmListMergeConflicts)
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

	// ========== 上传、服务器时间、配置、编译 ==========
	gm.GET("/server/time", logic.GmGetServerTime)
	gm.POST("/server/time", logic.GmSetServerTime)
	gm.POST("/upload", logic.GmUpload)
	gm.POST("/config/update", logic.GmUpdateConfig)
	gm.POST("/config/game-update", logic.GmGameUpdateConfig)
	gm.POST("/build", logic.GmBuildServer)
	gm.POST("/build/game", logic.GmGameBuildServer)
}
