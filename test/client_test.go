package tcpgate_test

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"xfx/pkg/gate/tcpgate/codec"
	"xfx/pkg/net/tcp"
	proto_id "xfx/proto"

	gogoproto "github.com/gogo/protobuf/proto"
)

// ═══════════════════════════════════════════════════════════════════
// 辅助工具
// ═══════════════════════════════════════════════════════════════════

// testClient 封装客户端连接，可发合法包也可发原始字节
type testClient struct {
	conn net.Conn
	t    *testing.T
}

// sendMsg 发送合法的 proto 消息（走正常编码路径）
func (c *testClient) sendMsg(msg gogoproto.Message) {
	c.t.Helper()
	data, err := gogoproto.Marshal(msg)
	if err != nil {
		c.t.Fatalf("marshal failed: %v", err)
	}
	protoId, err := proto_id.Router.MessageID(msg)
	if err != nil {
		c.t.Fatalf("get protoId failed: %v", err)
	}
	c.rawSend(protoId, data)
}

// rawSend 按协议格式手工写入 [dataLen 4B LE][protoId 4B LE][data]
func (c *testClient) rawSend(protoId uint32, data []byte) {
	c.t.Helper()
	buf := make([]byte, 4+4+len(data))
	binary.LittleEndian.PutUint32(buf[0:4], uint32(len(data)))
	binary.LittleEndian.PutUint32(buf[4:8], protoId)
	copy(buf[8:], data)
	c.conn.Write(buf)
}

// rawWriteBytes 直接写任意原始字节（用于构造畸形包）
func (c *testClient) rawWriteBytes(data []byte) {
	c.conn.Write(data)
}

// rawWriteUint32 写4字节小端整数
func (c *testClient) rawWriteUint32(v uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	c.conn.Write(buf)
}

// recvMsg 客户端接收一条服务端消息
func (c *testClient) recvMsg() (gogoproto.Message, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.conn, lenBuf); err != nil {
		return nil, err
	}
	dataLen := binary.LittleEndian.Uint32(lenBuf)
	body := make([]byte, dataLen+4)
	if _, err := io.ReadFull(c.conn, body); err != nil {
		return nil, err
	}
	protoId := binary.LittleEndian.Uint32(body[:4])
	msg, err := proto_id.Router.NewMessage(protoId)
	if err != nil {
		return nil, err
	}
	if err := gogoproto.Unmarshal(body[4:], msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *testClient) close() { c.conn.Close() }

// makeTestPair 用 net.Pipe 创建服务端 Session + 客户端 testClient
// net.Pipe 是内存管道，无需端口，速度快
func makeTestPair(t *testing.T) (*tcp.Session, *testClient) {
	t.Helper()
	serverConn, clientConn := net.Pipe()

	parser, err := codec.NewParser(serverConn, proto_id.Router)
	if err != nil {
		t.Fatal("NewParser failed:", err)
	}
	sess := tcp.NewSession(parser, 128)
	return sess, &testClient{conn: clientConn, t: t}
}

// startRecvLoop 后台协程不断 Receive，把消息和错误分别推入 channel
func startRecvLoop(sess *tcp.Session) (<-chan any, <-chan error) {
	msgCh := make(chan any, 64)
	errCh := make(chan error, 1)
	go func() {
		for {
			msg, err := sess.Receive()
			if err != nil {
				errCh <- err
				return
			}
			msgCh <- msg
		}
	}()
	return msgCh, errCh
}

// waitMsg 带超时的消息等待
func waitMsg(t *testing.T, msgCh <-chan any, errCh <-chan error, timeout time.Duration) any {
	t.Helper()
	select {
	case msg := <-msgCh:
		return msg
	case err := <-errCh:
		t.Fatalf("unexpected recv error: %v", err)
	case <-time.After(timeout):
		t.Fatal("timeout waiting for message")
	}
	return nil
}

// waitErr 带超时的错误等待
func waitErr(t *testing.T, msgCh <-chan any, errCh <-chan error, timeout time.Duration) error {
	t.Helper()
	select {
	case err := <-errCh:
		return err
	case msg := <-msgCh:
		t.Fatalf("expected error but received message: %T", msg)
	case <-time.After(timeout):
		t.Fatal("timeout waiting for error")
	}
	return nil
}

