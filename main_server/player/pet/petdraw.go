package pet

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"xfx/core/common"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/logic/activity/impl"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_equip"
	"xfx/proto/proto_pet"
	"xfx/proto/proto_public"
)

func PetDrawInit(pl *model.Player) {
	pl.PetDraw = new(model.PetDraw)
	pl.PetDraw.Awards = make([]int32, 0)
	pl.PetDraw.Pools = make(map[int32]*model.PetDrawPool)
}

func PetDrawSave(pl *model.Player, isSync bool) {
	//宠物
	j, err := json.Marshal(pl.PetDraw)
	if err != nil {
		log.Error("player[%v],save Pet marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("Save pet error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerPetDraw, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func PetDrawLoad(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load pet error, no this server:%v", err)
		return
	}

	//宠物
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerPetDraw, pl.Id))
	if err != nil {
		log.Error("player[%v],load pet error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.PetDraw)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load pet unmarshal error:%v", pl.Id, err)
	}

	pl.PetDraw = m

}

// 获取初始宠物抽卡
func ReqInitPetDraw(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPetDrawInit) {
	resp := new(proto_pet.S2CPetDrawInit)
	drawPoolRefreshState(pl)
	resp.PetDrawinfo = model.ToPetDrawProto(pl.PetDraw)
	ctx.Send(resp)
}

// 刷新卡池的状态
func drawPoolRefreshState(pl *model.Player) {
	confs := config.DrawPool.All()
	for k, v := range pl.PetDraw.Pools {
		conf := confs[int64(v.PoolId)]
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(conf.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			continue
		}
		if time.Now().Unix() >= endTime.Unix() && conf.ActivityType == 2 && conf.Type == 1 {
			delete(pl.PetDraw.Pools, k)
		}
	}

	//判断有没有新加的
	for _, v := range confs {
		if v.Type != define.CARDPOOL_PET {
			continue
		}

		if _, ok := pl.PetDraw.Pools[v.Id]; ok {
			continue
		}

		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			continue
		}
		if v.ActivityType == 2 {
			endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.EndTime), time.Local)
			if err != nil {
				log.Error("checkCfg parse endTime err:%v", err)
				continue
			}
			if time.Now().Unix() < startTime.Unix() || time.Now().Unix() >= endTime.Unix() {
				continue
			}
		}

		//添加新的活动
		pl.PetDraw.Pools[v.Id] = &model.PetDrawPool{
			PoolId:   v.Id,
			StarTime: startTime.Unix(),
			DrawNum:  0,
		}
	}
}

