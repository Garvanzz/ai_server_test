package db

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"xfx/pkg/agent"
	"xfx/pkg/log"
)

func TestMain(m *testing.M) {
	log.DefaultInit()
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Mock Redis
// ---------------------------------------------------------------------------

type mockConn struct {
	doFunc func(cmd string, args ...interface{}) (interface{}, error)
}

func (m *mockConn) Close() error                                           { return nil }
func (m *mockConn) Err() error                                             { return nil }
func (m *mockConn) Send(string, ...interface{}) error                      { return nil }
func (m *mockConn) Flush() error                                           { return nil }
func (m *mockConn) Receive() (interface{}, error)                          { return nil, nil }
func (m *mockConn) Do(cmd string, args ...interface{}) (interface{}, error) {
	if m.doFunc != nil {
		return m.doFunc(cmd, args...)
	}
	return "OK", nil
}

func mockPool(do func(string, ...interface{}) (interface{}, error)) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return &mockConn{doFunc: do}, nil
		},
		MaxIdle:   10,
		MaxActive: 50,
	}
}

func okPool() *redis.Pool {
	return mockPool(func(string, ...interface{}) (interface{}, error) {
		return "OK", nil
	})
}

func delayPool(d time.Duration, reply interface{}) *redis.Pool {
	return mockPool(func(string, ...interface{}) (interface{}, error) {
		time.Sleep(d)
		return reply, nil
	})
}

func errPool(e error) *redis.Pool {
	return mockPool(func(string, ...interface{}) (interface{}, error) {
		return nil, e
	})
}

// ---------------------------------------------------------------------------
// Capture Agent — receives messages from system.Cast into a channel
// ---------------------------------------------------------------------------

type captureAgent struct {
	received chan interface{}
	started  chan struct{}
}

func newCaptureAgent() *captureAgent {
	return &captureAgent{
		received: make(chan interface{}, 200),
		started:  make(chan struct{}),
	}
}

func (a *captureAgent) OnStart(agent.Context)                { close(a.started) }
func (a *captureAgent) OnStop()                              {}
func (a *captureAgent) OnTerminated(agent.PID, int)          {}
func (a *captureAgent) OnTick(time.Duration)                 {}
func (a *captureAgent) OnMessage(msg interface{}) interface{} {
	a.received <- msg
	return nil
}

func (a *captureAgent) wait(t *testing.T, timeout time.Duration) interface{} {
	t.Helper()
	select {
	case m := <-a.received:
		return m
	case <-time.After(timeout):
		t.Fatal("timeout waiting for actor message")
		return nil
	}
}

func (a *captureAgent) expectNoMsg(t *testing.T, wait time.Duration) {
	t.Helper()
	select {
	case m := <-a.received:
		t.Fatalf("unexpected message: %+v", m)
	case <-time.After(wait):
	}
}

// ---------------------------------------------------------------------------
// Test environment: agent.System + captureAgent + Go worker pool
// ---------------------------------------------------------------------------

type testEnv struct {
	sys   *agent.System
	ag    *captureAgent
	pid   agent.PID
	goSvc *Go
}

var testSeq int64

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	seq := atomic.AddInt64(&testSeq, 1)
	name := fmt.Sprintf("dbtest%d", seq)

	sys := agent.NewSystem(agent.WithName(name))
	sys.Start()

	ag := newCaptureAgent()
	pid, err := sys.Create("worker", ag)
	if err != nil {
		sys.Stop()
		t.Fatalf("create actor: %v", err)
	}

	select {
	case <-ag.started:
	case <-time.After(5 * time.Second):
		sys.Stop()
		t.Fatal("actor start timeout")
	}

	g := NewGo(sys)
	g.start()

	return &testEnv{sys: sys, ag: ag, pid: pid, goSvc: g}
}

func (e *testEnv) close() {
	e.goSvc.stop()
	e.sys.Stop()
}

// ---------------------------------------------------------------------------
// Tests: Worker pool mechanics
// ---------------------------------------------------------------------------

