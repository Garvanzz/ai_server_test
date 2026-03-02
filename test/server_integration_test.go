package tcpgate_test

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"xfx/pkg/agent"
	"xfx/pkg/gate/tcpgate/codec"
	"xfx/pkg/log"
	"xfx/pkg/net/tcp"
	proto_id "xfx/proto"
	"xfx/proto/proto_player"
	proto_public "xfx/proto/proto_public"

	gogoproto "github.com/gogo/protobuf/proto"
)

var initLogOnce sync.Once

func init() {
	initLogOnce.Do(func() {
		log.DefaultInit()
	})
}

// ═══════════════════════════════════════════════════════════════════
// 内部消息类型
// ═══════════════════════════════════════════════════════════════════

type iSessionMsg struct{ msg any }

type iLoginReq struct {
	Token   string
	Session agent.PID
}

type iLoginResp struct {
	PlayerId  int64
	PlayerPid agent.PID
	Ok        bool
}

type iLoginSuccessNotify struct{ Session agent.PID }

type iSessionReplace struct{ Session agent.PID }

type iPlayerDisconnect struct{}
type iPlayerLogout struct{}

// ═══════════════════════════════════════════════════════════════════
// Integration Gate Agent — 用真实 actor 系统的网关代理
// ═══════════════════════════════════════════════════════════════════

type iGateAgent struct {
	srv       *iServer
	sess      *tcp.Session
	ctx       agent.Context
	startedCh chan struct{}
	closeOnce sync.Once

	playerId  int64
	playerPid agent.PID
	pingTime  time.Duration
}

func (a *iGateAgent) OnStart(ctx agent.Context) {
	a.ctx = ctx
	a.pingTime = a.srv.pingTimeout
	close(a.startedCh)
}

func (a *iGateAgent) OnStop() {
	a.sess.Close()
	if a.playerId != 0 && a.srv.loginPid != nil {
		a.ctx.Cast(a.srv.loginPid, &iPlayerDisconnect{})
	}
}

func (a *iGateAgent) OnTerminated(_ agent.PID, _ int) {}

func (a *iGateAgent) OnTick(delta time.Duration) {
	a.pingTime -= delta
	if a.pingTime <= 0 {
		a.sess.Close()
	}
}

func (a *iGateAgent) OnMessage(msg any) any {
	switch m := msg.(type) {
	case *iSessionMsg:
		a.onSessionMessage(m.msg)
	case *proto_player.S2CKick:
		a.playerId = 0
		a.playerPid = nil
		a.sess.Send(m)
		go a.sess.CloseWithFlush(500 * time.Millisecond)
	default:
		a.sess.Send(m)
	}
	return nil
}

func (a *iGateAgent) onSessionMessage(msg any) {
	switch m := msg.(type) {
	case *proto_player.C2SLogin:
		result, err := a.ctx.Call(a.srv.loginPid, &iLoginReq{
			Token:   m.Token,
			Session: a.ctx.Self(),
		})
		if err != nil || result == nil {
			a.sess.Send(&proto_player.S2CLogin{State: proto_public.CommonState_Faild})
			return
		}
		resp := result.(*iLoginResp)
		if !resp.Ok {
			a.sess.Send(&proto_player.S2CLogin{State: proto_public.CommonState_Faild})
			return
		}
		a.playerId = resp.PlayerId
		a.playerPid = resp.PlayerPid
		a.ctx.Cast(resp.PlayerPid, &iLoginSuccessNotify{Session: a.ctx.Self()})

	case *proto_player.C2SPing:
		a.pingTime = a.srv.pingTimeout
		a.sess.Send(&proto_player.S2CPong{ZoneOffset: time.Now().Unix()})

	case *proto_player.C2SLogout:
		if a.playerId != 0 && a.srv.loginPid != nil {
			a.ctx.Cast(a.srv.loginPid, &iPlayerLogout{})
		}

	default:
		if a.playerPid != nil {
			a.ctx.Cast(a.playerPid, msg)
		} else {
			log.Error("integ gate: no player for msg %T", msg)
		}
	}
}

func (a *iGateAgent) onRecv(msg any) {
	select {
	case <-a.startedCh:
	case <-time.After(3 * time.Second):
		a.sess.Close()
		return
	}
	if a.ctx == nil {
		return
	}
	a.ctx.Cast(a.ctx.Self(), &iSessionMsg{msg: msg})
}

