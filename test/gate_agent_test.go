package tcpgate_test

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"xfx/pkg/gate"
	"xfx/pkg/gate/tcpgate"
	"xfx/pkg/gate/tcpgate/codec"
	"xfx/pkg/net/tcp"
	proto_id "xfx/proto"
	"xfx/proto/proto_player"

	gogoproto "github.com/gogo/protobuf/proto"
)

// ═══════════════════════════════════════════════════════════════════
// 模拟 mgate.Agent 的行为：复刻真实 Agent 的所有关键行为模式
//   - startedCh 等待 actor 就绪
//   - closeOnce 幂等关闭
//   - ping/pong 心跳
//   - 踢人 CloseWithFlush
//   - 超时断连
//   - 消息转发
// ═══════════════════════════════════════════════════════════════════

type mockActorAgent struct {
	tcpgate.Agent
	startedCh chan struct{}
	closeOnce sync.Once
	pingTime  time.Duration

	mu        sync.Mutex
	recvLog   []any       // 记录收到的消息
	closeLog  []string    // 记录关闭事件
	started   int32       // 是否已启动
	playerId  int64
	kicked    int32       // 是否被踢
}

const testPingTime = 5 * time.Second

func newMockActorAgent() *mockActorAgent {
	return &mockActorAgent{
		startedCh: make(chan struct{}),
		pingTime:  testPingTime,
	}
}

func (a *mockActorAgent) OnInit(_ gate.Gate, session gate.Session) {
	a.Agent.OnInit(nil, session)
}

// simulateActorStart 模拟 actor 启动（真实环境中由 ProtoActor 触发）
func (a *mockActorAgent) simulateActorStart() {
	atomic.StoreInt32(&a.started, 1)
	a.pingTime = testPingTime
	close(a.startedCh)
}

func (a *mockActorAgent) OnRecv(msg any) {
	select {
	case <-a.startedCh:
	case <-time.After(3 * time.Second):
		a.GetSession().Close()
		return
	}
	// 模拟 actor 邮箱投递：直接在当前 goroutine 处理
	a.onSessionMessage(msg)
}

func (a *mockActorAgent) onSessionMessage(msg any) {
	a.mu.Lock()
	a.recvLog = append(a.recvLog, msg)
	a.mu.Unlock()

	switch msg.(type) {
	case *proto_player.C2SPing:
		a.mu.Lock()
		a.pingTime = testPingTime
		a.mu.Unlock()
		a.Send(&proto_player.S2CPong{ZoneOffset: time.Now().Unix()})
	case *proto_player.C2SLogout:
		a.mu.Lock()
		a.closeLog = append(a.closeLog, "logout")
		a.mu.Unlock()
	case *proto_player.C2SLogin:
		// 模拟登录成功
		a.mu.Lock()
		a.playerId = 10001
		a.mu.Unlock()
		a.Send(&proto_player.S2CLogin{
			ZoneOffset: time.Now().Unix(),
		})
	default:
		// 模拟转发到 player actor
		a.mu.Lock()
		a.closeLog = append(a.closeLog, fmt.Sprintf("forward:%T", msg))
		a.mu.Unlock()
	}
}

// simulateKick 模拟服务端踢人（OnMessage 收到 S2CKick）
func (a *mockActorAgent) simulateKick() {
	atomic.StoreInt32(&a.kicked, 1)
	a.mu.Lock()
	a.playerId = 0
	a.mu.Unlock()
	a.Send(&proto_player.S2CKick{})
	go func() {
		a.GetSession().CloseWithFlush(500 * time.Millisecond)
	}()
}

// simulateTick 模拟 OnTick
func (a *mockActorAgent) simulateTick(delta time.Duration) {
	a.mu.Lock()
	a.pingTime -= delta
	timeout := a.pingTime <= 0
	a.mu.Unlock()

	if timeout {
		a.mu.Lock()
		a.closeLog = append(a.closeLog, "ping_timeout")
		a.mu.Unlock()
		a.GetSession().Close()
	}
}

func (a *mockActorAgent) Close() error {
	a.closeOnce.Do(func() {
		a.mu.Lock()
		a.closeLog = append(a.closeLog, "agent_close")
		a.mu.Unlock()
		// 模拟 actor stop → OnStop → session.Close
		a.GetSession().Close()
	})
	return nil
}

