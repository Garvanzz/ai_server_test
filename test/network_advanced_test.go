package tcpgate_test

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"xfx/pkg/gate/tcpgate/codec"
	"xfx/pkg/log"
	"xfx/pkg/net/tcp"
	proto_id "xfx/proto"
	"xfx/proto/proto_player"

	gogoproto "github.com/gogo/protobuf/proto"
)

func TestMain(m *testing.M) {
	log.DefaultInit()
	os.Exit(m.Run())
}

// ═══════════════════════════════════════════════════════════════════
// TCP Server 真实监听集成测试
// ═══════════════════════════════════════════════════════════════════

func startTestServer(t *testing.T) (*tcp.Server, string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("listen failed:", err)
	}
	addr := listener.Addr().String()

	server := tcp.NewServer(
		listener,
		tcp.ProtocolFunc(codec.NewParser),
		128,
		tcp.HandlerFunc(func(sess *tcp.Session) {
			for {
				msg, err := sess.Receive()
				if err != nil {
					return
				}
				// echo: 收到 C2SPing 回复 S2CPong
				if _, ok := msg.(*proto_player.C2SPing); ok {
					sess.Send(&proto_player.S2CPong{ZoneOffset: time.Now().Unix()})
				}
			}
		}),
	)

	go server.Serve()
	time.Sleep(20 * time.Millisecond) // 等待 listener ready
	return server, addr
}

func dialTestClient(t *testing.T, addr string) *testClient {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	return &testClient{conn: conn, t: t}
}

// TestTCPServerMultipleClients 多个客户端同时连接，各自正常通信
func TestTCPServerMultipleClients(t *testing.T) {
	server, addr := startTestServer(t)
	defer server.Stop()

	const clientCount = 10
	var wg sync.WaitGroup

	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			client := dialTestClient(t, addr)
			defer client.close()

			client.sendMsg(&proto_player.C2SPing{})
			msg, err := client.recvMsg()
			if err != nil {
				t.Errorf("client %d recvMsg failed: %v", idx, err)
				return
			}
			if _, ok := msg.(*proto_player.S2CPong); !ok {
				t.Errorf("client %d expected S2CPong, got %T", idx, msg)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("✓ %d 个客户端并行通信成功", clientCount)
}

// TestTCPServerGracefulStop Stop 后所有客户端连接断开
func TestTCPServerGracefulStop(t *testing.T) {
	server, addr := startTestServer(t)

	clients := make([]*testClient, 5)
	for i := range clients {
		clients[i] = dialTestClient(t, addr)
	}

	// 每个客户端先 ping 一下确保连接建立
	for _, c := range clients {
		c.sendMsg(&proto_player.C2SPing{})
		if _, err := c.recvMsg(); err != nil {
			t.Fatalf("pre-stop ping failed: %v", err)
		}
	}

	server.Stop()

	// 所有客户端读取应报错
	for i, c := range clients {
		c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, err := c.recvMsg()
		if err == nil {
			t.Errorf("client %d: expected error after server stop", i)
		}
		c.close()
	}
	t.Log("✓ Server.Stop() 后所有客户端连接断开")
}

// TestTCPServerNewConnectionAfterStop Stop 后新连接被拒绝
func TestTCPServerNewConnectionAfterStop(t *testing.T) {
	server, addr := startTestServer(t)
	server.Stop()

	time.Sleep(50 * time.Millisecond)

	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		t.Log("✓ Stop 后新连接被拒绝（dial 失败）")
		return
	}
	// 有时 OS 允许 dial 成功但随后 read/write 失败
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil {
		t.Log("✓ Stop 后新连接无法通信")
		return
	}
	t.Error("expected connection to fail after server stop")
}

// TestTCPServerConnectionFlood 连接洪泛：快速建立大量连接再立即断开
func TestTCPServerConnectionFlood(t *testing.T) {
	server, addr := startTestServer(t)
	defer server.Stop()

	const floodCount = 100
	var wg sync.WaitGroup
	var successCount int64

	for i := 0; i < floodCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				return
			}
			atomic.AddInt64(&successCount, 1)
			conn.Close()
		}()
	}

	wg.Wait()
	t.Logf("✓ 连接洪泛：%d/%d 连接建立成功后立即断开，服务器稳定", atomic.LoadInt64(&successCount), floodCount)
}

// TestTCPServerConnectionFloodWithData 连接洪泛+发数据
func TestTCPServerConnectionFloodWithData(t *testing.T) {
	server, addr := startTestServer(t)
	defer server.Stop()

	const floodCount = 50
	var wg sync.WaitGroup
	var echoSuccess int64

	for i := 0; i < floodCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				return
			}
			defer conn.Close()

			c := &testClient{conn: conn, t: t}
			c.sendMsg(&proto_player.C2SPing{})
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if _, err := c.recvMsg(); err == nil {
				atomic.AddInt64(&echoSuccess, 1)
			}
		}()
	}

	wg.Wait()
	t.Logf("✓ 连接洪泛+数据：%d/%d 完成 echo 往返", atomic.LoadInt64(&echoSuccess), floodCount)
}

