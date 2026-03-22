package main

import (
	"net/http"

	"xfx/login_server/conf"
	"xfx/login_server/internal/middleware"
	"xfx/login_server/logic"
	"xfx/login_server/router"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Init(conf.Server.Log)

	logic.AccountEngine = logic.NewMysqlEngine(conf.Server.AccountAddr)
	logic.InitRedis(conf.Server.RedisAddr, conf.Server.RedisPassword, conf.Server.RedisDbNum)

	// 正式服直接改成发布模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	router.Register(r)

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
