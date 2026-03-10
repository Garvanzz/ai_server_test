package server

import (
	"flag"
	"fmt"
	"xfx/game_server/logic/game"
	"xfx/game_server/pkg/kcp_server"
	"xfx/game_server/pkg/network"
	"xfx/game_server/pkg/packet/pb_packet"
	"xfx/pkg/env"
	"xfx/pkg/log"
)

// LockStepServer 帧同步服务器
type LockStepServer struct {
	udpServer *network.Server
	totalConn int64
	env       *env.Env
	mgr       *game.GameManager
}

// New 构造
func New() (*LockStepServer, error) {
	s := &LockStepServer{
		mgr: game.NewGameManager(),
	}

	//加载env
	e, err := env.LoadEnv()
	if err != nil {
		panic(fmt.Sprintf("load env error %v", err))
	}
	s.env = e

	log.Init(e.Log)
	address := flag.String("udp", string("0.0.0.0:10045"), "udp listen address ismistake")
	networkServer, err := kcp_server.ListenAndServe(*address, s, &pb_packet.MsgProtocol{})
	if err != nil {
		return nil, err
	}

	s.udpServer = networkServer
	return s, nil
}

// Manager 获取管理器
func (r *LockStepServer) GameManager() *game.GameManager {
	return r.mgr
}

// 获取配置
func (r *LockStepServer) GetEnv() *env.Env {
	return r.env
}

// Stop 停止服务
func (r *LockStepServer) Stop() {
	r.mgr.Stop()
	r.udpServer.Stop()
}
