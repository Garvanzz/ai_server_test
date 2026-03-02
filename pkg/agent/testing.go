package agent

import (
	"fmt"
	"sync"
	"time"
)

// CastRecord records a Cast call made through TestContext.
type CastRecord struct {
	PID PID
	Msg interface{}
}

// TestContext is a mock implementation of Context for unit testing agents
// without starting a real protoactor system.
type TestContext struct {
	mu          sync.Mutex
	self        PID
	metrics     *Metrics
	message     interface{}
	sender      PID
	filter      MessageFilter
	states      map[string]interface{}
	stopped     bool
	children    map[string]PID
	stashedMsgs []interface{}

	CastLog     []CastRecord
	CallHandler func(pid PID, msg interface{}) (interface{}, error)
}

func NewTestContext(name string) *TestContext {
	return &TestContext{
		self:     NewPID("test", name),
		metrics:  &Metrics{},
		states:   make(map[string]interface{}),
		children: make(map[string]PID),
	}
}

func (c *TestContext) Self() PID { return c.self }

func (c *TestContext) Cast(pid PID, msg interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CastLog = append(c.CastLog, CastRecord{PID: pid, Msg: msg})
}

func (c *TestContext) Call(pid PID, msg interface{}) (interface{}, error) {
	if c.CallHandler != nil {
		return c.CallHandler(pid, msg)
	}
	return nil, nil
}

func (c *TestContext) CallWithTimeout(pid PID, msg interface{}, _ time.Duration) (interface{}, error) {
	return c.Call(pid, msg)
}

func (c *TestContext) CallNR(pid PID, msg interface{}) error {
	_, err := c.Call(pid, msg)
	return err
}

func (c *TestContext) Watch(_ PID)   {}
func (c *TestContext) Unwatch(_ PID) {}
func (c *TestContext) Stop()         { c.stopped = true }

func (c *TestContext) Create(name string, _ Agent, _ ...Option) (PID, error) {
	pid := NewPID("test", name)
	c.mu.Lock()
	c.children[name] = pid
	c.mu.Unlock()
	return pid, nil
}

func (c *TestContext) SetMessageFilter(filter MessageFilter) { c.filter = filter }
func (c *TestContext) Message() interface{}                  { return c.message }
func (c *TestContext) Sender() PID                           { return c.sender }
func (c *TestContext) Metrics() *Metrics                     { return c.metrics }

func (c *TestContext) Stash() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.message != nil {
		c.stashedMsgs = append(c.stashedMsgs, c.message)
	}
}

func (c *TestContext) UnstashAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stashedMsgs = c.stashedMsgs[:0]
}

// --- Test helper methods ---

func (c *TestContext) SetMessage(msg interface{}) { c.message = msg }
func (c *TestContext) SetSender(pid PID)          { c.sender = pid }
func (c *TestContext) IsStopped() bool             { return c.stopped }

func (c *TestContext) GetCasts() []CastRecord {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]CastRecord, len(c.CastLog))
	copy(result, c.CastLog)
	return result
}

func (c *TestContext) GetStashed() []interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]interface{}, len(c.stashedMsgs))
	copy(result, c.stashedMsgs)
	return result
}

func (c *TestContext) Set(key string, value interface{}) { c.states[key] = value }

func (c *TestContext) Get(key string) (interface{}, bool) {
	v, ok := c.states[key]
	return v, ok
}

func (c *TestContext) Children() map[string]PID {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make(map[string]PID, len(c.children))
	for k, v := range c.children {
		result[k] = v
	}
	return result
}

func (c *TestContext) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CastLog = nil
	c.stopped = false
	c.message = nil
	c.sender = nil
	c.stashedMsgs = nil
}

// TestEnv provides a convenient test environment for an Agent.
type TestEnv struct {
	Ctx   *TestContext
	Agent Agent
}

// NewTestEnv creates a TestEnv and calls agent.OnStart with a TestContext.
func NewTestEnv(name string, agent Agent) *TestEnv {
	ctx := NewTestContext(name)
	agent.OnStart(ctx)
	return &TestEnv{Ctx: ctx, Agent: agent}
}

// Send delivers a message to the agent and returns the result.
func (e *TestEnv) Send(msg interface{}) interface{} {
	e.Ctx.SetMessage(msg)
	defer e.Ctx.SetMessage(nil)
	return e.Agent.OnMessage(msg)
}

// SendFrom delivers a message with a specified sender.
func (e *TestEnv) SendFrom(sender PID, msg interface{}) interface{} {
	e.Ctx.SetSender(sender)
	e.Ctx.SetMessage(msg)
	defer func() {
		e.Ctx.SetSender(nil)
		e.Ctx.SetMessage(nil)
	}()
	return e.Agent.OnMessage(msg)
}

// Tick triggers an OnTick call with the given delta.
func (e *TestEnv) Tick(delta time.Duration) {
	e.Agent.OnTick(delta)
}

// Terminate simulates a child termination event.
func (e *TestEnv) Terminate(pid PID, reason int) {
	e.Agent.OnTerminated(pid, reason)
}

// Stop calls the agent's OnStop.
func (e *TestEnv) Stop() {
	e.Agent.OnStop()
}

// AssertCastCount checks that exactly n Cast calls were made.
func (e *TestEnv) AssertCastCount(n int) error {
	casts := e.Ctx.GetCasts()
	if len(casts) != n {
		return fmt.Errorf("expected %d casts, got %d", n, len(casts))
	}
	return nil
}

// LastCast returns the most recent Cast record, or nil if none.
func (e *TestEnv) LastCast() *CastRecord {
	casts := e.Ctx.GetCasts()
	if len(casts) == 0 {
		return nil
	}
	r := casts[len(casts)-1]
	return &r
}
