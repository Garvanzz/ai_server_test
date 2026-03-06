package game

import (
	"fmt"
	"sync"
	"xfx/game_server/pkg/network"
	"xfx/pkg/log"
	proto_room "xfx/proto/proto_room"
)

// GameManager 游戏管理器
type GameManager struct {
	games map[uint64]*Game
	wg    sync.WaitGroup
	rw    sync.RWMutex
}

// NewRoomManager 构造
func NewGameManager() *GameManager {
	m := &GameManager{
		games: make(map[uint64]*Game),
	}
	return m
}

// CreateKcpGame 创建一局游戏
func (m *GameManager) CreateGame(info *proto_room.RoomInfo, logicServer string, conn *network.Conn) (*Game, error) {
	m.rw.Lock()
	defer m.rw.Unlock()

	//游戏ID
	roomId := uint64(info.GetRoomId())
	r, ok := m.games[roomId]
	if ok {
		return nil, fmt.Errorf("room id[%d] exists", roomId)
	}
	log.Info("***************")
	r = NewGame(roomId, info, logicServer, conn)
	m.games[roomId] = r
	go func() {
		m.wg.Add(1)
		defer func() {
			m.rw.Lock()
			log.Debug("游戏结束:%v", roomId)
			delete(m.games, roomId)
			m.rw.Unlock()

			m.wg.Done()
		}()
		r.Run()
	}()

	return r, nil
}

// GetGame 获得游戏
func (m *GameManager) GetGame(id uint64) *Game {

	m.rw.RLock()
	defer m.rw.RUnlock()

	r, _ := m.games[id]
	return r
}

// Num 获得游戏数量
func (m *GameManager) GameNum() int {

	m.rw.RLock()
	defer m.rw.RUnlock()

	return len(m.games)
}

// Stop 停止
func (m *GameManager) Stop() {

	m.rw.Lock()
	for _, v := range m.games {
		v.Stop()
	}
	m.games = make(map[uint64]*Game)
	m.rw.Unlock()

	m.wg.Wait()
}
