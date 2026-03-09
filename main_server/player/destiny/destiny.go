package destiny

import (
	"encoding/json"
	"fmt"
	"sort"
	"xfx/pkg/utils"
	"xfx/core/config"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_destiny"
)

func Init(pl *model.Player) {
	pl.Destiny = new(model.Destiny)
	pl.Destiny.Ids = make([]int32, 0)
	pl.Destiny.SelfIds = make([]int32, 0)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Destiny)
	if err != nil {
		log.Error("player[%v],save Destiny marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerDestiny, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerDestiny, pl.Id))
	if err != nil {
		log.Error("player[%v],load bag error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Destiny)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Equip unmarshal error:%v", pl.Id, err)
	}

	pl.Destiny = m
}

// 请求天命
func ReqInitDestiny(ctx global.IPlayer, pl *model.Player, req *proto_destiny.C2SReqDestinyInit) {
	res := &proto_destiny.S2CRespDestinyInit{}
	res.Ids = pl.Destiny.Ids
	res.Level = pl.Destiny.Level
	res.StageIds = pl.Destiny.SelfIds
	ctx.Send(res)
}

// 解锁天命
func ReqUnLockDestiny(ctx global.IPlayer, pl *model.Player, req *proto_destiny.C2SReqUnLockDestiny) {
	res := &proto_destiny.S2CRespUnLockDestiny{}

	//是否解锁
	if utils.ContainsInt32(pl.Destiny.Ids, req.Id) {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_AlGetAward
		ctx.Send(res)
		return
	}

	confs := config.DestinyLevel.All()
	if _, ok := confs[int64(req.Id)]; !ok {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_NOConfig
		ctx.Send(res)
		return
	}

	conf := confs[int64(req.Id)]

	//前置
	if conf.FrontId > 0 {
		if utils.ContainsInt32(pl.Destiny.Ids, conf.FrontId) == false {
			res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_ConditionNo
			ctx.Send(res)
			return
		}
	}

	cost := make(map[int32]int32)
	for _, v := range conf.CostItem {
		cost[v.ItemId] = v.ItemNum
	}

	//判断材料
	if internal.CheckItemsEnough(pl, cost) == false {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_NumNotEnough
		ctx.Send(res)
		return
	}

	internal.SubItems(ctx, pl, cost)

	pl.Destiny.Level = conf.Level
	pl.Destiny.Ids = append(pl.Destiny.Ids, req.Id)
	res.Code = proto_destiny.ERRORCODEDESTINY_ERR_Ok
	res.Ids = pl.Destiny.Ids
	res.Level = pl.Destiny.Level
	ctx.Send(res)
}

// 解锁天命独立属性
func ReqUnLockSelfDestiny(ctx global.IPlayer, pl *model.Player, req *proto_destiny.C2SReqUnlockSelfDestiny) {
	res := &proto_destiny.S2CRespUnlockSelfDestiny{}

	//是否解锁
	if utils.ContainsInt32(pl.Destiny.SelfIds, req.Id) {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_AlGetAward
		ctx.Send(res)
		return
	}

	confs := config.DestinyStage.All()
	if _, ok := confs[int64(req.Id)]; !ok {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_NOConfig
		ctx.Send(res)
		return
	}

	conf := confs[int64(req.Id)]

	cost := make(map[int32]int32)
	for _, v := range conf.CostItem {
		cost[v.ItemId] = v.ItemNum
	}

	//判断材料
	if internal.CheckItemsEnough(pl, cost) == false {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_NumNotEnough
		ctx.Send(res)
		return
	}

	if conf.Id > pl.Destiny.Level {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_ConditionNo
		ctx.Send(res)
		return
	}

	internal.SubItems(ctx, pl, cost)
	pl.Destiny.SelfIds = append(pl.Destiny.SelfIds, req.Id)
	res.Code = proto_destiny.ERRORCODEDESTINY_ERR_Ok
	res.StageIds = pl.Destiny.SelfIds
	ctx.Send(res)
}

// 一键解锁天命
func ReqOneKeyUnLockDestiny(ctx global.IPlayer, pl *model.Player, req *proto_destiny.C2SReqOneKeyUnLock) {
	res := &proto_destiny.S2CRespOneKeyUnLock{}

	//等级
	cost := make(map[int32]int32)
	ids := pl.Destiny.Ids
	level := int32(0)
	cacheId := make([]int32, 0)
	confIds := make([]int32, 0)
	confs := config.DestinyLevel.All()
	for _, v := range confs {
		confIds = append(confIds, v.Id)
	}
	sort.Slice(confIds, func(i, j int) bool {
		return confIds[i] < confIds[j]
	})

	for i := 0; i < len(confIds); i++ {
		cid := confIds[i]
		conf := config.DestinyLevel.All()[int64(cid)]

		//是否解锁
		if utils.ContainsInt32(ids, cid) {
			continue
		}

		//前置
		if conf.FrontId > 0 {
			if utils.ContainsInt32(ids, conf.FrontId) == false && utils.ContainsInt32(cacheId, conf.FrontId) == false {
				continue
			}
		}

		for _, v := range conf.CostItem {
			cost[v.ItemId] += v.ItemNum
		}

		//判断材料
		if internal.CheckItemsEnough(pl, cost) == false {
			break
		}

		cacheId = append(cacheId, cid)
		level = conf.Level
	}

	if len(cacheId) > 0 {
		internal.SubItems(ctx, pl, cost)
		pl.Destiny.Ids = append(pl.Destiny.Ids, cacheId...)
		pl.Destiny.Level = level
	}

	//阶级
	cost1 := make(map[int32]int32)
	ids1 := make([]int32, 0)
	confIds = make([]int32, 0)
	confs1 := config.DestinyStage.All()
	for _, v := range confs1 {
		confIds = append(confIds, v.Id)
	}
	sort.Slice(confIds, func(i, j int) bool {
		return confIds[i] < confIds[j]
	})

	for i := 0; i < len(confIds); i++ {
		cid := confIds[i]
		conf := config.DestinyStage.All()[int64(cid)]

		//是否解锁
		if utils.ContainsInt32(pl.Destiny.SelfIds, cid) {
			continue
		}

		//等级
		if conf.Id > pl.Destiny.Level {
			break
		}

		for _, v := range conf.CostItem {
			cost1[v.ItemId] += v.ItemNum
		}

		//判断材料
		if internal.CheckItemsEnough(pl, cost1) == false {
			break
		}

		ids1 = append(ids1, conf.Id)
	}

	if len(ids1) > 0 {
		internal.SubItems(ctx, pl, cost1)
		pl.Destiny.SelfIds = append(pl.Destiny.SelfIds, ids1...)
	}

	res.Code = proto_destiny.ERRORCODEDESTINY_ERR_Ok
	res.Ids = pl.Destiny.Ids
	res.Level = pl.Destiny.Level
	res.StageIds = pl.Destiny.SelfIds
	log.Debug("一键升级天命成功")
	ctx.Send(res)
}