// 宠物抽卡
func ReqDrawPet(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPetDrawCard) {
	res := &proto_pet.S2CPetDrawCard{}
	//判断一下卡池
	drawPoolRefreshState(pl)

	if _, ok := pl.PetDraw.Pools[req.PoolId]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	//消耗道具
	costItems := make(map[int32]int32)
	conf := config.DrawPool.All()[int64(req.PoolId)]
	if conf.Param == int32(define.PetDrawPoolType_Normal) {
		costItems[define.ItemIdYueshi] = req.Count
	} else {
		costItems[define.ItemIdShenshi] = req.Count
	}

	if !internal.CheckItemsEnough(pl, costItems) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	pool := pl.PetDraw.Pools[req.PoolId]

	resp, err := ctx.Invoke("Recruit", "RecruitPet", int32(define.CARDPOOL_PET), int32(conf.Param), req.Count, pl.PetDraw.BdCount)
	if err != nil {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
	}

	//扣道具
	internal.SubItems(ctx, pl, costItems)
	rect := resp.(*model.RecruitResp)

	//增加次数
	pool.DrawNum += req.Count
	pl.PetDraw.BdCount += rect.BdNum
	if conf.Param == int32(define.PetDrawPoolType_Normal) {
		pl.PetDraw.Score += req.Count
	} else {
		pl.PetDraw.Score += req.Count * 5
	}

	var items []conf2.ItemE
	resItems := make([]*proto_public.Item, 0)
	for _, v := range rect.Ids {
		conf_pet := config.PetDrawPool.All()[int64(v)]
		if conf_pet.Type == 3 {
			items = append(items, conf2.ItemE{
				ItemType: define.ItemTypeItem,
				ItemId:   conf_pet.Value,
				ItemNum:  conf_pet.Num,
			})

			resItems = append(resItems, &proto_public.Item{
				ItemId:   conf_pet.Value,
				ItemNum:  conf_pet.Num,
				ItemType: define.ItemTypeItem,
			})

			pool.Recores = append(pool.Recores, &model.DrawPetRecord{
				Id:   conf_pet.Value,
				Num:  conf_pet.Num,
				Type: define.ItemTypeItem,
			})
		} else if conf_pet.Type == 5 {
			if _, ok := pl.Collection.Collections[conf_pet.Value]; !ok {
				pl.Collection.Collections[conf_pet.Value] = new(model.CollectionOption)
				pl.Collection.Collections[conf_pet.Value].Id = conf_pet.Value
			} else {
				//转成碎片
				items := make(map[int32]int32)
				conf1 := config.Collection.All()[int64(conf_pet.Value)]
				confItem := config.Item.All()[int64(conf1.Fragment)]
				items[conf1.Fragment] += conf_pet.Num * confItem.CompositeNeed
				if len(items) > 0 {
					internal.AddItems(ctx, pl, items, false)
					resItems = append(resItems, &proto_public.Item{
						ItemId:   conf1.Fragment,
						ItemNum:  conf_pet.Num * confItem.CompositeNeed,
						ItemType: define.ItemTypeItem,
					})
				}
				continue
			}

			collection := pl.Collection.Collections[conf_pet.Value]
			pl.Collection.Collections[conf_pet.Value] = collection

			pushCol := make(map[int32]*proto_equip.CollectionOption)
			pushCol[collection.Id] = &proto_equip.CollectionOption{
				Id:   collection.Id,
				Star: 0,
			}

			//推送
			pushRes := &proto_equip.PushCollectionChange{
				Collection: pushCol,
			}

			ctx.Send(pushRes)

			resItems = append(resItems, &proto_public.Item{
				ItemId:   conf_pet.Value,
				ItemNum:  conf_pet.Num,
				ItemType: define.ItemTypeCollect,
			})

			pool.Recores = append(pool.Recores, &model.DrawPetRecord{
				Id:   conf_pet.Value,
				Num:  conf_pet.Num,
				Type: define.ItemTypeCollect,
			})
			//通告相关
			internal.SyncNotice_DrawCardPet(ctx, pl, conf.Type, conf_pet.Type, conf_pet.Value)
		}
	}

	pl.PetDraw.Pools[req.PoolId] = pool
	//添加道具
	bag.AddAward(ctx, pl, items, false)

	res.PetDrawinfo = model.ToPetDrawProto(pl.PetDraw)
	res.Ids = resItems
	res.PoolId = req.PoolId
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// ReqpetStageAward 请求宠物奖励
func ReqpetStageAward(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SGetPetDrawCardStageAward) {
	res := &proto_pet.S2CGetDrawCardStageAward{}

	if len(pl.PetDraw.Pools) <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	conf := config.DrawPool.All()[int64(req.PoolId)]
	if conf.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	confStage := config.DrawStageAward.All()[int64(conf.StageId)]
	if confStage.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	//判断领取没
	if common.IsHaveValueIntArray(confStage.Progress, req.Progress) == false {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if common.IsHaveValueIntArray(pl.PetDraw.Awards, req.Progress) {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	index := 0
	for k, v := range confStage.Progress {
		if v == req.Progress {
			index = k
			break
		}
	}

	awardstr := confStage.Award[index]
	awards := strings.Split(awardstr, ",")

	var items []conf2.ItemE
	items = append(items, conf2.ItemE{
		ItemType: int32(common.StringToInt64(awards[2])),
		ItemId:   int32(common.StringToInt64(awards[0])),
		ItemNum:  int32(common.StringToInt64(awards[1])),
	})
	//添加道具
	bag.AddAward(ctx, pl, items, true)
	pl.PetDraw.Awards = append(pl.PetDraw.Awards, req.Progress)

	res.Rewards = pl.PetDraw.Awards
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// ReqPetRecorde 请求宠物记录
func ReqPetRecorde(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SGetPetRecord) {
	res := &proto_pet.S2CGetPetRecord{}

	//获取卡池
	conf := config.DrawPool.All()
	if _, ok := conf[int64(req.PoolId)]; !ok {
		ctx.Send(res)
		return
	}

	if _, ok := pl.PetDraw.Pools[req.PoolId]; !ok {
		ctx.Send(res)
		return
	}

	if pl.PetDraw.Pools == nil {
		ctx.Send(res)
		return
	}

	arr := make([]*proto_pet.PetRecord, 0)
	for i := 0; i < len(pl.PetDraw.Pools[req.PoolId].Recores); i++ {
		arr = append(arr, &proto_pet.PetRecord{
			Id:  pl.PetDraw.Pools[req.PoolId].Recores[i].Id,
			Num: pl.PetDraw.Pools[req.PoolId].Recores[i].Num,
		})
	}

	res.PetRecords = arr
	ctx.Send(res)
}

// --------------------------------召唤--------------------------
// ReqPetCall 请求宠物召唤
func ReqPetCall(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPetCall) {
	res := new(proto_pet.S2CPetCall)

	//判断材料够不够
	if pl.Bag.Items[define.ItemIdPetZhaohuan] <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	awards := make([]conf2.ItemE, 0)
	cost := make(map[int32]int32)
	//随机召唤
	if req.Type == define.PetZhaohuanType_Range {
		costNum := config.Global.Get().PetDrawCardCostNum
		if pl.Bag.Items[define.ItemIdPetZhaohuan] <= costNum {
			res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
			ctx.Send(res)
			return
		}

		PetId := int32(0)
		//获取
		confs := config.Pet.All()
		for _, v := range confs {
			if v.Type == define.PetType_Shen {
				PetId = v.Id
				break
			}
		}

		if PetId <= 0 {
			res.Code = proto_public.CommonErrorCode_ERR_NoConfig
			ctx.Send(res)
			return
		}

		cost[define.ItemIdPetZhaohuan] = costNum

		awards = append(awards, conf2.ItemE{
			ItemType: define.ItemTypePet,
			ItemNum:  1,
			ItemId:   PetId,
		})
	} else if req.Type == define.PetZhaohuanType_Point {
		if req.Id <= 0 {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}

		//判断碎片够不够
		for k, v := range req.CostIds {
			cost[k] = v
		}

		if !internal.CheckItemsEnough(pl, cost) {
			res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
			ctx.Send(res)
			return
		}

		//灵宠碎片
		num := int32(0)
		if req.CostType == 1 {
			costNum := config.Global.Get().PetDrawCardCostNum
			if pl.Bag.Items[define.ItemIdPetZhaohuan] <= costNum {
				res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
				ctx.Send(res)
				return
			}

			for k, v := range cost {
				conf := config.Item.All()[int64(k)]
				confpet := config.Pet.All()[int64(conf.CompositeItem)]
				if confpet.Type != define.PetType_Ling {
					log.Debug("传入的碎片类型不对")
					res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
					ctx.Send(res)
					return
				}
				num += v
			}

			if num < config.Global.Get().PetDrawCardCostLingPetNum {
				log.Debug("传入的碎片数量不够")
				res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
				ctx.Send(res)
				return
			}
			cost[define.ItemIdPetZhaohuan] = costNum
		} else if req.CostType == 2 { //神宠碎片
			costNum := config.Global.Get().PetDrawCardCostNum
			dikouNum := config.Global.Get().PetDrawCardDikouNum
			if pl.Bag.Items[define.ItemIdPetZhaohuan] <= costNum-dikouNum {
				res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
				ctx.Send(res)
				return
			}

			for k, v := range cost {
				conf := config.Item.All()[int64(k)]
				confpet := config.Pet.All()[int64(conf.CompositeItem)]
				if confpet.Type != define.PetType_Shen {
					log.Debug("传入的灵宠碎片类型不对")
					res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
					ctx.Send(res)
					return
				}
				num += v
			}

			if num < config.Global.Get().PetDrawCardCostShenPetNum {
				log.Debug("传入的神宠碎片数量不够")
				res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
				ctx.Send(res)
				return
			}
			cost[define.ItemIdPetZhaohuan] = costNum - dikouNum
		}

		awards = append(awards, conf2.ItemE{
			ItemType: define.ItemTypePet,
			ItemNum:  1,
			ItemId:   req.Id,
		})
	}

	if len(awards) <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	//扣除道具
	internal.SubItems(ctx, pl, cost)

	//奖励进背包
	bag.AddAward(ctx, pl, awards, true)
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}
