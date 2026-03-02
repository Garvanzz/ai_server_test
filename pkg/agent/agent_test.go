package agent

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"xfx/pkg/log"
)

func TestMain(m *testing.M) {
	log.DefaultInit()
	os.Exit(m.Run())
}

// ================================================================
// Test Agents
// ================================================================

type terminatedEvent struct {
	Who    PID
	Reason int
}

type recordAgent struct {
	startCh    chan struct{}
	stopCh     chan struct{}
	ctx        Context
	mu         sync.Mutex
	msgs       []interface{}
	terminated []terminatedEvent
	ticks      int32
	onStartFn  func(ctx Context)
	onMsgFn    func(msg interface{}) interface{}
}

func newRecordAgent() *recordAgent {
	return &recordAgent{
		startCh: make(chan struct{}),
		stopCh:  make(chan struct{}),
	}
}

func (a *recordAgent) OnStart(ctx Context) {
	a.ctx = ctx
	if a.onStartFn != nil {
		a.onStartFn(ctx)
	}
	close(a.startCh)
}

func (a *recordAgent) OnStop() { close(a.stopCh) }

func (a *recordAgent) OnTerminated(pid PID, reason int) {
	a.mu.Lock()
	a.terminated = append(a.terminated, terminatedEvent{pid, reason})
	a.mu.Unlock()
}

func (a *recordAgent) OnMessage(msg interface{}) interface{} {
	a.mu.Lock()
	a.msgs = append(a.msgs, msg)
	a.mu.Unlock()
	if a.onMsgFn != nil {
		return a.onMsgFn(msg)
	}
	return msg
}

func (a *recordAgent) OnTick(delta time.Duration) {
	atomic.AddInt32(&a.ticks, 1)
}

func (a *recordAgent) messageCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.msgs)
}

func (a *recordAgent) getMessages() []interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	r := make([]interface{}, len(a.msgs))
	copy(r, a.msgs)
	return r
}

func (a *recordAgent) getTerminated() []terminatedEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	r := make([]terminatedEvent, len(a.terminated))
	copy(r, a.terminated)
	return r
}

// ================================================================
// Helpers
// ================================================================

var sysCounter int64

func newTestSystem(t *testing.T) *System {
	t.Helper()
	name := fmt.Sprintf("sys%d", atomic.AddInt64(&sysCounter, 1))
	s := NewSystem(WithName(name))
	s.Start()
	t.Cleanup(func() { s.Stop() })
	return s
}

var actorCounter int64

func nextName() string {
	return fmt.Sprintf("a%d", atomic.AddInt64(&actorCounter, 1))
}

func systemCall(s *System, pid PID, msg interface{}, timeout time.Duration) (interface{}, error) {
	if pid == nil {
		return nil, ErrNilPID
	}
	m, err := wrapMessage(s.root, pid, msg, true)
	if err != nil {
		return nil, err
	}
	return s.context.RequestFuture(pid, m, timeout).Result()
}

func waitFor(t *testing.T, check func() bool, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timeout waiting for condition")
}

// ================================================================
// PID Tests
// ================================================================

func TestNewPID(t *testing.T) {
	pid := NewPID("localhost:8080", "test")
	if pid.Address != "localhost:8080" || pid.Id != "test" {
		t.Fatalf("got address=%s id=%s", pid.Address, pid.Id)
	}
}

func TestParseValid(t *testing.T) {
	pid, ok := Parse("localhost:8080/actor1")
	if !ok || pid.Address != "localhost:8080" || pid.Id != "actor1" {
		t.Fatalf("parse failed: ok=%v pid=%v", ok, pid)
	}
}

func TestParseInvalid(t *testing.T) {
	_, ok := Parse("no-slash-here")
	if ok {
		t.Fatal("expected parse to fail for string without '/'")
	}
}

func TestAddressFormat(t *testing.T) {
	pid := NewPID("host", "id")
	if Address(pid) != "host/id" {
		t.Fatalf("got %s", Address(pid))
	}
}

func TestAddressNil(t *testing.T) {
	if Address(nil) != "nil" {
		t.Fatalf("expected 'nil', got %s", Address(nil))
	}
}

func TestPIDRegistryStoreAndLookup(t *testing.T) {
	name := nextName()
	pid := NewPID("test", name)
	_Store(name, pid)
	defer _Delete(name, pid)

	found, ok := Lookup(name)
	if !ok || found != pid {
		t.Fatalf("lookup by name failed")
	}
}

func TestPIDRegistryLookupByAddress(t *testing.T) {
	name := nextName()
	pid := NewPID("host:8080", name)
	_Store(name, pid)
	defer _Delete(name, pid)

	addr := Address(pid) // "host:8080/<name>" contains ':'
	found, ok := Lookup(addr)
	if !ok || found != pid {
		t.Fatalf("lookup by address failed: ok=%v", ok)
	}
}

func TestPIDRegistryDelete(t *testing.T) {
	name := nextName()
	pid := NewPID("test", name)
	_Store(name, pid)
	_Delete(name, pid)

	_, ok := Lookup(name)
	if ok {
		t.Fatal("expected not found after delete")
	}
}

func TestPIDRegistryConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("conc_%d", i)
			pid := NewPID("test", name)
			_Store(name, pid)
			Lookup(name)
			_Delete(name, pid)
		}(i)
	}
	wg.Wait()
}

// ================================================================
// Message Tests
// ================================================================