func (a *iGateAgent) closeAgent() {
	a.closeOnce.Do(func() {
		if a.ctx != nil {
			a.ctx.Stop()
		}
	})
}

// ═══════════════════════════════════════════════════════════════════
// Integration Player Actor — 模拟玩家 Actor
// ═══════════════════════════════════════════════════════════════════

type iPlayerActor struct {
	playerId int64
	session  agent.PID
	ctx      agent.Context
}

func (p *iPlayerActor) OnStart(ctx agent.Context) { p.ctx = ctx }
func (p *iPlayerActor) OnStop()                   {}
func (p *iPlayerActor) OnTerminated(_ agent.PID, _ int) {}
func (p *iPlayerActor) OnTick(_ time.Duration)    {}

func (p *iPlayerActor) OnMessage(msg any) any {
	switch m := msg.(type) {
	case *iLoginSuccessNotify:
		p.session = m.Session
		p.ctx.Cast(p.session, &proto_player.S2CLogin{
			State:      proto_public.CommonState_Success,
			ZoneOffset: time.Now().Unix(),
		})

	case *iSessionReplace:
		if p.session != nil {
			p.ctx.Cast(p.session, &proto_player.S2CKick{})
		}
		p.session = m.Session
		p.ctx.Cast(p.session, &proto_player.S2CLogin{
			State:      proto_public.CommonState_Success,
			ZoneOffset: time.Now().Unix(),
		})
		return "ok"

	case *proto_player.C2SChangeName:
		if p.session != nil {
			p.ctx.Cast(p.session, &proto_player.S2CChangeName{})
		}

	case *iPlayerDisconnect, *iPlayerLogout:
		return "ok"
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════
// Integration Login Actor — 模拟登录模块
// ═══════════════════════════════════════════════════════════════════

type iLoginActor struct {
	ctx     agent.Context
	tokens  map[string]int64
	players map[int64]agent.PID
	mu      sync.Mutex
}

func (l *iLoginActor) OnStart(ctx agent.Context) { l.ctx = ctx }
func (l *iLoginActor) OnStop()                   {}
func (l *iLoginActor) OnTerminated(_ agent.PID, _ int) {}
func (l *iLoginActor) OnTick(_ time.Duration)    {}

func (l *iLoginActor) OnMessage(msg any) any {
	switch m := msg.(type) {
	case *iLoginReq:
		l.mu.Lock()
		defer l.mu.Unlock()

		playerId, ok := l.tokens[m.Token]
		if !ok {
			return &iLoginResp{Ok: false}
		}

		if existingPid, online := l.players[playerId]; online {
			l.ctx.Call(existingPid, &iSessionReplace{Session: m.Session})
			return &iLoginResp{
				PlayerId:  playerId,
				PlayerPid: existingPid,
				Ok:        true,
			}
		}

		player := &iPlayerActor{playerId: playerId}
		pid, err := l.ctx.Create(
			fmt.Sprintf("player#%d", playerId),
			player,
		)
		if err != nil {
			return &iLoginResp{Ok: false}
		}
		l.players[playerId] = pid
		return &iLoginResp{
			PlayerId:  playerId,
			PlayerPid: pid,
			Ok:        true,
		}

	case *iPlayerDisconnect:
		l.mu.Lock()
		defer l.mu.Unlock()
		for id, pid := range l.players {
			l.ctx.Call(pid, &iPlayerDisconnect{})
			delete(l.players, id)
			break
		}

	case *iPlayerLogout:
		l.mu.Lock()
		defer l.mu.Unlock()
		for id, pid := range l.players {
			l.ctx.Call(pid, &iPlayerLogout{})
			delete(l.players, id)
			break
		}
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════
// Integration Server — 完整的测试服务器
// ═══════════════════════════════════════════════════════════════════

var iSysCounter int64

type iServer struct {
	system      *agent.System
	tcpServer   *tcp.Server
	addr        string
	loginPid    agent.PID
	agentCh     chan *iGateAgent
	pingTimeout time.Duration
	useTick     bool
	tickInterval time.Duration
}

type iServerOpt func(*iServer)

func withPingTimeout(d time.Duration) iServerOpt {
	return func(s *iServer) { s.pingTimeout = d }
}

func withTick(interval time.Duration) iServerOpt {
	return func(s *iServer) {
		s.useTick = true
		s.tickInterval = interval
	}
}

func newIServer(t *testing.T, tokens map[string]int64, opts ...iServerOpt) *iServer {
	t.Helper()

	sysName := fmt.Sprintf("isrv_%d", atomic.AddInt64(&iSysCounter, 1))
	sys := agent.NewSystem(agent.WithName(sysName))
	sys.Start()

	srv := &iServer{
		system:      sys,
		agentCh:     make(chan *iGateAgent, 128),
		pingTimeout: 30 * time.Second,
	}
	for _, o := range opts {
		o(srv)
	}

	loginActor := &iLoginActor{
		tokens:  tokens,
		players: make(map[int64]agent.PID),
	}
	loginPid, err := sys.Create("mock-login", loginActor)
	if err != nil {
		t.Fatal(err)
	}
	srv.loginPid = loginPid

	time.Sleep(50 * time.Millisecond)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv.addr = listener.Addr().String()

	srv.tcpServer = tcp.NewServer(
		listener,
		tcp.ProtocolFunc(codec.NewParser),
		2000,
		tcp.HandlerFunc(func(sess *tcp.Session) {
			ga := &iGateAgent{
				srv:       srv,
				sess:      sess,
				startedCh: make(chan struct{}),
			}

			var createOpts []agent.Option
			if srv.useTick {
				createOpts = append(createOpts, agent.WithTick(srv.tickInterval))
			}

			name := fmt.Sprintf("session#%d", sess.ID())
			_, err := sys.Create(name, ga, createOpts...)
			if err != nil {
				sess.Close()
				return
			}

			sess.AddCloseCallback(srv, sess.ID(), func() {
				ga.closeAgent()
			})

			srv.agentCh <- ga

			for {
				msg, err := sess.Receive()
				if err != nil {
					return
				}
				ga.onRecv(msg)
			}
		}),
	)
	go srv.tcpServer.Serve()
	time.Sleep(30 * time.Millisecond)
	return srv
}

func (s *iServer) stop() {
	s.tcpServer.Stop()
	time.Sleep(50 * time.Millisecond)
	s.system.Stop()
}

func (s *iServer) waitAgent(t *testing.T, timeout time.Duration) *iGateAgent {
	t.Helper()
	select {
	case a := <-s.agentCh:
		return a
	case <-time.After(timeout):
		t.Fatal("timeout waiting for gate agent")
		return nil
	}
}

// ═══════════════════════════════════════════════════════════════════
// 客户端辅助
// ═══════════════════════════════════════════════════════════════════

func dialInteg(t *testing.T, addr string) *testClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial %s failed: %v", addr, err)
	}
	return &testClient{conn: conn, t: t}
}

func integRecvWithTimeout(c *testClient, timeout time.Duration) (gogoproto.Message, error) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	msg, err := c.recvMsg()
	c.conn.SetReadDeadline(time.Time{})
	return msg, err
}

func integLogin(t *testing.T, c *testClient, token string) *proto_player.S2CLogin {
	t.Helper()
	c.sendMsg(&proto_player.C2SLogin{Token: token})
	msg, err := integRecvWithTimeout(c, 5*time.Second)
	if err != nil {
		t.Fatalf("login recv failed: %v", err)
	}
	resp, ok := msg.(*proto_player.S2CLogin)
	if !ok {
		t.Fatalf("expected S2CLogin, got %T", msg)
	}
	return resp
}

// ═══════════════════════════════════════════════════════════════════
// 正常流程测试
// ═══════════════════════════════════════════════════════════════════

func TestIntegLoginSuccess(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	resp := integLogin(t, c, "token_a")
	if resp.State != proto_public.CommonState_Success {
		t.Fatalf("expected Success, got %v", resp.State)
	}
}

func TestIntegLoginFailed(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	resp := integLogin(t, c, "bad_token")
	if resp.State != proto_public.CommonState_Faild {
		t.Fatalf("expected Faild, got %v", resp.State)
	}
}

func TestIntegPingPong(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	resp := integLogin(t, c, "token_a")
	if resp.State != proto_public.CommonState_Success {
		t.Fatal("login failed")
	}

	c.sendMsg(&proto_player.C2SPing{})
	msg, err := integRecvWithTimeout(c, 3*time.Second)
	if err != nil {
		t.Fatalf("recv pong failed: %v", err)
	}
	pong, ok := msg.(*proto_player.S2CPong)
	if !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}
	if pong.ZoneOffset == 0 {
		t.Fatal("ZoneOffset should not be 0")
	}
}

