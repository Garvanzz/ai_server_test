package player

import (
	"fmt"
	"xfx/core/config"
	"xfx/pkg/utils"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/base"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils/sensitive"
	"xfx/proto/proto_player"

	"github.com/gomodule/redigo/redis"
)

// OnSave 玩家数据落库
func (pl *PlayerAgent) OnSave(isSync bool) {
	dbData := fromProps(pl.model)

	//离线时间
	dbData.OfflineTime = utils.Now().Unix()

	rdb, err := db.GetEngineByPlayerId(pl.model.Id)
	if err != nil {
		log.Error("OnSave PlayerAgent error, no this server:%v", err)
		return
	}

	rdb.RedisExec("hmset", redis.Args{}.Add(fmt.Sprintf("%s:%d", define.Player, pl.model.Id)).AddFlat(dbData)...)

	base.Save(pl.model, isSync)
	bag.Save(pl.model, isSync)
}

func LoadPlayerData(id int64) (*model.Player, error) {
	rdb, err := db.GetEngineByPlayerId(id)
	if err != nil {
		log.Error("LoadPlayerData PlayerAgent error, no this server:%v", err)
		return nil, err
	}

	values, err := redis.Values(rdb.RedisExec("hgetall", fmt.Sprintf("%s:%d", define.Player, id)))
	if err != nil {
		return nil, err
	}

	dst := new(model.PlayerInfo)
	err = redis.ScanStruct(values, dst)
	if err != nil || dst.Id == 0 {
		return nil, fmt.Errorf("load player data scanStruct error:%v", err)
	}

	pl := new(model.Player)
	loadProps(pl, dst)
	base.Load(pl)
	bag.Load(pl)

	return pl, nil
}

func Born(uid string, serverId int) (*model.Player, error) {
	pl := new(model.Player)

	// 获取唯一id
	rdb, err := db.GetEngine(serverId)
	if err != nil {
		log.Error("LoadPlayerData PlayerAgent error, no this server:%v", err)
		return nil, err
	}

	id, err := rdb.GetPlayerId(serverId)
	if err != nil {
		return nil, fmt.Errorf("get unique id error:%v", err)
	}

	pl.Id = id
	pl.Uid = uid

	initPlayerProp(pl)
	base.Init(pl, uid)
	bag.Init(pl)

	return pl, nil
}

//============================Prop==============================

func loadProps(pl *model.Player, info *model.PlayerInfo) {
	pl.Id = info.Id
	pl.Uid = info.Uid

	// props
	pl.Props[define.PlayerPropLevel] = int64(info.Level)
	pl.Props[define.PlayerPropExp] = int64(info.Exp)
	pl.Props[define.PlayerPropFaceId] = int64(info.FaceId)
	pl.Props[define.PlayerPropFaceSlotId] = int64(info.FaceSlotId)
	pl.Props[define.PlayerPropOfflineTime] = info.OfflineTime
	pl.Props[define.PlayerPropRank] = int64(info.Rank)
	pl.Props[define.PlayerPropTitle] = int64(info.Title)
	pl.Props[define.PlayerPropJob] = int64(info.Job)
	pl.Props[define.PlayerPropSex] = int64(info.Sex)
	pl.Props[define.PlayerPropClanId] = int64(info.ClanId)
	pl.Props[define.PlayerPropHeroId] = int64(info.HeroId)
	pl.Props[define.PlayerPropBubbleId] = int64(info.BubbleId)
	pl.Props[define.PlayerPropPower] = int64(info.Power)
	pl.Props[define.PlayerPropServerId] = int64(info.ServerId)
}

func fromProps(pl *model.Player) *model.PlayerInfo {
	info := new(model.PlayerInfo)
	info.Id = pl.Id
	info.Uid = pl.Uid
	info.Name = pl.Base.Name

	// props
	info.FaceId = int32(pl.GetProp(define.PlayerPropFaceId))
	info.FaceSlotId = int32(pl.GetProp(define.PlayerPropFaceSlotId))
	info.Level = int32(pl.GetProp(define.PlayerPropLevel))
	info.Exp = int32(pl.GetProp(define.PlayerPropExp))
	info.OfflineTime = pl.GetProp(define.PlayerPropOfflineTime)
	info.Rank = int32(pl.GetProp(define.PlayerPropRank))
	info.Title = int32(pl.GetProp(define.PlayerPropTitle))
	info.Job = int32(pl.GetProp(define.PlayerPropJob))
	info.Sex = int32(pl.GetProp(define.PlayerPropSex))
	info.ClanId = int32(pl.GetProp(define.PlayerPropClanId))
	info.HeroId = int32(pl.GetProp(define.PlayerPropHeroId))
	info.BubbleId = int32(pl.GetProp(define.PlayerPropBubbleId))
	info.Power = pl.GetProp(define.PlayerPropPower)
	info.ServerId = int32(pl.GetProp(define.PlayerPropServerId))
	return info
}

