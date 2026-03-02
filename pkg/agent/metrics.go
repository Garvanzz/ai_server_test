package agent

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	msgCount    int64
	msgTotalNs  int64
	tickCount   int64
	tickTotalNs int64
	panicCount  int64
	lastMsgNs   int64
	lastTickNs  int64
}

func (m *Metrics) recordMsg(d time.Duration) {
	atomic.AddInt64(&m.msgCount, 1)
	atomic.AddInt64(&m.msgTotalNs, int64(d))
	atomic.StoreInt64(&m.lastMsgNs, time.Now().UnixNano())
}

func (m *Metrics) recordTick(d time.Duration) {
	atomic.AddInt64(&m.tickCount, 1)
	atomic.AddInt64(&m.tickTotalNs, int64(d))
	atomic.StoreInt64(&m.lastTickNs, time.Now().UnixNano())
}

func (m *Metrics) recordPanic() {
	atomic.AddInt64(&m.panicCount, 1)
}

func (m *Metrics) MessageCount() int64 { return atomic.LoadInt64(&m.msgCount) }
func (m *Metrics) TickCount() int64    { return atomic.LoadInt64(&m.tickCount) }
func (m *Metrics) PanicCount() int64   { return atomic.LoadInt64(&m.panicCount) }

func (m *Metrics) MessageAvgDuration() time.Duration {
	c := atomic.LoadInt64(&m.msgCount)
	if c == 0 {
		return 0
	}
	return time.Duration(atomic.LoadInt64(&m.msgTotalNs) / c)
}

func (m *Metrics) TickAvgDuration() time.Duration {
	c := atomic.LoadInt64(&m.tickCount)
	if c == 0 {
		return 0
	}
	return time.Duration(atomic.LoadInt64(&m.tickTotalNs) / c)
}

func (m *Metrics) LastMessageTime() time.Time {
	ns := atomic.LoadInt64(&m.lastMsgNs)
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

func (m *Metrics) LastTickTime() time.Time {
	ns := atomic.LoadInt64(&m.lastTickNs)
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, ns)
}

type MetricsSnapshot struct {
	MessageCount      int64
	MessageAvgDuration time.Duration
	TickCount         int64
	TickAvgDuration   time.Duration
	PanicCount        int64
	LastMessageTime   time.Time
	LastTickTime      time.Time
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		MessageCount:       m.MessageCount(),
		MessageAvgDuration: m.MessageAvgDuration(),
		TickCount:          m.TickCount(),
		TickAvgDuration:    m.TickAvgDuration(),
		PanicCount:         m.PanicCount(),
		LastMessageTime:    m.LastMessageTime(),
		LastTickTime:       m.LastTickTime(),
	}
}
