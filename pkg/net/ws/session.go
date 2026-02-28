package ws

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type SessionID = uint64

var gNextSessionID SessionID

func nextSessionID() SessionID {
	return atomic.AddUint64(&gNextSessionID, 1)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type SessionStates = map[string]interface{}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type ISessionHandler interface {
	HandleTextMessage(*Session, []byte)
	HandleBinaryMessage(*Session, []byte)
	HandleAfterSendTextMessage(*Session, []byte)
	HandleAfterSendBinaryMessage(*Session, []byte)
	HandleError(*Session, error)
	HandleClose(*Session, int, string) error
	HandleConnect(*Session) error
	HandleDisconnect(*Session)
	HandlePong(*Session)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type SessionHandler struct {
}

func (SessionHandler) HandleTextMessage(*Session, []byte) {
}

func (SessionHandler) HandleBinaryMessage(*Session, []byte) {
}

func (SessionHandler) HandleAfterSendTextMessage(*Session, []byte) {
}

func (SessionHandler) HandleAfterSendBinaryMessage(*Session, []byte) {
}

func (SessionHandler) HandleError(*Session, error) {
}

func (SessionHandler) HandleClose(*Session, int, string) error {
	return nil
}

func (SessionHandler) HandleConnect(*Session) error {
	return nil
}

func (SessionHandler) HandleDisconnect(*Session) {
}

func (SessionHandler) HandlePong(*Session) {
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type handleMessageFunc func(*Session, []byte)
type handleErrorFunc func(*Session, error)
type handleCloseFunc func(*Session, int, string) error
type handleConnectFunc func(*Session) error
type handleSessionFunc func(*Session)

type SessionHandler2 struct {
	textMessageHandler            handleMessageFunc
	binaryMessageHandler          handleMessageFunc
	afterSendTextMessageHandler   handleMessageFunc
	afterSendBinaryMessageHandler handleMessageFunc
	errorHandler                  handleErrorFunc
	closeHandler                  handleCloseFunc
	connectHandler                handleConnectFunc
	disconnectHandler             handleSessionFunc
	pongHandler                   handleSessionFunc
}

func NewSessionHandler2(
	textMessageHandler handleMessageFunc,
	binaryMessageHandler handleMessageFunc,
	afterSendTextMessageHandler handleMessageFunc,
	afterSendBinaryMessageHandler handleMessageFunc,
	errorHandler handleErrorFunc,
	closeHandler handleCloseFunc,
	connectHandler handleConnectFunc,
	disconnectHandler handleSessionFunc,
	pongHandler handleSessionFunc,
) *SessionHandler2 {
	h := &SessionHandler2{
		textMessageHandler:            textMessageHandler,
		binaryMessageHandler:          binaryMessageHandler,
		afterSendTextMessageHandler:   afterSendTextMessageHandler,
		afterSendBinaryMessageHandler: afterSendBinaryMessageHandler,
		errorHandler:                  errorHandler,
		closeHandler:                  closeHandler,
		connectHandler:                connectHandler,
		disconnectHandler:             disconnectHandler,
		pongHandler:                   pongHandler,
	}

	if h.textMessageHandler == nil {
		h.textMessageHandler = func(*Session, []byte) {}
	}
	if h.binaryMessageHandler == nil {
		h.binaryMessageHandler = func(*Session, []byte) {}
	}
	if h.afterSendBinaryMessageHandler == nil {
		h.afterSendBinaryMessageHandler = func(*Session, []byte) {}
	}
	if h.afterSendTextMessageHandler == nil {
		h.afterSendTextMessageHandler = func(*Session, []byte) {}
	}
	if h.errorHandler == nil {
		h.errorHandler = func(*Session, error) {}
	}
	if h.closeHandler == nil {
		h.closeHandler = func(*Session, int, string) error { return nil }
	}
	if h.connectHandler == nil {
		h.connectHandler = func(*Session) error { return nil }
	}
	if h.disconnectHandler == nil {
		h.disconnectHandler = func(*Session) {}
	}
	if h.pongHandler == nil {
		h.pongHandler = func(*Session) {}
	}

	return h
}

func (h SessionHandler2) HandleTextMessage(session *Session, data []byte) {
	h.textMessageHandler(session, data)
}

func (h SessionHandler2) HandleBinaryMessage(session *Session, data []byte) {
	h.binaryMessageHandler(session, data)
}

func (h SessionHandler2) HandleAfterSendTextMessage(session *Session, data []byte) {
	h.afterSendTextMessageHandler(session, data)
}

func (h SessionHandler2) HandleAfterSendBinaryMessage(session *Session, data []byte) {
	h.afterSendBinaryMessageHandler(session, data)
}

func (h SessionHandler2) HandleError(session *Session, err error) {
	h.errorHandler(session, err)
}

func (h SessionHandler2) HandleClose(session *Session, code int, s string) error {
	return h.closeHandler(session, code, s)
}

func (h SessionHandler2) HandleConnect(session *Session) error {
	return h.connectHandler(session)
}

func (h SessionHandler2) HandleDisconnect(session *Session) {
	h.disconnectHandler(session)
}

func (h SessionHandler2) HandlePong(session *Session) {
	h.pongHandler(session)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type ISessionHandlerProducer interface {
	ProduceSessionHandler(r *http.Request) ISessionHandler
}

type SessionHandlerProduceFunc func(r *http.Request) ISessionHandler

func (fn SessionHandlerProduceFunc) ProduceSessionHandler(r *http.Request) ISessionHandler {
	return fn(r)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Session wrapper around websocket connections.
type Session struct {
	request   *http.Request
	conn      *websocket.Conn
	sendQueue chan *envelope
	melody    *Melody
	opened    bool
	rwmutex   *sync.RWMutex
	handler   ISessionHandler
	id        SessionID
	states    SessionStates
}

func NewSession(
	conn *websocket.Conn,
	r *http.Request,
	m *Melody,
	handler ISessionHandler,
	states SessionStates,
) *Session {
	s := &Session{
		conn:      conn,
		request:   r,
		sendQueue: make(chan *envelope, m.Config.MessageBufferSize),
		melody:    m,
		opened:    true,
		rwmutex:   &sync.RWMutex{},
		handler:   handler,
		id:        nextSessionID(),
		states:    states,
	}
	return s
}

func (s Session) ID() SessionID {
	return s.id
}

func (s Session) Request() *http.Request {
	return s.request
}

func (s Session) URLPath() string {
	return s.request.URL.Path
}

func (s *Session) SetHandler(handler ISessionHandler) {
	if handler != nil {
		s.handler = handler
	}
}

func (s *Session) writeMessage(message *envelope) {
	if s.closed() {
		s.handler.HandleError(s, errors.New("tried to write to closed a session"))
		return
	}

	select {
	case s.sendQueue <- message:
	default:
		s.handler.HandleError(s, errors.New("session message buffer is full"))
	}
}

func (s *Session) writeRaw(message *envelope) error {
	if s.closed() {
		return errors.New("tried to write to a closed session")
	}

	s.conn.SetWriteDeadline(time.Now().Add(s.melody.Config.WriteWait))
	err := s.conn.WriteMessage(message.t, message.msg)
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) closed() bool {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	return !s.opened
}

func (s *Session) close() {
	if !s.closed() {
		s.rwmutex.Lock()
		s.opened = false
		s.conn.Close()
		close(s.sendQueue)
		s.rwmutex.Unlock()
	}
}

func (s *Session) ping() {
	s.writeRaw(&envelope{t: websocket.PingMessage, msg: []byte{}})
}

func (s *Session) writePump() {
	//ticker := time.NewTicker(s.melody.Config.PingPeriod)
	//defer ticker.Stop()

loop:
	for {
		select {
		case msg, ok := <-s.sendQueue:
			if !ok {
				break loop
			}

			err := s.writeRaw(msg)
			if err != nil {
				s.handler.HandleError(s, err)
				break loop
			}

			if msg.t == websocket.CloseMessage {
				break loop
			}

			switch msg.t {
			case websocket.TextMessage:
				s.handler.HandleAfterSendTextMessage(s, msg.msg)

			case websocket.BinaryMessage:
				s.handler.HandleAfterSendBinaryMessage(s, msg.msg)
			}

			//case <-ticker.C:
			//	s.ping()
		}
	}
}

func (s *Session) readPump() {
	//s.conn.SetReadLimit(s.melody.Config.MaxMessageSize)
	//s.conn.SetReadDeadline(time.Now().Add(s.melody.Config.PongWait))
	//
	//s.conn.SetPongHandler(func(string) error {
	//	s.conn.SetReadDeadline(time.Now().Add(s.melody.Config.PongWait))
	//	s.handler.HandlePong(s)
	//	return nil
	//})

	s.conn.SetCloseHandler(func(code int, text string) error {
		return s.handler.HandleClose(s, code, text)
	})

	for {
		t, message, err := s.conn.ReadMessage()
		if err != nil {
			s.handler.HandleError(s, err)
			break
		}

		switch t {
		case websocket.TextMessage:
			s.handler.HandleTextMessage(s, message)

		case websocket.BinaryMessage:
			s.handler.HandleBinaryMessage(s, message)
		}
	}
}

// WriteText writes message to session.
func (s *Session) WriteText(msg []byte) error {
	if s.closed() {
		return errors.New("session is closed")
	}

	s.writeMessage(&envelope{t: websocket.TextMessage, msg: msg})

	return nil
}

// WriteBinary writes a binary message to session.
func (s *Session) WriteBinary(msg []byte) error {
	if s.closed() {
		return errors.New("session is closed")
	}

	s.writeMessage(&envelope{t: websocket.BinaryMessage, msg: msg})

	return nil
}

// Close closes session.
func (s *Session) Close() error {
	if s.closed() {
		return errors.New("session is already closed")
	}

	s.writeMessage(&envelope{t: websocket.CloseMessage, msg: []byte{}})

	return nil
}

// CloseWithMsg closes the session with the provided payload.
// Use the FormatCloseMessage function to format a proper close message payload.
func (s *Session) CloseWithMsg(msg []byte) error {
	if s.closed() {
		return errors.New("session is already closed")
	}

	s.writeMessage(&envelope{t: websocket.CloseMessage, msg: msg})

	return nil
}

// Set is used to store a new key/value pair exclusivelly for this session.
// It also lazy initializes s.Keys if it was not used previously.
func (s *Session) Set(key string, value interface{}) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	if s.states == nil {
		s.states = make(map[string]interface{})
	}

	s.states[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (s *Session) Get(key string) (value interface{}, exists bool) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	if s.states != nil {
		value, exists = s.states[key]
	}

	return
}

func (s *Session) Delete(key string) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	if s.states != nil {
		delete(s.states, key)
	}
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (s *Session) MustGet(key string) interface{} {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()
	if value, exists := s.Get(key); exists {
		return value
	}

	panic("Key \"" + key + "\" does not exist")
}

// IsClosed returns the status of the connection.
func (s *Session) IsClosed() bool {
	return s.closed()
}
