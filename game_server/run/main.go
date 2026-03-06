package main

import (
	_ "embed"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"
	coreconfig "xfx/core/config"
	"xfx/game_server/server"
	"xfx/pkg/log"
)

func main() {
	flag.Parse()

	s, err := server.New()
	if err != nil {
		panic(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)
	ticker := time.NewTimer(time.Minute)
	defer ticker.Stop()

	//加载配置
	coreconfig.InitConfig(s.GetEnv().ConfPath)

	// 主循环
	log.Debug("main start")
QUIT:
	for {
		select {
		case sig := <-sigs:
			log.Debug("sig==%v", sig)
			break QUIT
		case <-ticker.C:
			//TODO
		}
	}
	s.Stop()
}