func TestGoBasicSubmitAndCallback(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	var gotReply atomic.Value
	done := make(chan struct{})

	pool := mockPool(func(cmd string, args ...interface{}) (interface{}, error) {
		return fmt.Sprintf("reply:%s", cmd), nil
	})

	err := env.goSvc.submitJob(pool, "PING", nil, func(res any, err error) {
		gotReply.Store(res)
		close(done)
	})
	if err != nil {
		t.Fatalf("submitJob: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("callback not called")
	}

	if gotReply.Load().(string) != "reply:PING" {
		t.Fatalf("unexpected reply: %v", gotReply.Load())
	}
}

func TestGoRedisErrorInCallback(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	expectedErr := errors.New("connection refused")
	done := make(chan struct{})
	var cbErr atomic.Value

	err := env.goSvc.submitJob(errPool(expectedErr), "GET", []any{"key"}, func(res any, err error) {
		cbErr.Store(err)
		close(done)
	})
	if err != nil {
		t.Fatalf("submitJob: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("callback not called")
	}

	if cbErr.Load().(error).Error() != expectedErr.Error() {
		t.Fatalf("expected error %q, got %q", expectedErr, cbErr.Load())
	}
}

func TestGoQueueFull(t *testing.T) {
	sys := agent.NewSystem(agent.WithName("qfull"))
	sys.Start()
	defer sys.Stop()

	g := &Go{
		taskQueue: make(chan *RedisJob, 2),
		system:    sys,
	}
	atomic.StoreInt32(&g.running, 1)

	slowPool := delayPool(time.Hour, "OK")

	for i := 0; i < 2; i++ {
		if err := g.submitJob(slowPool, "SET", nil, func(any, error) {}); err != nil {
			t.Fatalf("job %d should have been queued: %v", i, err)
		}
	}

	err := g.submitJob(slowPool, "SET", nil, func(any, error) {})
	if err == nil {
		t.Fatal("expected queue full error")
	}
}

func TestGoStoppedServiceReject(t *testing.T) {
	env := newTestEnv(t)
	env.goSvc.stop()

	err := env.goSvc.submitJob(okPool(), "PING", nil, func(any, error) {})
	if err == nil {
		t.Fatal("expected error from stopped service")
	}
	env.sys.Stop()
}

func TestGoDoubleStartPanics(t *testing.T) {
	sys := agent.NewSystem(agent.WithName("dstart"))
	sys.Start()
	defer sys.Stop()

	g := NewGo(sys)
	g.start()
	defer g.stop()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on double start")
		}
	}()
	g.start()
}

func TestGoDoubleStopSafe(t *testing.T) {
	env := newTestEnv(t)
	env.goSvc.stop()
	env.goSvc.stop()
	env.sys.Stop()
}

func TestGoPendingGoCounter(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	if n := atomic.LoadInt64(&env.goSvc.pendingGo); n != 0 {
		t.Fatalf("initial pendingGo should be 0, got %d", n)
	}

	blocker := make(chan struct{})
	pool := mockPool(func(string, ...interface{}) (interface{}, error) {
		<-blocker
		return "OK", nil
	})

	const N = 5
	for i := 0; i < N; i++ {
		env.goSvc.submitJob(pool, "SET", nil, func(any, error) {})
	}

	time.Sleep(100 * time.Millisecond)
	n := atomic.LoadInt64(&env.goSvc.pendingGo)
	if n != N {
		t.Fatalf("expected pendingGo=%d, got %d", N, n)
	}

	close(blocker)
	time.Sleep(200 * time.Millisecond)

	n = atomic.LoadInt64(&env.goSvc.pendingGo)
	if n != 0 {
		t.Fatalf("expected pendingGo=0 after completion, got %d", n)
	}
}

func TestGoPanicInCallbackRecovery(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	done := make(chan struct{})

	env.goSvc.submitJob(okPool(), "SET", nil, func(any, error) {
		panic("boom in callback")
	})

	time.Sleep(200 * time.Millisecond)

	env.goSvc.submitJob(okPool(), "PING", nil, func(any, error) {
		close(done)
	})

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("worker pool dead after panic")
	}
}

