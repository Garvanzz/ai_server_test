package equip

import (
	"xfx/core/common"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/proto/proto_equip"
)

// 请求背饰
func ReqInitEquipBrace(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SInitBraces) {
	res := &proto_equip.S2CInitBraces{}
	res.Braces = model.ToBraceItemProto(pl.Equip.Brace.BraceItems)
	res.Auras = model.ToBraceAuraProto(pl.Equip.Brace.BraceAuraItems)
	res.BraceTalent = model.ToBraceTalentIndexProto(pl.Equip.Brace.BraceTalentIndexs)
	res.GetAuraLevel = pl.Equip.Brace.GetAuraStageAward
	res.BraceTalentIndex = int32(pl.Equip.Brace.BraceTalentIndex)
	ctx.Send(res)
}

// 背饰灵韵升级
func ReqEquipBraceAuraUpLevel(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SLevelUpAura) {
	res := &proto_equip.S2CLevelUpAura{}

	if req.Type < 0 || req.Type > 4 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if _, ok := pl.Equip.Brace.BraceAuraItems[req.Type]; !ok {
		pl.Equip.Brace.BraceAuraItems[req.Type] = new(model.BraceAuraItem)
		pl.Equip.Brace.BraceAuraItems[req.Type].Type = req.Type
	}

	configs := config.BraceAura.All()
	var conf conf2.BraceAura
	for _, v := range configs {
		if v.Level == pl.Equip.Brace.BraceAuraItems[req.Type].Level {
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
	cost[define.ItemIdBraceAura] = req.Count

	if internal.CheckItemsEnough(pl, cost) == false {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}
	costNum := conf.CostNum[req.Type-1]
	if pl.Equip.Brace.BraceAuraItems[req.Type].Exp+req.Count >= costNum {
		pl.Equip.Brace.BraceAuraItems[req.Type].Level += 1
		pl.Equip.Brace.BraceAuraItems[req.Type].Exp = 0
		cost[define.ItemIdBraceAura] = costNum - pl.Equip.Brace.BraceAuraItems[req.Type].Exp
	} else {
		pl.Equip.Brace.BraceAuraItems[req.Type].Exp += req.Count
		cost[define.ItemIdBraceAura] = req.Count
	}

	//扣除道具
	internal.SubItems(ctx, pl, cost)
	isnewGet, newId := autoUnLockBrace(pl)
	if isnewGet {
		res.BraceItem = &proto_equip.BraceItem{
			Id:    pl.Equip.Brace.BraceItems[newId].Id,
			Num:   pl.Equip.Brace.BraceItems[newId].Num,
			Use:   pl.Equip.Brace.BraceItems[newId].IsUse,
			Level: pl.Equip.Brace.BraceItems[newId].Level,
		}
	}

	res.Auras = model.ToBraceAuraProto(pl.Equip.Brace.BraceAuraItems)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 自动解锁
func autoUnLockBrace(pl *model.Player) (bool, int32) {
	configs := config.Braces.All()
	for _, v := range configs {
		if v.UnLockAuraLevel > 0 {
			//判断等级
			isEnough := true
			for _, k := range pl.Equip.Brace.BraceAuraItems {
				if k.Level < v.UnLockAuraLevel {
					isEnough = false
					break
				}
			}

			if isEnough {
				//判断是否拥有
				if _, ok := pl.Equip.Brace.BraceItems[v.Id]; !ok {
					pl.Equip.Brace.BraceItems[v.Id] = &model.BraceItem{
						Id:    v.Id,
						Num:   1,
						Level: 1,
						IsUse: false,
					}

					return true, v.Id
				}
			}
		}
	}

	return false, 0
}

// 背饰领取共鸣等级
func ReqEquipBraceAuraStageAward(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SGetAuraStageAward) {
	res := &proto_equip.S2CGetAuraStageAward{}

	config, _ := config.BraceAuraStage.Find(int64(req.Id))
	if config.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//判断等级
	for i := int32(1); i < 5; i++ {
		if _, ok := pl.Equip.Brace.BraceAuraItems[i]; !ok {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
			ctx.Send(res)
			return
		} else {
			if pl.Equip.Brace.BraceAuraItems[i].Level < config.Stage {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
				ctx.Send(res)
				return
			}
		}
	}

	if common.IsHaveValueIntArray(pl.Equip.Brace.GetAuraStageAward, req.Id) == true {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	pl.Equip.Brace.GetAuraStageAward = append(pl.Equip.Brace.GetAuraStageAward, req.Id)
	res.GetAuraLevel = pl.Equip.Brace.GetAuraStageAward
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 背饰升级
func ReqEquipBraceUpLevel(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SLevelUpBrace) {
	res := &proto_equip.S2CLevelUpBrace{}

	if _, ok := pl.Equip.Brace.BraceItems[req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
		ctx.Send(res)
		return
	}

	configs := config.BracesLevel.All()
	conf := conf2.BracesLevel{}
	for _, v := range configs {
		if v.BracesId == req.Id && v.Level == pl.Equip.Brace.BraceItems[req.Id].Level {
			conf = v
			break
		}
	}

	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	if pl.Equip.Brace.BraceItems[req.Id].Num < conf.CostNum {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	pl.Equip.Brace.BraceItems[req.Id].Level += 1
	pl.Equip.Brace.BraceItems[req.Id].Num -= conf.CostNum
	res.Braces = model.ToBraceItemProto(pl.Equip.Brace.BraceItems)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 背饰使用
func ReqEquipBraceUse(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUseBrace) {
	res := &proto_equip.S2CUseBrace{}

	if req.Id != 0 {
		if _, ok := pl.Equip.Brace.BraceItems[req.Id]; !ok {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
			ctx.Send(res)
			return
		}

		//获取正在使用的
		useId := int32(0)
		for _, v := range pl.Equip.Brace.BraceItems {
			if v.IsUse {
				useId = v.Id
				break
			}
		}

		if useId > 0 {
			pl.Equip.Brace.BraceItems[useId].IsUse = false
		}

		//使用
		if req.IsUse {
			if req.Id == useId {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
				ctx.Send(res)
				return
			}

			pl.Equip.Brace.BraceItems[req.Id].IsUse = true
		}
	} else {
		for i, _ := range pl.Equip.Brace.BraceItems {
			pl.Equip.Brace.BraceItems[i].IsUse = false
		}
	}

	res.Braces = model.ToBraceItemProto(pl.Equip.Brace.BraceItems)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 背饰方案改名
func ReqEquipBraceChangeName(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SBraceIndexChangeName) {
	res := &proto_equip.S2CBraceIndexChangeName{}

	if req.Index > 5 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if _, ok := pl.Equip.Brace.BraceTalentIndexs[req.Index]; !ok {
		pl.Equip.Brace.BraceTalentIndexs[req.Index] = new(model.BraceTalentIndex)
		pl.Equip.Brace.BraceTalentIndexs[req.Index].Index = req.Index
		pl.Equip.Brace.BraceTalentIndexs[req.Index].BraceTalentJobs = make(map[int32]*model.BraceTalentJob)
	}

	pl.Equip.Brace.BraceTalentIndexs[req.Index].Name = req.Name
	names := make(map[int32]string)
	names[req.Index] = req.Name
	res.Name = names
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 切换背饰方案
func ReqTransformBraceIndex(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2STransformBraceIndex) {
	res := &proto_equip.S2CTransformBraceIndex{}

	if req.Index == int32(pl.Equip.Brace.BraceTalentIndex) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	pl.Equip.Brace.BraceTalentIndex = int(req.Index)
	res.Index = req.Index
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 背饰天赋升级
func ReqEquipBraceTalentUpLevel(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SLevelUpTalent) {
	res := &proto_equip.S2CLevelUpTalent{}

	conf, _ := config.BraceTalent.Find(int64(req.Id))
	if conf.Job != req.Job {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//判断是不是第一个
	if len(conf.FrontNode) <= 0 {
		item := getTalentItem(pl, req.Index, req.Job, conf.Group, req.Id)
		if item.Level >= define.BraceTalentLevelLimit {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_OUTLIMIT
			ctx.Send(res)
			return
		}
	} else {
		//判定前置解锁
		front := conf.FrontNode
		for _, v := range front {
			item := getTalentItem(pl, req.Index, req.Job, conf.Group, v)
			if item.Level < conf.UnLockLevel {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
				ctx.Send(res)
				return
			}
		}
	}
	item := getTalentItem(pl, req.Index, req.Job, conf.Group, req.Id)
	confLevels := config.BraceTalentLevel.All()
	conflevel := conf2.BraceTalentLevel{}
	for _, v := range confLevels {
		if v.TalentLevelId == conf.TalentLevelId && v.Level == item.Level {
			conflevel = v
			break
		}
	}

	if conflevel.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	cost := make(map[int32]int32)
	cost[conflevel.CostItem[0].ItemId] = req.Count

	if internal.CheckItemsEnough(pl, cost) == false {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	if item.Exp+req.Count >= conflevel.CostItem[0].ItemNum {
		item.Level += 1
		item.Exp = 0
		cost[conflevel.CostItem[0].ItemId] = conflevel.CostItem[0].ItemNum - item.Exp
	} else {
		item.Exp += req.Count
		cost[conflevel.CostItem[0].ItemId] = req.Count
	}

	internal.SubItems(ctx, pl, cost)
	pl.Equip.Brace.BraceTalentIndexs[req.Index].BraceTalentJobs[req.Job].BraceTalentGroups[req.Group].BraceTalentItems[req.Id] = item
	res.BraceTalent = model.ToBraceTalentSingleIndexProto(pl.Equip.Brace.BraceTalentIndexs, req.Index, req.Job, req.Group, req.Id)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 背饰天赋重置
func ReqEquipBraceTalentReSet(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SResetBraceTalent) {
	res := &proto_equip.S2CResetBraceTalent{}

	if _, ok := pl.Equip.Brace.BraceTalentIndexs[req.Index]; ok {
		if _, jobok := pl.Equip.Brace.BraceTalentIndexs[req.Index].BraceTalentJobs[req.Job]; jobok {
			//删除
			delete(pl.Equip.Brace.BraceTalentIndexs[req.Index].BraceTalentJobs, req.Job)
		}
	}
	res.Index = req.Index
	res.Job = req.Job
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 背饰图集奖励
func BraceHandbookAward(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SGetBraceHandBookAward) {
	res := &proto_equip.S2CGetBraceHandBookAward{}
	//算出等级
	level := int32(0)
	ids := make([]int32, 0)
	confs := config.HandbookAward.All()
	for _, v := range confs {
		if v.Type == define.HandbookAwardType_Brace {
			// 如果玩家经验大于等于配置所需经验，则更新等级
			if pl.Equip.Brace.HandbookExp >= v.Exp {
				ids = append(ids, v.Id)
				// 假设配置是按等级顺序的，取满足条件的最高等级
				if v.Level > level {
					level = v.Level
				}
			}
		}
	}

	//判断等级是否超出限制
	for _, v := range req.Id {
		if !common.IsHaveValueIntArray(ids, v) {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}

		//判断有没有领取的
		if pl.Equip.Brace.HandbookIds != nil {
			if common.IsHaveValueIntArray(pl.Equip.Brace.HandbookIds, v) {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
				ctx.Send(res)
				return
			}
		}
	}

	if pl.Equip.Brace.HandbookIds == nil {
		pl.Equip.Brace.HandbookIds = make([]int32, 0)
	}
	pl.Equip.Brace.HandbookIds = append(pl.Equip.Brace.HandbookIds, req.Id...)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 获取天赋
func getTalentItem(pl *model.Player, index, job, group, id int32) *model.BraceTalentItem {
	if _, ok := pl.Equip.Brace.BraceTalentIndexs[index]; !ok {
		pl.Equip.Brace.BraceTalentIndexs[index] = new(model.BraceTalentIndex)
		pl.Equip.Brace.BraceTalentIndexs[index].Index = index
		pl.Equip.Brace.BraceTalentIndexs[index].Name = ""
		pl.Equip.Brace.BraceTalentIndexs[index].BraceTalentJobs = make(map[int32]*model.BraceTalentJob)
	}

	if _, ok := pl.Equip.Brace.BraceTalentIndexs[index].BraceTalentJobs[job]; !ok {
		jobs := new(model.BraceTalentJob)
		jobs.Job = job
		jobs.BraceTalentGroups = make(map[int32]*model.BraceTalentGroup)
		pl.Equip.Brace.BraceTalentIndexs[index].BraceTalentJobs[job] = jobs
	}

	if _, ok := pl.Equip.Brace.BraceTalentIndexs[index].BraceTalentJobs[job].BraceTalentGroups[group]; !ok {
		groups := new(model.BraceTalentGroup)
		groups.Group = group
		groups.BraceTalentItems = make(map[int32]*model.BraceTalentItem)
		pl.Equip.Brace.BraceTalentIndexs[index].BraceTalentJobs[job].BraceTalentGroups[group] = groups
	}

	if _, ok := pl.Equip.Brace.BraceTalentIndexs[index].BraceTalentJobs[job].BraceTalentGroups[group].BraceTalentItems[id]; !ok {
		items := new(model.BraceTalentItem)
		items.Id = id
		items.Level = 0
		items.Exp = 0
		pl.Equip.Brace.BraceTalentIndexs[index].BraceTalentJobs[job].BraceTalentGroups[group].BraceTalentItems[id] = items
	}

	return pl.Equip.Brace.BraceTalentIndexs[index].BraceTalentJobs[job].BraceTalentGroups[group].BraceTalentItems[id]
}