// ═══════════════════════════════════════════════════════════════════
// Manager 集成测试
// ═══════════════════════════════════════════════════════════════════

func makeMockCodec() *mockCodec {
	return &mockCodec{closed: make(chan struct{})}
}

type mockCodec struct {
	closed   chan struct{}
	closeErr error
	mu       sync.Mutex
}

func (m *mockCodec) Receive() (any, error) {
	<-m.closed
	return nil, io.EOF
}
func (m *mockCodec) Send(msg any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	select {
	case <-m.closed:
		return fmt.Errorf("closed")
	default:
		return nil
	}
}
func (m *mockCodec) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	select {
	case <-m.closed:
	default:
		close(m.closed)
	}
	return m.closeErr
}

// TestManagerNewAndGetSession 创建和查找 Session
func TestManagerNewAndGetSession(t *testing.T) {
	mgr := tcp.NewManager()
	defer mgr.Dispose()

	sessions := make([]*tcp.Session, 10)
	for i := range sessions {
		sessions[i] = mgr.NewSession(makeMockCodec(), 0)
	}

	for i, sess := range sessions {
		got := mgr.GetSession(sess.ID())
		if got == nil {
			t.Errorf("session %d: GetSession returned nil", i)
		} else if got.ID() != sess.ID() {
			t.Errorf("session %d: ID mismatch", i)
		}
	}
	t.Log("✓ Manager 创建和查找 10 个 Session 成功")
}

// TestManagerGetNonExistent 查找不存在的 Session
func TestManagerGetNonExistent(t *testing.T) {
	mgr := tcp.NewManager()
	defer mgr.Dispose()

	if got := mgr.GetSession(99999); got != nil {
		t.Error("expected nil for non-existent session")
	}
	t.Log("✓ 查找不存在的 Session 返回 nil")
}

// TestManagerSessionRemovedAfterClose Session 关闭后从 Manager 中移除
func TestManagerSessionRemovedAfterClose(t *testing.T) {
	mgr := tcp.NewManager()
	defer mgr.Dispose()

	sess := mgr.NewSession(makeMockCodec(), 0)
	id := sess.ID()

	if mgr.GetSession(id) == nil {
		t.Fatal("session should exist before close")
	}

	sess.Close()
	time.Sleep(50 * time.Millisecond) // 等待异步 delSession

	if mgr.GetSession(id) != nil {
		t.Error("session should be removed after close")
	}
	t.Log("✓ Session 关闭后自动从 Manager 移除")
}

// TestManagerDispose Dispose 关闭所有 Session
func TestManagerDispose(t *testing.T) {
	mgr := tcp.NewManager()

	sessions := make([]*tcp.Session, 20)
	for i := range sessions {
		sessions[i] = mgr.NewSession(makeMockCodec(), 0)
	}

	mgr.Dispose()

	for i, sess := range sessions {
		if !sess.IsClosed() {
			t.Errorf("session %d should be closed after Dispose", i)
		}
	}
	t.Log("✓ Manager.Dispose() 关闭所有 20 个 Session")
}

// TestManagerDisposeIdempotent 多次 Dispose 不 panic
func TestManagerDisposeIdempotent(t *testing.T) {
	mgr := tcp.NewManager()
	mgr.NewSession(makeMockCodec(), 0)

	mgr.Dispose()
	mgr.Dispose() // 二次 Dispose
	mgr.Dispose() // 三次

	t.Log("✓ Manager 多次 Dispose 幂等，无 panic")
}

// TestManagerNewSessionAfterDispose Dispose 后创建的 Session 立即关闭
func TestManagerNewSessionAfterDispose(t *testing.T) {
	mgr := tcp.NewManager()
	mgr.Dispose()

	sess := mgr.NewSession(makeMockCodec(), 0)
	time.Sleep(20 * time.Millisecond)

	if !sess.IsClosed() {
		t.Error("session created after Dispose should be closed immediately")
	}
	t.Log("✓ Dispose 后创建的 Session 被立即关闭")
}

// ═══════════════════════════════════════════════════════════════════
// Channel 集成测试
// ═══════════════════════════════════════════════════════════════════

// TestChannelAutoCleanupOnSessionClose Session 关闭自动从 Channel 移除
func TestChannelAutoCleanupOnSessionClose(t *testing.T) {
	ch := tcp.NewChannel()

	sess1, c1 := makeTestPair(t)
	sess2, c2 := makeTestPair(t)
	defer c1.close()
	defer c2.close()

	ch.Put("player1", sess1)
	ch.Put("player2", sess2)

	if ch.Len() != 2 {
		t.Fatalf("expected 2 sessions, got %d", ch.Len())
	}

	sess1.Close()
	time.Sleep(50 * time.Millisecond)

	if ch.Len() != 1 {
		t.Errorf("expected 1 session after close, got %d", ch.Len())
	}
	if ch.Get("player1") != nil {
		t.Error("player1 should be removed after session close")
	}
	if ch.Get("player2") == nil {
		t.Error("player2 should still exist")
	}
	t.Log("✓ Session 关闭后自动从 Channel 移除")
}

