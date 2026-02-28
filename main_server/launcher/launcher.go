package launcher

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"net"
	"time"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	proto_id "xfx/proto"
	"xfx/proto/proto_game"
	"xfx/proto/proto_room"
)

// 启动模块, 链接战斗服

var Module = func() module.Module {
	return new(launcher)
}

type launcher struct {
	modules.BaseModule
	lastPingTime int64
	UDP          net.Conn
	wsconn       *websocket.Conn
}

func (m *launcher) Version() string { return "1.0.0" }
func (m *launcher) GetType() string { return "launcher" } //模块类型
func (m *launcher) OnTick(delta time.Duration) {
	if m.UDP != nil && (time.Now().Unix()-m.lastPingTime) > 10 {
		//发送心跳
		id, _ := proto_id.MessageID(&proto_game.C2SPing{})
		var info = NewPacket(id, &proto_game.C2SPing{}).Serialize()
		if _, e := m.UDP.Write(info); nil != e {
			panic(fmt.Sprintf("write error:%s", e.Error()))
		}
		m.lastPingTime = time.Now().Unix()
	}

	if m.wsconn != nil && (time.Now().Unix()-m.lastPingTime) > 10 {
		//发送心跳
		id, _ := proto_id.MessageID(&proto_game.C2SPing{})
		var info = NewPacket(uint32(id), &proto_game.C2SPing{}).Serialize()
		if e := m.wsconn.WriteMessage(websocket.BinaryMessage, info); nil != e {
			panic(fmt.Sprintf("write error:%s", e.Error()))
		}
		m.lastPingTime = time.Now().Unix()
	}
}

func (m *launcher) OnInit(app module.App) {
	m.BaseModule.OnInit(app)

	m.UDP = nil
	m.wsconn = nil
	m.lastPingTime = time.Now().Unix()
}

func (m *launcher) OnMessage(msg interface{}) interface{} {
	log.Debug("###### launcher receive msg: %v", msg)
	switch p := msg.(type) {
	case *proto_room.RoomInfo:
		m.OnKCPClientStartGameFunc(p)
	default:
		return nil
	}
	return nil
}

// ws连接
func (m *launcher) WsReadWrite() {
	for {
		_, msg, err := m.wsconn.ReadMessage()
		if err != nil {
			log.Error("OnKCPClientStartGameFunc error %v", err)
			return
		}

		ms := &MsgProtocol{}
		n, err := ms.ReadPacketByByte(msg)
		if err != nil {
			log.Error("OnKCPClientStartGameFunc error %v", err)
			return
		}

		if n != nil {
			//id, _ := proto_id.MessageID(&proto_room.S2CStartGame{})
			//settleId, _ := proto_id.MessageID(&proto_public.S2SGameSettleInfo{})
			//if n.id == id {
			//	var data = &proto_room.S2CStartGame{}
			//	n.Unmarshal(data)
			//	// TODO: 报错
			//	m.Invoke("Room", "GameServerStartGame", data)
			//} else if n.id == settleId {
			//	var data = &proto_public.S2SGameSettleInfo{}
			//	n.Unmarshal(data)
			//	m.Cast("Room", data)
			//}
		}
	}
}

// kcp连接
func (m *launcher) KcpReadWrite() {
	ms := &MsgProtocol{}
	for {
		n, err := ms.ReadPacket(m.UDP)
		if err != nil {
			log.Error("OnKCPClientStartGameFunc error %v", err)
			return
		}

		if n != nil {
			id, _ := proto_id.MessageID(&proto_room.S2CStartGame{})
			if n.id == id {
				var data = &proto_room.S2CStartGame{}
				n.Unmarshal(data)
				log.Debug("回调大厅，开始游戏")
				invoke.RoomClient(m).GameServerStartGame(data)
			}
		}
	}
}

// 建立连接
func (m *launcher) OnCreateWsConnnect() {
	if m.wsconn != nil {
		return
	}

	url := "ws://127.0.0.1:10045/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Error("连接失败:", err)
		//panic(err)
	} else {
		m.wsconn = conn
		go m.WsReadWrite() // 开启连接教程
	}
}

// 建立KCP连接
func (m *launcher) OnCreateKcpConnnect() {
	if m.UDP != nil {
		return
	}

	// 创建一个 UDP 连接
	udpConn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10045})
	if err != nil {
		fmt.Println("创建UDP连接失败：", err)
		return
	}

	// 创建 KCP 对象
	kcpConn, err := kcp.Dial(udpConn.RemoteAddr().String())
	if err != nil {
		fmt.Println("创建KCP连接失败：", err)
		return
	}

	m.UDP = kcpConn
	log.Debug("开启udp连接:%v", udpConn.RemoteAddr().String())
	go m.KcpReadWrite() // 开启连接教程
}

// 开始游戏
func (m *launcher) OnKCPClientStartGameFunc(msg *proto_room.RoomInfo) {
	log.Debug("开始游戏")

	//if m.wsconn == nil {
	//	m.OnCreateConnnect()
	//}

	if m.UDP == nil {
		m.OnCreateKcpConnnect()
	}

	// 发送消息
	id, _ := proto_id.MessageID(&proto_room.S2SMSGStartGame{})
	info := NewPacket(id, &proto_room.S2SMSGStartGame{
		RoomInfo:    msg,
		LogicServer: "",
	}).Serialize()
	_, err := m.UDP.Write(info)
	//err := m.wsconn.WriteMessage(websocket.BinaryMessage, info)
	if err != nil {
		log.Debug("发送消息失败:", err)
		return
	}
}