func TestIntegPingPongMultiple(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh
	integLogin(t, c, "token_a")

	for i := 0; i < 10; i++ {
		c.sendMsg(&proto_player.C2SPing{})
		msg, err := integRecvWithTimeout(c, 3*time.Second)
		if err != nil {
			t.Fatalf("ping %d failed: %v", i, err)
		}
		if _, ok := msg.(*proto_player.S2CPong); !ok {
			t.Fatalf("ping %d: expected S2CPong, got %T", i, msg)
		}
	}
}

func TestIntegPlayerInteraction(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	resp := integLogin(t, c, "token_a")
	if resp.State != proto_public.CommonState_Success {
		t.Fatal("login failed")
	}

	c.sendMsg(&proto_player.C2SChangeName{Name: "newname"})
	msg, err := integRecvWithTimeout(c, 3*time.Second)
	if err != nil {
		t.Fatalf("recv change name resp failed: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CChangeName); !ok {
		t.Fatalf("expected S2CChangeName, got %T", msg)
	}
}

func TestIntegLogout(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	integLogin(t, c, "token_a")
	c.sendMsg(&proto_player.C2SLogout{})
	time.Sleep(200 * time.Millisecond)
}

func TestIntegMultiplePlayersLogin(t *testing.T) {
	tokens := map[string]int64{
		"t1": 2001, "t2": 2002, "t3": 2003, "t4": 2004, "t5": 2005,
	}
	srv := newIServer(t, tokens)
	defer srv.stop()

	type result struct {
		idx   int
		state proto_public.CommonState
		err   error
	}
	resultCh := make(chan result, len(tokens))

	var wg sync.WaitGroup
	i := 0
	for token := range tokens {
		wg.Add(1)
		go func(idx int, tk string) {
			defer wg.Done()
			c := dialInteg(t, srv.addr)
			defer c.close()

			ga := srv.waitAgent(t, 3*time.Second)
			<-ga.startedCh

			resp := integLogin(t, c, tk)
			resultCh <- result{idx: idx, state: resp.State}
		}(i, token)
		i++
	}

	wg.Wait()
	close(resultCh)

	for r := range resultCh {
		if r.state != proto_public.CommonState_Success {
			t.Errorf("player %d login failed: %v", r.idx, r.state)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════
// 连接断开场景
// ═══════════════════════════════════════════════════════════════════

func TestIntegClientDisconnectBeforeLogin(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	c.close()
	time.Sleep(200 * time.Millisecond)
}

func TestIntegClientDisconnectAfterLogin(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	integLogin(t, c, "token_a")
	c.close()
	time.Sleep(300 * time.Millisecond)
}

func TestIntegConnectNoData(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	time.Sleep(50 * time.Millisecond)
	c.close()
	time.Sleep(200 * time.Millisecond)
}

func TestIntegConnectSendPartialThenClose(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	srv.waitAgent(t, 3*time.Second)

	c.rawWriteUint32(100)
	c.rawWriteBytes(make([]byte, 5))
	c.close()
	time.Sleep(200 * time.Millisecond)
}

func TestIntegDoubleLogin(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 3001})
	defer srv.stop()

	c1 := dialInteg(t, srv.addr)
	defer c1.close()
	ga1 := srv.waitAgent(t, 3*time.Second)
	<-ga1.startedCh

	resp1 := integLogin(t, c1, "token_a")
	if resp1.State != proto_public.CommonState_Success {
		t.Fatal("first login failed")
	}

	c2 := dialInteg(t, srv.addr)
	defer c2.close()
	ga2 := srv.waitAgent(t, 3*time.Second)
	<-ga2.startedCh

	c1.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	kickCh := make(chan gogoproto.Message, 1)
	go func() {
		msg, _ := c1.recvMsg()
		kickCh <- msg
	}()

	resp2 := integLogin(t, c2, "token_a")
	if resp2.State != proto_public.CommonState_Success {
		t.Fatal("second login failed")
	}

	select {
	case msg := <-kickCh:
		if _, ok := msg.(*proto_player.S2CKick); !ok {
			t.Fatalf("expected S2CKick for old client, got %T", msg)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for kick on old client")
	}
}

// ═══════════════════════════════════════════════════════════════════
// 协议错误场景
// ═══════════════════════════════════════════════════════════════════

func TestIntegUnknownProtoID(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	srv.waitAgent(t, 3*time.Second)

	const unknownID uint32 = 0xDEADBEEF
	c.rawSend(unknownID, []byte{0x01, 0x02})
	time.Sleep(200 * time.Millisecond)
}

func TestIntegOversizedMessage(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	srv.waitAgent(t, 3*time.Second)

	c.rawWriteUint32(200 * 1024) // 200KB > MaxSize(128KB)
	time.Sleep(200 * time.Millisecond)
}

func TestIntegMalformedProtobuf(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	srv.waitAgent(t, 3*time.Second)

	loginID, _ := proto_id.Router.MessageID(&proto_player.C2SLogin{})
	garbage := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	c.rawSend(loginID, garbage)
	time.Sleep(200 * time.Millisecond)
}

func TestIntegZeroLengthBody(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	pingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})
	c.rawSend(pingID, []byte{})
	msg, err := integRecvWithTimeout(c, 3*time.Second)
	if err != nil {
		t.Fatalf("recv failed: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}
}

func TestIntegZeroProtoID(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	srv.waitAgent(t, 3*time.Second)

	c.rawSend(0, []byte{})
	time.Sleep(200 * time.Millisecond)
}

func TestIntegMultipleProtocolErrorsSequential(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	for i := 0; i < 5; i++ {
		c := dialInteg(t, srv.addr)
		srv.waitAgent(t, 3*time.Second)

		c.rawSend(0xDEADBEEF, []byte{0x01})
		time.Sleep(100 * time.Millisecond)
		c.close()
		time.Sleep(50 * time.Millisecond)
	}
}

// ═══════════════════════════════════════════════════════════════════
// Ping 超时测试
// ═══════════════════════════════════════════════════════════════════

func TestIntegPingTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test")
	}

	srv := newIServer(t, map[string]int64{"token_a": 1001},
		withPingTimeout(600*time.Millisecond),
		withTick(100*time.Millisecond),
	)
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	integLogin(t, c, "token_a")

	c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err := c.recvMsg()
	if err == nil {
		t.Fatal("expected connection to be closed due to ping timeout")
	}
}

func TestIntegPingKeepsAlive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test")
	}

	srv := newIServer(t, map[string]int64{"token_a": 1001},
		withPingTimeout(400*time.Millisecond),
		withTick(100*time.Millisecond),
	)
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh
	integLogin(t, c, "token_a")

	for i := 0; i < 5; i++ {
		time.Sleep(200 * time.Millisecond)
		c.sendMsg(&proto_player.C2SPing{})
		msg, err := integRecvWithTimeout(c, 2*time.Second)
		if err != nil {
			t.Fatalf("ping %d failed: %v", i, err)
		}
		if _, ok := msg.(*proto_player.S2CPong); !ok {
			t.Fatalf("ping %d: expected S2CPong, got %T", i, msg)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════
// 稳定性与压力测试
// ═══════════════════════════════════════════════════════════════════

func TestIntegRapidConnectDisconnect(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	const iterations = 50
	for i := 0; i < iterations; i++ {
		c := dialInteg(t, srv.addr)
		if i%3 == 0 {
			pingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})
			c.rawSend(pingID, []byte{})
		}
		c.close()
	}
	time.Sleep(500 * time.Millisecond)
}

