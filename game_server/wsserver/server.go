package wsserver

//
//import (
//	"fmt"
//	"github.com/gorilla/websocket"
//	"net/http"
//	"xfx/game_server/logic/game"
//	"xfx/game_server/pkg/packet/pb_packet"
//	"xfx/game_server/pkg/ws_server"
//	"xfx/game_server/pkg/wsnetwork"
//	"xfx/pkg/env"
//	"xfx/pkg/log"
//)
//
//// LockStepServer 帧同步服务器
//type LockStepServer struct {
//	udpServer *wsnetwork.Server
//	totalConn int64
//	env       *env.Env
//	mgr       *game.GameManager
//}
//
//func (l *LockStepServer) GetEnv() *env.Env {
//	return l.env
//}
//
//// New 构造
//func New(envs []byte) (*LockStepServer, error) {
//	s := &LockStepServer{
//		mgr: game.NewGameManager(),
//	}
//
//	//加载env
//	e, err := env.LoadEnv(envs)
//	if err != nil {
//		panic(fmt.Sprintf("load env error %v", err))
//	}
//	s.env = e
//
//	////初始mongo
//	//conf := &mongo.Config{
//	//	Url:             e.Mongo.Url,
//	//	MaxPoolSize:     uint64(e.Mongo.MaxPoolSize),
//	//	MinPoolSize:     uint64(e.Mongo.MinPoolSize),
//	//	MaxConnIdleTime: e.Mongo.MaxConnIdleTime,
//	//}
//	//mongo.NewMongo(conf)
//
//	log.Init(e)
//
//	networkServer, err := ws_server.ListenAndWsServe(s, &pb_packet.MsgProtocol{})
//	if err != nil {
//		return nil, err
//	}
//
//	s.udpServer = networkServer
//
//	go func() {
//		fmt.Println("Server started at :10045")
//		http.HandleFunc("/ws", s.handleConnection)
//		if err := http.ListenAndServe(":10045", nil); err != nil {
//			fmt.Println("Error starting server:", err)
//		}
//	}()
//
//	return s, nil
//}
//
//var upgrader = websocket.Upgrader{
//	CheckOrigin: func(r *http.Request) bool {
//		return true // 允许所有跨域请求
//	},
//}
//
//func (l *LockStepServer) handleConnection(w http.ResponseWriter, r *http.Request) {
//	conn, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		fmt.Println("Error during connection upgrade:", err)
//		return
//	}
//	//defer conn.Close()
//
//	fmt.Println("Client connected")
//
//	go l.udpServer.Start(conn, func(conn *websocket.Conn, i *wsnetwork.Server) *wsnetwork.Conn {
//		return wsnetwork.NewConn(conn, l.udpServer)
//	})
//}
//
//// Manager 获取管理器
//func (r *LockStepServer) GameManager() *game.GameManager {
//	return r.mgr
//}
//
//// Stop 停止服务
//func (r *LockStepServer) Stop() {
//	r.mgr.Stop()
//	r.udpServer.Stop()
//}