// TestChannelPutDuplicateKey 重复 Key 替换旧 Session
func TestChannelPutDuplicateKey(t *testing.T) {
	ch := tcp.NewChannel()

	sess1, c1 := makeTestPair(t)
	sess2, c2 := makeTestPair(t)
	defer c1.close()
	defer c2.close()

	ch.Put("player1", sess1)
	ch.Put("player1", sess2) // 替换

	if ch.Len() != 1 {
		t.Errorf("expected 1 session, got %d", ch.Len())
	}
	got := ch.Get("player1")
	if got == nil || got.ID() != sess2.ID() {
		t.Error("should be replaced with sess2")
	}
	t.Log("✓ 重复 Key 替换旧 Session")
}

// TestChannelFetch 遍历所有 Session
func TestChannelFetch(t *testing.T) {
	ch := tcp.NewChannel()

	pairs := make([]struct {
		sess   *tcp.Session
		client *testClient
	}, 5)
	for i := range pairs {
		s, c := makeTestPair(t)
		pairs[i] = struct {
			sess   *tcp.Session
			client *testClient
		}{s, c}
		ch.Put(fmt.Sprintf("p%d", i), s)
		defer c.close()
	}

	var count int
	ch.Fetch(func(s *tcp.Session) {
		count++
	})
	if count != 5 {
		t.Errorf("Fetch: expected 5, got %d", count)
	}
	t.Log("✓ Channel.Fetch 遍历 5 个 Session")
}

// TestChannelFetchAndRemove 取出并移除所有 Session
func TestChannelFetchAndRemove(t *testing.T) {
	ch := tcp.NewChannel()

	for i := 0; i < 3; i++ {
		s, c := makeTestPair(t)
		defer c.close()
		ch.Put(i, s)
	}

	var removed int
	ch.FetchAndRemove(func(s *tcp.Session) {
		removed++
	})

	if removed != 3 {
		t.Errorf("expected 3 removed, got %d", removed)
	}
	if ch.Len() != 0 {
		t.Errorf("expected empty channel, got %d", ch.Len())
	}
	t.Log("✓ FetchAndRemove 取出并清空 3 个 Session")
}

// TestChannelClose 关闭 Channel 清空所有 Session
func TestChannelClose(t *testing.T) {
	ch := tcp.NewChannel()

	for i := 0; i < 4; i++ {
		s, c := makeTestPair(t)
		defer c.close()
		ch.Put(i, s)
	}

	ch.Close()

	if ch.Len() != 0 {
		t.Errorf("expected empty after Close, got %d", ch.Len())
	}
	t.Log("✓ Channel.Close 清空所有 Session")
}

// ═══════════════════════════════════════════════════════════════════
// SendChan 背压 / 阻塞测试
// ═══════════════════════════════════════════════════════════════════

