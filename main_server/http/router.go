package http

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *HttpModule) register() {
	m.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	// 健康检查（无需鉴权），gm_server 通过 TCP dial 此端口判定大厅服是否在线
	m.router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
	})

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
		gm.POST("/activity/adjust_time", m.GMActivityAdjustTime)
		gm.POST("/activity/force_start", m.GMActivityForceStart)
		gm.POST("/activity/player_count", m.GMActivityPlayerCount)
	}
}

// gmAuth GM 接口鉴权中间件。
// 优先使用 Token 鉴权：env.toml 中设置 GmToken，请求须携带 "X-GM-Token: <token>" 请求头。
// 若未配置 Token，则回落到 IP 白名单（GmAllowIPs）；两者均未配置则拦截全部请求。
func (m *HttpModule) gmAuth() gin.HandlerFunc {
	env := m.GetApp().GetEnv()
	token := strings.TrimSpace(env.GmToken)
	allowIPs := env.GmAllowIPs

	return func(c *gin.Context) {
		if token != "" {
			if c.GetHeader("X-GM-Token") != token {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "forbidden"})
				return
			}
			c.Next()
			return
		}

		// 无 Token 配置时走 IP 白名单
		if len(allowIPs) > 0 {
			clientIP := c.ClientIP()
			allowed := false
			for _, cidr := range allowIPs {
				cidr = strings.TrimSpace(cidr)
				if strings.Contains(cidr, "/") {
					_, network, err := net.ParseCIDR(cidr)
					if err == nil && network.Contains(net.ParseIP(clientIP)) {
						allowed = true
						break
					}
				} else if cidr == clientIP {
					allowed = true
					break
				}
			}
			if !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "forbidden"})
				return
			}
			c.Next()
			return
		}

		// Token 和白名单均未配置，拒绝所有请求（安全默认值）
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "gm auth not configured"})
	}
}