//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 1：正常收发
//// ═══════════════════════════════════════════════════════════════════
//
//// TestNormalClientToServer 客户端发消息，服务端正确接收
//func TestNormalClientToServer(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	client.sendMsg(&proto_player.C2SPing{})
//
//	msg := waitMsg(t, msgCh, errCh, 2*time.Second)
//	if _, ok := msg.(*proto_player.C2SPing); !ok {
//		t.Errorf("expected *C2SPing, got %T", msg)
//	}
//	t.Log("✓ 客户端→服务端 正常收发")
//}
//
//// TestNormalServerToClient 服务端发消息，客户端正确接收
//func TestNormalServerToClient(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	pong := &proto_player.S2CPong{ZoneOffset: 99999}
//	if err := sess.Send(pong); err != nil {
//		t.Fatalf("server Send failed: %v", err)
//	}
//
//	msg, err := client.recvMsg()
//	if err != nil {
//		t.Fatalf("client recvMsg failed: %v", err)
//	}
//	got, ok := msg.(*proto_player.S2CPong)
//	if !ok {
//		t.Fatalf("expected *S2CPong, got %T", msg)
//	}
//	if got.ZoneOffset != 99999 {
//		t.Errorf("ZoneOffset mismatch: want 99999, got %d", got.ZoneOffset)
//	}
//	t.Log("✓ 服务端→客户端 正常收发，字段值正确")
//}
//
//// TestMultipleMessagesInOrder 连续发送多条消息，验证消息边界和顺序
//func TestMultipleMessagesInOrder(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	// 三条不同类型消息连续发出
//	client.sendMsg(&proto_player.C2SPing{})
//	client.sendMsg(&proto_player.C2SLogin{Token: "hello_token"})
//	client.sendMsg(&proto_player.C2SLogout{})
//
//	received := make([]any, 0, 3)
//	deadline := time.After(3 * time.Second)
//	for len(received) < 3 {
//		select {
//		case msg := <-msgCh:
//			received = append(received, msg)
//		case err := <-errCh:
//			t.Fatalf("recv error after %d msgs: %v", len(received), err)
//		case <-deadline:
//			t.Fatalf("timeout: got %d/3", len(received))
//		}
//	}
//
//	if _, ok := received[0].(*proto_player.C2SPing); !ok {
//		t.Errorf("[0] expected C2SPing, got %T", received[0])
//	}
//	if login, ok := received[1].(*proto_player.C2SLogin); !ok {
//		t.Errorf("[1] expected C2SLogin, got %T", received[1])
//	} else if login.Token != "hello_token" {
//		t.Errorf("Token mismatch: want 'hello_token', got '%s'", login.Token)
//	}
//	if _, ok := received[2].(*proto_player.C2SLogout); !ok {
//		t.Errorf("[2] expected C2SLogout, got %T", received[2])
//	}
//	t.Log("✓ 3条消息边界正确，顺序正确，字段正确")
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 2：客户端主动断开
//// ═══════════════════════════════════════════════════════════════════
//
//// TestClientActiveDisconnect 客户端主动断开，服务端 Receive 返回 EOF
//func TestClientActiveDisconnect(t *testing.T) {
//	sess, client := makeTestPair(t)
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	// 先发一条消息，再断开
//	client.sendMsg(&proto_player.C2SPing{})
//	waitMsg(t, msgCh, errCh, 2*time.Second)
//
//	// 客户端主动关闭
//	client.close()
//
//	err := waitErr(t, msgCh, errCh, 2*time.Second)
//	if err != io.EOF && err != io.ErrUnexpectedEOF {
//		t.Errorf("expected EOF-like error, got: %v", err)
//	}
//	t.Logf("✓ 客户端主动断开，服务端收到: %v", err)
//
//	if !sess.IsClosed() {
//		t.Error("session.IsClosed() should be true after client disconnect")
//	}
//}
//
//// TestClientDisconnectMidPacket 客户端在包体发送到一半时断开（截断包）
//func TestClientDisconnectMidPacket(t *testing.T) {
//	sess, client := makeTestPair(t)
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	// 写 dataLen=100, protoId=C2SLogin，但只写 5 字节 body 就断开
//	client.rawWriteUint32(100) // dataLen
//	c2sLoginID, _ := proto_id.Router.MessageID(&proto_player.C2SLogin{})
//	client.rawWriteUint32(c2sLoginID)     // protoId
//	client.rawWriteBytes(make([]byte, 5)) // 只写5字节
//	client.close()                        // 未写完就断开
//
//	err := waitErr(t, msgCh, errCh, 2*time.Second)
//	if err == nil {
//		t.Error("expected error for truncated packet")
//	}
//	t.Logf("✓ 截断包触发错误: %v", err)
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 3：服务端主动断开
//// ═══════════════════════════════════════════════════════════════════
//
//// TestServerActiveClose 服务端主动 Close，客户端读取报错
//func TestServerActiveClose(t *testing.T) {
//	sess, client := makeTestPair(t)
//
//	if err := sess.Close(); err != nil {
//		t.Fatalf("sess.Close() failed: %v", err)
//	}
//
//	// 客户端尝试读，应报错
//	buf := make([]byte, 4)
//	_, err := io.ReadFull(client.conn, buf)
//	if err == nil {
//		t.Error("client Read should fail after server close")
//	}
//	t.Logf("✓ 服务端主动关闭，客户端读取报错: %v", err)
//
//	// 服务端再 Send 应立即返回错误
//	sendErr := sess.Send(&proto_player.S2CPong{})
//	if sendErr == nil {
//		t.Error("Send to closed session should fail")
//	}
//	t.Logf("✓ 已关闭 session Send 返回: %v", sendErr)
//}
//
//// TestServerCloseWhileClientConnected 服务端关闭后，回收 loop 中的 recv 也报错
//func TestServerCloseWhileClientConnected(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	// 服务端关闭
//	sess.Close()
//
//	// recv loop 应当退出并报错
//	select {
//	case err := <-errCh:
//		t.Logf("✓ 服务端关闭后 recv loop 退出: %v", err)
//	case <-msgCh:
//		t.Error("should not receive message after session closed")
//	case <-time.After(2 * time.Second):
//		t.Error("recv loop should have exited")
//	}
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 4：协议违规——包体过大
//// ═══════════════════════════════════════════════════════════════════
//
//// TestMessageTooLong 声明包大小超过 MaxSize(128KB)，应被拒绝
//func TestMessageTooLong(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	// 写超大 dataLen（200KB）
//	client.rawWriteUint32(200 * 1024)
//
//	err := waitErr(t, msgCh, errCh, 2*time.Second)
//	t.Logf("✓ 超大包被拒绝: %v", err)
//
//	if !sess.IsClosed() {
//		t.Error("session should be closed after oversized message")
//	}
//}
//
//// TestExactlyMaxSize 恰好等于最大长度（应通过长度检查，但 proto 解析可能失败）
//func TestExactlyMaxSize(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	// dataLen = 128 * 1024 - 1 字节（刚好不超）
//	c2sPingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})
//	body := make([]byte, 128*1024-1)
//	client.rawSend(c2sPingID, body)
//
//	select {
//	case err := <-errCh:
//		t.Logf("最大边界包（可能解析失败）: %v", err)
//	case msg := <-msgCh:
//		t.Logf("✓ 最大边界包解析成功: %T", msg)
//	case <-time.After(2 * time.Second):
//		t.Fatal("timeout")
//	}
//}
//
//// TestOneByteOverMaxSize 超过最大长度1字节
//func TestOneByteOverMaxSize(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	client.rawWriteUint32(128*1024 + 1) // 超过 MaxSize
//
//	err := waitErr(t, msgCh, errCh, 2*time.Second)
//	t.Logf("✓ 超最大长度+1 被拒绝: %v", err)
//	_ = sess
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 5：协议违规——未知 proto ID
//// ═══════════════════════════════════════════════════════════════════
//
//// TestUnknownProtoID 发送未注册的协议号
//func TestUnknownProtoID(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	const unknownID uint32 = 0xDEADBEEF
//	client.rawSend(unknownID, []byte{0x01, 0x02, 0x03})
//
//	err := waitErr(t, msgCh, errCh, 2*time.Second)
//	t.Logf("✓ 未知协议号被拒绝: %v", err)
//
//	if !sess.IsClosed() {
//		t.Error("session should be closed after unknown proto id")
//	}
//}
//
//// TestZeroProtoID 协议号为 0（未注册）
//func TestZeroProtoID(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	client.rawSend(0, []byte{})
//
//	err := waitErr(t, msgCh, errCh, 2*time.Second)
//	t.Logf("✓ 协议号 0 被拒绝: %v", err)
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 6：协议违规——body 无法反序列化
//// ═══════════════════════════════════════════════════════════════════
//
//// TestCorruptedProtoBody 合法协议号 + 乱码 body
//func TestCorruptedProtoBody(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	c2sLoginID, _ := proto_id.Router.MessageID(&proto_player.C2SLogin{})
//	// 构造非法 protobuf 数据（varint 非法截断）
//	corruptedBody := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
//	client.rawSend(c2sLoginID, corruptedBody)
//
//	select {
//	case err := <-errCh:
//		t.Logf("✓ 乱码 body 被拒绝: %v", err)
//	case msg := <-msgCh:
//		// proto 对某些乱码有容错，允许通过
//		t.Logf("注意: 乱码被 proto 容错解析为 %T（proto 宽容性）", msg)
//	case <-time.After(2 * time.Second):
//		t.Fatal("timeout")
//	}
//}
//
//// TestEmptyBody 空 body（对于空消息如 C2SPing 是合法的）
//func TestEmptyBody(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	c2sPingID, _ := proto_id.Router.MessageID(&proto_player.C2SPing{})
//	client.rawSend(c2sPingID, []byte{}) // 空 body 合法
//
//	msg := waitMsg(t, msgCh, errCh, 2*time.Second)
//	if _, ok := msg.(*proto_player.C2SPing); !ok {
//		t.Errorf("expected *C2SPing, got %T", msg)
//	}
//	t.Log("✓ 空 body 包正常解析")
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 7：仅写包头不写 body（纯垃圾数据）
//// ═══════════════════════════════════════════════════════════════════
//
//// TestOnlyLenHeader 只写 dataLen 不写后续内容后关闭
//func TestOnlyLenHeader(t *testing.T) {
//	sess, client := makeTestPair(t)
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	// 写一个合法长度（比如 4），但不写后续字节，直接关闭
//	client.rawWriteUint32(4)
//	client.close()
//
//	err := waitErr(t, msgCh, errCh, 2*time.Second)
//	t.Logf("✓ 只有包头后断开，错误: %v", err)
//}
//
//// TestZeroBytesThenClose 建连后立即断开，不发任何数据
//func TestZeroBytesThenClose(t *testing.T) {
//	sess, client := makeTestPair(t)
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	client.close() // 不发任何数据
//
//	err := waitErr(t, msgCh, errCh, 2*time.Second)
//	if err != io.EOF && err != io.ErrUnexpectedEOF {
//		t.Errorf("expected EOF, got: %v", err)
//	}
//	t.Logf("✓ 空连接断开: %v", err)
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 8：Session 状态和幂等性
//// ═══════════════════════════════════════════════════════════════════
//
//// TestSessionIsClosedFlag IsClosed 状态正确反映
//func TestSessionIsClosedFlag(t *testing.T) {
//	sess, client := makeTestPair(t)
//
//	if sess.IsClosed() {
//		t.Error("session should not be closed initially")
//	}
//
//	// 启动 recv loop 触发 EOF 检测
//	go sess.Receive()
//
//	client.close()
//	time.Sleep(100 * time.Millisecond)
//
//	if !sess.IsClosed() {
//		t.Error("session should be closed after client disconnect")
//	}
//	t.Log("✓ IsClosed 状态正确")
//}
//
//// TestDoubleClose 重复关闭幂等（第二次应返回 SessionClosedError）
//func TestDoubleClose(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	err1 := sess.Close()
//	err2 := sess.Close()
//
//	if err1 != nil && err1.Error() != "" {
//		// net.Pipe 关闭时可能有 "io: read/write on closed pipe"，忽略
//	}
//	if err2 != tcp.SessionClosedError {
//		t.Errorf("second Close() should return SessionClosedError, got: %v", err2)
//	}
//	t.Log("✓ 重复关闭幂等，第二次返回 SessionClosedError")
//}
//
//// TestSendAfterClose 关闭后发送立即返回错误
//func TestSendAfterClose(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	sess.Close()
//
//	err := sess.Send(&proto_player.S2CPong{})
//	if err == nil {
//		t.Error("Send after Close should fail")
//	}
//	t.Logf("✓ 关闭后 Send 返回: %v", err)
//}
//
//// TestReceiveAfterClose 关闭后 Receive 应立即返回
//func TestReceiveAfterClose(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	sess.Close()
//
//	done := make(chan struct{})
//	go func() {
//		sess.Receive() // 应该立即返回（conn 已关闭）
//		close(done)
//	}()
//
//	select {
//	case <-done:
//		t.Log("✓ 关闭后 Receive 立即返回")
//	case <-time.After(2 * time.Second):
//		t.Error("Receive after Close hung")
//	}
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 9：关闭回调机制
//// ═══════════════════════════════════════════════════════════════════
//
//// TestCloseCallbackTriggeredOnSessionClose session.Close() 触发回调
//func TestCloseCallbackTriggeredOnSessionClose(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	done := make(chan struct{})
//	sess.AddCloseCallback("owner", "key1", func() {
//		close(done)
//	})
//
//	sess.Close()
//
//	select {
//	case <-done:
//		t.Log("✓ session.Close() 触发关闭回调")
//	case <-time.After(2 * time.Second):
//		t.Error("close callback not triggered")
//	}
//}
//
//// TestCloseCallbackTriggeredOnClientDisconnect 客户端断开也触发回调
//func TestCloseCallbackTriggeredOnClientDisconnect(t *testing.T) {
//	sess, client := makeTestPair(t)
//
//	done := make(chan struct{})
//	sess.AddCloseCallback("owner", "key2", func() {
//		close(done)
//	})
//
//	go sess.Receive() // 启动 recv 检测 EOF
//
//	client.close()
//
//	select {
//	case <-done:
//		t.Log("✓ 客户端断开触发关闭回调")
//	case <-time.After(2 * time.Second):
//		t.Error("close callback not triggered after client disconnect")
//	}
//}
//
//// TestMultipleCloseCallbacks 多个回调都被触发
//func TestMultipleCloseCallbacks(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	var count int64
//	wg := sync.WaitGroup{}
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		key := i
//		sess.AddCloseCallback("owner", key, func() {
//			atomic.AddInt64(&count, 1)
//			wg.Done()
//		})
//	}
//
//	sess.Close()
//
//	done := make(chan struct{})
//	go func() { wg.Wait(); close(done) }()
//
//	select {
//	case <-done:
//		if n := atomic.LoadInt64(&count); n != 5 {
//			t.Errorf("expected 5 callbacks, got %d", n)
//		}
//		t.Log("✓ 5个关闭回调全部触发")
//	case <-time.After(2 * time.Second):
//		t.Errorf("timeout: only %d/5 callbacks fired", atomic.LoadInt64(&count))
//	}
//}
//
//// TestRemoveCloseCallback 移除回调后不再触发
//func TestRemoveCloseCallback(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	var called int64
//	sess.AddCloseCallback("owner", "removable", func() {
//		atomic.StoreInt64(&called, 1)
//	})
//	sess.RemoveCloseCallback("owner", "removable")
//
//	sess.Close()
//	time.Sleep(100 * time.Millisecond)
//
//	if atomic.LoadInt64(&called) != 0 {
//		t.Error("removed callback should not be triggered")
//	}
//	t.Log("✓ 移除的回调未被触发")
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 10：并发安全
//// ═══════════════════════════════════════════════════════════════════
//
//// TestConcurrentSend 多个 goroutine 并发发送，不应 panic 或 race
//func TestConcurrentSend(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	// 客户端消费数据
//	go func() {
//		for {
//			if _, err := client.recvMsg(); err != nil {
//				return
//			}
//		}
//	}()
//
//	const goroutines = 20
//	const msgsEach = 50
//	var wg sync.WaitGroup
//	var errCount int64
//
//	for i := 0; i < goroutines; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			for j := 0; j < msgsEach; j++ {
//				if err := sess.Send(&proto_player.S2CPong{ZoneOffset: int64(j)}); err != nil {
//					atomic.AddInt64(&errCount, 1)
//					return
//				}
//			}
//		}()
//	}
//
//	wg.Wait()
//	t.Logf("✓ %d 个 goroutine 各发 %d 条，发送错误数: %d", goroutines, msgsEach, atomic.LoadInt64(&errCount))
//}
//
//// TestConcurrentClientSend 客户端多 goroutine 并发发，服务端顺序收
//func TestConcurrentClientSend(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	var received int64
//	msgCh, errCh := startRecvLoop(sess)
//
//	const total = 100
//	var wg sync.WaitGroup
//	for i := 0; i < total; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			client.sendMsg(&proto_player.C2SPing{})
//		}()
//	}
//
//	// 等待所有包发完后关闭
//	go func() {
//		wg.Wait()
//		time.Sleep(200 * time.Millisecond)
//		client.close()
//	}()
//
//	deadline := time.After(5 * time.Second)
//	for {
//		select {
//		case <-msgCh:
//			atomic.AddInt64(&received, 1)
//		case <-errCh:
//			t.Logf("✓ 并发收包完成，共收到: %d/%d", atomic.LoadInt64(&received), total)
//			return
//		case <-deadline:
//			t.Errorf("timeout: received %d/%d", atomic.LoadInt64(&received), total)
//			return
//		}
//	}
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 11：快速连接断开（压测场景）
//// ═══════════════════════════════════════════════════════════════════
//
//// TestRapidConnectDisconnect 快速建连断连 30 次，无泄漏
//func TestRapidConnectDisconnect(t *testing.T) {
//	const iterations = 30
//	for i := 0; i < iterations; i++ {
//		sess, client := makeTestPair(t)
//
//		recvDone := make(chan error, 1)
//		go func() {
//			_, err := sess.Receive()
//			recvDone <- err
//		}()
//
//		// 随机发或不发消息
//		if i%2 == 0 {
//			client.sendMsg(&proto_player.C2SPing{})
//		}
//		client.close()
//
//		select {
//		case <-recvDone:
//		case <-time.After(500 * time.Millisecond):
//			t.Errorf("iteration %d: recv did not return after client close", i)
//		}
//	}
//	t.Logf("✓ 快速连接断开 %d 次完成，无泄漏", iterations)
//}
//
//// ═══════════════════════════════════════════════════════════════════
//// 用例 12：错误包后继续发合法包（连接已关闭，不应继续处理）
//// ═══════════════════════════════════════════════════════════════════
//
//// TestNoMessageAfterProtocolViolation 协议违规后 session 关闭，后续包不被处理
//func TestNoMessageAfterProtocolViolation(t *testing.T) {
//	sess, client := makeTestPair(t)
//	defer client.close()
//
//	msgCh, errCh := startRecvLoop(sess)
//
//	// 先发一个未知 proto id（触发 session 关闭）
//	client.rawSend(0xDEADBEEF, []byte{1, 2, 3})
//
//	// 等 session 关闭
//	waitErr(t, msgCh, errCh, 2*time.Second)
//
//	// 再发一条合法消息
//	client.sendMsg(&proto_player.C2SPing{})
//
//	// 不应收到任何消息
//	select {
//	case msg := <-msgCh:
//		t.Errorf("should not receive message after session closed, got %T", msg)
//	case <-time.After(200 * time.Millisecond):
//		t.Log("✓ 协议违规后 session 关闭，后续合法包不被处理")
//	}
//}

//# 进入目录运行全部测试
//cd pkg/gate/tcpgate
//go test -v -race ./...
//
//# 只运行某一类测试
//go test -v -race -run TestClient
//go test -v -race -run TestProtocol
//go test -v -race -run TestConcurrent
//
//# 显示覆盖率
//go test -v -race -cover ./...