func TestIntegRapidLoginDisconnect(t *testing.T) {
	tokens := make(map[string]int64)
	for i := 0; i < 20; i++ {
		tokens[fmt.Sprintf("rapid_%d", i)] = int64(5000 + i)
	}
	srv := newIServer(t, tokens)
	defer srv.stop()

	var wg sync.WaitGroup
	for token := range tokens {
		wg.Add(1)
		go func(tk string) {
			defer wg.Done()
			c := dialInteg(t, srv.addr)
			ga := srv.waitAgent(t, 3*time.Second)
			<-ga.startedCh

			c.sendMsg(&proto_player.C2SLogin{Token: tk})
			integRecvWithTimeout(c, 3*time.Second)
			c.close()
		}(token)
	}
	wg.Wait()
	time.Sleep(300 * time.Millisecond)
}

func TestIntegConcurrentConnections(t *testing.T) {
	tokens := make(map[string]int64)
	for i := 0; i < 30; i++ {
		tokens[fmt.Sprintf("conc_%d", i)] = int64(6000 + i)
	}
	srv := newIServer(t, tokens)
	defer srv.stop()

	var (
		wg         sync.WaitGroup
		successCnt int64
		failCnt    int64
	)

	i := 0
	for token := range tokens {
		wg.Add(1)
		go func(idx int, tk string) {
			defer wg.Done()
			c := dialInteg(t, srv.addr)
			defer c.close()

			ga := srv.waitAgent(t, 5*time.Second)
			<-ga.startedCh

			resp := integLogin(t, c, tk)
			if resp.State == proto_public.CommonState_Success {
				atomic.AddInt64(&successCnt, 1)

				c.sendMsg(&proto_player.C2SPing{})
				msg, err := integRecvWithTimeout(c, 3*time.Second)
				if err == nil {
					if _, ok := msg.(*proto_player.S2CPong); !ok {
						atomic.AddInt64(&failCnt, 1)
					}
				}
			} else {
				atomic.AddInt64(&failCnt, 1)
			}
		}(i, token)
		i++
	}

	wg.Wait()

	s := atomic.LoadInt64(&successCnt)
	f := atomic.LoadInt64(&failCnt)
	t.Logf("concurrent connections: %d success, %d fail out of %d", s, f, len(tokens))
	if f > 0 {
		t.Errorf("unexpected failures: %d", f)
	}
}

