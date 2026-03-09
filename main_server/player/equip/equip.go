package equip

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
	"xfx/pkg/utils/sensitive"
	"xfx/proto/proto_equip"
)

func Init(pl *model.Player) {
	pl.Equip = new(model.Equip)
	pl.Equip.Equips = make([]*model.EquipOption, 0)
	pl.Equip.Mount = new(model.MountOption)
	pl.Equip.Weaponry = new(model.WeaponryOption)
	pl.Equip.Enchant = make(map[int32]*model.EnchantOption)
	pl.Equip.Succinct = new(model.SuccinctOption)
	pl.Equip.Succinct.LevelAward = make([]int32, 0)
	pl.Equip.Succinct.SuccinctIndexs = make(map[int]*model.SuccinctIndex)
	pl.Equip.Succinct.CacheSuccinctIndexs = make(map[int]*model.CacheSuccinctIndex)

	//初始小白马
	pl.Equip.Mount.Mount = make(map[int32]*model.MountItemOption, 0)
	pl.Equip.Mount.Mount[2001] = &model.MountItemOption{
		Id:    2001,
		Level: 1,
	}
	pl.Equip.Mount.HandbookExp = 0
	pl.Equip.Mount.HandbookIds = make([]int32, 0)

	//背饰
	pl.Equip.Brace = new(model.BraceOption)
	pl.Equip.Brace.BraceAuraItems = make(map[int32]*model.BraceAuraItem)
	pl.Equip.Brace.BraceItems = make(map[int32]*model.BraceItem)
	pl.Equip.Brace.GetAuraStageAward = make([]int32, 0)
	pl.Equip.Brace.BraceTalentIndexs = make(map[int32]*model.BraceTalentIndex)
	pl.Equip.Brace.BraceTalentIndex = 1 //默认方案1
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Equip)
	if err != nil {
		log.Error("player[%v],save Equip marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerEquip, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerEquip, pl.Id))
	if err != nil {
		log.Error("player[%v],load bag error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Equip)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Equip unmarshal error:%v", pl.Id, err)
	}

	pl.Equip = m
}

// 请求装备
func ReqInitEquip(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SInitEquip) {
	res := &proto_equip.S2CInitEquip{}
	res.EquipOption = model.ToEquipProto(pl.Equip.Equips)
	ctx.Send(res)
}

// 穿装备
func ReqWearEquip(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SWearEquip) {
	res := &proto_equip.S2CWearEquip{}
	log.Debug("穿戴装备id:%d,%d,%v", req.Id, req.Index, req.IsSell)
	curEquip := new(model.EquipOption)
	for _, v := range pl.Equip.Equips {
		if v.Id == req.Id {
			curEquip = v
			break
		}
	}

	if curEquip.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
		ctx.Send(res)
		return
	}

	fop := &model.EquipOption{}
	_index := 0
	for i, v := range pl.Equip.Equips {
		if v.Index == req.Index && v.IsUse == true {
			fop = v
			_index = i
			break
		}
	}

	//自动出售
	if req.IsSell && fop.Id > 0 {
		//获取品质
		equipConfs := config.Equip.All()
		conf := equipConfs[int64(fop.CId)]
		equipSellConfs := config.EquipSell.All()
		equipSellConf := conf2.EquipSell{}
		for _, v := range equipSellConfs {
			if v.Rate == conf.Rate && v.Index == conf.Index {
				equipSellConf = v
				break
			}
		}
		//出售
		if equipSellConf.Id > 0 {
			for _, v := range equipSellConf.Award {
				v.ItemNum = v.ItemNum * fop.Num
			}
			log.Debug("出售装备id:%d", equipSellConf.Id)
			bag.AddAward(ctx, pl, equipSellConf.Award, false)
		}
		pl.Equip.Equips = append(pl.Equip.Equips[:_index], pl.Equip.Equips[_index+1:]...)
	} else if fop.Id > 0 {
		for _, v := range pl.Equip.Equips {
			if v.Id == fop.Id {
				v.IsUse = false
				break
			}
		}
	}

	for _, v := range pl.Equip.Equips {
		if v.Id == req.Id {
			log.Debug("使用装备id:%d", v.Id)
			v.IsUse = true
			v.Index = req.Index
			break
		}
	}

	//同步变化
	pushRes := &proto_equip.PushEquipChange{
		EquipOption: model.ToEquipProto(pl.Equip.Equips),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 出售装备
func ReqSellEquip(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SSellEquip) {
	res := &proto_equip.S2CSellEquip{}
	log.Debug("出售装备id:%v", req.Id)
	equipConfs := config.Equip.All()
	equipSellConfs := config.EquipSell.All()
	Items := make([]conf2.ItemE, 0)
	removeIndex := make([]int, 0)
	for i := 0; i < len(req.Id); i++ {
		curEquip := new(model.EquipOption)
		_index := 0
		for k, v := range pl.Equip.Equips {
			if v.Id == req.Id[i] {
				_index = k
				curEquip = v
				break
			}
		}

		if curEquip.Id <= 0 {
			log.Error("没有这件装备:%v", req.Id[i])
			continue
		}

		//判断是否穿戴
		if curEquip.IsUse {
			log.Error("已经穿戴这件装备:%v", req.Id[i])
			continue
		}

		//获取品质
		conf := equipConfs[int64(curEquip.CId)]
		equipSellConf := conf2.EquipSell{}
		for _, v := range equipSellConfs {
			if v.Rate == conf.Rate && v.Index == conf.Index {
				equipSellConf = v
				break
			}
		}

		//出售
		if equipSellConf.Id > 0 {
			for _, v := range equipSellConf.Award {
				v.ItemNum = v.ItemNum * curEquip.Num
			}
			Items = append(Items, equipSellConf.Award...)
			removeIndex = append(removeIndex, _index)
		}

		log.Debug("出售装备:%v", req.Id[i])
	}

	//统一移除
	if len(removeIndex) > 0 {
		// Sort indices in descending order to avoid shifting issues
		sort.Sort(sort.Reverse(sort.IntSlice(removeIndex)))

		for _, idx := range removeIndex {
			// Check if index is valid
			if idx < 0 || idx >= len(pl.Equip.Equips) {
				continue // or return an error
			}

			pl.Equip.Equips = append(pl.Equip.Equips[:idx], pl.Equip.Equips[idx+1:]...)
		}
	}

	//出售
	if len(Items) > 0 {
		bag.AddAward(ctx, pl, Items, false)
	}

	//同步变化
	pushRes := &proto_equip.PushEquipChange{
		EquipOption: model.ToEquipProto(pl.Equip.Equips),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求坐骑
func ReqInitMount(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SInitMount) {
	res := &proto_equip.S2CInitMount{}
	res.MountOption = model.ToMountProto(pl.Equip.Mount)
	ctx.Send(res)
}

// 升级坐骑
func ReqLevelUpMount(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUpLevelMount) {
	res := &proto_equip.S2CUpLevelMount{}

	conf := config.MountStage.All()
	stage := conf2.MountStage{}
	for _, v := range conf {
		if v.Stage == pl.Equip.Mount.Stage && v.Star == pl.Equip.Mount.Star {
			stage = v
			break
		}
	}

	if stage.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	needNum := int32(0)
	needId := int32(0)
	//判断材料够不够
	costs := make(map[int32]int32)
	for _, v := range stage.UpStarCondition {
		needNum = v.ItemNum
		needId = v.ItemId
		costs[v.ItemId] = req.Count
	}
	//判断道具是否足够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	offseNum := needNum - pl.Equip.Mount.Exp
	if req.Count >= offseNum {
		costs[needId] = offseNum
		//升星
		pl.Equip.Mount.Star += 1
		if pl.Equip.Mount.Star >= 10 {
			pl.Equip.Mount.Stage += 1
			if pl.Equip.Mount.Stage >= 5 {
				pl.Equip.Mount.Stage = 5
			}
			pl.Equip.Mount.Star = 0
		}
		pl.Equip.Mount.Exp = 0
	} else {
		costs[needId] = req.Count
		pl.Equip.Mount.Exp += req.Count
	}

	internal.SubItems(ctx, pl, costs)

	//解锁新坐骑
	MountUnLock(ctx, pl)

	//同步变化
	pushRes := &proto_equip.PushMountChange{
		MountOption: model.ToMountProto(pl.Equip.Mount),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 升级单个坐骑的等级
func ReqLevelUpMountItem(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUpLevelMountItem) {
	res := &proto_equip.S2CUpLevelMountItem{}

	if _, ok := pl.Equip.Mount.Mount[req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
		ctx.Send(res)
		return
	}

	mount := pl.Equip.Mount.Mount[req.Id]

	conf := config.MountLevel.All()
	stage := conf2.MountLevel{}
	for _, v := range conf {
		if v.Level == mount.Level && v.MountId == mount.Id {
			stage = v
			break
		}
	}

	if stage.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	if mount.Num < stage.CostNum {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	mount.Num -= stage.CostNum
	mount.Level += 1
	pl.Equip.Mount.Mount[req.Id] = mount

	//同步变化
	pushRes := &proto_equip.PushMountChange{
		MountOption: model.ToMountProto(pl.Equip.Mount),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 坐骑图集奖励
func MountHandbookAward(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SGetMountHandBookAward) {
	res := &proto_equip.S2CGetMountHandBookAward{}
	//算出等级
	level := int32(0)
	ids := make([]int32, 0)
	confs := config.HandbookAward.All()
	for _, v := range confs {
		if v.Type == define.HandbookAwardType_Mount {
			// 如果玩家经验大于等于配置所需经验，则更新等级
			if pl.Equip.Mount.HandbookExp >= v.Exp {
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
		if !utils.ContainsInt32(ids, v) {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}

		//判断有没有领取的
		if pl.Equip.Mount.HandbookIds != nil {
			if utils.ContainsInt32(pl.Equip.Mount.HandbookIds, v) {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
				ctx.Send(res)
				return
			}
		}
	}

	if pl.Equip.Mount.HandbookIds == nil {
		pl.Equip.Mount.HandbookIds = make([]int32, 0)
	}
	pl.Equip.Mount.HandbookIds = append(pl.Equip.Mount.HandbookIds, req.Id...)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 升星升级解锁新坐骑
func MountUnLock(ctx global.IPlayer, pl *model.Player) {
	confs := config.Mount.All()
	for _, v := range confs {
		//升级和升星
		if v.UnLock == 1 && v.UnlockValue[0] == pl.Equip.Mount.Stage && v.UnlockValue[1] == pl.Equip.Mount.Star {
			pl.Equip.Mount.Mount[v.Id] = &model.MountItemOption{
				Id:    v.Id,
				Level: 1,
			}
			break
		}
	}
}

// 使用坐骑
func ReqUseMount(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUseMount) {
	res := &proto_equip.S2CUseMount{}

	if pl.Equip.Mount.UseId == req.Id {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if req.Id != 0 {
		has := false
		for _, v := range pl.Equip.Mount.Mount {
			if v.Id == req.Id {
				has = true
				break
			}
		}

		if !has {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
			ctx.Send(res)
			return
		}
	}

	pl.Equip.Mount.UseId = req.Id

	//同步变化
	pushRes := &proto_equip.PushMountChange{
		MountOption: model.ToMountProto(pl.Equip.Mount),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 坐骑改名
func ReqMountChangeName(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SChangeMountName) {
	res := &proto_equip.S2CChangeMountName{}

	if _, ok := pl.Equip.Mount.Mount[req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
		ctx.Send(res)
		return
	}

	//判断材料够不够
	costs := make(map[int32]int32)
	cost := config.Global.Get().MountChangeName
	costs[cost[0].ItemId] = cost[0].ItemNum

	//判断道具是否足够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	//筛查敏感字
	if sensitive.Filter.IsSensitive(req.Name) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	pl.Equip.Mount.Mount[req.Id].Name = req.Name
	internal.SubItems(ctx, pl, costs)

	//同步变化
	pushRes := &proto_equip.PushMountChange{
		MountOption: model.ToMountProto(pl.Equip.Mount),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 坐骑赋能
func ReqMountUpEnergy(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUpEnergyLevel) {
	res := &proto_equip.S2CUpEnergyLevel{}

	//总等级
	sumLevel := int32(0)
	for _, v := range pl.Equip.Mount.MountEnergy {
		sumLevel += v
	}

	confs := config.MountEnergy.All()
	conf := conf2.MountEnergy{}
	for _, v := range confs {
		if v.Level == sumLevel {
			conf = v
			break
		}
	}

	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//随机类型
	rangeTyp := utils.WeightIndex(conf.Weight)

	//随机成功率
	rangeSuc := utils.RandInt(0, 100)

	if int32(rangeSuc) < conf.SuccessRate || conf.SuccessRate == 100 {
		costs := make(map[int32]int32)
		costs[conf.UpLevelCondition[0].ItemId] = conf.UpLevelCondition[0].ItemNum

		//判断道具是否足够
		if !internal.CheckItemsEnough(pl, costs) {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
			ctx.Send(res)
			return
		}

		if pl.Equip.Mount.MountEnergy == nil {
			pl.Equip.Mount.MountEnergy = make(map[int32]int32)
		}

		level := pl.Equip.Mount.MountEnergy[int32(rangeTyp+1)]
		level += 1
		pl.Equip.Mount.MountEnergy[int32(rangeTyp+1)] = level

		internal.SubItems(ctx, pl, costs)
	}

	//同步变化
	pushRes := &proto_equip.PushMountChange{
		MountOption: model.ToMountProto(pl.Equip.Mount),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 请求神兵
func ReqInitWeaponry(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SInitWeaponry) {
	res := &proto_equip.S2CInitWeaponry{}
	res.WeaponryOption = model.ToWeaponryProto(pl.Equip.Weaponry)
	ctx.Send(res)
}

// 升级神兵
func ReqLevelUpWeaponry(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUpLevelWeaponry) {
	res := &proto_equip.S2CUpLevelWeaponry{}

	conf := config.WeaponryStar.All()
	stage := conf2.WeaponryStar{}
	for _, v := range conf {
		if v.Star == pl.Equip.Weaponry.Star {
			stage = v
			break
		}
	}

	if stage.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	needNum := int32(0)
	needId := int32(0)
	//判断材料够不够
	costs := make(map[int32]int32)
	for _, v := range stage.UpStarCondition {
		needNum = v.ItemNum
		needId = v.ItemId
		costs[v.ItemId] = req.Count
	}
	//判断道具是否足够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	offseNum := needNum - pl.Equip.Weaponry.Exp
	if req.Count >= offseNum {
		costs[needId] = offseNum
		//升星
		pl.Equip.Weaponry.Star += 1
		pl.Equip.Weaponry.Exp = 0
	} else {
		costs[needId] = req.Count
		pl.Equip.Weaponry.Exp += req.Count
	}

	internal.SubItems(ctx, pl, costs)

	//同步变化
	pushRes := &proto_equip.PushWeaponryChange{
		WeaponryOption: model.ToWeaponryProto(pl.Equip.Weaponry),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 升级单个神兵
func ReqLevelUpWeaponryItem(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUpLevelWeaponryItem) {
	res := &proto_equip.S2CUpLevelWeaponryItem{}

	if _, ok := pl.Equip.Weaponry.WeaponryItems[req.Id]; !ok {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOEQUIP
		ctx.Send(res)
		return
	}

	weaponry := pl.Equip.Weaponry.WeaponryItems[req.Id]

	conf := config.WeaponryLevel.All()
	stage := conf2.WeaponryLevel{}
	for _, v := range conf {
		if v.WeaponryId == req.Id && v.Level == weaponry.Level {
			stage = v
			break
		}
	}

	if stage.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	if weaponry.Num < stage.CostNum {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	weaponry.Num -= stage.CostNum
	weaponry.Level += 1

	pl.Equip.Weaponry.WeaponryItems[weaponry.Id] = weaponry

	//同步变化
	pushRes := &proto_equip.PushWeaponryChange{
		WeaponryOption: model.ToWeaponryProto(pl.Equip.Weaponry),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 神兵图集奖励
func WeaponHandbookAward(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SGetWeaponryHandBookAward) {
	res := &proto_equip.S2CGetMountHandBookAward{}
	//算出等级
	level := int32(0)
	ids := make([]int32, 0)
	confs := config.HandbookAward.All()
	for _, v := range confs {
		if v.Type == define.HandbookAwardType_Weapon {
			// 如果玩家经验大于等于配置所需经验，则更新等级
			if pl.Equip.Weaponry.HandbookExp >= v.Exp {
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
		if !utils.ContainsInt32(ids, v) {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}

		//判断有没有领取的
		if pl.Equip.Weaponry.HandbookIds != nil {
			if utils.ContainsInt32(pl.Equip.Weaponry.HandbookIds, v) {
				res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
				ctx.Send(res)
				return
			}
		}
	}

	if pl.Equip.Weaponry.HandbookIds == nil {
		pl.Equip.Weaponry.HandbookIds = make([]int32, 0)
	}
	pl.Equip.Weaponry.HandbookIds = append(pl.Equip.Weaponry.HandbookIds, req.Id...)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 使用神兵
func ReqUseWeaponry(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUseWeaponry) {
	res := &proto_equip.S2CUseWeaponry{}

	if pl.Equip.Weaponry.UseId == req.Id {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if req.Id > 0 {
		has := false
		for _, v := range pl.Equip.Weaponry.WeaponryItems {
			if v.Id == req.Id {
				has = true
				break
			}
		}

		if has {
			pl.Equip.Weaponry.UseId = req.Id
		}
	} else {
		pl.Equip.Weaponry.UseId = req.Id
	}

	//同步变化
	pushRes := &proto_equip.PushWeaponryChange{
		WeaponryOption: model.ToWeaponryProto(pl.Equip.Weaponry),
	}
	ctx.Send(pushRes)

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}