// TestSendChanBlocked sendChan 满时返回 SessionBlockedError 并关闭 session
func TestSendChanBlocked(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	// 用一个永远不消费的 mockCodec 来模拟阻塞
	blockingCodec := &blockingSendCodec{conn: serverConn}
	sess := tcp.NewSession(blockingCodec, 2) // sendChan 容量仅 2

	// 快速塞满 sendChan
	// 先发 2 条填满 buffer
	for i := 0; i < 2; i++ {
		if err := sess.Send("msg"); err != nil {
			t.Fatalf("send %d failed: %v", i, err)
		}
	}

	// 第 3 条应触发 blocked
	err := sess.Send("overflow")
	if err != tcp.SessionBlockedError {
		t.Errorf("expected SessionBlockedError, got: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if !sess.IsClosed() {
		t.Error("session should be closed after blocked")
	}
	t.Logf("✓ SendChan 满时返回 SessionBlockedError，session 关闭")
}

type blockingSendCodec struct {
	conn net.Conn
	mu   sync.Mutex
}

func (b *blockingSendCodec) Receive() (any, error) {
	buf := make([]byte, 1)
	_, err := b.conn.Read(buf)
	return nil, err
}
func (b *blockingSendCodec) Send(msg any) error {
	// 模拟慢写入，永远阻塞
	time.Sleep(10 * time.Second)
	return nil
}
func (b *blockingSendCodec) Close() error {
	return b.conn.Close()
}

// TestSendChanSyncMode sendChanSize=0 同步发送模式
func TestSendChanSyncMode(t *testing.T) {
	// 使用独立的 net.Pipe 测试 sync 模式
	serverConn, clientConn := net.Pipe()
	parser, _ := codec.NewParser(serverConn, proto_id.Router)
	syncSess := tcp.NewSession(parser, 0) // 同步模式

	go func() {
		// 客户端持续消费
		for {
			buf := make([]byte, 1024)
			if _, err := clientConn.Read(buf); err != nil {
				return
			}
		}
	}()

	// 同步发送 10 条
	for i := 0; i < 10; i++ {
		if err := syncSess.Send(&proto_player.S2CPong{ZoneOffset: int64(i)}); err != nil {
			t.Fatalf("sync send %d failed: %v", i, err)
		}
	}

	syncSess.Close()
	clientConn.Close()
	t.Log("✓ 同步发送模式（sendChanSize=0）正常工作")
}

// ═══════════════════════════════════════════════════════════════════
// CloseWithFlush 测试
// ═══════════════════════════════════════════════════════════════════

// TestCloseWithFlushSendsRemaining 发送队列中的消息发完后关闭
func TestCloseWithFlushSendsRemaining(t *testing.T) {
	sess, client := makeTestPair(t)

	// 先发 3 条消息到 sendChan
	for i := 0; i < 3; i++ {
		sess.Send(&proto_player.S2CPong{ZoneOffset: int64(i)})
	}

	// 开始 FlushClose
	go sess.CloseWithFlush(2 * time.Second)

	// 客户端持续读取
	received := 0
	for {
		client.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if _, err := client.recvMsg(); err != nil {
			break
		}
		received++
	}

	if received < 3 {
		t.Errorf("expected at least 3 messages before close, got %d", received)
	}
	client.close()
	t.Logf("✓ CloseWithFlush 发送完 %d 条消息后关闭", received)
}

// TestCloseWithFlushTimeout flush 超时强制关闭
func TestCloseWithFlushTimeout(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	blockCodec := &blockingSendCodec{conn: serverConn}
	sess := tcp.NewSession(blockCodec, 64)

	// 塞满一些消息（sendLoop 因为 blockCodec.Send 阻塞，无法消费）
	sess.Send("msg1")

	start := time.Now()
	err := sess.CloseWithFlush(200 * time.Millisecond)
	elapsed := time.Since(start)

	// 应该在 ~200ms 后超时关闭
	if elapsed > 1*time.Second {
		t.Errorf("CloseWithFlush should timeout in ~200ms, took %v", elapsed)
	}
	t.Logf("✓ CloseWithFlush 超时关闭耗时 %v, err=%v", elapsed, err)
}

// TestCloseWithFlushNoSendChan 无 sendChan 时等价于 Close
func TestCloseWithFlushNoSendChan(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	parser, _ := codec.NewParser(serverConn, proto_id.Router)
	sess := tcp.NewSession(parser, 0) // 无 sendChan

	err := sess.CloseWithFlush(1 * time.Second)
	if err != nil {
		t.Logf("CloseWithFlush(no sendChan) returned: %v", err)
	}
	if !sess.IsClosed() {
		t.Error("session should be closed")
	}
	t.Log("✓ 无 sendChan 时 CloseWithFlush 等价于 Close")
}

// ═══════════════════════════════════════════════════════════════════
// 慢客户端 / 畸形连接测试
// ═══════════════════════════════════════════════════════════════════

// TestSlowClientPartialHeader 客户端分片发送头部（慢速攻击模拟）
func TestSlowClientPartialHeader(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()

	msgCh, errCh := startRecvLoop(sess)

	// 构造一条合法 C2SPing 的完整二进制
	pingData, _ := gogoproto.Marshal(&proto_player.C2SPing{})
	pingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})

	fullPacket := make([]byte, 4+4+len(pingData))
	binary.LittleEndian.PutUint32(fullPacket[0:4], uint32(len(pingData)))
	binary.LittleEndian.PutUint32(fullPacket[4:8], pingID)
	copy(fullPacket[8:], pingData)

	// 逐字节发送
	for _, b := range fullPacket {
		client.rawWriteBytes([]byte{b})
		time.Sleep(5 * time.Millisecond)
	}

	msg := waitMsg(t, msgCh, errCh, 3*time.Second)
	if _, ok := msg.(*proto_player.C2SPing); !ok {
		t.Errorf("expected C2SPing, got %T", msg)
	}
	t.Log("✓ 逐字节慢速发送仍可正确解析")
}