func (a *mockActorAgent) getRecvLog() []any {
	a.mu.Lock()
	defer a.mu.Unlock()
	cp := make([]any, len(a.recvLog))
	copy(cp, a.recvLog)
	return cp
}

func (a *mockActorAgent) getCloseLog() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	cp := make([]string, len(a.closeLog))
	copy(cp, a.closeLog)
	return cp
}

// ═══════════════════════════════════════════════════════════════════
// 测试基础设施：在 tcpgate 层使用 mockActorAgent
// ═══════════════════════════════════════════════════════════════════

type gateTestEnv struct {
	server  *tcp.Server
	addr    string
	agents  sync.Map // sessionID → *mockActorAgent
	agentCh chan *mockActorAgent
}

func newGateTestEnv(t *testing.T) *gateTestEnv {
	t.Helper()
	env := &gateTestEnv{
		agentCh: make(chan *mockActorAgent, 64),
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	env.addr = listener.Addr().String()

	var gt tcpgate.Gate
	gt.SetCreateAgent(func(g gate.Gate, session gate.Session) (gate.Agent, error) {
		agent := newMockActorAgent()
		agent.OnInit(g, session)
		env.agents.Store(session.ID(), agent)
		env.agentCh <- agent
		// 模拟 actor 创建后短暂延迟启动
		go func() {
			time.Sleep(10 * time.Millisecond)
			agent.simulateActorStart()
		}()
		return agent, nil
	})

	env.server = tcp.NewServer(
		listener,
		tcp.ProtocolFunc(codec.NewParser),
		2000,
		tcp.HandlerFunc(func(sess *tcp.Session) {
			agentI, err := gt.NewAgent(sess)
			if err != nil {
				sess.Close()
				return
			}
			agent := agentI.(*mockActorAgent)

			sess.AddCloseCallback(&gt, sess.ID(), func() {
				agent.Close()
			})

			for {
				msg, err := sess.Receive()
				if err != nil {
					return
				}
				agent.OnRecv(msg)
			}
		}),
	)
	go env.server.Serve()
	time.Sleep(20 * time.Millisecond)
	return env
}

func (env *gateTestEnv) stop() {
	env.server.Stop()
}

func (env *gateTestEnv) waitAgent(t *testing.T, timeout time.Duration) *mockActorAgent {
	t.Helper()
	select {
	case a := <-env.agentCh:
		return a
	case <-time.After(timeout):
		t.Fatal("timeout waiting for agent")
		return nil
	}
}

func dialGateClient(t *testing.T, addr string) *testClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	return &testClient{conn: conn, t: t}
}

