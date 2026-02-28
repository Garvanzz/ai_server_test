package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (m *HttpModule) register() {
	//无页面处理
	m.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	encryptGroup := m.router.Group("api")
	{
		encryptGroup.POST("/GMSendMail", m.GMSendMail)     //GM邮件
		encryptGroup.POST("/GMSendNotice", m.GMSendNotice) //GM紧急公告
		encryptGroup.POST("/GMSendHorse", m.GMSendHorse)   //GM紧急公告
	}
}
