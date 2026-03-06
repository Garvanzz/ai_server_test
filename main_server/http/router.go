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

	// GM 接口：建议内网访问 + 鉴权（见 gmAuth 中间件）。按需扩展踢人、封禁、重载配置等。
	gm := m.router.Group("/gm")
	gm.Use(m.gmAuth())
	{
		gm.POST("/mail", m.GMSendMail)
		gm.POST("/notice", m.GMSendNotice)
		gm.POST("/horse", m.GMSendHorse)
		gm.POST("/kick", m.GMKick)
		gm.POST("/item", m.GMGrantItem)
	}

	// 兼容旧路径，后续可删
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
