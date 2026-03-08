package divine

import (
	"encoding/json"
	"fmt"
	"sort"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_equip"
)

func Init(pl *model.Player) {
	pl.Divine = new(model.Divine)
	pl.Divine.Divines = make(map[int32]map[int32]*model.DivineOption)
	pl.Divine.Learning = make(map[int32]*model.LearningOption)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Divine)
	if err != nil {
		log.Error("player[%v],save Divine marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save Divine error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerDivine, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load Divine error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerDivine, pl.Id))
	if err != nil {
		log.Error("player[%v],load Divine error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Divine)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Equip unmarshal error:%v", pl.Id, err)
	}

	pl.Divine = m
}

// 请求领悟心得
func ReqInitDivine(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SInitDivine) {
	res := &proto_equip.S2CInitDivine{}
	res.DivineIndexItems = model.ToDivineProto(pl.Divine.Divines)
	res.LearningOptions = model.ToLearningProto(pl.Divine.Learning)
	ctx.Send(res)
}

// 请求升级
func ReqDivineLevelUp(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SReqDivineUpLevel) {
	res := &proto_equip.S2CRespDivineUpLevel{}

	if req.Index <= 0 || req.Index > 6 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	conf := config.Divine.All()[int64(req.Id)]
	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//判断类型
	if conf.Type != 1 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//获取序列
	if _, ok := pl.Divine.Divines[req.Index]; !ok {
		pl.Divine.Divines[req.Index] = make(map[int32]*model.DivineOption)
	}

	//判断存不存在
	if _, ok := pl.Divine.Divines[req.Index][req.Id]; !ok {
		//判断前置
		front := conf.FrontNode
		for _, v := range front {
			if _, ok = pl.Divine.Divines[req.Index][v]; !ok {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
				ctx.Send(res)
				return
			} else {
				conf1 := config.Divine.All()[int64(v)]
				if conf1.Type == 1 && pl.Divine.Divines[req.Index][v].Level < conf.Level {
					res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
					ctx.Send(res)
					return
				}
			}
		}

		pl.Divine.Divines[req.Index][req.Id] = new(model.DivineOption)
		pl.Divine.Divines[req.Index][req.Id].Id = req.Id
	}

	data := pl.Divine.Divines[req.Index][req.Id]

	//判断满级没有
	if conf.Type == 1 && data.Level >= conf.Level {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_MAX
		ctx.Send(res)
		return
	}

	//判断材料够不够
	cost := make(map[int32]int32)
	level := data.Level
	costNum := conf.LevelCost[level]
	cost[define.ItemIdLiLian] = costNum

	if internal.CheckItemsEnough(pl, cost) == false {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	internal.SubItems(ctx, pl, cost)

	data.Level += 1
	pl.Divine.Divines[req.Index][req.Id] = data
	log.Debug("INDEX:%v", req.Index)
	//推送变化
	pushChangeDivine(ctx, pl, req.Index, req.Id)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求一键升级
func ReqOneKeyDivineLevelUp(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SOneKeyUpLevelDivine) {
	res := &proto_equip.S2COneKeyUpLevelDivine{}
	resp := &proto_equip.PushDivineChange{}

	if req.Index <= 0 || req.Index > 6 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	confs := config.Divine.All()
	//获取槽位配置
	conf := make([]conf2.Divine, 0)
	for _, v := range confs {
		if v.Slot == req.Index {
			conf = append(conf, v)
		}
	}

	//排序
	sort.Slice(conf, func(a, b int) bool {
		return conf[a].Id < conf[b].Id
	})

	//获取序列
	if _, ok := pl.Divine.Divines[req.Index]; !ok {
		pl.Divine.Divines[req.Index] = make(map[int32]*model.DivineOption)
	}

	cost := make(map[int32]int32)
	cost[define.ItemIdLiLian] = 0

	resp.DivineIndexItems = make(map[int32]*proto_equip.DivineIndexItem)
	divineItemItems := make(map[int32]*proto_equip.DivineItem)
	for i := 0; i < len(conf); i++ {
		//当前满级
		id := conf[i].Id

		divineOption := new(model.DivineOption)
		//判断存不存在
		if _, ok := pl.Divine.Divines[req.Index][id]; !ok {
			divineOption.Id = id
		} else {
			divineOption = pl.Divine.Divines[req.Index][id]
		}

		if conf[i].Type == 1 && divineOption.Level >= conf[i].Level {
			continue
		}

		//判断类型
		if conf[i].Type == 1 {
			//等级
			for k := divineOption.Level; k < conf[i].Level; k++ {
				costNum := conf[i].LevelCost[k]
				cost[define.ItemIdLiLian] += costNum
				if internal.CheckItemsEnough(pl, cost) == false {
					cost[define.ItemIdLiLian] -= costNum
					break
				}
				divineOption.Level = k + 1
				pl.Divine.Divines[req.Index][id] = divineOption
				item := model.ToDivineSingleProto(pl.Divine.Divines[req.Index][id])
				for o, l := range item {
					divineItemItems[o] = l
				}
			}
		} else {
			//解锁
			cost[define.ItemIdLiLian] += conf[i].UnLockCost
			if internal.CheckItemsEnough(pl, cost) == false {
				cost[define.ItemIdLiLian] -= conf[i].UnLockCost
				break
			}

			pl.Divine.Divines[req.Index][id] = divineOption
			item := model.ToDivineSingleProto(pl.Divine.Divines[req.Index][id])
			for o, l := range item {
				divineItemItems[o] = l
			}
		}
	}
	log.Debug(" cost[define.ItemIdLiLian]:%v", cost[define.ItemIdLiLian])
	if cost[define.ItemIdLiLian] > 0 {
		internal.SubItems(ctx, pl, cost)

		resp.DivineIndexItems[req.Index] = new(proto_equip.DivineIndexItem)
		resp.DivineIndexItems[req.Index].Index = req.Index
		resp.DivineIndexItems[req.Index].DivineItems = divineItemItems
		ctx.Send(resp)
	}

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求解锁
func ReqDivineUnLock(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SReqUnLockDivine) {
	res := &proto_equip.S2CRespUnLockDivine{}

	if req.Index <= 0 || req.Index > 6 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	conf := config.Divine.All()[int64(req.Id)]
	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//判断类型
	if conf.Type != 2 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//获取序列
	if _, ok := pl.Divine.Divines[req.Index]; !ok {
		pl.Divine.Divines[req.Index] = make(map[int32]*model.DivineOption)
	}

	//已经解锁
	if _, ok := pl.Divine.Divines[req.Index][req.Id]; ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	//判断前置
	front := conf.FrontNode
	for _, v := range front {
		if _, ok := pl.Divine.Divines[req.Index][v]; !ok {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
			ctx.Send(res)
			return
		} else {
			if pl.Divine.Divines[req.Index][v].Level < conf.Level {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
				ctx.Send(res)
				return
			}
		}
	}

	pl.Divine.Divines[req.Index][req.Id] = new(model.DivineOption)
	data := pl.Divine.Divines[req.Index][req.Id]

	//判断材料够不够
	cost := make(map[int32]int32)
	costNum := conf.UnLockCost
	cost[define.ItemIdLiLian] = costNum

	if internal.CheckItemsEnough(pl, cost) == false {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	internal.SubItems(ctx, pl, cost)

	data.Id = req.Id
	pl.Divine.Divines[req.Index][req.Id] = data

	//推送变化
	pushChangeDivine(ctx, pl, req.Index, req.Id)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求装配
func ReqDivineWear(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SReqWearDivine) {
	res := &proto_equip.S2CRespWearDivine{}

	if req.Index <= 0 || req.Index > 6 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	conf := config.Divine.All()[int64(req.Id)]
	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//判断类型
	if conf.Type != 2 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//获取序列
	if _, ok := pl.Divine.Divines[req.Index]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	if _, ok := pl.Divine.Divines[req.Index][req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	//判断是否有心得
	if _, ok := pl.Divine.Learning[req.SId]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	//获取心得类型
	conf1 := config.Learning.All()[int64(req.SId)]
	if conf1.Type == 1 {
		if req.Index <= 0 || req.Index > 3 {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}
	} else if conf1.Type == 2 {
		if req.Index <= 3 || req.Index > 6 {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}
	}

	//判断是否装备了心得
	sids := pl.Divine.Divines[req.Index]
	for _, v := range sids {
		if v.Sid == req.SId {
			log.Debug("已经装备了")
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}
	}

	data := pl.Divine.Divines[req.Index][req.Id]
	data.Sid = req.SId
	pl.Divine.Divines[req.Index][req.Id] = data

	//推送变化
	pushChangeDivine(ctx, pl, req.Index, req.Id)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求卸载
func ReqDivineRemove(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SReqRemoveDivine) {
	res := &proto_equip.S2CRespRemoveDivine{}

	if req.Index <= 0 || req.Index > 6 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	conf := config.Divine.All()[int64(req.Id)]
	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//判断类型
	if conf.Type != 2 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//获取序列
	if _, ok := pl.Divine.Divines[req.Index]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	if _, ok := pl.Divine.Divines[req.Index][req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	//判断是否装备了心得
	data := pl.Divine.Divines[req.Index][req.Id]
	if data.Sid <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	data.Sid = 0
	pl.Divine.Divines[req.Index][req.Id] = data

	//推送变化
	pushChangeDivine(ctx, pl, req.Index, req.Id)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求重置
func ReqDivineReset(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SReqResetDivine) {
	res := &proto_equip.S2CReqResetDivine{}

	if req.Index <= 0 || req.Index > 6 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	confs := config.Divine.All()
	//获取槽位配置
	conf := make([]conf2.Divine, 0)
	for _, v := range confs {
		if v.Slot == req.Index && req.Id <= v.Id {
			conf = append(conf, v)
		}
	}

	//排序
	sort.Slice(conf, func(a, b int) bool {
		return conf[a].Id < conf[b].Id
	})

	//获取序列
	if _, ok := pl.Divine.Divines[req.Index]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	if _, ok := pl.Divine.Divines[req.Index][req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	award := make([]conf2.ItemE, 0)
	for i := 0; i < len(conf); i++ {
		//当前满级
		id := conf[i].Id
		log.Debug("重置领悟ID:%v", id)
		//判断存不存在
		if _, ok := pl.Divine.Divines[req.Index][id]; !ok {
			continue
		} else {
			divineOption := pl.Divine.Divines[req.Index][id]
			//判断类型
			if conf[i].Type == 1 {
				//等级
				for k := int32(0); k < divineOption.Level; k++ {
					if k >= conf[i].Level {
						delete(pl.Divine.Divines[req.Index], id)
						continue
					}
					award = append(award, conf2.ItemE{
						ItemType: define.ItemTypeItem,
						ItemId:   define.ItemIdLiLian,
						ItemNum:  conf[i].LevelCost[k],
					})
				}
				delete(pl.Divine.Divines[req.Index], id)
			} else {
				//解锁
				award = append(award, conf2.ItemE{
					ItemType: define.ItemTypeItem,
					ItemId:   define.ItemIdLiLian,
					ItemNum:  conf[i].UnLockCost,
				})
				delete(pl.Divine.Divines[req.Index], id)
			}
		}
	}

	if len(award) > 0 {
		bag.AddAward(ctx, pl, award, true)

		divMap := pl.Divine.Divines[req.Index]
		n := make(map[int32]*proto_equip.DivineItem, 0)
		for key, l := range divMap {
			n[key] = &proto_equip.DivineItem{
				ID:    l.Id,
				Level: l.Level,
				SId:   l.Sid,
			}
		}
		res.DivineItems = n
	}

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求合成
func ReqLearnCompose(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SReqDivineCompose) {
	res := &proto_equip.S2SRespDivineCompose{}

	//数量
	if len(req.Id) != 4 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
		ctx.Send(res)
		return
	}

	//特殊合成
	if req.Type == 2 {
		//读取配置
		conf, ok := config.LearningCompose.Find(int64(req.CfgId))
		if !ok {
			log.Error("register new activity get config id error:%v", req.CfgId)
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
			ctx.Send(res)
			return
		}

		//判断心得存不存在
		for _, v := range conf.Condition {
			if _, ok := pl.Divine.Learning[v]; !ok {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
				ctx.Send(res)
				return
			} else {
				if pl.Divine.Learning[v].Num <= 0 {
					res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
					ctx.Send(res)
					return
				}
			}
		}

		changeIds := make([]int32, 0)
		//删除之前
		for _, v := range conf.Condition {
			pl.Divine.Learning[v].Num -= 1
			if pl.Divine.Learning[v].Num <= 0 {
				delete(pl.Divine.Learning, v)
			}
			changeIds = append(changeIds, v)
		}

		//增加新增加
		if _, ok := pl.Divine.Learning[conf.Value]; !ok {
			pl.Divine.Learning[conf.Value] = new(model.LearningOption)
			pl.Divine.Learning[conf.Value].Id = conf.Value
		}
		pl.Divine.Learning[conf.Value].Num += 1
		changeIds = append(changeIds, conf.Value)

		//推送
		pushChangeLearning(ctx, pl, changeIds)
		res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
		ctx.Send(res)
		return
	}

	//id转map
	ids := make(map[int32]int32)
	for _, v := range req.Id {
		if _, ok := ids[v]; !ok {
			ids[v] = 0
		}
		ids[v] += 1
	}

	//判断品质
	rare := int32(0)
	for k, _ := range ids {
		conf := config.Learning.All()[int64(k)]
		if rare == 0 {
			rare = conf.Rare
		} else {
			if conf.Rare != rare {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
				ctx.Send(res)
				return
			}
		}
	}

	//判断心得存不存在
	for k, v := range ids {
		if _, ok := pl.Divine.Learning[k]; !ok {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
			ctx.Send(res)
			return
		} else {
			if pl.Divine.Learning[k].Num < v {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
				ctx.Send(res)
				return
			}
		}
	}

	//处理合成
	logicCompose := func(ids map[int32]int32, rare int32, tye int32) (bool, int32) {
		composeId := int32(0)
		//定向合成
		if tye == 1 {
			confs := config.Learning.All()
			for _, v := range confs {
				if v.Rare == rare+1 {
					composeId = v.Id
					break
				}
			}
		} else if tye == 3 {
			ok := false
			for k, v := range ids {
				if v >= 2 {
					ok = true
					composeId = k
					break
				}
			}

			if !ok {
				return false, 0
			}
		}

		if composeId <= 0 {
			return false, 0
		}

		changeIds := make([]int32, 0)
		//删除之前
		for k, v := range ids {
			pl.Divine.Learning[k].Num -= v
			if pl.Divine.Learning[k].Num <= 0 {
				delete(pl.Divine.Learning, k)
			}
			changeIds = append(changeIds, k)
		}

		//增加新增加
		if _, ok := pl.Divine.Learning[composeId]; !ok {
			pl.Divine.Learning[composeId] = new(model.LearningOption)
			pl.Divine.Learning[composeId].Id = composeId
		}
		pl.Divine.Learning[composeId].Num += 1
		changeIds = append(changeIds, composeId)

		//推送
		pushChangeLearning(ctx, pl, changeIds)
		return true, composeId
	}

	//史诗以下品质
	if req.Type == 1 {
		conf := config.Learning.All()[int64(req.Id[0])]
		if conf.Rare > 4 {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
			ctx.Send(res)
			return
		}

		isOk, id := logicCompose(ids, conf.Rare, 1)
		if !isOk {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}

		//奖励弹窗
		internal.PushLearningReawardPop(ctx, pl, []int32{id})

	} else if req.Type == 3 { //卓越以上合成
		conf := config.Learning.All()[int64(req.Id[0])]
		if conf.Rare < 5 {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_CONDITIONNOTENOUGH
			ctx.Send(res)
			return
		}

		isOk, id := logicCompose(ids, conf.Rare, 3)
		if !isOk {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}

		//奖励弹窗
		internal.PushLearningReawardPop(ctx, pl, []int32{id})
	}

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求一键合成
func ReqOneKeyLearnCompose(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SReqOnekeyDivineCompose) {
	res := &proto_equip.S2CReqOnekeyDivineCompose{}

	//分类集合
	ids := make(map[int32]map[int32]int32)
	for _, v := range pl.Divine.Learning {
		conf := config.Learning.All()[int64(v.Id)]
		if _, ok := ids[conf.Rare]; !ok {
			ids[conf.Rare] = make(map[int32]int32)
		}

		if _, ok := ids[conf.Rare][v.Id]; !ok {
			ids[conf.Rare][v.Id] = 0
		}

		ids[conf.Rare][v.Id] += v.Num
	}

	//处理合成
	logicCompose := func(idds [][]int32, rare int32, tye int32) ([]int32, []int32) {
		changeIds := make([]int32, 0)
		getIds := make([]int32, 0)
		for i := 0; i < len(idds); i++ {
			//id转map
			ids := make(map[int32]int32)
			for _, v := range idds[i] {
				if _, ok := ids[v]; !ok {
					ids[v] = 0
				}
				ids[v] += 1
			}

			composeId := int32(0)
			//定向合成
			if tye == 1 {
				confs := config.Learning.All()
				for _, v := range confs {
					if v.Rare == rare+1 {
						composeId = v.Id
						break
					}
				}
			} else if tye == 3 {
				ok := false
				for k, v := range ids {
					if v >= 2 {
						ok = true
						composeId = k
						break
					}
				}

				if !ok {
					return getIds, changeIds
				}
			}

			if composeId <= 0 {
				return getIds, changeIds
			}

			//删除之前
			for k, v := range ids {
				pl.Divine.Learning[k].Num -= v
				if pl.Divine.Learning[k].Num <= 0 {
					delete(pl.Divine.Learning, k)
				}
				changeIds = append(changeIds, k)
			}

			//增加新增加
			if _, ok := pl.Divine.Learning[composeId]; !ok {
				pl.Divine.Learning[composeId] = new(model.LearningOption)
				pl.Divine.Learning[composeId].Id = composeId
			}
			pl.Divine.Learning[composeId].Num += 1
			changeIds = append(changeIds, composeId)
			getIds = append(getIds, composeId)
		}

		return getIds, changeIds
	}

	//奖励
	getIds := make([]int32, 0)
	changeIds := make([]int32, 0)
	for _, v := range req.Type {
		//过滤
		if v <= 0 || v > 3 {
			continue
		}

		//史诗-卓越
		if v == 1 {
			comIds := make([][]int32, 0)
			if _, ok := ids[v]; ok {
				group := make([]int32, 0)
				for l, n := range ids[v] {
					for index := int32(0); index < n; index++ {
						group = append(group, l)

						if len(group) >= 4 {
							comIds = append(comIds, group)
							group = make([]int32, 0)
						}
					}
				}

				_getIds, _changeIds := logicCompose(comIds, 4, 1)
				changeIds = append(changeIds, _changeIds...)
				getIds = append(getIds, _getIds...)
			}
		} else {
			comIds := make([][]int32, 0)
			if _, ok := ids[v]; ok {
				group := make([]int32, 0)
				for l, n := range ids[v] {
					for index := int32(0); index < n; index++ {
						group = append(group, l)

						if len(group) >= 4 {
							if ok, _ := utils.HasDuplicateInt32(group); ok {
								comIds = append(comIds, group)
							}
							group = make([]int32, 0)
						}
					}
				}

				_getIds, _changeIds := logicCompose(comIds, 4, 1)
				changeIds = append(changeIds, _changeIds...)
				getIds = append(getIds, _getIds...)
			}
		}
	}

	//奖励弹窗
	internal.PushLearningReawardPop(ctx, pl, getIds)
	pushChangeLearning(ctx, pl, changeIds)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 推送变化
func pushChangeDivine(ctx global.IPlayer, pl *model.Player, index, id int32) {
	res := &proto_equip.PushDivineChange{}
	item := model.ToDivineSingleProto(pl.Divine.Divines[index][id])
	res.DivineIndexItems = make(map[int32]*proto_equip.DivineIndexItem)
	res.DivineIndexItems[index] = new(proto_equip.DivineIndexItem)
	res.DivineIndexItems[index].Index = index
	res.DivineIndexItems[index].DivineItems = item
	log.Debug("INDEX:%v", res)
	ctx.Send(res)
}

// 推送变化
func pushChangeLearning(ctx global.IPlayer, pl *model.Player, id []int32) {
	res := &proto_equip.PushLearningOptionChange{}
	res.LearningOptions = make(map[int32]*proto_equip.LearningOption)
	for _, v := range id {
		if _, ok := pl.Divine.Learning[v]; ok {
			res.LearningOptions[v] = &proto_equip.LearningOption{
				Id:  pl.Divine.Learning[v].Id,
				Num: pl.Divine.Learning[v].Num,
			}
		} else {
			res.LearningOptions[v] = &proto_equip.LearningOption{
				Id:  v,
				Num: 0,
			}
		}
	}
	ctx.Send(res)
}