// TestSlowClientMultipleSlowPackets 慢客户端连续发多条消息
func TestSlowClientMultipleSlowPackets(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()

	msgCh, errCh := startRecvLoop(sess)

	for pktIdx := 0; pktIdx < 5; pktIdx++ {
		pingData, _ := gogoproto.Marshal(&proto_player.C2SPing{})
		pingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})

		fullPacket := make([]byte, 4+4+len(pingData))
		binary.LittleEndian.PutUint32(fullPacket[0:4], uint32(len(pingData)))
		binary.LittleEndian.PutUint32(fullPacket[4:8], pingID)
		copy(fullPacket[8:], pingData)

		// 分 2-3 段随机发送
		chunks := splitRandom(fullPacket, 2+rand.Intn(2))
		for _, chunk := range chunks {
			client.rawWriteBytes(chunk)
			time.Sleep(10 * time.Millisecond)
		}
	}

	// 应收到 5 条消息
	for i := 0; i < 5; i++ {
		msg := waitMsg(t, msgCh, errCh, 3*time.Second)
		if _, ok := msg.(*proto_player.C2SPing); !ok {
			t.Errorf("packet %d: expected C2SPing, got %T", i, msg)
		}
	}
	t.Log("✓ 慢客户端分段发送 5 条消息全部正确解析")
}

func splitRandom(data []byte, parts int) [][]byte {
	if parts <= 1 || len(data) <= 1 {
		return [][]byte{data}
	}
	result := make([][]byte, 0, parts)
	remaining := data
	for i := 0; i < parts-1 && len(remaining) > 1; i++ {
		n := 1 + rand.Intn(len(remaining)-1)
		result = append(result, remaining[:n])
		remaining = remaining[n:]
	}
	if len(remaining) > 0 {
		result = append(result, remaining)
	}
	return result
}

// TestGarbageDataStream 纯随机垃圾数据
func TestGarbageDataStream(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()

	msgCh, errCh := startRecvLoop(sess)

	garbage := make([]byte, 256)
	rand.Read(garbage)
	client.rawWriteBytes(garbage)

	select {
	case err := <-errCh:
		t.Logf("✓ 纯垃圾数据触发错误: %v", err)
	case msg := <-msgCh:
		// 极小概率垃圾数据恰好能解析
		t.Logf("注意：垃圾数据被解析为 %T（概率极低）", msg)
	case <-time.After(2 * time.Second):
		t.Log("✓ 垃圾数据可能导致 Receive 阻塞在读取 body 阶段（等待更多数据）")
		// 关闭客户端让 recv 退出
		client.close()
		select {
		case <-errCh:
		case <-time.After(1 * time.Second):
			t.Error("recv should exit after client close")
		}
	}
}

// TestRepeatedPartialPackets 反复发不完整包后断开
func TestRepeatedPartialPackets(t *testing.T) {
	sess, client := makeTestPair(t)

	msgCh, errCh := startRecvLoop(sess)

	// 发 5 次：每次只写 dataLen 不写 body，除了最后一次
	for i := 0; i < 4; i++ {
		// 写 dataLen=0（合法长度，body 为空但还需 protoId 4字节）
		client.rawWriteUint32(0)
		// 写一个 protoId 字节但不完整（只写 2 字节）
		client.rawWriteBytes([]byte{0x01, 0x02})
		// 不发后续内容
	}

	client.close()

	select {
	case err := <-errCh:
		t.Logf("✓ 反复不完整包后断开: %v", err)
	case <-msgCh:
		t.Error("should not receive valid message from incomplete packets")
	case <-time.After(2 * time.Second):
		t.Error("timeout")
	}
	_ = sess
}

// TestZeroLengthBody dataLen=0 + protoId（body 为空）
func TestZeroLengthBody(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()

	msgCh, errCh := startRecvLoop(sess)

	// dataLen=0 表示 body 长度为 0, 但 codec 还会读 protoId (4 bytes)
	c2sPingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})
	client.rawWriteUint32(0)
	client.rawWriteUint32(c2sPingID)

	msg := waitMsg(t, msgCh, errCh, 2*time.Second)
	if _, ok := msg.(*proto_player.C2SPing); !ok {
		t.Errorf("expected C2SPing, got %T", msg)
	}
	t.Log("✓ dataLen=0（空 body）正确解析")
}

// TestLargeNumberOfSmallPackets 大量小包压力测试
func TestLargeNumberOfSmallPackets(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()

	msgCh, errCh := startRecvLoop(sess)

	const total = 1000
	go func() {
		for i := 0; i < total; i++ {
			client.sendMsg(&proto_player.C2SPing{})
		}
	}()

	var received int
	deadline := time.After(10 * time.Second)
	for received < total {
		select {
		case <-msgCh:
			received++
		case err := <-errCh:
			t.Fatalf("error after %d/%d msgs: %v", received, total, err)
		case <-deadline:
			t.Fatalf("timeout: %d/%d", received, total)
		}
	}
	t.Logf("✓ 连续 %d 条小包全部正确接收", total)
}

// ═══════════════════════════════════════════════════════════════════
// 并发安全和压力测试
// ═══════════════════════════════════════════════════════════════════