func initPlayerProp(pl *model.Player) {
	pl.Props[define.PlayerPropLevel] = 1
	pl.Props[define.PlayerPropExp] = 0
	pl.Props[define.PlayerPropFaceId] = define.PlayerHeadIcon
	pl.Props[define.PlayerPropFaceSlotId] = define.PlayerHeadFrame
	pl.Props[define.PlayerPropHeroId] = define.PlayerHeroID
	pl.Props[define.PlayerPropOfflineTime] = utils.Now().Unix()
	pl.Props[define.PlayerPropBubbleId] = define.PlayerBubbleID
	pl.Props[define.PlayerPropServerId] = pl.Id / define.PlayerIdBase
}

//============================Prop END==============================

// ReqChangePlayerName 改名
func ReqChangePlayerName(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SChangeName) {
	res := &proto_player.S2CChangeName{}

	conf := config.Global.Get().PlayerRename
	costItems := make(map[int32]int32)
	costItems[conf[0].ItemId] = conf[0].ItemNum
	if internal.CheckItemsEnough(pl, costItems) == false {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_NONAMECARD
		ctx.Send(res)
		return
	}

	//筛查敏感字
	if sensitive.Filter.IsSensitive(req.Name) {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_SENSITIVEWORDS
		ctx.Send(res)
		return
	}

	//扣除道具
	internal.SubItems(ctx, pl, costItems)

	pl.Base.Name = req.Name
	res.CODE = proto_player.LOGINERRORCODE_ERROR_OK
	ctx.Send(res)
}

// ReqChangeTitle 改称号
func ReqChangeTitle(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SChangeTitle) {
	res := &proto_player.S2CChangeTitle{}
	//判断有没有这个称号
	has := false
	for _, v := range pl.PlayerProp.Titles {
		if v == req.Id {
			has = true
			break
		}
	}

	if !has {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_NOTITLE
		ctx.Send(res)
		return
	}

	pl.Props[define.PlayerPropTitle] = int64(req.Id)
	res.CODE = proto_player.LOGINERRORCODE_ERROR_OK
	ctx.Send(res)
}

// 改头像
func ReqChangeHead(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SChangeHead) {
	res := &proto_player.S2CChangeHead{}

	//判断有没有这个英雄
	if _, ok := pl.Hero.Hero[req.Id]; !ok {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_NOHERO
		ctx.Send(res)
		return
	}

	pl.Props[define.PlayerPropFaceId] = int64(req.Id)
	res.CODE = proto_player.LOGINERRORCODE_ERROR_OK
	ctx.Send(res)
}

// 改头像框
func ReqChangeHeadFrame(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SChangeHeadFrame) {
	res := &proto_player.S2CChangeHeadFrame{}

	//判断有没有这个头像框
	has := false
	for _, v := range pl.PlayerProp.HeadFrames {
		if v == req.Frame {
			has = true
			break
		}
	}

	if !has {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_NOTITLE
		ctx.Send(res)
		return
	}

	pl.Props[define.PlayerPropFaceSlotId] = int64(req.Frame)
	res.CODE = proto_player.LOGINERRORCODE_ERROR_OK
	ctx.Send(res)
}

// 改泡泡
func ReqChangeBubble(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SChangeBubble) {
	res := &proto_player.S2CChangeBubble{}

	//判断有没有这个泡泡
	has := false
	for _, v := range pl.PlayerProp.Bubbles {
		if v == req.Id {
			has = true
			break
		}
	}

	if !has {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_NOTITLE
		ctx.Send(res)
		return
	}

	pl.Props[define.PlayerPropBubbleId] = int64(req.Id)
	res.CODE = proto_player.LOGINERRORCODE_ERROR_OK
	ctx.Send(res)
}

// 改性别
func ReqChangeSex(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SChangeSex) {
	res := &proto_player.S2CChangeBubble{}

	if int32(pl.GetProp(define.PlayerPropSex)) == req.Id {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_NOTITLE
		ctx.Send(res)
		return
	}

	pl.Props[define.PlayerPropSex] = int64(req.Id)
	res.CODE = proto_player.LOGINERRORCODE_ERROR_OK
	ctx.Send(res)
}