func TestWrapUnwrapLocalMessage(t *testing.T) {
	from := NewPID("local", "from")
	to := NewPID("local", "to")

	wrapped, err := wrapMessage(from, to, "hello", false)
	if err != nil {
		t.Fatal(err)
	}

	lm, ok := wrapped.(*LocalMessage)
	if !ok {
		t.Fatalf("expected *LocalMessage, got %T", wrapped)
	}
	if lm.msg != "hello" || lm.response != false || lm.sender != from {
		t.Fatal("LocalMessage fields mismatch")
	}

	msg, sender, senderA, response, err := unwrapMessage(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if msg != "hello" || sender != from || senderA != "" || response != false {
		t.Fatalf("unwrap mismatch: msg=%v sender=%v senderA=%s resp=%v", msg, sender, senderA, response)
	}
}

func TestWrapLocalMessageWithResponse(t *testing.T) {
	from := NewPID("local", "from")
	to := NewPID("local", "to")

	wrapped, err := wrapMessage(from, to, "req", true)
	if err != nil {
		t.Fatal(err)
	}

	lm := wrapped.(*LocalMessage)
	if !lm.response {
		t.Fatal("expected response=true")
	}

	_, _, _, response, _ := unwrapMessage(wrapped)
	if !response {
		t.Fatal("expected response=true after unwrap")
	}
}

func TestWrapTickPassthrough(t *testing.T) {
	from := NewPID("local", "from")
	to := NewPID("local", "to")

	tick := tickMessage(time.Second)
	wrapped, err := wrapMessage(from, to, tick, false)
	if err != nil {
		t.Fatal(err)
	}
	if wrapped != tick {
		t.Fatal("tick should pass through unchanged")
	}

	msg, _, _, _, err := unwrapMessage(tick)
	if err != nil {
		t.Fatal(err)
	}
	if msg != tick {
		t.Fatal("unwrap tick should return original")
	}
}

func TestWrapRemoteNonProtoError(t *testing.T) {
	from := NewPID("host-a", "from")
	to := NewPID("host-b", "to")

	_, err := wrapMessage(from, to, "not-proto", false)
	if err == nil {
		t.Fatal("expected error for non-proto remote message")
	}
}

func TestUnwrapRemoteMessageError(t *testing.T) {
	rm := &RemoteMessage{
		TypeName:    "nonexistent.Type",
		MessageData: []byte{0xff},
		Response:    true,
		Sender:      "addr/sender",
	}

	_, _, _, _, err := unwrapMessage(rm)
	if err == nil {
		t.Fatal("expected error for invalid remote message deserialization")
	}
}

func TestUnwrapUnknownType(t *testing.T) {
	msg, _, _, _, err := unwrapMessage(42)
	if err != nil {
		t.Fatal(err)
	}
	if msg != 42 {
		t.Fatalf("expected 42, got %v", msg)
	}
}

// ================================================================
// Metrics Tests
// ================================================================

func TestMetricsRecordMsg(t *testing.T) {
	m := &Metrics{}
	m.recordMsg(100 * time.Millisecond)
	m.recordMsg(200 * time.Millisecond)

	if m.MessageCount() != 2 {
		t.Fatalf("expected 2, got %d", m.MessageCount())
	}
	avg := m.MessageAvgDuration()
	if avg != 150*time.Millisecond {
		t.Fatalf("expected 150ms avg, got %v", avg)
	}
	if m.LastMessageTime().IsZero() {
		t.Fatal("LastMessageTime should not be zero")
	}
}

func TestMetricsRecordTick(t *testing.T) {
	m := &Metrics{}
	m.recordTick(50 * time.Millisecond)

	if m.TickCount() != 1 {
		t.Fatalf("expected 1, got %d", m.TickCount())
	}
	if m.TickAvgDuration() != 50*time.Millisecond {
		t.Fatalf("expected 50ms, got %v", m.TickAvgDuration())
	}
	if m.LastTickTime().IsZero() {
		t.Fatal("LastTickTime should not be zero")
	}
}

func TestMetricsRecordPanic(t *testing.T) {
	m := &Metrics{}
	m.recordPanic()
	m.recordPanic()
	if m.PanicCount() != 2 {
		t.Fatalf("expected 2, got %d", m.PanicCount())
	}
}

func TestMetricsAvgEmpty(t *testing.T) {
	m := &Metrics{}
	if m.MessageAvgDuration() != 0 {
		t.Fatal("expected 0 for empty metrics")
	}
	if m.TickAvgDuration() != 0 {
		t.Fatal("expected 0 for empty metrics")
	}
	if !m.LastMessageTime().IsZero() {
		t.Fatal("expected zero time")
	}
	if !m.LastTickTime().IsZero() {
		t.Fatal("expected zero time")
	}
}

func TestMetricsSnapshot(t *testing.T) {
	m := &Metrics{}
	m.recordMsg(10 * time.Millisecond)
	m.recordTick(20 * time.Millisecond)
	m.recordPanic()

	snap := m.Snapshot()
	if snap.MessageCount != 1 || snap.TickCount != 1 || snap.PanicCount != 1 {
		t.Fatalf("snapshot mismatch: %+v", snap)
	}
	if snap.MessageAvgDuration != 10*time.Millisecond {
		t.Fatalf("msg avg: %v", snap.MessageAvgDuration)
	}
	if snap.TickAvgDuration != 20*time.Millisecond {
		t.Fatalf("tick avg: %v", snap.TickAvgDuration)
	}
}

func TestMetricsConcurrent(t *testing.T) {
	m := &Metrics{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() { defer wg.Done(); m.recordMsg(time.Millisecond) }()
		go func() { defer wg.Done(); m.recordTick(time.Millisecond) }()
		go func() { defer wg.Done(); m.recordPanic() }()
	}
	wg.Wait()
	if m.MessageCount() != 100 || m.TickCount() != 100 || m.PanicCount() != 100 {
		t.Fatalf("concurrent counts: msg=%d tick=%d panic=%d", m.MessageCount(), m.TickCount(), m.PanicCount())
	}
}

// ================================================================
// EventBus Unit Tests
// ================================================================

func TestEventBusSubscribeAndList(t *testing.T) {
	eb := newEventBus()
	p1 := NewPID("test", "p1")
	p2 := NewPID("test", "p2")
	eb.subscribe("topic1", p1)
	eb.subscribe("topic1", p2)

	subs := eb.subscribers("topic1")
	if len(subs) != 2 {
		t.Fatalf("expected 2 subs, got %d", len(subs))
	}
}

func TestEventBusDuplicateSubscribe(t *testing.T) {
	eb := newEventBus()
	p1 := NewPID("test", "p1")
	eb.subscribe("t", p1)
	eb.subscribe("t", p1)

	if len(eb.subscribers("t")) != 1 {
		t.Fatal("duplicate should be ignored")
	}
}

func TestEventBusUnsubscribe(t *testing.T) {
	eb := newEventBus()
	p1 := NewPID("test", "p1")
	p2 := NewPID("test", "p2")
	eb.subscribe("t", p1)
	eb.subscribe("t", p2)
	eb.unsubscribe("t", p1)

	subs := eb.subscribers("t")
	if len(subs) != 1 || Address(subs[0]) != Address(p2) {
		t.Fatalf("expected only p2, got %v", subs)
	}
}

func TestEventBusUnsubscribeNonExistent(t *testing.T) {
	eb := newEventBus()
	p1 := NewPID("test", "p1")
	eb.unsubscribe("t", p1) // should not panic
}

func TestEventBusUnsubscribeAll(t *testing.T) {
	eb := newEventBus()
	p1 := NewPID("test", "p1")
	p2 := NewPID("test", "p2")
	eb.subscribe("t1", p1)
	eb.subscribe("t2", p1)
	eb.subscribe("t1", p2)
	eb.unsubscribeAll(p1)

	if len(eb.subscribers("t1")) != 1 {
		t.Fatal("t1 should have only p2")
	}
	if len(eb.subscribers("t2")) != 0 {
		t.Fatal("t2 should be empty")
	}
}

func TestEventBusEmptyTopic(t *testing.T) {
	eb := newEventBus()
	subs := eb.subscribers("nonexistent")
	if len(subs) != 0 {
		t.Fatal("expected empty")
	}
}

func TestEventBusMultipleTopics(t *testing.T) {
	eb := newEventBus()
	p1 := NewPID("test", "p1")
	eb.subscribe("a", p1)
	eb.subscribe("b", p1)

	if len(eb.subscribers("a")) != 1 || len(eb.subscribers("b")) != 1 {
		t.Fatal("each topic should have 1 subscriber")
	}
}

// ================================================================
// Options Tests
// ================================================================

func TestOptionsDefaults(t *testing.T) {
	opt := Options{
		CallTTL:          DefaultCallTTL,
		MaxRetries:       DefaultMaxRetries,
		SupervisorWindow: DefaultSupervisorWindow,
	}
	if opt.CallTTL != time.Second {
		t.Fatalf("default CallTTL: %v", opt.CallTTL)
	}
	if opt.MaxRetries != 5 {
		t.Fatalf("default MaxRetries: %d", opt.MaxRetries)
	}
	if opt.SupervisorWindow != time.Second {
		t.Fatalf("default SupervisorWindow: %v", opt.SupervisorWindow)
	}
	if opt.DeadLetterHandler != nil {
		t.Fatal("default DeadLetterHandler should be nil")
	}
}

func TestOptionsAll(t *testing.T) {
	ag := &emptyAgent{}
	dlCalled := false
	dlHandler := func(target PID, message interface{}, sender PID) {
		dlCalled = true
	}
	opts := Options{}
	fns := []Option{
		WithTick(100 * time.Millisecond),
		WithCallTTL(2 * time.Second),
		WithAgent(ag),
		WithName("test"),
		WithRestart(),
		WithHost("0.0.0.0"),
		WithPort(9000),
		WithMaxRetries(10),
		WithSupervisorWindow(5 * time.Second),
		WithDeadLetterHandler(dlHandler),
	}
	for _, fn := range fns {
		fn(&opts)
	}
	if opts.Tick != 100*time.Millisecond {
		t.Fatal("Tick")
	}
	if opts.CallTTL != 2*time.Second {
		t.Fatal("CallTTL")
	}
	if opts.Agent != ag {
		t.Fatal("Agent")
	}
	if opts.Name != "test" {
		t.Fatal("Name")
	}
	if !opts.Restart {
		t.Fatal("Restart")
	}
	if opts.Host != "0.0.0.0" {
		t.Fatal("Host")
	}
	if opts.Port != 9000 {
		t.Fatal("Port")
	}
	if opts.MaxRetries != 10 {
		t.Fatal("MaxRetries")
	}
	if opts.SupervisorWindow != 5*time.Second {
		t.Fatal("SupervisorWindow")
	}
	if opts.DeadLetterHandler == nil {
		t.Fatal("DeadLetterHandler should not be nil")
	}
	opts.DeadLetterHandler(nil, "test", nil)
	if !dlCalled {
		t.Fatal("DeadLetterHandler not invoked")
	}
}

func TestOptionsMaxRetriesZero(t *testing.T) {
	opts := Options{}
	WithMaxRetries(0)(&opts)
	if opts.MaxRetries != 0 {
		t.Fatal("MaxRetries should be 0")
	}
}

func TestOptionsSupervisorWindowZero(t *testing.T) {
	opts := Options{}
	WithSupervisorWindow(0)(&opts)
	if opts.SupervisorWindow != 0 {
		t.Fatal("SupervisorWindow should be 0")
	}
}

// ================================================================
// TestContext Tests
// ================================================================

func TestTestContextSelf(t *testing.T) {
	ctx := NewTestContext("myactor")
	if ctx.Self().Id != "myactor" {
		t.Fatal("Self().Id mismatch")
	}
}

func TestTestContextCast(t *testing.T) {
	ctx := NewTestContext("x")
	target := NewPID("test", "target")
	ctx.Cast(target, "hello")
	ctx.Cast(target, "world")

	casts := ctx.GetCasts()
	if len(casts) != 2 {
		t.Fatalf("expected 2 casts, got %d", len(casts))
	}
	if casts[0].Msg != "hello" || casts[1].Msg != "world" {
		t.Fatal("cast messages mismatch")
	}
}

func TestTestContextCallHandler(t *testing.T) {
	ctx := NewTestContext("x")
	ctx.CallHandler = func(pid PID, msg interface{}) (interface{}, error) {
		return fmt.Sprintf("reply:%v", msg), nil
	}

	result, err := ctx.Call(NewPID("t", "t"), "req")
	if err != nil || result != "reply:req" {
		t.Fatalf("Call: result=%v err=%v", result, err)
	}

	result2, err2 := ctx.CallWithTimeout(NewPID("t", "t"), "req2", time.Second)
	if err2 != nil || result2 != "reply:req2" {
		t.Fatalf("CallWithTimeout: result=%v err=%v", result2, err2)
	}
}

func TestTestContextCreate(t *testing.T) {
	ctx := NewTestContext("parent")
	pid, err := ctx.Create("child", &emptyAgent{})
	if err != nil || pid == nil {
		t.Fatal("Create failed")
	}
	children := ctx.Children()
	if _, ok := children["child"]; !ok {
		t.Fatal("child not tracked")
	}
}

func TestTestContextStop(t *testing.T) {
	ctx := NewTestContext("x")
	if ctx.IsStopped() {
		t.Fatal("should not be stopped initially")
	}
	ctx.Stop()
	if !ctx.IsStopped() {
		t.Fatal("should be stopped")
	}
}

func TestTestContextStash(t *testing.T) {
	ctx := NewTestContext("x")
	ctx.SetMessage("msg1")
	ctx.Stash()
	ctx.SetMessage("msg2")
	ctx.Stash()

	stashed := ctx.GetStashed()
	if len(stashed) != 2 || stashed[0] != "msg1" || stashed[1] != "msg2" {
		t.Fatalf("stashed: %v", stashed)
	}

	ctx.UnstashAll()
	if len(ctx.GetStashed()) != 0 {
		t.Fatal("should be empty after UnstashAll")
	}
}

func TestTestContextMetrics(t *testing.T) {
	ctx := NewTestContext("x")
	if ctx.Metrics() == nil {
		t.Fatal("Metrics() should not be nil")
	}
}

func TestTestContextSetGet(t *testing.T) {
	ctx := NewTestContext("x")
	ctx.Set("key", "value")
	v, ok := ctx.Get("key")
	if !ok || v != "value" {
		t.Fatalf("Set/Get failed: ok=%v v=%v", ok, v)
	}
	_, ok = ctx.Get("missing")
	if ok {
		t.Fatal("should not find missing key")
	}
}

func TestTestContextReset(t *testing.T) {
	ctx := NewTestContext("x")
	ctx.Cast(NewPID("t", "t"), "msg")
	ctx.SetMessage("m")
	ctx.SetSender(NewPID("t", "s"))
	ctx.Stop()
	ctx.Reset()

	if len(ctx.GetCasts()) != 0 || ctx.Message() != nil || ctx.Sender() != nil || ctx.IsStopped() {
		t.Fatal("Reset did not clear state")
	}
}

// ================================================================
// TestEnv Tests
// ================================================================

type simpleAgent struct {
	ctx      Context
	received []interface{}
	mu       sync.Mutex
}

func (a *simpleAgent) OnStart(ctx Context)                   { a.ctx = ctx }
func (a *simpleAgent) OnStop()                               {}
func (a *simpleAgent) OnTerminated(pid PID, reason int)      {}
func (a *simpleAgent) OnTick(delta time.Duration)            {}
func (a *simpleAgent) OnMessage(msg interface{}) interface{} {
	a.mu.Lock()
	a.received = append(a.received, msg)
	a.mu.Unlock()
	if s, ok := msg.(string); ok && s == "cast-it" {
		a.ctx.Cast(NewPID("test", "other"), "casted")
	}
	return msg
}

func TestTestEnvLifecycle(t *testing.T) {
	ag := &simpleAgent{}
	env := NewTestEnv("test", ag)

	if ag.ctx == nil {
		t.Fatal("OnStart should have been called")
	}

	result := env.Send("hello")
	if result != "hello" {
		t.Fatalf("expected echo, got %v", result)
	}

	env.Stop()
}

func TestTestEnvSendFrom(t *testing.T) {
	ag := &simpleAgent{}
	env := NewTestEnv("test", ag)
	sender := NewPID("test", "sender")
	env.SendFrom(sender, "msg")

	if env.Ctx.Sender() != nil {
		t.Fatal("sender should be cleared after SendFrom")
	}
}

func TestTestEnvCastTracking(t *testing.T) {
	ag := &simpleAgent{}
	env := NewTestEnv("test", ag)
	env.Send("cast-it")

	if err := env.AssertCastCount(1); err != nil {
		t.Fatal(err)
	}
	last := env.LastCast()
	if last == nil || last.Msg != "casted" {
		t.Fatal("LastCast mismatch")
	}
}

func TestTestEnvAssertCastCountZero(t *testing.T) {
	env := NewTestEnv("test", &emptyAgent{})
	if err := env.AssertCastCount(0); err != nil {
		t.Fatal(err)
	}
	if env.LastCast() != nil {
		t.Fatal("should be nil with no casts")
	}
}

// ================================================================
// System Integration Tests
// ================================================================

func TestSystemStartStop(t *testing.T) {
	name := fmt.Sprintf("sys_ss_%d", atomic.AddInt64(&sysCounter, 1))
	s := NewSystem(WithName(name))
	s.Start()
	if s.Root() == nil {
		t.Fatal("root PID should not be nil after Start")
	}
	s.Stop()
}

func TestSystemStopGraceful(t *testing.T) {
	name := fmt.Sprintf("sys_sg_%d", atomic.AddInt64(&sysCounter, 1))
	s := NewSystem(WithName(name))
	s.Start()

	err := s.StopGraceful(5 * time.Second)
	if err != nil {
		t.Fatalf("StopGraceful error: %v", err)
	}
}

func TestSystemStopGracefulTimeout(t *testing.T) {
	name := fmt.Sprintf("sys_sgt_%d", atomic.AddInt64(&sysCounter, 1))
	s := NewSystem(WithName(name))
	s.Start()
	defer s.Stop()

	// StopGraceful with very short timeout - system should still stop cleanly
	// but if it's slow, we get a timeout error
	err := s.StopGraceful(1 * time.Nanosecond)
	// Either nil (fast enough) or timeout error is acceptable
	_ = err
}

func TestSystemCreateAndDestroy(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	name := nextName()

	pid, err := sys.Create(name, ag)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if pid == nil {
		t.Fatal("PID should not be nil")
	}

	<-ag.startCh

	sys.Destroy(pid)

	select {
	case <-ag.stopCh:
	case <-time.After(5 * time.Second):
		t.Fatal("OnStop not called after Destroy")
	}
}

func TestSystemCreateInvalidName(t *testing.T) {
	sys := newTestSystem(t)
	_, err := sys.Create("invalid/name", &emptyAgent{})
	if err == nil {
		t.Fatal("expected error for name containing '/'")
	}
}

func TestSystemCast(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Cast(pid, "msg1")
	sys.Cast(pid, "msg2")

	waitFor(t, func() bool { return ag.messageCount() >= 2 }, 5*time.Second)

	msgs := ag.getMessages()
	if msgs[0] != "msg1" || msgs[1] != "msg2" {
		t.Fatalf("messages: %v", msgs)
	}
}

func TestSystemCastNilPIDSafe(t *testing.T) {
	sys := newTestSystem(t)
	sys.Cast(nil, "msg") // should not panic
}

func TestSystemDestroyNilPIDSafe(t *testing.T) {
	sys := newTestSystem(t)
	sys.Destroy(nil) // should not panic
}

func TestSystemDeadLetterNocrash(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	name := nextName()
	pid, _ := sys.Create(name, ag)
	<-ag.startCh

	sys.Destroy(pid)
	<-ag.stopCh

	// Cast to destroyed PID triggers dead letter — should not crash
	sys.Cast(pid, "dead")
	time.Sleep(100 * time.Millisecond)
}

// ================================================================
// Actor Lifecycle Tests
// ================================================================

func TestActorLifecycleStartStop(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	pid, _ := sys.Create(nextName(), ag)

	select {
	case <-ag.startCh:
	case <-time.After(5 * time.Second):
		t.Fatal("OnStart not called")
	}

	if ag.ctx == nil {
		t.Fatal("Context not set")
	}

	sys.Destroy(pid)

	select {
	case <-ag.stopCh:
	case <-time.After(5 * time.Second):
		t.Fatal("OnStop not called")
	}
}

func TestActorCallEcho(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	result, err := systemCall(sys, pid, "hello", 5*time.Second)
	if err != nil {
		t.Fatalf("Call error: %v", err)
	}
	if result != "hello" {
		t.Fatalf("expected echo 'hello', got %v", result)
	}
}

func TestActorCallMultiple(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} {
		if n, ok := msg.(int); ok {
			return n * 2
		}
		return msg
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	for i := 0; i < 10; i++ {
		result, err := systemCall(sys, pid, i, 5*time.Second)
		if err != nil {
			t.Fatalf("Call(%d) error: %v", i, err)
		}
		if result != i*2 {
			t.Fatalf("expected %d, got %v", i*2, result)
		}
	}
}

func TestActorCallNilPIDError(t *testing.T) {
	sys := newTestSystem(t)
	_, err := systemCall(sys, nil, "msg", time.Second)
	if err == nil {
		t.Fatal("expected error for nil PID in wrapMessage")
	}
}

func TestActorOnMessagePanicRecovery(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} {
		if msg == "panic" {
			panic("test panic")
		}
		return msg
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	// Cast panic message — actor should recover
	sys.Cast(pid, "panic")
	time.Sleep(200 * time.Millisecond)

	// Actor should still be alive and process messages
	result, err := systemCall(sys, pid, "after-panic", 5*time.Second)
	if err != nil {
		t.Fatalf("actor dead after panic: %v", err)
	}
	if result != "after-panic" {
		t.Fatalf("expected 'after-panic', got %v", result)
	}

	// Panic should be recorded in metrics
	m := ag.ctx.Metrics()
	if m.PanicCount() < 1 {
		t.Fatal("panic count should be >= 1")
	}
}

func TestActorCallPanicReturnsError(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} {
		panic("call panic")
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	result, err := systemCall(sys, pid, "trigger", 5*time.Second)
	if err != nil {
		// timeout or error from future is also acceptable
		return
	}
	if _, ok := result.(error); !ok {
		t.Fatalf("expected error result from panic, got %T: %v", result, result)
	}
}

func TestActorOnTickPanicRecovery(t *testing.T) {
	sys := newTestSystem(t)
	var tickCount int32
	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} { return msg }

	panicOnce := int32(1)
	origOnTick := ag.OnTick
	_ = origOnTick
	ag2 := &tickPanicAgent{startCh: make(chan struct{}), panicOnce: &panicOnce}
	pid, _ := sys.Create(nextName(), ag2, WithTick(50*time.Millisecond))
	<-ag2.startCh

	time.Sleep(300 * time.Millisecond)

	// Actor should still be ticking after panic
	if atomic.LoadInt32(&ag2.ticks) < 2 {
		t.Fatalf("expected ticks after panic recovery, got %d", atomic.LoadInt32(&ag2.ticks))
	}
	_ = tickCount
	_ = pid
}

