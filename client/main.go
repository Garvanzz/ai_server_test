package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"xfx/pkg/gate/tcpgate/codec"
	"xfx/pkg/log"
	"xfx/pkg/net/tcp"
	proto_id1 "xfx/proto"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"
)

const (
	ip      = "0.0.0.0:8082"
	Key     = "12348578902223367877723456789012"
	httpUrl = "http://127.0.0.1:9033"
	count   = 2
)

var wg sync.WaitGroup
var closeChan = make(chan struct{})

type Client struct {
	codec      tcp.Codec
	writeChan  chan any
	mu         sync.Mutex
	closeFlag  bool
	uid, token string
	id         int
}

func NewClient(codec tcp.Codec) *Client {
	return &Client{
		codec:     codec,
		writeChan: make(chan any, 1024),
	}
}

func register(account, password string) {
	type RegisterUser struct {
		Account  string
		Password string
		Platform int // 1/pc 2/ios 3/安卓
	}

	b, _ := json.Marshal(&RegisterUser{
		Account:  account,
		Password: password,
		Platform: 1,
	})

	parsedUrl, err := url.Parse(httpUrl + "/register")

	resp, err := http.Post(parsedUrl.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		log.Debug("post failed, err:%v", err)
		return
	}
	defer resp.Body.Close()
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Debug("get resp failed, err:%v", err)
		return
	}

	m := make(map[string]any)
	err = json.Unmarshal(r, &m)
	if err != nil {
		log.Debug("json unmarshal error:%v", err)
		return
	}
	//log.Debug("register success:%v", m)
}

func login(account, password string) (token string, uid string, err error) {

	// TODO:暂时跳过加密
	//_passByte, err := crypto.AesPkcs7Encrypt([]byte(password), []byte(Key))
	//
	//passByte := hex.EncodeToString(_passByte)
	//if err != nil {
	//	err = fmt.Errorf("login decrypt password hex err : %v", err)
	//	return
	//}

	type LoginUser struct {
		Account  string
		Password string
		Version  string
		ServerId int
		Platform int //平台1pc 2ios 3安卓
	}
	b, _ := json.Marshal(&LoginUser{
		Account:  account,
		Password: password,
		Version:  "0.1",
		Platform: 1,
		ServerId: 1,
	})

	parsedUrl, err := url.Parse(httpUrl + "/login")
	resp, err := http.Post(parsedUrl.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		err = fmt.Errorf("login post failed, err:%v", err)
		return
	}
	defer resp.Body.Close()
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("login get resp failed, err:%v", err)
		return
	}

	m := make(map[string]any)
	err = json.Unmarshal(r, &m)
	if err != nil {
		err = fmt.Errorf("json unmarshal error:%v", err)
		return
	}

	_token, ok := m["token"]
	if !ok {
		err = fmt.Errorf("login resp no token:%v", m)
		return
	}
	token = _token.(string)
	uid = m["uid"].(string)
	return
}

func (c *Client) sendMsg(msg any) error {
	c.mu.Lock()
	if c.closeFlag {
		c.mu.Unlock()
		return errors.New("client closed")
	}

	select {
	case c.writeChan <- msg:
		c.mu.Unlock()
		return nil
	default:
		log.Error("client send chan is full ,", c.uid)
		c.mu.Unlock()
		return errors.New("client send chan full")
	}
}

func (c *Client) write() {
	defer c.close()

	for {
		select {
		case msg, ok := <-c.writeChan:
			if !ok {
				return
			}

			if err := c.codec.Send(msg); err != nil {
				log.Debug("client:%v, write send err:%v", c.id, err)
				return
			}
		case <-closeChan:
			log.Debug("client:%v, receive close signal", c.id)
			return
		}
	}
}

func (c *Client) close() {
	//buf := make([]byte, 1024)
	//l := runtime.Stack(buf, false)
	//log.Debug("codec closed:%v \n", fmt.Sprintf("%s", buf[:l]))

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closeFlag {
		if err := c.codec.Close(); err != nil {
			log.Debug("client codec close error:%v", err)
		}

		close(c.writeChan)
		c.closeFlag = true

		wg.Done()
	}
}

func (c *Client) gameTest() {
	if err := c.sendMsg(&proto_player.C2SLogin{Token: c.token}); err != nil {
		log.Error("client:%v, send msg err:%v", c.id, err)
		return
	}

	//time.Sleep(time.Second * 2)
	//if err := c.sendMsg(&proto_activity.C2SLadderRaceSetLineUp{}); err != nil {
	//	log.Error("client:%v, send msg err:%v", c.id, err)
	//	return
	//}
	//
	//time.Sleep(time.Second * 2)
	//
	//_ = c.sendMsg(&proto_activity.C2SArenaSetLineUp{})

	//if err := c.sendMsg(&proto_player.C2SLogin{
	//	Token: c.token,
	//}); err != nil {
	//	return
	//}
}

func (c *Client) run() {
	log.Debug("client:%v, start run", c.id)
	go c.write()
	go c.tick()
	go c.gameTest()

	defer c.close()

	for {
		data, err := c.codec.Receive()
		if err != nil {
			log.Debug("client:%v, read receive error:%v", c.id, err)
			return
		}

		switch msg := data.(type) {
		case *proto_player.S2CLogin:
			log.Debug("client:%v, receive login resp:%v,%v", c.id, msg.State, msg.Player)
		case *proto_player.S2CPong:
			log.Debug("client:%v, receive pong:%v", c.id, msg.ZoneOffset)
		case *proto_player.S2CKick:
			log.Debug("client:%v, receive kick", c.id)
		case *proto_activity.S2CArenaSetLineUp:
			log.Debug("client:%v, receive S2CArenaSetLineUp", c.id)
		default:
			log.Debug("client:%v, client read receive unknown msg:%v", c.id, data)
		}
	}
}

func (c *Client) tick() {
	ticker := time.Tick(time.Second * 5)
	for {
		select {
		case <-ticker:
			log.Debug("client %v,send tick", c.id)
			err := c.sendMsg(&proto_player.C2SPing{})
			if err != nil {
				log.Debug("client:%v,send tick err:%v", c.id, err)
				return
			}
		}
	}
}

func main() {
	log.DefaultInit()

	//flag := false

	for i := 0; i < count; i++ {

		account := fmt.Sprintf("account:%d", i+2)
		password := fmt.Sprintf("account:%d", i+2)

		register(account, password)
		token, uid, err := login(account, password)
		if err != nil {
			log.Error("client:%v,login error:%v", i+1, err)
			continue
		}

		conn, err := net.Dial("tcp", ip)
		if err != nil {
			log.Debug("client:%v, dial tcp error:%v", i+1, err)
			continue
		}

		var parser tcp.Codec

		//if flag {
		//	parser, err = codec.NewParser(conn, proto_id.Router) // error proto id
		//} else {
		parser, err = codec.NewParser(conn, proto_id1.Router)
		//}
		//flag = !flag

		client := NewClient(parser)
		client.token = token
		client.uid = uid
		client.id = i + 1

		wg.Add(1)

		//log.Debug("client:%v ,start run", client.id)
		go client.run()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
	<-c

	close(closeChan)

	log.Debug("test close wait")
	wg.Wait()
	log.Debug("test close success")
}