// 战力变化
func ReqPlayerChangePower(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SPlayerPowerChange) {
	res := &proto_player.S2CPlayerPowerChange{}

	//要做战力校验

	pl.Props[define.PlayerPropPower] = req.Power
	res.CODE = proto_player.LOGINERRORCODE_ERROR_OK
	ctx.Send(res)
}

// 切换职业
func ReqTransformJob(ctx global.IPlayer, pl *model.Player, req *proto_player.C2STransformJob) {
	res := &proto_player.S2CTransformJob{}

	//判定灵玉够不够
	costItems := make(map[int32]int32)
	costItems[define.ItemIdBoxLingyu] = config.Global.Get().TransformJob
	if internal.CheckItemsEnough(pl, costItems) == false {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_ITEMNOENGTH
		ctx.Send(res)
		return
	}

	//获取职业
	confs := config.Hero.All()
	heroId := int32(0)
	for _, v := range confs {
		if v.Job == req.Job && v.Type == 1 {
			heroId = v.Id
			break
		}
	}

	if heroId <= 0 {
		res.CODE = proto_player.LOGINERRORCODE_ERROR_NOHERO
		ctx.Send(res)
		return
	}

	//同步消息
	mainHero := pl.Hero.Hero[int32(pl.GetProp(define.PlayerPropHeroId))]

	if _, ok := pl.Hero.Hero[heroId]; !ok {
		pl.Hero.Hero[heroId] = &model.HeroOption{
			Id:          heroId,
			Star:        mainHero.Star,
			Level:       mainHero.Level,
			Stage:       mainHero.Stage,
			Exp:         mainHero.Exp,
			Skin:        mainHero.Skin,
			Cultivation: mainHero.Cultivation,
		}
	} else {
		pl.Hero.Hero[heroId].Id = heroId
		pl.Hero.Hero[heroId].Star = mainHero.Star
		pl.Hero.Hero[heroId].Level = mainHero.Level
		pl.Hero.Hero[heroId].Stage = mainHero.Stage
		pl.Hero.Hero[heroId].Exp = mainHero.Exp
		pl.Hero.Hero[heroId].Skin = mainHero.Skin
		pl.Hero.Hero[heroId].Cultivation = mainHero.Cultivation
	}

	pl.Props[define.PlayerPropJob] = int64(req.Job)
	pl.Props[define.PlayerPropHeroId] = int64(heroId)

	internal.SubItems(ctx, pl, costItems)

	res.CODE = proto_player.LOGINERRORCODE_ERROR_OK
	ctx.Send(res)
}

// 获取个人信息
func ReqGetPlayerInfoById(ctx global.IPlayer, pl *model.Player, req *proto_player.C2SGetPlayerById) {
	res := &proto_player.S2CGetPlayerById{}

	if req.Id == pl.Id {
		return
	}

	//info
	info := global.GetPlayerInfo(req.Id)
	if info == nil {
		return
	}
	cInfo := info.ToCommonPlayer()
	cInfo.Job = info.Job
	cInfo.Sex = info.Sex
	cInfo.Clan = info.ClanId
	cInfo.ClanName = info.Clan
	cInfo.ServerId = info.ServerId
	res.Info = cInfo

	//装备相关
	equips, mountId, weaId, braceId := global.GetPlayerEquipBindInfo(req.Id)
	res.EquipInfo = equips
	res.MountId = mountId
	res.Weapont = weaId
	res.BracesId = braceId

	//藏品
	collect := global.GetPlayerCollectInfo(req.Id)
	res.Collections = collect

	//布阵
	lineup := global.GetPlayerLineUpInfo(req.Id)
	res.Cards = lineup

	_lineup := lineup[int32(define.LINEUP_STAGE)]
	res.Data = global.GetBattlePlayerData(info.ToToContext(), _lineup.HeroId)

	//战力
	mainHero := global.GetMainHeroByBattleHeroData(res.Data)
	if mainHero != nil {
		cInfo.Power = global.GetBattlePower(mainHero)
	}

	ctx.Send(res)
}

// OnRet redis回调
func OnRet(ctx global.IPlayer, pl *model.Player, ret *db.RedisRet) {
	if ret.Err != nil {
		log.Error("player on ret error:%v", ret.Err)
		return
	}

	switch ret.OpType {
	default:
	}
}
