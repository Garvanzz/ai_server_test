package wsnetwork

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
	"xfx/game_server/pkg/packet/pb_packet"
)

type Config struct {
	PacketSendChanLimit    uint32        // the limit of packet send channel
	PacketReceiveChanLimit uint32        // the limit of packet receive channel
	ConnReadTimeout        time.Duration // read timeout
	ConnWriteTimeout       time.Duration // write timeout
}

type Server struct {
	config    *Config            // server configuration
	callback  ConnCallback       // message callbacks in connection
	protocol  pb_packet.Protocol // customize packet protocol
	exitChan  chan struct{}      // notify all goroutines to shutdown
	waitGroup *sync.WaitGroup    // wait for all goroutines
	closeOnce sync.Once
	conn      *websocket.Conn
}

// NewServer creates a server
func NewServer(config *Config, callback ConnCallback, protocol pb_packet.Protocol) *Server {
	return &Server{
		config:    config,
		callback:  callback,
		protocol:  protocol,
		exitChan:  make(chan struct{}),
		waitGroup: &sync.WaitGroup{},
	}
}

type ConnectionCreator func(*websocket.Conn, *Server) *Conn

// Start starts service
func (s *Server) Start(conn *websocket.Conn, create ConnectionCreator) {
	s.conn = conn
	s.waitGroup.Add(1)
	defer func() {
		s.waitGroup.Done()
	}()

	s.waitGroup.Add(1)
	go func() {
		create(conn, s).Do()
		s.waitGroup.Done()
	}()
}

// Stop stops service
func (s *Server) Stop() {
	s.closeOnce.Do(func() {
		close(s.exitChan)
		s.conn.Close()
	})

	s.waitGroup.Wait()
}
