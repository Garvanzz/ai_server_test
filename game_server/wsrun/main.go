package main

//
//import (
//	_ "embed"
//	"os"
//	"os/signal"
//	"syscall"
//	"time"
//	coreconfig "xfx/core/config"
//	"xfx/game_server/wsserver"
//	"xfx/pkg/log"
//)
//
////go:embed wsenv.toml
//var env []byte
//
//func main() {
//	s, err := wsserver.New(env)
//	if err != nil {
//		panic(err)
//	}
//
//	sigs := make(chan os.Signal, 1)
//	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)
//	ticker := time.NewTimer(time.Minute)
//	defer ticker.Stop()
//	//加载配置
//	coreconfig.InitConfig(s.GetEnv().ConfPath) // todo:json path
//
//	// 主循环
//	log.Debug("main start")
//QUIT:
//	for {
//		select {
//		case sig := <-sigs:
//			log.Debug("sig==%v", sig)
//			break QUIT
//		case <-ticker.C:
//			//TODO
//		}
//	}
//	s.Stop()
//}