func TestGoPanicInRedisDo(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	panicPool := mockPool(func(string, ...interface{}) (interface{}, error) {
		panic("redis exploded")
	})

	env.goSvc.submitJob(panicPool, "SET", nil, func(any, error) {})
	time.Sleep(200 * time.Millisecond)

	done := make(chan struct{})
	env.goSvc.submitJob(okPool(), "PING", nil, func(any, error) {
		close(done)
	})

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("worker pool dead after Redis panic")
	}
}

func TestGoStopWaitsForInFlight(t *testing.T) {
	sys := agent.NewSystem(agent.WithName("stopwait"))
	sys.Start()
	defer sys.Stop()

	g := NewGo(sys)
	g.start()

	var completed int32
	pool := delayPool(300*time.Millisecond, "OK")

	for i := 0; i < 4; i++ {
		g.submitJob(pool, "SET", nil, func(any, error) {
			atomic.AddInt32(&completed, 1)
		})
	}

	time.Sleep(50 * time.Millisecond)
	g.stop()

	if c := atomic.LoadInt32(&completed); c != 4 {
		t.Fatalf("stop should wait for completion, got %d/4", c)
	}
}

// ---------------------------------------------------------------------------
// Tests: Worker pool → system.Cast → Actor message loop
// ---------------------------------------------------------------------------

func TestGoCallbackToActor(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	pool := mockPool(func(cmd string, args ...interface{}) (interface{}, error) {
		return "saved", nil
	})

	env.goSvc.submitJob(pool, "SET", []any{"k", "v"}, func(res any, err error) {
		env.sys.Cast(env.pid, &RedisRet{
			OpType: 42,
			Params: []int64{100, 200},
			Reply:  res,
			Err:    err,
		})
	})

	msg := env.ag.wait(t, 5*time.Second)
	ret, ok := msg.(*RedisRet)
	if !ok {
		t.Fatalf("expected *RedisRet, got %T", msg)
	}
	if ret.OpType != 42 {
		t.Fatalf("OpType: want 42, got %d", ret.OpType)
	}
	if ret.Reply.(string) != "saved" {
		t.Fatalf("Reply: want 'saved', got %v", ret.Reply)
	}
	if ret.Err != nil {
		t.Fatalf("Err: want nil, got %v", ret.Err)
	}
	if len(ret.Params) != 2 || ret.Params[0] != 100 || ret.Params[1] != 200 {
		t.Fatalf("Params: want [100,200], got %v", ret.Params)
	}
}

func TestGoRedisErrorReachesActor(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	redisErr := errors.New("READONLY")

	env.goSvc.submitJob(errPool(redisErr), "SET", []any{"k", "v"}, func(res any, err error) {
		env.sys.Cast(env.pid, &RedisRet{
			OpType: 1,
			Reply:  res,
			Err:    err,
		})
	})

	msg := env.ag.wait(t, 5*time.Second)
	ret := msg.(*RedisRet)
	if ret.Err == nil || ret.Err.Error() != "READONLY" {
		t.Fatalf("expected READONLY error, got %v", ret.Err)
	}
	if ret.Reply != nil {
		t.Fatalf("expected nil reply, got %v", ret.Reply)
	}
}

func TestGoConcurrentJobsAllReachActor(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	const N = 100
	pool := mockPool(func(cmd string, args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			return args[0], nil
		}
		return "OK", nil
	})

	for i := 0; i < N; i++ {
		idx := i
		env.goSvc.submitJob(pool, "SET", []any{idx}, func(res any, err error) {
			env.sys.Cast(env.pid, &RedisRet{
				OpType: idx,
				Reply:  res,
			})
		})
	}

	seen := make(map[int]bool)
	for i := 0; i < N; i++ {
		msg := env.ag.wait(t, 10*time.Second)
		ret := msg.(*RedisRet)
		seen[ret.OpType] = true
	}

	for i := 0; i < N; i++ {
		if !seen[i] {
			t.Fatalf("missing OpType=%d", i)
		}
	}
}

