package main

import (
	"errors"
	"net"
	"sync"
	"time"

	"xfx/pkg/gate/tcpgate/codec"
	"xfx/pkg/log"
	"xfx/pkg/net/tcp"
	proto_id "xfx/proto"
	"xfx/proto/proto_player"
)

// GameClient 游戏服 TCP 客户端：连接、收发 proto 消息
type GameClient struct {
	conn      net.Conn
	codec     tcp.Codec
	writeChan chan interface{}
	mu        sync.Mutex
	closed    bool
	// 登录态
	Token string
	UID   string
	ID    int // 客户端序号，日志用
}

// NewGameClient 建立 TCP 连接并创建编解码器
func NewGameClient(addr string, id int) (*GameClient, error) {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, err
	}
	parser, err := codec.NewParser(conn, proto_id.Router)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	c := &GameClient{
		conn:      conn,
		codec:     parser,
		writeChan: make(chan interface{}, 256),
		ID:        id,
	}
	return c, nil
}

// Send 投递一条消息到发送队列（非阻塞）
func (c *GameClient) Send(msg interface{}) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errors.New("client closed")
	}
	c.mu.Unlock()
	select {
	case c.writeChan <- msg:
		return nil
	default:
		log.Debug("game client %d write chan full", c.ID)
		return errors.New("write chan full")
	}
}

// writeLoop 在 goroutine 中运行，从 writeChan 取消息发送
func (c *GameClient) writeLoop() {
	for msg := range c.writeChan {
		if err := c.codec.Send(msg); err != nil {
			log.Debug("game client %d send err: %v", c.ID, err)
			return
		}
	}
}

// Close 关闭连接与写通道
func (c *GameClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	close(c.writeChan)
	return c.codec.Close()
}

// readLoop 在 goroutine 中循环收包
func (c *GameClient) readLoop() {
	for {
		c.mu.Lock()
		closed := c.closed
		c.mu.Unlock()
		if closed {
			return
		}
		_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		msg, err := c.codec.Receive()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Debug("game client %d receive err: %v", c.ID, err)
			return
		}
		if msg != nil {
			handleGameMessage(c.ID, msg)
		}
	}
}

// Run 阻塞：启动写/读循环，发 C2SLogin，再按间隔随机打接口直到 stopCh 或超时
func (c *GameClient) Run(loginResult *LoginResult, cfg *Config, stopCh <-chan struct{}, onStop func()) {
	defer func() {
		_ = c.Close()
		if onStop != nil {
			onStop()
		}
	}()

	c.Token = loginResult.Token
	c.UID = loginResult.UID

	go c.writeLoop()
	go c.readLoop()

	// 先发登录
	if err := c.Send(&proto_player.C2SLogin{Token: c.Token}); err != nil {
		log.Debug("game client %d send login err: %v", c.ID, err)
		return
	}

	ticker := time.NewTicker(cfg.TestInterval)
	defer ticker.Stop()
	// 心跳：定期发送 C2SPing，避免服务端超时断开（默认 30s）
	pingTicker := time.NewTicker(15 * time.Second)
	defer pingTicker.Stop()
	runDeadline := time.Time{}
	if cfg.RunDuration > 0 {
		runDeadline = time.Now().Add(cfg.RunDuration)
	}

	for {
		select {
		case <-stopCh:
			return
		case <-pingTicker.C:
			if err := c.Send(&proto_player.C2SPing{}); err != nil {
				log.Debug("game client %d send ping err: %v", c.ID, err)
				return
			}
		case <-ticker.C:
			if !runDeadline.IsZero() && time.Now().After(runDeadline) {
				return
			}
			sendRandomC2S(c)
		}
	}
}

// handleGameMessage 根据类型打日志或统计
func handleGameMessage(clientID int, msg interface{}) {
	switch m := msg.(type) {
	case *proto_player.S2CLogin:
		log.Debug("client %d S2CLogin state=%v player=%v", clientID, m.State, m.Player != nil)
	case *proto_player.S2CPong:
		log.Debug("client %d S2CPong zoneOffset=%v", clientID, m.ZoneOffset)
	case *proto_player.S2CKick:
		log.Debug("client %d S2CKick", clientID)
	default:
		log.Debug("client %d recv %T", clientID, msg)
	}
}
