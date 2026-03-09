// 模拟客户端：先登录 login_server 获取 token，再连 main_server TCP 发 C2SLogin，并随机请求部分游戏接口。
// 用法（示例）：
//   go run . -login http://127.0.0.1:9033 -main 127.0.0.1:8082 -n 1 -register -interval 2s
//   go run . -account myuser -password mypass -main 127.0.0.1:8082 -register=false
package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"xfx/pkg/log"
)

func main() {
	log.DefaultInit()

	cfg := LoadConfig()
	loginClient := NewLoginClient(cfg.LoginServerURL)

	stopCh := make(chan struct{})
	var wg sync.WaitGroup

	for i := 1; i <= cfg.ClientCount; i++ {
		account := cfg.AccountForIndex(i)
		password := cfg.PasswordForIndex(i)

		if cfg.DoRegister {
			if err := loginClient.Register(account, password, 1, cfg.ServerID); err != nil {
				log.Debug("register skip (may exist): %v", err)
			}
		}

		loginResult, err := loginClient.Login(account, password, cfg.ServerID, 1, "0.1")
		if err != nil {
			log.Error("client %d login failed: %v", i, err)
			continue
		}
		log.Debug("client %d login ok, uid=%s token=%s", i, loginResult.UID, loginResult.Token[:8]+"...")

		gameClient, err := NewGameClient(cfg.MainServerAddr, i)
		if err != nil {
			log.Error("client %d connect main_server failed: %v", i, err)
			continue
		}

		wg.Add(1)
		client := gameClient
		result := loginResult
		go func() {
			defer wg.Done()
			client.Run(result, cfg, stopCh, nil)
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Debug("shutting down...")
	close(stopCh)
	wg.Wait()
	log.Debug("exit")
}
