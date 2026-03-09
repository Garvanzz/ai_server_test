package main

import (
	"math/rand"
	"xfx/proto/proto_friend"
	"xfx/proto/proto_handbook"
	"xfx/proto/proto_hero"
	"xfx/proto/proto_idlebox"
	"xfx/proto/proto_item"
	"xfx/proto/proto_lineup"
	"xfx/proto/proto_mail"
	"xfx/proto/proto_player"
	"xfx/proto/proto_rank"
	"xfx/proto/proto_shop"
	"xfx/proto/proto_skill"
	"xfx/proto/proto_stage"
	"xfx/proto/proto_task"
	"xfx/proto/proto_welfare"
)

// noArgC2S 无参或仅默认参数的 C2S 消息，用于随机压测
var noArgC2S = []interface{}{
	&proto_item.C2SBag{},
	&proto_task.C2SGetTasks{},
	&proto_player.C2SGetPlayerProp{},
	&proto_mail.C2SMailList{Page: 0},
	&proto_rank.C2SRankData{}, // Type/Id 默认 0
	&proto_stage.C2SInitStage{},
	&proto_shop.C2SShopData{},
	&proto_welfare.C2SDaySign{},
	&proto_welfare.C2SDayAward{},
	&proto_hero.C2SInitHero{},
	&proto_lineup.C2SInitLineUp{},
	&proto_handbook.C2SHandBookData{},
	&proto_idlebox.C2SGetIdleBoxData{},
	&proto_skill.C2SInitSkill{},
	&proto_friend.C2SReqFriendList{},
	&proto_welfare.C2SOnlineAwardInit{},
	&proto_welfare.C2SFunctionOpenInit{},
	&proto_welfare.C2SMonthCardInit{},
}

func sendRandomC2S(c *GameClient) {
	if len(noArgC2S) == 0 {
		return
	}
	msg := noArgC2S[rand.Intn(len(noArgC2S))]
	_ = c.Send(msg)
}
