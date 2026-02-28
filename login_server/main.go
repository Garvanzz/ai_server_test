package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"xfx/login_server/conf"
	"xfx/login_server/logic"
	"xfx/pkg/log"
)

func main() {
	log.Init(conf.Server.Log)

	logic.AccountEngine = logic.NewMysqlEngine(conf.Server.AccountAddr)
	logic.InitRedis(conf.Server.RedisAddr, conf.Server.RedisPassword, conf.Server.RedisDbNum)

	//正式服直接改成发布模式
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

	//用group做 直接替代中间件 需要解码通讯的 游戏内专用
	encryptGroup := r.Group("")
	encryptGroup.Use(logic.AesDecryptMiddleFuncForGame) //使用解码中间件
	{
	}

	//不需要解码的
	r.POST("/login", logic.Login)                 //登录
	r.POST("/register", logic.Accountregister)    //注册
	r.POST("/forceupdate", logic.Forceupdate)     //判断更新
	r.POST("/getserverlist", logic.GetServerList) //获取服务器列表
	r.POST("/getnotices", logic.GetNotices)       //获取公告

	//官网相关接口
	//用group做 直接替代中间件 需要解码通讯的 这里是官网主页专用
	homeWebGroup := r.Group("")
	homeWebGroup.Use(logic.AesDecryptMiddleFuncForHomeWeb) //使用解码中间件
	{

	}

	log.Debug("http service listen at %v", conf.Server.HttpPort)
	if err := http.ListenAndServe(conf.Server.HttpPort, r); err != nil {
		log.Fatal("ListenAndServe err : ", err)
	}
}
