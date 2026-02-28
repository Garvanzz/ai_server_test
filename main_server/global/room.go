package global

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/proto/proto_room"
)

func ToRoomPlayerProto(p *model.RoomPlayer, leader int64) *proto_room.RoomPlayers {
	player := GetPlayerInfo(p.PlayerId)
	return &proto_room.RoomPlayers{
		CommonPlayerInfo: player.ToCommonPlayer(),
		HeroInfo:         p.Heros,
		IsMonster:        p.IsMonster,
		IsReady:          p.IsReady,
		Group:            p.Group,
		Isleader:         p.PlayerId == leader,
	}
}

func ToRoomInfoProto(r *model.Room) *proto_room.RoomInfo {
	players := make([]*proto_room.RoomPlayers, 0)
	for _, roomPlayer := range r.Players {
		if roomPlayer != nil {
			player := ToRoomPlayerProto(roomPlayer, r.Owner)
			players = append(players, player)
		}
	}

	return &proto_room.RoomInfo{
		RoomId:       r.Id,
		Type:         r.Type,
		IsPermitLook: r.IsPermitLook,
		IsGame:       r.State == define.RoomStateGame,
		IsOpen:       r.IsOpen,
		Name:         r.Name,
		Password:     r.Password,
		Players:      players,
	}
}
