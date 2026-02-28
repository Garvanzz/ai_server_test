package wsgate

import (
	"github.com/gogo/protobuf/proto"
	"strconv"
	"strings"
	"xfx/pkg/gate"
	"xfx/pkg/gate/wsgate/codec"
	"xfx/pkg/log"
	"xfx/pkg/net/ws"
	"xfx/pkg/serialize"
	"xfx/pkg/serialize/pb"
)

type Session struct {
	gate.Session
	s          *ws.Session
	ip         string
	id         string
	serializer serialize.Serializer
	decoder    *codec.Parser
	agent      gate.Agent
}

func NewSession(s *ws.Session) *Session {
	ip := ""
	parts := strings.Split(s.Request().RemoteAddr, ":")
	if len(parts) > 0 {
		ip = parts[0]
	}

	session := &Session{
		s:          s,
		id:         strconv.FormatUint(s.ID(), 10),
		ip:         ip,
		serializer: pb.NewSerializer(),
		decoder:    codec.NewParser(),
	}
	return session
}

func (s *Session) GetIP() string { return s.ip }
func (s *Session) ID() uint64 {
	value, _ := strconv.ParseUint(s.id, 10, 64)
	return value
}
func (s *Session) IsClosed() bool { return s.s.IsClosed() }

func (s *Session) Send(msg interface{}) error {
	data, err := s.decoder.EncodeMsg(msg)
	if err != nil {
		return err
	}
	s.s.WriteBinary(data)
	return nil
}

func (s *Session) doRecv(data []byte) {
	msg, err := s.decoder.DecodeMsg(data)
	if err != nil {
		// error
		log.Error("Session: recv %v", err)
		return
	}

	if s.agent == nil {
		value, ok := s.Get("#agent")
		if !ok {
			log.Error("Session: no agent")
			return
		}
		s.agent, _ = value.(gate.Agent)
	}

	pbs, ok := msg.([]proto.Message)
	if !ok {
		log.Error("doRecv pbs error")
		return
	}

	for _, pb := range pbs {
		s.agent.OnRecv(pb)
	}
}

func (s *Session) Close() error {
	return s.s.Close()
}

func (s *Session) Set(key string, value interface{}) {
	s.s.Set(key, value)
}
func (s *Session) Get(key string) (value interface{}, ok bool) {
	return s.s.Get(key)
}