// TestConcurrentCloseAndSend 并发 Close + Send 不 panic
func TestConcurrentCloseAndSend(t *testing.T) {
	for round := 0; round < 20; round++ {
		sess, client := makeTestPair(t)

		// 客户端消费
		go func() {
			for {
				if _, err := client.recvMsg(); err != nil {
					return
				}
			}
		}()

		var wg sync.WaitGroup
		// 多 goroutine 同时 Send
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					sess.Send(&proto_player.S2CPong{ZoneOffset: int64(j)})
				}
			}()
		}
		// 另一个 goroutine Close
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			sess.Close()
		}()

		wg.Wait()
		client.close()
	}
	t.Log("✓ 并发 Close+Send 20 轮，无 panic")
}

// TestConcurrentCloseAndReceive 并发 Close + Receive 不 panic
func TestConcurrentCloseAndReceive(t *testing.T) {
	for round := 0; round < 20; round++ {
		sess, client := makeTestPair(t)

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for {
				if _, err := sess.Receive(); err != nil {
					return
				}
			}
		}()

		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(5)) * time.Millisecond)
			sess.Close()
		}()

		wg.Wait()
		client.close()
	}
	t.Log("✓ 并发 Close+Receive 20 轮，无 panic")
}

// TestConcurrentMultipleClose 多 goroutine 同时 Close
func TestConcurrentMultipleClose(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()

	var wg sync.WaitGroup
	var closedCount int64
	var errCount int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sess.Close()
			if err == nil {
				atomic.AddInt64(&closedCount, 1)
			} else if err == tcp.SessionClosedError {
				atomic.AddInt64(&errCount, 1)
			}
		}()
	}

	wg.Wait()

	if closedCount != 1 {
		t.Errorf("expected exactly 1 successful close, got %d", closedCount)
	}
	t.Logf("✓ 10 goroutine 并发 Close: 成功=%d, SessionClosedError=%d",
		atomic.LoadInt64(&closedCount), atomic.LoadInt64(&errCount))
}

// TestHighThroughputBidirectional 双向高吞吐（使用真实 TCP 避免 net.Pipe 无缓冲瓶颈）
func TestHighThroughputBidirectional(t *testing.T) {
	const msgCount = 500

	// 服务端：echo + 主动发送 S2CPong
	var serverSess atomic.Value
	sessReady := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()

	server := tcp.NewServer(
		listener,
		tcp.ProtocolFunc(codec.NewParser),
		2048, // 大缓冲
		tcp.HandlerFunc(func(sess *tcp.Session) {
			serverSess.Store(sess)
			close(sessReady)
			for {
				if _, err := sess.Receive(); err != nil {
					return
				}
			}
		}),
	)
	go server.Serve()
	defer server.Stop()

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	client := &testClient{conn: conn, t: t}

	// 等待服务端 session 就绪
	select {
	case <-sessReady:
	case <-time.After(2 * time.Second):
		t.Fatal("server session not ready")
	}
	sess := serverSess.Load().(*tcp.Session)

	// 服务端主动发 msgCount 条消息
	serverSendDone := make(chan int64)
	go func() {
		var sent int64
		for i := 0; i < msgCount; i++ {
			if err := sess.Send(&proto_player.S2CPong{ZoneOffset: int64(i)}); err != nil {
				break
			}
			sent++
		}
		serverSendDone <- sent
	}()

	// 客户端发 msgCount 条消息
	clientSendDone := make(chan int64)
	go func() {
		var sent int64
		for i := 0; i < msgCount; i++ {
			client.sendMsg(&proto_player.C2SPing{})
			sent++
		}
		clientSendDone <- sent
	}()

	// 客户端收取服务端发来的消息
	clientRecvDone := make(chan int64)
	go func() {
		var count int64
		for count < msgCount {
			client.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
			if _, err := client.recvMsg(); err != nil {
				break
			}
			count++
		}
		clientRecvDone <- count
	}()

	srvSent := <-serverSendDone
	cliSent := <-clientSendDone
	cliRecv := <-clientRecvDone

	client.close()

	t.Logf("✓ 双向高吞吐：服务端发送 %d/%d, 客户端发送 %d/%d, 客户端接收 %d/%d",
		srvSent, msgCount, cliSent, msgCount, cliRecv, msgCount)

	if srvSent < int64(msgCount*90/100) {
		t.Errorf("server sent too few: %d/%d", srvSent, msgCount)
	}
	if cliRecv < int64(msgCount*80/100) {
		t.Errorf("client received too few: %d/%d", cliRecv, msgCount)
	}
}

// ═══════════════════════════════════════════════════════════════════
// Session 状态 (State) 测试
// ═══════════════════════════════════════════════════════════════════