func TestIntegMessageFlood(t *testing.T) {
	srv := newIServer(t, map[string]int64{"flood": 7001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	integLogin(t, c, "flood")

	const count = 100
	go func() {
		for i := 0; i < count; i++ {
			c.sendMsg(&proto_player.C2SPing{})
		}
	}()

	received := 0
	deadline := time.After(10 * time.Second)
	for received < count {
		c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, err := c.recvMsg()
		if err != nil {
			break
		}
		received++

		select {
		case <-deadline:
			t.Fatalf("timeout: received %d/%d", received, count)
			return
		default:
		}
	}
	c.conn.SetReadDeadline(time.Time{})
	t.Logf("message flood: received %d/%d pongs", received, count)
	if received < count/2 {
		t.Errorf("too few responses: %d/%d", received, count)
	}
}

func TestIntegProtocolErrorDoesNotAffectOthers(t *testing.T) {
	srv := newIServer(t, map[string]int64{"good": 8001})
	defer srv.stop()

	goodClient := dialInteg(t, srv.addr)
	defer goodClient.close()
	gaGood := srv.waitAgent(t, 3*time.Second)
	<-gaGood.startedCh

	resp := integLogin(t, goodClient, "good")
	if resp.State != proto_public.CommonState_Success {
		t.Fatal("good client login failed")
	}

	for i := 0; i < 5; i++ {
		bad := dialInteg(t, srv.addr)
		srv.waitAgent(t, 3*time.Second)
		bad.rawSend(0xDEADBEEF, []byte{0xFF})
		time.Sleep(50 * time.Millisecond)
		bad.close()
		time.Sleep(50 * time.Millisecond)
	}

	goodClient.sendMsg(&proto_player.C2SPing{})
	msg, err := integRecvWithTimeout(goodClient, 3*time.Second)
	if err != nil {
		t.Fatalf("good client ping after bad clients failed: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}
}

func TestIntegMixedTraffic(t *testing.T) {
	tokens := map[string]int64{
		"mix_1": 9001, "mix_2": 9002, "mix_3": 9003,
	}
	srv := newIServer(t, tokens)
	defer srv.stop()

	var wg sync.WaitGroup

	for token := range tokens {
		wg.Add(1)
		go func(tk string) {
			defer wg.Done()
			c := dialInteg(t, srv.addr)
			defer c.close()

			ga := srv.waitAgent(t, 3*time.Second)
			<-ga.startedCh

			resp := integLogin(t, c, tk)
			if resp.State != proto_public.CommonState_Success {
				return
			}

			for j := 0; j < 5; j++ {
				c.sendMsg(&proto_player.C2SPing{})
				integRecvWithTimeout(c, 3*time.Second)
			}

			c.sendMsg(&proto_player.C2SChangeName{Name: "test"})
			integRecvWithTimeout(c, 3*time.Second)
		}(token)
	}

	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			defer wg.Done()
			c := dialInteg(t, srv.addr)
			srv.waitAgent(t, 3*time.Second)
			c.rawSend(0xBAD, []byte{0xFF})
			time.Sleep(50 * time.Millisecond)
			c.close()
		}()
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)
}

func TestIntegSendAfterLoginThenImmediateDisconnect(t *testing.T) {
	srv := newIServer(t, map[string]int64{"quick": 10001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	c.sendMsg(&proto_player.C2SLogin{Token: "quick"})
	c.sendMsg(&proto_player.C2SPing{})
	c.sendMsg(&proto_player.C2SChangeName{Name: "x"})
	c.close()
	time.Sleep(300 * time.Millisecond)
}

func TestIntegHalfOpenConnections(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	conns := make([]net.Conn, 10)
	for i := range conns {
		conn, err := net.DialTimeout("tcp", srv.addr, 2*time.Second)
		if err != nil {
			t.Fatalf("dial %d failed: %v", i, err)
		}
		conns[i] = conn
	}

	time.Sleep(100 * time.Millisecond)
	for _, conn := range conns {
		conn.Close()
	}
	time.Sleep(200 * time.Millisecond)
}

func TestIntegRandomGarbage(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	for i := 0; i < 10; i++ {
		c := dialInteg(t, srv.addr)
		garbage := make([]byte, rand.Intn(200)+1)
		rand.Read(garbage)
		c.rawWriteBytes(garbage)
		time.Sleep(30 * time.Millisecond)
		c.close()
		time.Sleep(30 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
}

func TestIntegLargeTokenLogin(t *testing.T) {
	bigToken := make([]byte, 4096)
	for i := range bigToken {
		bigToken[i] = 'A'
	}
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	resp := integLogin(t, c, string(bigToken))
	if resp.State != proto_public.CommonState_Faild {
		t.Fatalf("expected Faild for unknown large token, got %v", resp.State)
	}
}

func TestIntegPingBeforeLogin(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	c.sendMsg(&proto_player.C2SPing{})
	msg, err := integRecvWithTimeout(c, 3*time.Second)
	if err != nil {
		t.Fatalf("recv pong before login failed: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}
}

func TestIntegSendBusinessMsgBeforeLogin(t *testing.T) {
	srv := newIServer(t, map[string]int64{})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	c.sendMsg(&proto_player.C2SChangeName{Name: "test"})

	c.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, err := c.recvMsg()
	c.conn.SetReadDeadline(time.Time{})
	if err == nil {
		t.Fatal("should not receive response for business msg before login")
	}
}

// ═══════════════════════════════════════════════════════════════════
// 写入辅助——构造原始二进制帧（避免与 testClient 上已有方法冲突）
// ═══════════════════════════════════════════════════════════════════

func integRawFrame(protoId uint32, body []byte) []byte {
	buf := make([]byte, 4+4+len(body))
	binary.LittleEndian.PutUint32(buf[0:4], uint32(len(body)))
	binary.LittleEndian.PutUint32(buf[4:8], protoId)
	copy(buf[8:], body)
	return buf
}

func TestIntegFragmentedSend(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	pingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})
	frame := integRawFrame(pingID, []byte{})

	for _, b := range frame {
		c.conn.Write([]byte{b})
		time.Sleep(5 * time.Millisecond)
	}

	msg, err := integRecvWithTimeout(c, 3*time.Second)
	if err != nil {
		t.Fatalf("fragmented send recv failed: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}
}

func TestIntegBackToBackFrames(t *testing.T) {
	srv := newIServer(t, map[string]int64{"token_a": 1001})
	defer srv.stop()

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	pingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})
	var combined []byte
	for i := 0; i < 5; i++ {
		combined = append(combined, integRawFrame(pingID, []byte{})...)
	}
	c.conn.Write(combined)

	for i := 0; i < 5; i++ {
		msg, err := integRecvWithTimeout(c, 3*time.Second)
		if err != nil {
			t.Fatalf("back-to-back frame %d recv failed: %v", i, err)
		}
		if _, ok := msg.(*proto_player.S2CPong); !ok {
			t.Fatalf("frame %d: expected S2CPong, got %T", i, msg)
		}
	}
}

func TestIntegServerStopWhileClientsConnected(t *testing.T) {
	srv := newIServer(t, map[string]int64{"stop_test": 11001})

	c := dialInteg(t, srv.addr)
	defer c.close()
	ga := srv.waitAgent(t, 3*time.Second)
	<-ga.startedCh

	integLogin(t, c, "stop_test")

	srv.stop()

	c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err := c.recvMsg()
	if err == nil {
		t.Fatal("expected error after server stop")
	}
}
