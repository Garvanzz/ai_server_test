package model

import (
	"xfx/core/define"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"
)

type PlayerProps [define.PlayerPropMax]int64

// Player 玩家数据
type Player struct {
	Cache         `json:"-"`
	Id            int64
	Uid           string
	Props         PlayerProps
	Base          *PlayerBase
	Bag           *Bag
	Shop          *PlayerShop
	Task          *PlayerTask
	Handbook      *Handbook      //图鉴
	Hero          *Hero          //角色相关
	Draw          *DrawHero      //抽卡
	Stage         *Stage         //关卡
	Lineup        *LineUp        //布阵
	Equip         *Equip         //装备
	Fashion       *Fashion       //时装
	Welfare       *Welfare       //福利
	OpenBox       *OpenBox       //开箱子
	IdleBox       *IdleBox       //挂机宝箱
	Skill         *Skill         //技能
	Magic         *Magic         //法术
	Destiny       *Destiny       //天命
	ShenjiDraw    *ShenjiDraw    //神机抽取
	Collection    *Collection    //收集
	Divine        *Divine        //领悟心得
	GemAppraisal  *GemAppraisal  //鉴宝
	Pet           *Pet           //宠物
	PetEquip      *PetEquip      //宠物装备
	PetHandbook   *PetHandbook   //宠物图鉴
	PetDraw       *PetDraw       //宠物抽卡
	PlayerProp    *PlayerProp    //玩家道具相关，头像框/称号/泡泡
	Danaotiangong *Danaotiangong //大闹天宫
	Mission       *Mission       //副本
	Transaction   *Transaction   //交易所
	Paradise      *Paradise      //乐园
	Cdkey         *PlayerCdkey    //兑换码
}

func (pl *Player) InGame() bool {
	return pl.GameRun.PID != nil
}

type PlayerInfo struct {
	Id          int64  `redis:"id"`
	Uid         string `redis:"uid"`
	Name        string `redis:"name"`         // 名字
	Level       int32  `redis:"level"`        // 等级
	Exp         int32  `redis:"exp"`          // 经验
	FaceId      int32  `redis:"face_id"`      // 头像ID
	FaceSlotId  int32  `redis:"face_slot_id"` // 头像框id
	OfflineTime int64  `redis:"offline_time"` // 上次登录时间
	Rank        int32  `redis:"rank"`         // 段位0：无
	Title       int32  `redis:"title"`        // 称号
	Job         int32  `redis:"job"`          // 职业
	Sex         int32  `redis:"sex"`          // 性别
	Clan        string `redis:"clan"`         // 帮会
	ClanId      int32  `redis:"clanid"`       // 帮会ID
	HeroId      int32  `redis:"heroId"`       // 主角ID
	BubbleId    int32  `redis:"bubbleId"`     // 泡泡ID
	Power       int64  `redis:"power"`        // 战力
	ServerId    int32  `redis:"serverId"`     // 服务器ID
}

// GetProp 获取prop
func (pl *Player) GetProp(index int) int64 {
	return pl.Props[index]
}

// SetProp 设置prop
func (pl *Player) SetProp(index int, value int64, add bool) (int64, bool) {
	if add {
		pl.Props[index] += value
	} else {
		pl.Props[index] = value
	}

	return pl.Props[index], true
}

func (pl *Player) ToContext() *proto_player.Context {
	ctx := new(proto_player.Context)
	ctx.Id = pl.Id
	ctx.Uid = pl.Uid
	ctx.Name = pl.Base.Name
	ctx.Level = pl.GetProp(define.PlayerPropLevel)
	ctx.Exp = pl.GetProp(define.PlayerPropExp)
	ctx.FaceId = pl.GetProp(define.PlayerPropFaceId)
	ctx.FaceSlotId = pl.GetProp(define.PlayerPropFaceSlotId)
	ctx.OfflineTime = pl.GetProp(define.PlayerPropOfflineTime)
	ctx.Rank = pl.GetProp(define.PlayerPropRank)
	ctx.Chapter = pl.Stage.CurChapter
	ctx.Stage = pl.Stage.CurStage
	ctx.OpenBoxLevel = pl.OpenBox.Level
	ctx.ServerId = int32(pl.GetProp(define.PlayerPropServerId))

	return ctx
}

func (info *PlayerInfo) ToCommonPlayer() *proto_public.CommonPlayerInfo {
	ret := new(proto_public.CommonPlayerInfo)
	ret.PlayerId = info.Id
	ret.Name = info.Name
	ret.Level = info.Level
	ret.FaceId = info.FaceId
	ret.FaceSlotId = info.FaceSlotId
	ret.Title = info.Title
	ret.Bubble = info.BubbleId
	ret.MainHero = info.HeroId
	ret.IsRobot = false
	ret.Job = info.Job
	ret.Score = int64(info.Rank)
	ret.ServerId = info.ServerId
	return ret
}

func (pl *PlayerInfo) ToToContext() *proto_player.Context {
	ctx := new(proto_player.Context)
	ctx.Id = pl.Id
	ctx.Uid = pl.Uid
	ctx.Name = pl.Name
	ctx.Level = int64(pl.Level)
	ctx.Exp = int64(pl.Exp)
	ctx.FaceId = int64(pl.FaceId)
	ctx.FaceSlotId = int64(pl.FaceSlotId)
	ctx.OfflineTime = pl.OfflineTime
	ctx.Rank = int64(pl.Rank)
	ctx.Job = pl.Job
	ctx.ServerId = pl.ServerId
	return ctx
}

func ToCommonPlayerByParam(Id int64, Name string) *proto_public.CommonPlayerInfo {
	ret := new(proto_public.CommonPlayerInfo)
	ret.PlayerId = Id
	ret.Name = Name
	return ret
}
