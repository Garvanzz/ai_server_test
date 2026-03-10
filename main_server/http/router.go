package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (m *HttpModule) register() {
	m.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	// GM 接口：建议内网访问 + 鉴权（见 gmAuth 中间件）
	gm := m.router.Group("/gm")
	gm.Use(m.gmAuth())
	{
		// 邮件
		gm.POST("/mail", m.GMSendMail)

		// 公告
		gm.POST("/notice", m.GMSendNotice)
		gm.POST("/horse", m.GMSendHorse)

		// 玩家管理
		gm.POST("/kick", m.GMKick)

		// 背包
		gm.POST("/item", m.GMGrantItem)
		gm.POST("/bag", m.GMGetBag)
		gm.POST("/item/delete", m.GMDeleteItem)

		// 装备
		gm.POST("/equip", m.GMGetEquip)
		gm.POST("/equip/set", m.GMSetEquip)
		gm.POST("/equip/delete", m.GMDeleteEquip)

		// 关卡
		gm.POST("/stage", m.GMGetStage)
		gm.POST("/stage/set", m.GMSetStage)

		// 英雄
		gm.POST("/hero", m.GMGetHero)
		gm.POST("/hero/set", m.GMSetHero)

		// 玩家游戏信息（Redis Player / 批量）
		gm.POST("/player/game-info", m.GMGetPlayerGameInfo)
		gm.POST("/player/info", m.GMGetPlayerInfo)

		// 时间调试（游戏逻辑时间偏移）
		gm.GET("/time", m.GMTimeGet)
		gm.POST("/time/set_offset", m.GMTimeSetOffset)

		// 活动 GM
		gm.GET("/activity/list", m.GMActivityList)
		gm.POST("/activity/get_by_act_id", m.GMActivityGetByActId)
		gm.POST("/activity/get_by_cfg_id", m.GMActivityGetByCfgId)
		gm.POST("/activity/stop", m.GMActivityStop)
		gm.POST("/activity/recover", m.GMActivityRecover)
		gm.POST("/activity/close", m.GMActivityClose)
		gm.POST("/activity/restart", m.GMActivityRestart)
		gm.POST("/activity/remove", m.GMActivityRemove)
		gm.POST("/activity/close_by_cfg_id", m.GMActivityCloseByCfgId)
		gm.POST("/activity/stop_by_type", m.GMActivityStopByType)
	}

	// TODO:兼容旧路径，后续可删
	api := m.router.Group("api")
	api.Use(m.gmAuth())
	{
		api.POST("/GMSendMail", m.GMSendMail)
		api.POST("/GMSendNotice", m.GMSendNotice)
		api.POST("/GMSendHorse", m.GMSendHorse)
		api.POST("/GMGrantItem", m.GMGrantItem)
	}
}

// gmAuth GM 鉴权中间件，目前仅占位，生产环境应校验 token/白名单等
func (m *HttpModule) gmAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 校验 X-GM-Token 或 IP 白名单，失败则 c.AbortWithStatus(403)
		c.Next()
	}
}
