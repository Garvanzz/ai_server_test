package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xfx/core/config"
	"xfx/gm_server/conf"
	"xfx/gm_server/logic"
	"xfx/pkg/log"
)

func main() {
	log.Init(conf.Server.Log)
	config.InitConfig("./json")
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	// ========== 鉴权（仅管理系统） ==========
	auth := r.Group("/gm/auth")
	{
		auth.POST("/login", logic.GmLogin)             // GM 登录，校验账号密码并写 session/token
		auth.POST("/logout", logic.GmLoginout)         // GM 登出，清除登录态
		auth.POST("/user-info", logic.GmAdminUserInfo) // 获取当前登录 GM 用户信息与权限
	}

	// ========== 区服与进程（仅管理系统） ==========
	servers := r.Group("/gm/servers")
	{
		servers.POST("/list", logic.GmGetServerList)                  // 获取大厅服（main_server）列表及状态
		servers.POST("/start", logic.GmStartServer)                   // 启动指定大厅服进程
		servers.POST("/stop", logic.GmStopServer)                     // 停止指定大厅服进程
		servers.POST("/restart", logic.GmReStartServer)               // 重启指定大厅服进程
		servers.POST("/game-list", logic.GmGetGameServerList)         // 获取游戏服列表（含运行状态、开服时间等）
		servers.POST("/game-start", logic.GmStartGameServer)          // 启动指定游戏服进程
		servers.POST("/game-stop", logic.GmStopGameServer)            // 停止指定游戏服进程
		servers.POST("/game-restart", logic.GmReStartGameServer)      // 重启指定游戏服进程
		servers.POST("/game-list-all", logic.GmGetGameGameServerList) // 获取所有游戏服简要列表
	}

	// ========== 玩家数据查询（读 Redis/DB，仅管理系统） ==========
	players := r.Group("/gm/players")
	{
		players.POST("/info", logic.GmGetPlayerInfo)           // 按区服/uid 查询玩家基础信息（可批量）
		players.POST("/game-info", logic.GmGetPlayerGameInfo)  // 按区服/uid 查询玩家游戏内数据（等级、英雄等）
		players.POST("/bag", logic.GmItem)                     // 查询指定玩家背包道具列表
		players.POST("/equip", logic.GmEquip)                  // 查询指定玩家装备列表
		players.POST("/equip-delete", logic.GmDeleteEquip)     // 删除指定玩家装备（直写 Redis）
		players.POST("/hero", logic.GmHero)                    // 查询指定玩家英雄/阵容数据
		players.POST("/hero-edit", logic.GmEditHero)           // 编辑指定玩家英雄数据（直写 Redis）
		players.POST("/stage-info", logic.GmGetStageInfo)      // 查询指定玩家关卡进度
		players.POST("/stage-set", logic.GmSetStageInfo)       // 设置指定玩家关卡信息（直写 Redis）
		players.POST("/stage-add", logic.GmAddStageInfo)       // 为指定玩家添加关卡进度
		players.POST("/stage-delete", logic.GmDeleteStageInfo) // 删除指定玩家关卡数据
	}

	// ========== 需转发 main_server 的接口 ==========
	r.POST("/gm/mail/send", logic.GmCreateAdminMail)            // 创建并发送邮件（转发 main_server /gm/mail，支持延时与附件）
	r.POST("/gm/notice/horse", logic.GmSendHorse)               // 发送跑马灯（转发 main_server /gm/horse，全服推送）
	r.POST("/gm/notice/send", logic.GmSendNotice)               // 发送公告（当前仅入库；即时下发需再调 main_server /gm/notice）
	r.POST("/gm/players/grant-item", logic.GmAddItem)           // 给指定玩家发放道具（转发 main_server /gm/item，玩家进程内执行）
	r.POST("/gm/players/grant-item-all", logic.GmOneKeyAddItem) // 一键给指定玩家发放全部配置道具（转发 main_server /gm/item）
	r.POST("/gm/players/item-delete", logic.GmDeleteItem)       // 删除指定玩家背包道具（当前直写 Redis，未走 main_server）

	// ========== 订单、帮会==========
	r.POST("/gm/orders/list", logic.GmGetOrderList)            // 查询充值订单列表（支持按订单号、uid 筛选）
	r.POST("/gm/orders/cache-list", logic.GmGetCacheOrderList) // 查询缓存充值订单列表
	r.POST("/gm/guild/list", logic.GmGetGuidList)              // 查询帮会列表（名称、人数、会长等）

	// ========== 热更==========
	hotfix := r.Group("/gm/hotfix")
	{
		hotfix.POST("/list", logic.GmGetHotUpdate)                     // 获取热更版本列表
		hotfix.POST("/version-edit", logic.GmEditHotUpdateVersion)     // 编辑指定热更版本
		hotfix.POST("/version-create", logic.GmCreateHotUpdateVersion) // 创建热更版本
		hotfix.POST("/version-delete", logic.GmDeleteHotUpdateVersion) // 删除热更版本
		hotfix.POST("/path-create", logic.GmCreateHotUpdatePath)       // 创建热更路径/渠道
	}

	// ========== 上传、服务器时间、配置、编译 ==========
	r.POST("/gm/upload", logic.Gmupload)                       // 上传文件（如热更包、配置等）
	r.POST("/gm/server/time", logic.GmSetServerTime)           // 设置 GM 所在机器系统时间（执行 date 命令，慎用）
	r.POST("/gm/config/update", logic.GmUpdateConfig)          // 拉取配置仓库并更新大厅服配置
	r.POST("/gm/config/game-update", logic.GmGameUpdateConfig) // 拉取配置仓库并更新游戏服配置
	r.POST("/gm/build", logic.GmBuildServer)                   // 拉取代码并编译大厅服（main_server）
	r.POST("/gm/build/game", logic.GmGameBuildServer)          // 拉取代码并编译游戏服

	log.Debug("http service listen at %v", conf.Server.HttpPort)
	if err := http.ListenAndServe(conf.Server.HttpPort, r); err != nil {
		log.Fatal("ListenAndServe err : ", err)
	}
}