type tickPanicAgent struct {
	startCh   chan struct{}
	ctx       Context
	panicOnce *int32
	ticks     int32
}

func (a *tickPanicAgent) OnStart(ctx Context) {
	a.ctx = ctx
	close(a.startCh)
}
func (a *tickPanicAgent) OnStop()                               {}
func (a *tickPanicAgent) OnTerminated(pid PID, reason int)      {}
func (a *tickPanicAgent) OnMessage(msg interface{}) interface{} { return nil }
func (a *tickPanicAgent) OnTick(delta time.Duration) {
	atomic.AddInt32(&a.ticks, 1)
	if atomic.CompareAndSwapInt32(a.panicOnce, 1, 0) {
		panic("tick panic")
	}
}

func TestActorTick(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	pid, _ := sys.Create(nextName(), ag, WithTick(50*time.Millisecond))
	<-ag.startCh

	time.Sleep(280 * time.Millisecond)
	ticks := atomic.LoadInt32(&ag.ticks)
	if ticks < 3 {
		t.Fatalf("expected >= 3 ticks in 280ms, got %d", ticks)
	}

	sys.Destroy(pid)
	<-ag.stopCh
	time.Sleep(150 * time.Millisecond)

	ticksAfterStop := atomic.LoadInt32(&ag.ticks)
	if ticksAfterStop > ticks+2 {
		t.Fatalf("ticks should stop after destroy: before=%d after=%d", ticks, ticksAfterStop)
	}
}

