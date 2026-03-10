package ws_server

import (
	"time"
	"xfx/game_server/pkg/packet/pb_packet"
	"xfx/game_server/pkg/wsnetwork"
)

func ListenAndWsServe(callback wsnetwork.ConnCallback, protocol pb_packet.Protocol) (*wsnetwork.Server, error) {
	dupConfig := &wsnetwork.Config{
		PacketReceiveChanLimit: 1024,
		PacketSendChanLimit:    1024,
		ConnReadTimeout:        time.Second * 15,
		ConnWriteTimeout:       time.Second * 15,
	}

	server := wsnetwork.NewServer(dupConfig, callback, protocol)
	return server, nil
}
