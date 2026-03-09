package main

import (
	"net/http"

	"xfx/login_server/conf"
	"xfx/login_server/internal/middleware"
	"xfx/login_server/logic"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Init(conf.Server.Log)

	logic.AccountEngine = logic.NewMysqlEngine(conf.Server.AccountAddr)
	logic.InitRedis(conf.Server.RedisAddr, conf.Server.RedisPassword, conf.Server.RedisDbNum)

	// 确保表的存在
	logic.EnsureServerTables()

	// 正式服直接改成发布模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	//无页面处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "PAGE_NOT_FOUND",
			"message": "Page not found",
		})
	})

	r.POST("/login", logic.Login)                 // 登录
	r.POST("/register", logic.Register)           // 注册
	r.POST("/forceupdate", logic.ForceUpdate)     // 判断更新
	r.POST("/getserverlist", logic.GetServerList) // 获取服务器列表
	r.POST("/getnotices", logic.GetNotices)       // 获取公告

	aesKey := []byte(conf.Server.AesKey)
	if len(aesKey) > 0 {
		r.Group("").Use(middleware.AesDecryptGame(aesKey))
		r.Group("").Use(middleware.AesDecryptHomeWeb(aesKey))
	}

	log.Debug("http service listen at %v", conf.Server.HttpPort)
	if err := http.ListenAndServe(conf.Server.HttpPort, r); err != nil {
		log.Fatal("ListenAndServe err: %v", err)
	}
}
