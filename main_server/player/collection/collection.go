package collection

import (
	"encoding/json"
	"fmt"
	"xfx/pkg/utils"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_equip"
)

func Init(pl *model.Player) {
	pl.Collection = new(model.Collection)
	pl.Collection.Collections = make(map[int32]*model.CollectionOption)
	pl.Collection.CollectionSlots = make(map[int32]map[int32]int32)
	pl.Collection.Heros = make([]int32, 5)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Collection)
	if err != nil {
		log.Error("player[%v],save Collection marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save Collection error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerCollection, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load Collection error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerCollection, pl.Id))
	if err != nil {
		log.Error("player[%v],load Collection error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Collection)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Collection unmarshal error:%v", pl.Id, err)
	}

	pl.Collection = m
}

// 请求初始藏品
func ReqInitCollection(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SInitCollection) {
	res := &proto_equip.S2CInitCollection{}

	//这里去判断一下布阵
	internal.CollectionSlotHero(ctx, pl)
	res.Collections = model.ToCollectionProto(pl.Collection.Collections)
	res.HeroIds = pl.Collection.Heros
	res.Slots = model.ToCollectionSlotProto(pl.Collection.CollectionSlots)
	ctx.Send(res)
}

// 请求藏品升星
func ReqUpStarCollection(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SCollectionUpStar) {
	res := &proto_equip.S2CCollectionUpStar{}

	//判断激活没有
	if _, ok := pl.Collection.Collections[req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
		ctx.Send(res)
		return
	}

	collection := pl.Collection.Collections[req.Id]
	conf := conf2.CollectionUpStar{}
	configs := config.CollectionUpStar.All()
	for _, v := range configs {
		if v.CollectionId == req.Id && v.Star == collection.Star {
			conf = v
			break
		}
	}

	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	cost := make(map[int32]int32)
	for _, v := range conf.UpStarCondition {
		cost[v.ItemId] = v.ItemNum
	}

	//碎片
	if conf.NeedFragNum > 0 {
		confItem := conf2.Item{}
		configItems := config.Item.All()
		for _, v := range configItems {
			if v.Type == define.BagItemTypeCollectionPiece && v.IsComposite && v.CompositeItem == req.Id {
				confItem = v
				break
			}
		}

		if confItem.Id <= 0 {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
			ctx.Send(res)
			return
		}

		cost[confItem.Id] = conf.NeedFragNum
	}

	//判断够不够
	if internal.CheckItemsEnough(pl, cost) == false {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	internal.SubItems(ctx, pl, cost)

	collection.Star += 1
	pl.Collection.Collections[req.Id] = collection

	//推送
	pushCol := make(map[int32]*proto_equip.CollectionOption)
	pushCol[collection.Id] = &proto_equip.CollectionOption{
		Id:   collection.Id,
		Star: collection.Star,
	}
	res.Collection = pushCol
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求藏品穿戴
func ReqWearCollection(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SCollectionWear) {
	res := &proto_equip.S2CCollectionWear{}

	if req.Index < 1 || req.Index > 2 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if req.Slot <= 0 || req.Slot > 5 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//判断激活没有
	if _, ok := pl.Collection.Collections[req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
		ctx.Send(res)
		return
	}

	//判断穿戴之前的有没有
	slot := int32(0)
	index := int32(0)
	for k, v := range pl.Collection.CollectionSlots {
		if slot > 0 {
			break
		}

		for a, b := range v {
			if b == req.Id {
				slot = k
				index = a
				break
			}
		}
	}

	//判断当前槽位有没有藏品
	curId := int32(0)
	if _, ok := pl.Collection.CollectionSlots[req.Slot]; ok {
		if _, ok = pl.Collection.CollectionSlots[req.Slot][req.Index]; ok {
			curId = pl.Collection.CollectionSlots[req.Slot][req.Index]
		}
	} else {
		pl.Collection.CollectionSlots[req.Slot] = make(map[int32]int32)
	}

	pl.Collection.CollectionSlots[req.Slot][req.Index] = req.Id

	if curId > 0 && slot > 0 {
		pl.Collection.CollectionSlots[slot][index] = curId
	}

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	res.Slots = model.ToCollectionSlotProto(pl.Collection.CollectionSlots)
	ctx.Send(res)
}

// 请求藏品卸下
func ReqRemoveCollection(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SCollectionRemove) {
	res := &proto_equip.S2CCollectionRemove{}

	if req.Index < 1 || req.Index > 2 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if req.Slot <= 0 || req.Slot > 5 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//判断当前槽位有没有藏品
	if _, ok := pl.Collection.CollectionSlots[req.Slot]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if _, ok := pl.Collection.CollectionSlots[req.Slot][req.Index]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	delete(pl.Collection.CollectionSlots[req.Slot], req.Index)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	res.Slots = model.ToCollectionSlotProto(pl.Collection.CollectionSlots)
	ctx.Send(res)
}

// 设置槽位英雄
func ReqSetSlotHeroCollection(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SSetCollectionSlotHero) {
	res := &proto_equip.S2CSetCollectionSlotHero{}

	lineup := pl.Lineup.LineUps[define.LINEUP_STAGE]
	for j := 0; j < len(req.HeroIds); j++ {
		heroId := req.HeroIds[j]
		if heroId == 0 {
			continue
		}

		if !utils.ContainsInt32(lineup.HeroId, heroId) {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}
	}
	pl.Collection.Heros = req.HeroIds
	res.HeroIds = req.HeroIds
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}
