package agent

import "sync"

type eventBus struct {
	mu   sync.RWMutex
	subs map[string][]PID
}

func newEventBus() *eventBus {
	return &eventBus{subs: make(map[string][]PID)}
}

func (eb *eventBus) subscribe(topic string, pid PID) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	addr := Address(pid)
	for _, p := range eb.subs[topic] {
		if Address(p) == addr {
			return
		}
	}
	eb.subs[topic] = append(eb.subs[topic], pid)
}

func (eb *eventBus) unsubscribe(topic string, pid PID) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	addr := Address(pid)
	pids := eb.subs[topic]
	n := 0
	for _, p := range pids {
		if Address(p) != addr {
			pids[n] = p
			n++
		}
	}
	eb.subs[topic] = pids[:n]
}

func (eb *eventBus) unsubscribeAll(pid PID) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	addr := Address(pid)
	for topic, pids := range eb.subs {
		n := 0
		for _, p := range pids {
			if Address(p) != addr {
				pids[n] = p
				n++
			}
		}
		eb.subs[topic] = pids[:n]
	}
}

func (eb *eventBus) subscribers(topic string) []PID {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	pids := eb.subs[topic]
	result := make([]PID, len(pids))
	copy(result, pids)
	return result
}

// Subscribe registers a PID to receive messages published on the topic.
// Duplicate subscriptions for the same PID+topic are ignored.
func (s *System) Subscribe(topic string, pid PID) {
	s.eventBus.subscribe(topic, pid)
}

// Unsubscribe removes a PID from a specific topic.
func (s *System) Unsubscribe(topic string, pid PID) {
	s.eventBus.unsubscribe(topic, pid)
}

// UnsubscribeAll removes a PID from all topics. Useful in OnStop handlers.
func (s *System) UnsubscribeAll(pid PID) {
	s.eventBus.unsubscribeAll(pid)
}

// Publish sends a message to all subscribers of the given topic via Cast.
func (s *System) Publish(topic string, msg interface{}) {
	for _, pid := range s.eventBus.subscribers(topic) {
		s.Cast(pid, msg)
	}
}
