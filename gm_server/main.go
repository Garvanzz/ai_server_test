package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xfx/gm_server/conf"
	"xfx/gm_server/router"
	"xfx/pkg/log"
)

func main() {
	log.Init(conf.Server.Log)
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	router.Register(r)

	log.Debug("http service listen at %v", conf.Server.HttpPort)
	if err := http.ListenAndServe(conf.Server.HttpPort, r); err != nil {
		log.Fatal("ListenAndServe err : ", err)
	}
}