func TestGoActorReceivesAllMessages(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	const N = 20
	pool := okPool()

	customAgent := newCaptureAgent()
	seq := atomic.AddInt64(&testSeq, 1)
	pid2, err := env.sys.Create(fmt.Sprintf("order%d", seq), customAgent)
	if err != nil {
		t.Fatal(err)
	}
	<-customAgent.started

	for i := 0; i < N; i++ {
		idx := i
		env.goSvc.submitJob(pool, "SET", nil, func(any, error) {
			env.sys.Cast(pid2, &RedisRet{OpType: idx})
		})
	}

	seen := make(map[int]bool)
	for i := 0; i < N; i++ {
		msg := customAgent.wait(t, 5*time.Second)
		ret := msg.(*RedisRet)
		seen[ret.OpType] = true
	}

	if len(seen) != N {
		t.Fatalf("expected %d unique messages, got %d", N, len(seen))
	}
}

// ---------------------------------------------------------------------------
// Tests: RedisAsyncExec integration (uses global asyncGo)
// ---------------------------------------------------------------------------

func TestRedisAsyncExecFullLoop(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	origAsyncGo := asyncGo
	asyncGo = env.goSvc
	defer func() { asyncGo = origAsyncGo }()

	pool := mockPool(func(cmd string, args ...interface{}) (interface{}, error) {
		if cmd == "SET" {
			return "OK", nil
		}
		return nil, fmt.Errorf("unknown cmd: %s", cmd)
	})

	origEngine := Engine
	Engine = &CDBEngine{Redis: pool}
	defer func() { Engine = origEngine }()

	RedisAsyncExec(env.pid, 7, []int64{111, 222}, "SET", "mykey", "myval")

	msg := env.ag.wait(t, 5*time.Second)
	ret := msg.(*RedisRet)
	if ret.OpType != 7 {
		t.Fatalf("OpType: want 7, got %d", ret.OpType)
	}
	if ret.Reply.(string) != "OK" {
		t.Fatalf("Reply: want OK, got %v", ret.Reply)
	}
	if ret.Err != nil {
		t.Fatalf("Err: want nil, got %v", ret.Err)
	}
	if len(ret.Params) != 2 || ret.Params[0] != 111 || ret.Params[1] != 222 {
		t.Fatalf("Params: want [111,222], got %v", ret.Params)
	}
}

func TestRedisAsyncExecWithError(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	origAsyncGo := asyncGo
	asyncGo = env.goSvc
	defer func() { asyncGo = origAsyncGo }()

	origEngine := Engine
	Engine = &CDBEngine{Redis: errPool(errors.New("NOPERM"))}
	defer func() { Engine = origEngine }()

	RedisAsyncExec(env.pid, 3, nil, "DEL", "forbidden")

	msg := env.ag.wait(t, 5*time.Second)
	ret := msg.(*RedisRet)
	if ret.Err == nil || ret.Err.Error() != "NOPERM" {
		t.Fatalf("expected NOPERM, got %v", ret.Err)
	}
}

func TestRedisAsyncExecQueueFullSilent(t *testing.T) {
	sys := agent.NewSystem(agent.WithName("asyncfull"))
	sys.Start()
	defer sys.Stop()

	ag := newCaptureAgent()
	pid, _ := sys.Create("w", ag)
	<-ag.started

	g := &Go{
		taskQueue: make(chan *RedisJob, 1),
		system:    sys,
	}
	atomic.StoreInt32(&g.running, 1)

	origAsyncGo := asyncGo
	asyncGo = g
	defer func() { asyncGo = origAsyncGo }()

	slowPool := delayPool(time.Hour, "OK")
	asyncGo.submitJob(slowPool, "BLOCK", nil, func(any, error) {})

	origEngine := Engine
	Engine = &CDBEngine{Redis: okPool()}
	defer func() { Engine = origEngine }()

	RedisAsyncExec(pid, 1, nil, "SET", "k", "v")

	ag.expectNoMsg(t, 500*time.Millisecond)
}

// ---------------------------------------------------------------------------
// Tests: Multiple actors, each receives own RedisRet
// ---------------------------------------------------------------------------