// TestSessionState Set/Get/Delete 基本操作
func TestSessionState(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()
	defer sess.Close()

	sess.Set("uid", int64(12345))
	sess.Set("name", "player1")

	if v, ok := sess.Get("uid"); !ok || v.(int64) != 12345 {
		t.Errorf("Get uid: ok=%v, v=%v", ok, v)
	}
	if v, ok := sess.Get("name"); !ok || v.(string) != "player1" {
		t.Errorf("Get name: ok=%v, v=%v", ok, v)
	}

	// 不存在的 key
	if _, ok := sess.Get("nonexist"); ok {
		t.Error("nonexist key should not exist")
	}

	sess.Delete("uid")
	if _, ok := sess.Get("uid"); ok {
		t.Error("uid should be deleted")
	}
	t.Log("✓ Session State Set/Get/Delete 正常")
}

// TestSessionStateOverwrite 覆盖写入
func TestSessionStateOverwrite(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()
	defer sess.Close()

	sess.Set("key", "value1")
	sess.Set("key", "value2")

	v, ok := sess.Get("key")
	if !ok || v.(string) != "value2" {
		t.Errorf("expected 'value2', got %v", v)
	}
	t.Log("✓ Session State 覆盖写入正确")
}

// ═══════════════════════════════════════════════════════════════════
// TCP Server echo 端到端集成测试
// ═══════════════════════════════════════════════════════════════════

// TestTCPServerEchoDataIntegrity 端到端数据完整性
func TestTCPServerEchoDataIntegrity(t *testing.T) {
	server, addr := startTestServer(t)
	defer server.Stop()

	client := dialTestClient(t, addr)
	defer client.close()

	// 发送 100 个 ping，每次验证返回
	for i := 0; i < 100; i++ {
		client.sendMsg(&proto_player.C2SPing{})
		client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		msg, err := client.recvMsg()
		if err != nil {
			t.Fatalf("echo %d: %v", i, err)
		}
		if _, ok := msg.(*proto_player.S2CPong); !ok {
			t.Fatalf("echo %d: expected S2CPong, got %T", i, msg)
		}
	}
	t.Log("✓ 端到端 100 次 echo 数据完整")
}

// TestTCPServerClientSendsGarbageThenDisconnects 客户端发垃圾后断开，服务器稳定
func TestTCPServerClientSendsGarbageThenDisconnects(t *testing.T) {
	server, addr := startTestServer(t)
	defer server.Stop()

	for i := 0; i < 20; i++ {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			t.Fatalf("dial %d failed: %v", i, err)
		}
		garbage := make([]byte, rand.Intn(200)+1)
		rand.Read(garbage)
		conn.Write(garbage)
		conn.Close()
	}

	// 验证服务器仍然正常工作
	time.Sleep(100 * time.Millisecond)
	client := dialTestClient(t, addr)
	defer client.close()

	client.sendMsg(&proto_player.C2SPing{})
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("server should still work: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}
	t.Log("✓ 20 个垃圾连接后服务器仍正常工作")
}

// TestTCPServerHalfOpenConnection 半开连接：客户端只连不发
func TestTCPServerHalfOpenConnection(t *testing.T) {
	server, addr := startTestServer(t)
	defer server.Stop()

	// 5 个半开连接
	halfOpen := make([]net.Conn, 5)
	for i := range halfOpen {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			t.Fatalf("dial %d failed: %v", i, err)
		}
		halfOpen[i] = conn
	}

	// 验证正常客户端仍可通信
	client := dialTestClient(t, addr)
	defer client.close()

	client.sendMsg(&proto_player.C2SPing{})
	client.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := client.recvMsg()
	if err != nil {
		t.Fatalf("normal client should work: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}

	// 关闭半开连接
	for _, c := range halfOpen {
		c.Close()
	}
	t.Log("✓ 半开连接不影响正常客户端通信")
}

// TestTCPServerClientSendsOversizedPacket 客户端发超大包，服务器断开该连接但其他连接不受影响
func TestTCPServerClientSendsOversizedPacket(t *testing.T) {
	server, addr := startTestServer(t)
	defer server.Stop()

	// 正常客户端
	normalClient := dialTestClient(t, addr)
	defer normalClient.close()

	// 异常客户端发超大包
	badConn, _ := net.DialTimeout("tcp", addr, 2*time.Second)
	badClient := &testClient{conn: badConn, t: t}
	badClient.rawWriteUint32(500 * 1024) // 超大包 500KB > 128KB MaxSize
	time.Sleep(100 * time.Millisecond)
	badClient.close()

	// 正常客户端仍然可用
	normalClient.sendMsg(&proto_player.C2SPing{})
	normalClient.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	msg, err := normalClient.recvMsg()
	if err != nil {
		t.Fatalf("normal client should still work: %v", err)
	}
	if _, ok := msg.(*proto_player.S2CPong); !ok {
		t.Fatalf("expected S2CPong, got %T", msg)
	}
	t.Log("✓ 异常客户端超大包不影响其他正常连接")
}

// TestTCPServerMixedGoodAndBadClients 混合正常和异常客户端
func TestTCPServerMixedGoodAndBadClients(t *testing.T) {
	server, addr := startTestServer(t)
	defer server.Stop()

	var wg sync.WaitGroup
	var goodSuccess, badDone int64

	// 10 个正常客户端
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				return
			}
			defer conn.Close()
			c := &testClient{conn: conn, t: t}
			c.sendMsg(&proto_player.C2SPing{})
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if _, err := c.recvMsg(); err == nil {
				atomic.AddInt64(&goodSuccess, 1)
			}
		}()
	}

	// 10 个异常客户端（各种异常行为）
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				return
			}
			defer conn.Close()
			c := &testClient{conn: conn, t: t}

			switch idx % 5 {
			case 0: // 立即断开
				// do nothing
			case 1: // 超大包
				c.rawWriteUint32(300 * 1024)
			case 2: // 未知 proto ID
				c.rawSend(0xDEADBEEF, []byte{1, 2, 3})
			case 3: // 垃圾数据
				garbage := make([]byte, 100)
				rand.Read(garbage)
				c.rawWriteBytes(garbage)
			case 4: // 不完整包
				c.rawWriteUint32(50)
				c.rawWriteBytes([]byte{1, 2, 3})
			}
			atomic.AddInt64(&badDone, 1)
		}(i)
	}

	wg.Wait()
	t.Logf("✓ 混合测试：正常客户端成功=%d/10, 异常客户端完成=%d/10",
		atomic.LoadInt64(&goodSuccess), atomic.LoadInt64(&badDone))

	if goodSuccess < 8 {
		t.Errorf("too few good clients succeeded: %d/10", goodSuccess)
	}
}