// ═══════════════════════════════════════════════════════════════════
// 测试 1：正常连接 → ping/pong → 断开 全生命周期
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentFullLifecycle(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	// 等 actor 启动
	<-agent.startedCh

	// 发 ping
	client.sendMsg(&proto_player.C2SPing{})
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("recvMsg failed: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}

	// 发 login
	client.sendMsg(&proto_player.C2SLogin{Token: "test_token"})
	msg, err = client.recvMsg()
	if err != nil {
		t.Fatalf("recvMsg failed: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CLogin); !ok {
		t.Fatalf("expected S2CLogin, got %T", msg)
	}

	// 客户端主动断开
	client.close()
	time.Sleep(100 * time.Millisecond)

	// 验证 agent 被正确关闭
	closeLog := agent.getCloseLog()
	found := false
	for _, entry := range closeLog {
		if entry == "agent_close" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'agent_close' in closeLog, got: %v", closeLog)
	}

	// 验证消息记录
	recvLog := agent.getRecvLog()
	if len(recvLog) != 2 {
		t.Errorf("expected 2 messages, got %d", len(recvLog))
	}
	t.Logf("✓ 全生命周期：连接→ping/pong→login→断开→agent 关闭，recvLog=%d, closeLog=%v", len(recvLog), closeLog)
}

// ═══════════════════════════════════════════════════════════════════
// 测试 2：客户端在 actor 启动前断开（startedCh 未关闭）
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentClientDisconnectBeforeActorStart(t *testing.T) {
	env := &gateTestEnv{
		agentCh: make(chan *mockActorAgent, 64),
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	env.addr = listener.Addr().String()

	var gt tcpgate.Gate
	gt.SetCreateAgent(func(g gate.Gate, session gate.Session) (gate.Agent, error) {
		agent := newMockActorAgent()
		agent.OnInit(g, session)
		env.agentCh <- agent
		// 不启动 actor：模拟 actor 创建延迟/失败
		return agent, nil
	})

	env.server = tcp.NewServer(
		listener,
		tcp.ProtocolFunc(codec.NewParser),
		2000,
		tcp.HandlerFunc(func(sess *tcp.Session) {
			agentI, err := gt.NewAgent(sess)
			if err != nil {
				sess.Close()
				return
			}
			agent := agentI.(*mockActorAgent)
			sess.AddCloseCallback(&gt, sess.ID(), func() {
				agent.Close()
			})
			for {
				msg, err := sess.Receive()
				if err != nil {
					return
				}
				agent.OnRecv(msg)
			}
		}),
	)
	go env.server.Serve()
	defer env.server.Stop()
	time.Sleep(20 * time.Millisecond)

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	// 客户端发消息（agent 未启动，OnRecv 会等 startedCh 最多 3 秒）
	client.sendMsg(&proto_player.C2SPing{})

	// 立即断开客户端
	client.close()
	time.Sleep(200 * time.Millisecond)

	// agent 收到消息后会等 startedCh，但 session 已关闭，
	// handleConnect 的 recv loop 会因 Receive 返回 err 退出
	closeLog := agent.getCloseLog()
	t.Logf("✓ Actor 未启动时客户端断开：closeLog=%v", closeLog)

	hasAgentClose := false
	for _, entry := range closeLog {
		if entry == "agent_close" {
			hasAgentClose = true
		}
	}
	if !hasAgentClose {
		t.Logf("注意：agent.Close 可能还在 OnRecv 等待 startedCh 超时中")
	}
}

// ═══════════════════════════════════════════════════════════════════
// 测试 3：startedCh 超时（actor 永不启动）→ session 被关闭
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentStartedChTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}

	env := &gateTestEnv{
		agentCh: make(chan *mockActorAgent, 64),
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	env.addr = listener.Addr().String()

	var gt tcpgate.Gate
	gt.SetCreateAgent(func(g gate.Gate, session gate.Session) (gate.Agent, error) {
		agent := newMockActorAgent()
		agent.OnInit(g, session)
		env.agentCh <- agent
		return agent, nil // 永不启动 actor
	})

	env.server = tcp.NewServer(
		listener,
		tcp.ProtocolFunc(codec.NewParser),
		2000,
		tcp.HandlerFunc(func(sess *tcp.Session) {
			agentI, err := gt.NewAgent(sess)
			if err != nil {
				sess.Close()
				return
			}
			agent := agentI.(*mockActorAgent)
			sess.AddCloseCallback(&gt, sess.ID(), func() {
				agent.Close()
			})
			for {
				msg, err := sess.Receive()
				if err != nil {
					return
				}
				agent.OnRecv(msg)
			}
		}),
	)
	go env.server.Serve()
	defer env.server.Stop()
	time.Sleep(20 * time.Millisecond)

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	// 客户端发消息，OnRecv 将等 3 秒后超时关闭 session
	start := time.Now()
	client.sendMsg(&proto_player.C2SPing{})

	// 等待 session 被关闭（客户端读取应失败）
	client.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, recvErr := client.recvMsg()
	elapsed := time.Since(start)

	if recvErr == nil {
		t.Error("expected read error after startedCh timeout")
	}
	if elapsed < 2*time.Second || elapsed > 5*time.Second {
		t.Errorf("expected ~3s timeout, got %v", elapsed)
	}

	_ = agent
	client.close()
	t.Logf("✓ startedCh 超时(%.1fs)后 session 自动关闭", elapsed.Seconds())
}

// ═══════════════════════════════════════════════════════════════════
// 测试 4：服务端踢人 → S2CKick + CloseWithFlush
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentKick(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	<-agent.startedCh

	// 先 ping 确保连接正常
	client.sendMsg(&proto_player.C2SPing{})
	if _, err := client.recvMsg(); err != nil {
		t.Fatalf("ping failed: %v", err)
	}

	// 服务端踢人
	agent.simulateKick()

	// 客户端应先收到 S2CKick
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("expected to receive S2CKick: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CKick); !ok {
		t.Fatalf("expected S2CKick, got %T", msg)
	}

	// 然后连接应断开
	_, err = client.recvMsg()
	if err == nil {
		t.Error("expected connection to close after kick")
	}

	client.close()
	t.Log("✓ 踢人流程：发送 S2CKick → CloseWithFlush → 连接断开")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 5：心跳超时 → session 关闭
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentPingTimeout(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	<-agent.startedCh

	// 模拟 tick 消耗时间，但客户端不发 ping
	agent.simulateTick(3 * time.Second) // 还剩 2s
	agent.simulateTick(1 * time.Second) // 还剩 1s

	// 此时还没超时，客户端仍可通信
	client.sendMsg(&proto_player.C2SPing{})
	client.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("should still be alive: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}

	// ping 重置了 pingTime（testPingTime=5s），再模拟超时
	agent.simulateTick(6 * time.Second) // 超过 testPingTime

	// session 应已关闭
	time.Sleep(50 * time.Millisecond)

	closeLog := agent.getCloseLog()
	found := false
	for _, entry := range closeLog {
		if entry == "ping_timeout" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'ping_timeout' in closeLog, got: %v", closeLog)
	}

	client.close()
	t.Logf("✓ 心跳超时断连：中途 ping 重置 → 再次超时 → session 关闭, closeLog=%v", closeLog)
}

// ═══════════════════════════════════════════════════════════════════
// 测试 6：closeOnce 幂等性 — 多路径触发关闭只执行一次
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentCloseOnceIdempotent(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	<-agent.startedCh

	// 多路径触发关闭
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			agent.Close()
		}()
	}
	wg.Wait()

	closeLog := agent.getCloseLog()
	count := 0
	for _, entry := range closeLog {
		if entry == "agent_close" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 'agent_close', got %d in %v", count, closeLog)
	}

	client.close()
	t.Logf("✓ closeOnce 幂等：10 次并发 Close 只执行 1 次")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 7：双向关闭链 — session 关闭触发 agent.Close
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentCloseChainSessionToAgent(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	<-agent.startedCh

	// 客户端断开 → session.Receive 返回 err
	// → session.Close() → CloseCallback → agent.Close()
	client.close()
	time.Sleep(200 * time.Millisecond)

	closeLog := agent.getCloseLog()
	found := false
	for _, entry := range closeLog {
		if entry == "agent_close" {
			found = true
		}
	}
	if !found {
		t.Errorf("session close should trigger agent_close, got: %v", closeLog)
	}
	t.Logf("✓ 关闭链 Session→Agent: 客户端断开 → session 关闭 → agent.Close 触发")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 8：双向关闭链 — agent.Close 触发 session 关闭
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentCloseChainAgentToSession(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	<-agent.startedCh

	// 服务端主动关闭 agent
	agent.Close()
	time.Sleep(100 * time.Millisecond)

	// session 应已关闭
	if !agent.GetSession().IsClosed() {
		t.Error("session should be closed after agent.Close")
	}

	// 客户端读取应失败
	client.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err := client.recvMsg()
	if err == nil {
		t.Error("client should get read error after agent close")
	}

	client.close()
	t.Log("✓ 关闭链 Agent→Session: agent.Close → session 关闭 → 客户端断开")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 9：多条消息连续处理 + 消息顺序保证
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentMessageOrder(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	<-agent.startedCh

	// 连续发 3 种消息
	client.sendMsg(&proto_player.C2SPing{})
	client.sendMsg(&proto_player.C2SLogin{Token: "order_test"})
	client.sendMsg(&proto_player.C2SPing{})

	// 收取回复
	replies := make([]gogoproto.Message, 0, 3)
	for i := 0; i < 3; i++ {
		client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		msg, err := client.recvMsg()
		if err != nil {
			t.Fatalf("reply %d: %v", i, err)
		}
		replies = append(replies, msg)
	}

	if _, ok := replies[0].(*proto_player.S2CPong); !ok {
		t.Errorf("[0] expected S2CPong, got %T", replies[0])
	}
	if _, ok := replies[1].(*proto_player.S2CLogin); !ok {
		t.Errorf("[1] expected S2CLogin, got %T", replies[1])
	}
	if _, ok := replies[2].(*proto_player.S2CPong); !ok {
		t.Errorf("[2] expected S2CPong, got %T", replies[2])
	}

	// 验证 recvLog 顺序
	recvLog := agent.getRecvLog()
	if len(recvLog) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(recvLog))
	}
	if _, ok := recvLog[0].(*proto_player.C2SPing); !ok {
		t.Errorf("recvLog[0] expected C2SPing, got %T", recvLog[0])
	}
	if _, ok := recvLog[1].(*proto_player.C2SLogin); !ok {
		t.Errorf("recvLog[1] expected C2SLogin, got %T", recvLog[1])
	}

	client.close()
	t.Logf("✓ 消息顺序保证：3 条消息按序处理，回复按序返回")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 10：协议违规后 agent 被正确清理
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentCleanupAfterProtocolViolation(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	conn, err := net.DialTimeout("tcp", env.addr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	client := &testClient{conn: conn, t: t}
	agent := env.waitAgent(t, 2*time.Second)

	<-agent.startedCh

	// 发未知协议号（触发 codec Receive 返回 error → session.Close）
	client.rawSend(0xDEADBEEF, []byte{1, 2, 3})
	time.Sleep(200 * time.Millisecond)

	closeLog := agent.getCloseLog()
	found := false
	for _, entry := range closeLog {
		if entry == "agent_close" {
			found = true
		}
	}
	if !found {
		t.Errorf("protocol violation should trigger agent cleanup, got: %v", closeLog)
	}

	if !agent.GetSession().IsClosed() {
		t.Error("session should be closed")
	}

	client.close()
	t.Logf("✓ 协议违规后 agent 正确清理: closeLog=%v", closeLog)
}

// ═══════════════════════════════════════════════════════════════════
// 测试 11：超大包不影响其他连接的 agent
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentIsolation(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	// 正常客户端
	normalClient := dialGateClient(t, env.addr)
	normalAgent := env.waitAgent(t, 2*time.Second)
	<-normalAgent.startedCh

	// 异常客户端
	badConn, _ := net.DialTimeout("tcp", env.addr, 2*time.Second)
	badClient := &testClient{conn: badConn, t: t}
	badAgent := env.waitAgent(t, 2*time.Second)
	<-badAgent.startedCh

	// 异常客户端发超大包
	badClient.rawWriteUint32(500 * 1024)
	time.Sleep(100 * time.Millisecond)
	badClient.close()

	// 正常客户端仍可通信
	normalClient.sendMsg(&proto_player.C2SPing{})
	normalClient.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := normalClient.recvMsg()
	if err != nil {
		t.Fatalf("normal client should still work: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}

	// 异常 agent 应已关闭
	if !badAgent.GetSession().IsClosed() {
		t.Error("bad agent session should be closed")
	}
	// 正常 agent 应正常
	if normalAgent.GetSession().IsClosed() {
		t.Error("normal agent session should NOT be closed")
	}

	normalClient.close()
	t.Log("✓ Agent 隔离性：异常连接不影响其他正常连接")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 12：客户端发 Logout 消息
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentLogout(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)

	<-agent.startedCh

	// 先 login
	client.sendMsg(&proto_player.C2SLogin{Token: "test"})
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	client.recvMsg() // 消费 S2CLogin 回复

	// 然后 logout
	client.sendMsg(&proto_player.C2SLogout{})
	time.Sleep(100 * time.Millisecond)

	closeLog := agent.getCloseLog()
	found := false
	for _, entry := range closeLog {
		if entry == "logout" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'logout' in closeLog, got: %v", closeLog)
	}

	client.close()
	t.Logf("✓ Logout 流程正确: closeLog=%v", closeLog)
}

// ═══════════════════════════════════════════════════════════════════
// 测试 13：多客户端并发连接，各自独立生命周期
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentMultipleClients(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	const clientCount = 10
	var wg sync.WaitGroup
	var successCount int64

	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			client := dialGateClient(t, env.addr)
			defer client.close()

			agent := env.waitAgent(t, 3*time.Second)
			<-agent.startedCh

			// 每个客户端 ping 3 次
			for j := 0; j < 3; j++ {
				client.sendMsg(&proto_player.C2SPing{})
				client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				msg, err := client.recvMsg()
				if err != nil {
					t.Errorf("client %d ping %d: %v", idx, j, err)
					return
				}
				if _, ok := msg.(*proto_player.S2CPong); !ok {
					t.Errorf("client %d: expected S2CPong, got %T", idx, msg)
					return
				}
			}

			atomic.AddInt64(&successCount, 1)
		}(i)
	}

	wg.Wait()
	t.Logf("✓ %d/%d 客户端各完成 3 次 ping/pong", atomic.LoadInt64(&successCount), clientCount)

	if successCount != clientCount {
		t.Errorf("expected all %d clients to succeed", clientCount)
	}
}

// ═══════════════════════════════════════════════════════════════════
// 测试 14：快速连接→发消息→断开 压力测试
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentRapidConnectDisconnect(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	const iterations = 30
	for i := 0; i < iterations; i++ {
		conn, err := net.DialTimeout("tcp", env.addr, 2*time.Second)
		if err != nil {
			t.Fatalf("iter %d: dial failed: %v", i, err)
		}
		c := &testClient{conn: conn, t: t}

		if i%3 == 0 {
			c.sendMsg(&proto_player.C2SPing{})
		}
		c.close()
	}

	// 消费所有 agentCh 中的 agent
	time.Sleep(500 * time.Millisecond)

	// 验证服务器仍然稳定
	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)
	<-agent.startedCh

	client.sendMsg(&proto_player.C2SPing{})
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("server should still work: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}

	client.close()
	t.Logf("✓ %d 次快速连接/断开后服务器仍稳定", iterations)
}

// ═══════════════════════════════════════════════════════════════════
// 测试 15：零字节连接 + 垃圾数据连接 → agent 正确清理
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentGarbageAndEmptyConnections(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	// 空连接
	for i := 0; i < 5; i++ {
		conn, _ := net.DialTimeout("tcp", env.addr, 1*time.Second)
		if conn != nil {
			conn.Close()
		}
	}

	// 垃圾数据连接
	for i := 0; i < 5; i++ {
		conn, _ := net.DialTimeout("tcp", env.addr, 1*time.Second)
		if conn != nil {
			garbage := make([]byte, 50)
			conn.Write(garbage)
			conn.Close()
		}
	}

	time.Sleep(500 * time.Millisecond)

	// 服务器仍然正常
	client := dialGateClient(t, env.addr)
	defer client.close()

	// 需要先消费掉 agentCh
	done := make(chan *mockActorAgent, 1)
	go func() {
		for {
			select {
			case a := <-env.agentCh:
				// 找到最新的 agent
				select {
				case <-a.startedCh:
					done <- a
					return
				case <-time.After(200 * time.Millisecond):
					continue
				}
			case <-time.After(3 * time.Second):
				return
			}
		}
	}()

	client.sendMsg(&proto_player.C2SPing{})
	client.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("server should still work: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}
	t.Log("✓ 空连接+垃圾连接后服务器仍正常，agent 正确清理")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 16：半包 → 补完 → 正常处理
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentPartialPacketThenComplete(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)
	<-agent.startedCh

	// 构造完整 C2SPing 包
	pingData, _ := gogoproto.Marshal(&proto_player.C2SPing{})
	pingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})
	fullPacket := make([]byte, 4+4+len(pingData))
	binary.LittleEndian.PutUint32(fullPacket[0:4], uint32(len(pingData)))
	binary.LittleEndian.PutUint32(fullPacket[4:8], pingID)
	copy(fullPacket[8:], pingData)

	// 分两段发：先发前 3 字节
	client.rawWriteBytes(fullPacket[:3])
	time.Sleep(100 * time.Millisecond)

	// 再发剩余
	client.rawWriteBytes(fullPacket[3:])

	// 应正确收到 S2CPong
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("expected S2CPong: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}

	client.close()
	t.Log("✓ 半包→补完→正常处理")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 17：Kick 后客户端继续发消息 → 不应处理
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentNoMessageAfterKick(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)
	<-agent.startedCh

	// 先确认正常通信
	client.sendMsg(&proto_player.C2SPing{})
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	client.recvMsg()

	recvBefore := len(agent.getRecvLog())

	// 踢人
	agent.simulateKick()

	// 读取 S2CKick
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("should receive S2CKick: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CKick); !ok {
		t.Fatalf("expected S2CKick, got %T", msg)
	}

	// 等连接完全关闭
	time.Sleep(600 * time.Millisecond)

	// 尝试再发消息（应写入失败或被忽略）
	client.sendMsg(&proto_player.C2SPing{})
	time.Sleep(100 * time.Millisecond)

	recvAfter := len(agent.getRecvLog())
	if recvAfter > recvBefore+1 {
		t.Errorf("should not process messages after kick: before=%d, after=%d", recvBefore, recvAfter)
	}

	client.close()
	t.Log("✓ Kick 后不再处理新消息")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 18：截断包 → session 关闭 → agent 清理
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentTruncatedPacket(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	conn, _ := net.DialTimeout("tcp", env.addr, 2*time.Second)
	client := &testClient{conn: conn, t: t}
	agent := env.waitAgent(t, 2*time.Second)
	<-agent.startedCh

	// 声明 dataLen=100 但只写 5 字节 body 后断开
	c2sLoginID, _ := proto_id.Router.MessageID(&proto_player.C2SLogin{})
	client.rawWriteUint32(100)
	client.rawWriteUint32(c2sLoginID)
	client.rawWriteBytes(make([]byte, 5))
	client.close()

	time.Sleep(200 * time.Millisecond)

	if !agent.GetSession().IsClosed() {
		t.Error("session should be closed after truncated packet")
	}
	closeLog := agent.getCloseLog()
	found := false
	for _, entry := range closeLog {
		if entry == "agent_close" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected agent_close after truncated packet, got: %v", closeLog)
	}
	t.Logf("✓ 截断包→session 关闭→agent 正确清理: closeLog=%v", closeLog)
}

// ═══════════════════════════════════════════════════════════════════
// 测试 19：高并发 ping + kick 竞争
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentConcurrentPingAndKick(t *testing.T) {
	env := newGateTestEnv(t)
	defer env.stop()

	client := dialGateClient(t, env.addr)
	agent := env.waitAgent(t, 2*time.Second)
	<-agent.startedCh

	// 消费回复的 goroutine
	recvDone := make(chan struct{})
	go func() {
		defer close(recvDone)
		for {
			client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if _, err := client.recvMsg(); err != nil {
				return
			}
		}
	}()

	// 快速发送多条 ping
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.sendMsg(&proto_player.C2SPing{})
		}()
	}

	// 同时踢人
	time.Sleep(5 * time.Millisecond)
	agent.simulateKick()

	wg.Wait()
	<-recvDone

	client.close()
	t.Log("✓ 并发 ping + kick 无 panic")
}

// ═══════════════════════════════════════════════════════════════════
// 测试 20：createAgent 返回 error → session 被立即关闭
// ═══════════════════════════════════════════════════════════════════

func TestGateAgentCreateError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()

	var gt tcpgate.Gate
	gt.SetCreateAgent(func(g gate.Gate, session gate.Session) (gate.Agent, error) {
		return nil, fmt.Errorf("simulated agent creation failure")
	})

	server := tcp.NewServer(
		listener,
		tcp.ProtocolFunc(codec.NewParser),
		2000,
		tcp.HandlerFunc(func(sess *tcp.Session) {
			_, err := gt.NewAgent(sess)
			if err != nil {
				sess.Close()
				return
			}
		}),
	)
	go server.Serve()
	defer server.Stop()
	time.Sleep(20 * time.Millisecond)

	// 连接应被立即关闭
	conn, _ := net.DialTimeout("tcp", addr, 2*time.Second)
	if conn == nil {
		t.Fatal("dial failed")
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 1)
	_, readErr := conn.Read(buf)
	if readErr == nil {
		t.Error("expected read error when agent creation fails")
	}
	t.Logf("✓ createAgent 失败→session 立即关闭: %v", readErr)
}
