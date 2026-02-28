package tcp

import (
	"errors"
	"sync"
	"sync/atomic"
)

var SessionClosedError = errors.New("session Closed")
var SessionBlockedError = errors.New("session Blocked")

var globalSessionId uint64

type Session struct {
	id        uint64
	codec     Codec
	manager   *Manager
	sendChan  chan any
	recvMutex sync.Mutex
	sendMutex sync.RWMutex

	closeFlag          int32
	closeChan          chan int
	closeMutex         sync.Mutex
	firstCloseCallback *closeCallback
	lastCloseCallback  *closeCallback

	state map[string]any
}

func NewSession(codec Codec, sendChanSize int) *Session {
	return newSession(nil, codec, sendChanSize)
}

func newSession(manager *Manager, codec Codec, sendChanSize int) *Session {
	session := &Session{
		codec:     codec,
		manager:   manager,
		closeChan: make(chan int),
		id:        atomic.AddUint64(&globalSessionId, 1),
		state:     make(map[string]any),
	}
	if sendChanSize > 0 {
		session.sendChan = make(chan any, sendChanSize)
		go session.sendLoop()
	}
	return session
}

func (session *Session) ID() uint64 {
	return session.id
}

func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

func (session *Session) Close() error {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		close(session.closeChan)

		if session.sendChan != nil {
			session.sendMutex.Lock()
			close(session.sendChan)
			if clear, ok := session.codec.(ClearSendChan); ok {
				clear.ClearSendChan(session.sendChan)
			}
			session.sendMutex.Unlock()
		}

		err := session.codec.Close()

		go func() {
			session.invokeCloseCallbacks()

			if session.manager != nil {
				session.manager.delSession(session)
			}
		}()
		return err
	}
	return SessionClosedError
}

func (session *Session) Codec() Codec {
	return session.codec
}

func (session *Session) Receive() (any, error) {
	session.recvMutex.Lock()
	defer session.recvMutex.Unlock()

	msg, err := session.codec.Receive()
	if err != nil {
		session.Close()
	}
	return msg, err
}

func (session *Session) sendLoop() {
	defer session.Close()
	for {
		select {
		case msg, ok := <-session.sendChan:
			if !ok || session.codec.Send(msg) != nil {
				return
			}
		case <-session.closeChan:
			return
		}
	}
}

func (session *Session) Send(msg any) error {
	if session.sendChan == nil {
		if session.IsClosed() {
			return SessionClosedError
		}

		session.sendMutex.Lock()
		defer session.sendMutex.Unlock()

		err := session.codec.Send(msg)
		if err != nil {
			session.Close()
		}
		return err
	}

	session.sendMutex.RLock()
	if session.IsClosed() {
		session.sendMutex.RUnlock()
		return SessionClosedError
	}

	select {
	case session.sendChan <- msg:
		session.sendMutex.RUnlock()
		return nil
	default:
		session.sendMutex.RUnlock()
		session.Close()
		return SessionBlockedError
	}
}

type closeCallback struct {
	Handler any
	Key     any
	Func    func()
	Next    *closeCallback
}

func (session *Session) AddCloseCallback(handler, key any, callback func()) {
	if session.IsClosed() {
		return
	}

	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	newItem := &closeCallback{handler, key, callback, nil}

	if session.firstCloseCallback == nil {
		session.firstCloseCallback = newItem
	} else {
		session.lastCloseCallback.Next = newItem
	}
	session.lastCloseCallback = newItem
}

func (session *Session) RemoveCloseCallback(handler, key any) {
	if session.IsClosed() {
		return
	}

	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	var prev *closeCallback
	for callback := session.firstCloseCallback; callback != nil; prev, callback = callback, callback.Next {
		if callback.Handler == handler && callback.Key == key {
			if session.firstCloseCallback == callback {
				session.firstCloseCallback = callback.Next
			} else {
				prev.Next = callback.Next
			}
			if session.lastCloseCallback == callback {
				session.lastCloseCallback = prev
			}
			return
		}
	}
}

func (session *Session) invokeCloseCallbacks() {
	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	for callback := session.firstCloseCallback; callback != nil; callback = callback.Next {
		callback.Func()
	}
}

func (session *Session) Set(key string, value any) {
	session.state[key] = value
}

func (session *Session) Get(key string) (any, bool) {
	value, exist := session.state[key]
	return value, exist
}

func (session *Session) Delete(key string) {
	delete(session.state, key)
}
