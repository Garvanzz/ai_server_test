package ws

import "sync"

type sessionManager struct {
	sessions map[SessionID]*Session
	open     bool
	rwmutex  *sync.RWMutex
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		sessions: make(map[SessionID]*Session),
		open:     true,
		rwmutex:  &sync.RWMutex{},
	}
}

func (mgr *sessionManager) register(s *Session) {
	mgr.rwmutex.Lock()
	mgr.sessions[s.ID()] = s
	mgr.rwmutex.Unlock()
}

func (mgr *sessionManager) unregister(s *Session) {
	mgr.rwmutex.Lock()
	delete(mgr.sessions, s.ID())
	mgr.rwmutex.Unlock()
}

func (mgr *sessionManager) broadcast(m *envelope) {
	mgr.rwmutex.RLock()
	for _, s := range mgr.sessions {
		if m.filter != nil {
			if m.filter(s) {
				s.writeMessage(m)
			}
		} else {
			s.writeMessage(m)
		}
	}
	mgr.rwmutex.RUnlock()
}

func (mgr *sessionManager) close(m *envelope) {
	mgr.rwmutex.Lock()
	for id, s := range mgr.sessions {
		s.writeMessage(m)
		delete(mgr.sessions, id)
		s.Close()
	}
	mgr.open = false
	mgr.rwmutex.Unlock()
}

func (mgr *sessionManager) closed() bool {
	mgr.rwmutex.RLock()
	defer mgr.rwmutex.RUnlock()
	return !mgr.open
}

func (mgr *sessionManager) len() int {
	mgr.rwmutex.RLock()
	defer mgr.rwmutex.RUnlock()
	return len(mgr.sessions)
}
