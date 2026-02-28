package ws

import (
	"net"
	"net/http"
)

var S = NewServer(DefaultConfig)

type Server struct {
	*Melody
	listener   net.Listener
	httpServer *http.Server
}

func NewServer(cfg Config) *Server {
	svr := &Server{
		Melody: NewMelody(cfg),
	}
	svr.httpServer = &http.Server{Handler: svr}
	return svr
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.HandleRequest(w, r)
}

func (s *Server) Serve(addr string) (err error) {
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}
	go s.httpServer.Serve(s.listener)
	return
}

func (s *Server) ServeTLS(addr string, certFile string, keyFile string) (err error) {
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}
	go s.httpServer.ServeTLS(s.listener, certFile, keyFile)
	return
}

func (s Server) ListenAddr() string {
	return s.listener.Addr().String()
}

func (s Server) SessionNum() int {
	return s.sessionMgr.len()
}
