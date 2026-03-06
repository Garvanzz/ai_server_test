package kcp_server

import (
	"github.com/xtaci/kcp-go"
	"net"
	"time"
	"xfx/game_server/pkg/network"
	"xfx/game_server/pkg/packet/pb_packet"
)

func ListenAndServe(addr string, callback network.ConnCallback, protocol pb_packet.Protocol) (*network.Server, error) {
	dupConfig := &network.Config{
		PacketReceiveChanLimit: 1024,
		PacketSendChanLimit:    1024,
		ConnReadTimeout:        time.Second * 15,
		ConnWriteTimeout:       time.Second * 15,
	}

	l, err := kcp.Listen(addr)
	if nil != err {
		return nil, err
	}

	server := network.NewServer(dupConfig, callback, protocol)
	go server.Start(l, func(conn net.Conn, i *network.Server) *network.Conn {

		// 普通模式
		// setKCPConfig(32, 32, 0, 40, 0, 0, 100, 1400)

		// 极速模式
		// setKCPConfig(32, 32, 1, 10, 2, 1, 30, 1400)

		// 普通模式：ikcp_nodelay(kcp, 0, 40, 0, 0); 极速模式： ikcp_nodelay(kcp, 1, 10, 2, 1);

		kcpConn := conn.(*kcp.UDPSession)
		kcpConn.SetNoDelay(1, 10, 2, 1)
		kcpConn.SetStreamMode(true)
		kcpConn.SetWindowSize(4096, 4096)
		kcpConn.SetReadBuffer(4 * 1024 * 1024)
		kcpConn.SetWriteBuffer(4 * 1024 * 1024)
		kcpConn.SetACKNoDelay(true)

		return network.NewConn(conn, server)
	})

	return server, nil
}