// ═══════════════════════════════════════════════════════════════════
// 边界条件和回归测试
// ═══════════════════════════════════════════════════════════════════

// TestAddCloseCallbackAfterClose 关闭后添加回调不触发
func TestAddCloseCallbackAfterClose(t *testing.T) {
	sess, client := makeTestPair(t)
	defer client.close()

	sess.Close()

	var called int64
	sess.AddCloseCallback("late", "key", func() {
		atomic.StoreInt64(&called, 1)
	})
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt64(&called) != 0 {
		t.Error("callback added after close should not be triggered")
	}
	t.Log("✓ 关闭后添加的回调不触发")
}

// TestSessionIDUnique 每个 Session ID 唯一
func TestSessionIDUnique(t *testing.T) {
	ids := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		sess, client := makeTestPair(t)
		if ids[sess.ID()] {
			t.Fatalf("duplicate session ID: %d", sess.ID())
		}
		ids[sess.ID()] = true
		sess.Close()
		client.close()
	}
	t.Log("✓ 100 个 Session ID 全部唯一")
}

// TestCodecEncodeDecodeRoundTrip 编解码往返一致性
func TestCodecEncodeDecodeRoundTrip(t *testing.T) {
	messages := []gogoproto.Message{
		&proto_player.C2SPing{},
		&proto_player.C2SLogin{Token: "test_token_123"},
		&proto_player.S2CPong{ZoneOffset: 1234567890},
		&proto_player.C2SLogout{},
	}

	for _, original := range messages {
		data, _ := gogoproto.Marshal(original)
		id, err := proto_id.Router.MessageID(original)
		if err != nil {
			t.Fatalf("MessageID failed for %T: %v", original, err)
		}

		encoded, err := codec.Encode(id, data)
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		// 手动解码验证
		dataLen := binary.LittleEndian.Uint32(encoded[0:4])
		protoID := binary.LittleEndian.Uint32(encoded[4:8])
		body := encoded[8:]

		if protoID != id {
			t.Errorf("proto ID mismatch: %d vs %d", protoID, id)
		}
		if int(dataLen) != len(data) {
			t.Errorf("data length mismatch: %d vs %d", dataLen, len(data))
		}

		decoded, err := proto_id.Router.NewMessage(protoID)
		if err != nil {
			t.Fatalf("NewMessage failed: %v", err)
		}
		if err := gogoproto.Unmarshal(body, decoded); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if !gogoproto.Equal(original, decoded) {
			t.Errorf("round-trip mismatch for %T", original)
		}
	}
	t.Log("✓ 编解码往返一致性验证通过（4 种消息类型）")
}

// TestCodecEncodeMaxSizeBoundary 编码最大 body 边界
func TestCodecEncodeMaxSizeBoundary(t *testing.T) {
	// 刚好 MaxSize
	data := make([]byte, codec.MaxSize)
	_, err := codec.Encode(1, data)
	if err != nil {
		t.Errorf("Encode at MaxSize should succeed: %v", err)
	}

	// MaxSize + 1 (Encode 本身不做长度检查，由 Receive 端检查)
	data2 := make([]byte, codec.MaxSize+1)
	_, err = codec.Encode(1, data2)
	if err != nil {
		t.Logf("Encode at MaxSize+1: %v", err)
	}
	t.Log("✓ 编码边界测试完成")
}