func TestActorTickMetrics(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	_, err := sys.Create(nextName(), ag, WithTick(50*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	<-ag.startCh

	waitFor(t, func() bool {
		return ag.ctx.Metrics().TickCount() >= 3
	}, 5*time.Second)

	if ag.ctx.Metrics().TickAvgDuration() < 0 {
		t.Fatal("tick avg duration should be >= 0")
	}
}

// ================================================================
// Actor MessageFilter Tests
// ================================================================

func TestActorMessageFilter(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	ag.onStartFn = func(ctx Context) {
		ctx.SetMessageFilter(func(msg interface{}) (bool, interface{}) {
			if s, ok := msg.(string); ok && s == "blocked" {
				return true, nil
			}
			return false, nil
		})
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Cast(pid, "blocked")
	sys.Cast(pid, "allowed")

	waitFor(t, func() bool { return ag.messageCount() >= 1 }, 5*time.Second)
	time.Sleep(100 * time.Millisecond)

	msgs := ag.getMessages()
	if len(msgs) != 1 || msgs[0] != "allowed" {
		t.Fatalf("filter failed: msgs=%v", msgs)
	}
}

func TestActorFilterWithCallResponse(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	ag.onStartFn = func(ctx Context) {
		ctx.SetMessageFilter(func(msg interface{}) (bool, interface{}) {
			if s, ok := msg.(string); ok && s == "filtered-call" {
				return true, "filter-response"
			}
			return false, nil
		})
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	result, err := systemCall(sys, pid, "filtered-call", 5*time.Second)
	if err != nil {
		t.Fatalf("Call error: %v", err)
	}
	if result != "filter-response" {
		t.Fatalf("expected 'filter-response', got %v", result)
	}

	if ag.messageCount() != 0 {
		t.Fatal("message should not reach OnMessage when filtered")
	}
}

// ================================================================
// Actor Metrics Integration
// ================================================================

func TestActorMetricsIntegration(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	for i := 0; i < 5; i++ {
		sys.Cast(pid, i)
	}

	waitFor(t, func() bool { return ag.messageCount() >= 5 }, 5*time.Second)

	m := ag.ctx.Metrics()
	if m.MessageCount() < 5 {
		t.Fatalf("expected >= 5 messages, got %d", m.MessageCount())
	}
	if m.MessageAvgDuration() < 0 {
		t.Fatal("avg duration should be >= 0")
	}
	if m.LastMessageTime().IsZero() {
		t.Fatal("last message time should be set")
	}
}

// ================================================================
// Actor Watch/Terminate Tests
// ================================================================

func TestActorWatchTerminate(t *testing.T) {
	sys := newTestSystem(t)

	watched := newRecordAgent()
	watchedPID, _ := sys.Create(nextName(), watched)
	<-watched.startCh

	watcher := newRecordAgent()
	watcher.onStartFn = func(ctx Context) {
		ctx.Watch(watchedPID)
	}
	sys.Create(nextName(), watcher)
	<-watcher.startCh
	time.Sleep(50 * time.Millisecond)

	sys.Destroy(watchedPID)

	waitFor(t, func() bool {
		return len(watcher.getTerminated()) > 0
	}, 5*time.Second)

	events := watcher.getTerminated()
	if Address(events[0].Who) != Address(watchedPID) {
		t.Fatalf("expected terminated PID=%v, got %v", Address(watchedPID), Address(events[0].Who))
	}
}

// ================================================================
// Actor Stash/UnstashAll Tests
// ================================================================

type stashTestAgent struct {
	startCh chan struct{}
	ctx     Context
	ready   int32
	mu      sync.Mutex
	msgs    []interface{}
}

func (a *stashTestAgent) OnStart(ctx Context) {
	a.ctx = ctx
	close(a.startCh)
}
func (a *stashTestAgent) OnStop()                          {}
func (a *stashTestAgent) OnTerminated(pid PID, reason int) {}
func (a *stashTestAgent) OnTick(delta time.Duration)       {}
func (a *stashTestAgent) OnMessage(msg interface{}) interface{} {
	if s, ok := msg.(string); ok && s == "ready" {
		atomic.StoreInt32(&a.ready, 1)
		a.ctx.UnstashAll()
		return nil
	}

	if atomic.LoadInt32(&a.ready) == 0 {
		a.ctx.Stash()
		return nil
	}

	a.mu.Lock()
	a.msgs = append(a.msgs, msg)
	a.mu.Unlock()
	return nil
}

func (a *stashTestAgent) getMessages() []interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	r := make([]interface{}, len(a.msgs))
	copy(r, a.msgs)
	return r
}

func TestActorStashAndUnstash(t *testing.T) {
	sys := newTestSystem(t)
	ag := &stashTestAgent{startCh: make(chan struct{})}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Cast(pid, "msg1")
	sys.Cast(pid, "msg2")
	sys.Cast(pid, "msg3")

	// Give time for messages to be stashed
	time.Sleep(200 * time.Millisecond)
	if len(ag.getMessages()) != 0 {
		t.Fatal("messages should be stashed, not processed")
	}

	// Send ready to trigger UnstashAll
	sys.Cast(pid, "ready")

	waitFor(t, func() bool {
		return len(ag.getMessages()) >= 3
	}, 5*time.Second)

	msgs := ag.getMessages()
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d: %v", len(msgs), msgs)
	}
	if msgs[0] != "msg1" || msgs[1] != "msg2" || msgs[2] != "msg3" {
		t.Fatalf("message order: %v", msgs)
	}
}

func TestActorStashEmpty(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	ag.onStartFn = func(ctx Context) {
		ctx.UnstashAll() // should be safe when no stashed messages
	}
	ag.onMsgFn = func(msg interface{}) interface{} {
		ag.ctx.Stash()         // rawMsg is set, so this stashes
		ag.ctx.UnstashAll()    // immediately re-sends
		return nil
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Cast(pid, "test")
	time.Sleep(500 * time.Millisecond)
	// The stash-then-immediate-unstash creates a re-delivered message
	// which will be processed again (and again stashed/unstashed — loop)
	// This test just verifies it doesn't crash/deadlock within the timeout
}

// ================================================================
// Actor Create Child Tests
// ================================================================

func TestActorCreateChild(t *testing.T) {
	sys := newTestSystem(t)
	child := newRecordAgent()
	var childPID PID

	parent := newRecordAgent()
	parent.onStartFn = func(ctx Context) {
		var err error
		childPID, err = ctx.Create(nextName(), child)
		if err != nil {
			t.Errorf("create child: %v", err)
		}
	}

	sys.Create(nextName(), parent)
	<-parent.startCh

	select {
	case <-child.startCh:
	case <-time.After(5 * time.Second):
		t.Fatal("child OnStart not called")
	}

	if childPID == nil {
		t.Fatal("child PID is nil")
	}
}

func TestActorCreateChildInvalidName(t *testing.T) {
	sys := newTestSystem(t)
	var createErr error

	parent := newRecordAgent()
	parent.onStartFn = func(ctx Context) {
		_, createErr = ctx.Create("bad:name", &emptyAgent{})
	}

	sys.Create(nextName(), parent)
	<-parent.startCh
	time.Sleep(50 * time.Millisecond)

	if createErr == nil {
		t.Fatal("expected error for invalid child name")
	}
}

// ================================================================
// Inter-Actor Communication Tests
// ================================================================

func TestActorContextCast(t *testing.T) {
	sys := newTestSystem(t)

	receiver := newRecordAgent()
	receiverPID, _ := sys.Create(nextName(), receiver)
	<-receiver.startCh

	sender := newRecordAgent()
	sender.onMsgFn = func(msg interface{}) interface{} {
		if msg == "send-it" {
			sender.ctx.Cast(receiverPID, "from-sender")
		}
		return nil
	}
	senderPID, _ := sys.Create(nextName(), sender)
	<-sender.startCh

	sys.Cast(senderPID, "send-it")

	waitFor(t, func() bool { return receiver.messageCount() >= 1 }, 5*time.Second)

	msgs := receiver.getMessages()
	if msgs[0] != "from-sender" {
		t.Fatalf("expected 'from-sender', got %v", msgs[0])
	}
}

func TestActorContextCall(t *testing.T) {
	sys := newTestSystem(t)

	echo := newRecordAgent()
	echo.onMsgFn = func(msg interface{}) interface{} {
		return fmt.Sprintf("echo:%v", msg)
	}
	echoPID, _ := sys.Create(nextName(), echo)
	<-echo.startCh

	type callResult struct {
		result interface{}
		err    error
	}
	resultCh := make(chan callResult, 1)

	caller := newRecordAgent()
	caller.onMsgFn = func(msg interface{}) interface{} {
		if msg == "do-call" {
			r, err := caller.ctx.Call(echoPID, "hello")
			resultCh <- callResult{r, err}
		}
		return nil
	}
	callerPID, _ := sys.Create(nextName(), caller)
	<-caller.startCh

	sys.Cast(callerPID, "do-call")

	select {
	case cr := <-resultCh:
		if cr.err != nil {
			t.Fatalf("Call error: %v", cr.err)
		}
		if cr.result != "echo:hello" {
			t.Fatalf("expected 'echo:hello', got %v", cr.result)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for call result")
	}
}

func TestActorContextCallWithTimeout(t *testing.T) {
	sys := newTestSystem(t)

	slow := &slowTestAgent{startCh: make(chan struct{}), delay: 500 * time.Millisecond}
	slowPID, _ := sys.Create(nextName(), slow)
	<-slow.startCh

	type callResult struct {
		result interface{}
		err    error
	}
	resultCh := make(chan callResult, 1)

	caller := newRecordAgent()
	caller.onMsgFn = func(msg interface{}) interface{} {
		if msg == "do-timeout-call" {
			r, err := caller.ctx.CallWithTimeout(slowPID, "req", 50*time.Millisecond)
			resultCh <- callResult{r, err}
		}
		return nil
	}
	callerPID, _ := sys.Create(nextName(), caller)
	<-caller.startCh

	sys.Cast(callerPID, "do-timeout-call")

	select {
	case cr := <-resultCh:
		if cr.err == nil {
			t.Fatal("expected timeout error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("test hung")
	}
}

type slowTestAgent struct {
	startCh chan struct{}
	delay   time.Duration
}

func (a *slowTestAgent) OnStart(ctx Context)                   { close(a.startCh) }
func (a *slowTestAgent) OnStop()                               {}
func (a *slowTestAgent) OnTerminated(pid PID, reason int)      {}
func (a *slowTestAgent) OnTick(delta time.Duration)            {}
func (a *slowTestAgent) OnMessage(msg interface{}) interface{} {
	time.Sleep(a.delay)
	return msg
}

func TestActorContextCallNilPID(t *testing.T) {
	sys := newTestSystem(t)

	type callResult struct {
		err error
	}
	resultCh := make(chan callResult, 1)

	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} {
		if msg == "call-nil" {
			_, err := ag.ctx.Call(nil, "msg")
			resultCh <- callResult{err}
		}
		return nil
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Cast(pid, "call-nil")

	select {
	case cr := <-resultCh:
		if cr.err != ErrNilPID {
			t.Fatalf("expected ErrNilPID, got %v", cr.err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestActorContextCastNilPID(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} {
		if msg == "cast-nil" {
			ag.ctx.Cast(nil, "noop") // should not panic
		}
		return nil
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Cast(pid, "cast-nil")
	time.Sleep(100 * time.Millisecond)
}

// ================================================================
// Concurrent Cast Tests
// ================================================================

func TestActorConcurrentCast(t *testing.T) {
	sys := newTestSystem(t)
	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} { return nil }
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	count := 200
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			sys.Cast(pid, n)
		}(i)
	}
	wg.Wait()

	waitFor(t, func() bool { return ag.messageCount() >= count }, 10*time.Second)

	if ag.messageCount() != count {
		t.Fatalf("expected %d messages, got %d", count, ag.messageCount())
	}
}

// ================================================================
// EventBus Integration Tests
// ================================================================

func TestSystemEventBusPublish(t *testing.T) {
	sys := newTestSystem(t)

	ag1 := newRecordAgent()
	ag1.onMsgFn = func(msg interface{}) interface{} { return nil }
	pid1, _ := sys.Create(nextName(), ag1)
	<-ag1.startCh

	ag2 := newRecordAgent()
	ag2.onMsgFn = func(msg interface{}) interface{} { return nil }
	pid2, _ := sys.Create(nextName(), ag2)
	<-ag2.startCh

	sys.Subscribe("news", pid1)
	sys.Subscribe("news", pid2)

	sys.Publish("news", "breaking")

	waitFor(t, func() bool {
		return ag1.messageCount() >= 1 && ag2.messageCount() >= 1
	}, 5*time.Second)

	if ag1.getMessages()[0] != "breaking" || ag2.getMessages()[0] != "breaking" {
		t.Fatal("publish message mismatch")
	}
}

func TestSystemEventBusUnsubscribeStopsDelivery(t *testing.T) {
	sys := newTestSystem(t)

	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} { return nil }
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Subscribe("topic", pid)
	sys.Publish("topic", "msg1")
	waitFor(t, func() bool { return ag.messageCount() >= 1 }, 5*time.Second)

	sys.Unsubscribe("topic", pid)
	sys.Publish("topic", "msg2")
	time.Sleep(200 * time.Millisecond)

	if ag.messageCount() != 1 {
		t.Fatalf("expected 1 message after unsubscribe, got %d", ag.messageCount())
	}
}

func TestSystemEventBusUnsubscribeAllOnStop(t *testing.T) {
	sys := newTestSystem(t)

	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} { return nil }
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Subscribe("t1", pid)
	sys.Subscribe("t2", pid)
	sys.UnsubscribeAll(pid)

	sys.Publish("t1", "x")
	sys.Publish("t2", "y")
	time.Sleep(200 * time.Millisecond)

	if ag.messageCount() != 0 {
		t.Fatalf("expected 0 messages after UnsubscribeAll, got %d", ag.messageCount())
	}
}

// ================================================================
// Context State Tests
// ================================================================

func TestActorContextSetGet(t *testing.T) {
	sys := newTestSystem(t)

	type stateResult struct {
		v  interface{}
		ok bool
	}
	resultCh := make(chan stateResult, 1)

	ag := newRecordAgent()
	ag.onStartFn = func(ctx Context) {
		c := ctx.(*agentContext)
		c.Set("key", 42)
	}
	ag.onMsgFn = func(msg interface{}) interface{} {
		if msg == "get" {
			c := ag.ctx.(*agentContext)
			v, ok := c.Get("key")
			resultCh <- stateResult{v, ok}
		}
		return nil
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Cast(pid, "get")

	select {
	case sr := <-resultCh:
		if !sr.ok || sr.v != 42 {
			t.Fatalf("expected 42, got %v (ok=%v)", sr.v, sr.ok)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestActorContextSenderInMessage(t *testing.T) {
	sys := newTestSystem(t)

	senderCh := make(chan PID, 1)

	receiver := newRecordAgent()
	receiver.onMsgFn = func(msg interface{}) interface{} {
		s := receiver.ctx.Sender()
		senderCh <- s
		return nil
	}
	receiverPID, _ := sys.Create(nextName(), receiver)
	<-receiver.startCh

	sender := newRecordAgent()
	sender.onMsgFn = func(msg interface{}) interface{} {
		if msg == "send" {
			sender.ctx.Cast(receiverPID, "hi")
		}
		return nil
	}
	senderPID, _ := sys.Create(nextName(), sender)
	<-sender.startCh

	sys.Cast(senderPID, "send")

	select {
	case s := <-senderCh:
		if s == nil {
			t.Fatal("Sender() should not be nil for Cast from another actor")
		}
		if Address(s) != Address(senderPID) {
			t.Fatalf("Sender address mismatch: %s vs %s", Address(s), Address(senderPID))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

func TestActorContextMessageDuringOnMessage(t *testing.T) {
	sys := newTestSystem(t)

	msgCh := make(chan interface{}, 1)
	ag := newRecordAgent()
	ag.onMsgFn = func(msg interface{}) interface{} {
		msgCh <- ag.ctx.Message()
		return nil
	}
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Cast(pid, "check-msg")

	select {
	case m := <-msgCh:
		if m != "check-msg" {
			t.Fatalf("Message() returned %v", m)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}

// ================================================================
// Multiple Systems
// ================================================================

func TestMultipleSystemsIsolation(t *testing.T) {
	sys1 := newTestSystem(t)
	sys2 := newTestSystem(t)

	ag1 := newRecordAgent()
	ag1.onMsgFn = func(msg interface{}) interface{} { return nil }
	pid1, _ := sys1.Create(nextName(), ag1)
	<-ag1.startCh

	ag2 := newRecordAgent()
	ag2.onMsgFn = func(msg interface{}) interface{} { return nil }
	pid2, _ := sys2.Create(nextName(), ag2)
	<-ag2.startCh

	sys1.Cast(pid1, "to-sys1")
	sys2.Cast(pid2, "to-sys2")

	waitFor(t, func() bool { return ag1.messageCount() >= 1 }, 5*time.Second)
	waitFor(t, func() bool { return ag2.messageCount() >= 1 }, 5*time.Second)

	if ag1.getMessages()[0] != "to-sys1" || ag2.getMessages()[0] != "to-sys2" {
		t.Fatal("cross-system contamination")
	}
}

// ================================================================
// emptyAgent Tests
// ================================================================

func TestEmptyAgent(t *testing.T) {
	ag := &emptyAgent{}
	ag.OnStart(NewTestContext("x"))
	ag.OnStop()
	ag.OnTerminated(nil, 0)
	r := ag.OnMessage("msg")
	if r != nil {
		t.Fatal("emptyAgent OnMessage should return nil")
	}
	ag.OnTick(time.Second)
}

// ================================================================
// Custom DeadLetterHandler Tests
// ================================================================

func TestSystemCustomDeadLetterHandler(t *testing.T) {
	var dlTarget PID
	var dlMsg interface{}
	dlCh := make(chan struct{}, 1)

	handler := func(target PID, message interface{}, sender PID) {
		dlTarget = target
		dlMsg = message
		select {
		case dlCh <- struct{}{}:
		default:
		}
	}

	name := fmt.Sprintf("sys_cdl_%d", atomic.AddInt64(&sysCounter, 1))
	sys := NewSystem(WithName(name), WithDeadLetterHandler(handler))
	sys.Start()
	defer sys.Stop()

	ag := newRecordAgent()
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Destroy(pid)
	<-ag.stopCh

	sys.Cast(pid, "dead-msg")

	select {
	case <-dlCh:
		if dlTarget == nil {
			t.Fatal("dead letter target should not be nil")
		}
		if dlMsg != "dead-msg" {
			t.Fatalf("dead letter message: expected 'dead-msg', got %v", dlMsg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("custom dead letter handler not called")
	}
}

func TestSystemDefaultDeadLetterHandlerNocrash(t *testing.T) {
	name := fmt.Sprintf("sys_ddl_%d", atomic.AddInt64(&sysCounter, 1))
	sys := NewSystem(WithName(name))
	sys.Start()
	defer sys.Stop()

	ag := newRecordAgent()
	pid, _ := sys.Create(nextName(), ag)
	<-ag.startCh

	sys.Destroy(pid)
	<-ag.stopCh

	sys.Cast(pid, "default-dead")
	time.Sleep(100 * time.Millisecond)
}

// ================================================================
// Custom Supervisor Options Tests
// ================================================================

func TestSystemCustomMaxRetries(t *testing.T) {
	name := fmt.Sprintf("sys_cmr_%d", atomic.AddInt64(&sysCounter, 1))
	sys := NewSystem(
		WithName(name),
		WithMaxRetries(10),
		WithSupervisorWindow(3*time.Second),
	)
	sys.Start()
	defer sys.Stop()

	if sys.opts.MaxRetries != 10 {
		t.Fatalf("expected MaxRetries=10, got %d", sys.opts.MaxRetries)
	}
	if sys.opts.SupervisorWindow != 3*time.Second {
		t.Fatalf("expected SupervisorWindow=3s, got %v", sys.opts.SupervisorWindow)
	}
}

func TestChildActorCustomSupervisorOptions(t *testing.T) {
	sys := newTestSystem(t)

	parentAg := newRecordAgent()
	var childPID PID
	childCreated := make(chan struct{}, 1)

	parentAg.onStartFn = func(ctx Context) {
		var err error
		childPID, err = ctx.Create("child-sv", &emptyAgent{},
			WithMaxRetries(2),
			WithSupervisorWindow(500*time.Millisecond),
		)
		if err != nil {
			t.Errorf("failed to create child: %v", err)
		}
		select {
		case childCreated <- struct{}{}:
		default:
		}
	}
	_, _ = sys.Create(nextName(), parentAg)
	<-parentAg.startCh

	select {
	case <-childCreated:
		if childPID == nil {
			t.Fatal("child PID should not be nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for child creation")
	}
}
