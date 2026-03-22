package internal

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_player"
)

// PushPlayerData 推送玩家数据
func PushPlayerData(ctx global.IPlayer, pl *model.Player) {
	res := &proto_player.PushPlayerInfo{}
	player := new(proto_player.Player)
	player.Id = pl.Id
	player.Uid = pl.Uid
	player.Name = pl.Base.Name
	player.CreateTime = pl.Base.CreateTime
	player.Level = int32(pl.GetProp(define.PlayerPropLevel))
	player.HeadIcon = int32(pl.GetProp(define.PlayerPropFaceId))
	player.Exp = int32(pl.GetProp(define.PlayerPropExp))
	player.HeadFrame = int32(pl.GetProp(define.PlayerPropFaceSlotId))
	player.Title = int32(pl.GetProp(define.PlayerPropTitle))
	player.Rank = int32(pl.GetProp(define.PlayerPropRank))
	player.Job = int32(pl.GetProp(define.PlayerPropJob))
	player.Sex = int32(pl.GetProp(define.PlayerPropSex))
	player.Clan = int32(pl.GetProp(define.PlayerPropClan))
	player.HeroId = int32(pl.GetProp(define.PlayerPropHeroId))
	player.Bubble = int32(pl.GetProp(define.PlayerPropBubbleId))
	res.Player = player
	ctx.Send(res)

	log.Debug("PushPlayerData: %v", res)
}

// 同步头像框/称号/泡泡
func PushPlayerProp(ctx global.IPlayer, pl *model.Player) {
	res := &proto_player.PushChangePlayerProp{}
	res.Titles = pl.PlayerProp.Titles
	res.HeadFrames = pl.PlayerProp.HeadFrames
	res.Bubbles = pl.PlayerProp.Bubbles
	ctx.Send(res)

	log.Debug("PushPlayerProp: %v", res)
}

// 添加头像框
func AddPlayerPropHeadFrame(ctx global.IPlayer, pl *model.Player, id, num int32) {
	//判断有没有这个头像框
	has := false
	for _, v := range pl.PlayerProp.HeadFrames {
		if v == id {
			has = true
			break
		}
	}

	if !has {
		pl.PlayerProp.HeadFrames = append(pl.PlayerProp.HeadFrames, id)
		PushPlayerProp(ctx, pl)
	}
}

// 添加称号
func AddPlayerPropTitle(ctx global.IPlayer, pl *model.Player, id, num int32) {
	//判断有没有这个称号
	has := false
	for _, v := range pl.PlayerProp.Titles {
		if v == id {
			has = true
			break
		}
	}

	if !has {
		pl.PlayerProp.Titles = append(pl.PlayerProp.Titles, id)
		PushPlayerProp(ctx, pl)
	}
}

// 添加泡泡
func AddPlayerPropBubble(ctx global.IPlayer, pl *model.Player, id, num int32) {
	//判断有没有这个泡泡
	has := false
	for _, v := range pl.PlayerProp.Bubbles {
		if v == id {
			has = true
			break
		}
	}

	if !has {
		pl.PlayerProp.Bubbles = append(pl.PlayerProp.Bubbles, id)
		PushPlayerProp(ctx, pl)
	}
}
