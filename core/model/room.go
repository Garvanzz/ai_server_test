package model

import (
	"math"
	"xfx/proto/proto_public"
)

// Invite 房间邀请
type Invite struct {
	Id         int32
	Sender     int64
	SenderName string
	Receiver   int64
	InviteTime int64
	Group      int32
	RoomId     int32
	RoomName   string
}

type Room struct {
	Id           int32
	Type         int32
	State        int32
	Name         string
	Password     string
	Players      []*RoomPlayer
	Owner        int64
	IsPermitLook bool // 是否可以观战
	IsOpen       bool // 是否开放房间
}

type RoomPlayer struct {
	PlayerId  int64
	IsMonster bool
	IsReady   bool
	Group     int32
	Heros     *proto_public.BattleHeroData
	Rank      int32
}

func (r *Room) SetPlayer(player *RoomPlayer) bool {
	limit := 8
	for i := 0; i < limit; i++ {
		if r.Players[i] == nil {
			r.Players[i] = player
			return true
		}
	}
	return false
}

func (r *Room) RemovePlayer(playerId int64) bool {
	for index, v := range r.Players {
		if v == nil {
			v.PlayerId = playerId
			r.Players[index] = nil
			return true
		}
	}
	return false
}

func (r *Room) IsPlayerInRoom(playerId int64) bool {
	for _, v := range r.Players {
		if v.PlayerId == playerId {
			return true
		}
	}
	return false
}

func (r *Room) GetPlayer(playerId int64) *RoomPlayer {
	for _, v := range r.Players {
		if v.PlayerId == playerId {
			return v
		}
	}
	return nil
}

// CanMatch 是否可以开始匹配
func (r *Room) CanMatch() bool {
	for i := 0; i < len(r.Players); i++ {
		roomPlayer := r.Players[i]
		if roomPlayer != nil {
			// 判断玩家上阵数据
			if roomPlayer.Heros == nil || len(roomPlayer.Heros.Items) <= 0 {
				return false
			}

			// 判断玩家是否准备
			if !roomPlayer.IsReady {
				return false
			}
		}
	}
	return true
}

// AverageRank 获取平均段位
func (r *Room) AverageRank() int32 {
	sum := int32(0)
	count := 0
	for i := 0; i < len(r.Players); i++ {
		roomPlayer := r.Players[i]
		if roomPlayer != nil {
			sum += roomPlayer.Rank
			count++
		}
	}

	return int32(math.Ceil(float64(sum) / float64(count)))
}