func TestGoMultipleActorsIsolation(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	ag2 := newCaptureAgent()
	seq := atomic.AddInt64(&testSeq, 1)
	pid2, err := env.sys.Create(fmt.Sprintf("iso%d", seq), ag2)
	if err != nil {
		t.Fatal(err)
	}
	<-ag2.started

	pool := okPool()

	env.goSvc.submitJob(pool, "SET", nil, func(res any, err error) {
		env.sys.Cast(env.pid, &RedisRet{OpType: 1, Reply: "for-actor1"})
	})
	env.goSvc.submitJob(pool, "SET", nil, func(res any, err error) {
		env.sys.Cast(pid2, &RedisRet{OpType: 2, Reply: "for-actor2"})
	})

	msg1 := env.ag.wait(t, 5*time.Second)
	msg2 := ag2.wait(t, 5*time.Second)

	r1 := msg1.(*RedisRet)
	r2 := msg2.(*RedisRet)

	if r1.OpType != 1 || r1.Reply.(string) != "for-actor1" {
		t.Fatalf("actor1 got wrong message: %+v", r1)
	}
	if r2.OpType != 2 || r2.Reply.(string) != "for-actor2" {
		t.Fatalf("actor2 got wrong message: %+v", r2)
	}

	env.ag.expectNoMsg(t, 200*time.Millisecond)
	ag2.expectNoMsg(t, 200*time.Millisecond)
}

// ---------------------------------------------------------------------------
// Tests: Cast to nil PID (graceful)
// ---------------------------------------------------------------------------

func TestGoCastToNilPIDNocrash(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	done := make(chan struct{})
	env.goSvc.submitJob(okPool(), "SET", nil, func(res any, err error) {
		env.sys.Cast(nil, &RedisRet{OpType: 99})
		close(done)
	})

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("callback hung on nil PID cast")
	}
}

// ---------------------------------------------------------------------------
// Tests: High concurrency stress
// ---------------------------------------------------------------------------

func TestGoHighConcurrencyStress(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	const N = 500
	pool := mockPool(func(cmd string, args ...interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond)
		if len(args) > 0 {
			return args[0], nil
		}
		return "OK", nil
	})

	var wg sync.WaitGroup
	wg.Add(N)

	for i := 0; i < N; i++ {
		idx := i
		env.goSvc.submitJob(pool, "SET", []any{idx}, func(res any, err error) {
			env.sys.Cast(env.pid, &RedisRet{OpType: idx, Reply: res, Err: err})
			wg.Done()
		})
	}

	wg.Wait()

	received := 0
	timeout := time.After(15 * time.Second)
	for received < N {
		select {
		case <-env.ag.received:
			received++
		case <-timeout:
			t.Fatalf("stress test: only received %d/%d", received, N)
		}
	}
}

// ---------------------------------------------------------------------------
// Tests: Verify worker count bounds concurrency
// ---------------------------------------------------------------------------

func TestGoWorkerCountLimitsConcurrency(t *testing.T) {
	env := newTestEnv(t)
	defer env.close()

	var concurrent int64
	var maxConcurrent int64
	var mu sync.Mutex

	pool := mockPool(func(string, ...interface{}) (interface{}, error) {
		cur := atomic.AddInt64(&concurrent, 1)
		mu.Lock()
		if cur > maxConcurrent {
			maxConcurrent = cur
		}
		mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt64(&concurrent, -1)
		return "OK", nil
	})

	const N = 30
	done := make(chan struct{})
	var remaining int32 = N

	for i := 0; i < N; i++ {
		env.goSvc.submitJob(pool, "SET", nil, func(any, error) {
			if atomic.AddInt32(&remaining, -1) == 0 {
				close(done)
			}
		})
	}

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("timeout")
	}

	mu.Lock()
	mc := maxConcurrent
	mu.Unlock()

	if mc > int64(defaultWorkerCount) {
		t.Fatalf("max concurrency %d exceeded worker count %d", mc, defaultWorkerCount)
	}
	if mc == 0 {
		t.Fatal("no concurrency observed")
	}
	t.Logf("max observed concurrency: %d (worker count: %d)", mc, defaultWorkerCount)
}
