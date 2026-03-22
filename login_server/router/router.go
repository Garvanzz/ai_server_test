package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xfx/login_server/logic"
)

// Register 注册 login_server 路由。
func Register(r *gin.Engine) {
	r.Use(gin.Recovery())
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/login", logic.Login)
		auth.POST("/register", logic.Register)
	}

	updates := r.Group("/updates")
	{
		updates.POST("/force", logic.ForceUpdate)
	}

	servers := r.Group("/servers")
	{
		servers.POST("/list", logic.GetServerList)
	}

	notices := r.Group("/notices")
	{
		notices.POST("/list", logic.GetNotices)
	}
}
