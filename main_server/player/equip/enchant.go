package equip

import (
	"strings"
	"xfx/pkg/utils"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/proto/proto_equip"
)

// 请求附魔
func ReqInitEquipEnchant(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SInitEquipEnchant) {
	res := &proto_equip.S2CInitEquipEnchant{}
	res.Enchants = model.ToEnchantProto(pl.Equip.Enchant)
	ctx.Send(res)
}

// 使用符咒
func ReqUseEnchant(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SUseEnchant) {
	res := &proto_equip.S2CUseEnchant{}
	//获取道具类型
	itemconf := config.Item.All()[int64(req.Id)]
	if itemconf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//先判断类型
	enLimitconf := config.Global.Get().EnchantLimit
	enlimittyps := strings.Split(enLimitconf, "|")
	maps := make(map[int32][]int32)
	for i := 0; i < len(enlimittyps); i++ {
		typss := strings.Split(enlimittyps[i], "=")
		values := strings.Split(typss[1], ",")
		typInt := utils.MustParseInt64(typss[0])
		arr := make([]int32, 0)
		for j := 0; j < len(values); j++ {
			arr = append(arr, int32(utils.MustParseInt64(values[j])))
		}
		maps[int32(typInt)] = arr
	}

	typ := define.EnchantTypeById[itemconf.Id]
	islimit := true
	enlimits := maps[req.Index]
	for i := 0; i < len(enlimits); i++ {
		if enlimits[i] == int32(typ) {
			islimit = false
			break
		}
	}

	//限制类型
	if islimit {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if pl.Equip.Enchant == nil {
		pl.Equip.Enchant = make(map[int32]*model.EnchantOption)
	}

	enchant := &model.EnchantOption{}
	//判断是不是自己
	if _, ok := pl.Equip.Enchant[req.Index]; ok {
		enchant = pl.Equip.Enchant[req.Index]
	}

	//开始是空的
	if enchant.Id <= 0 {
		pl.Equip.Enchant[req.Index] = &model.EnchantOption{
			Id:    req.Id,
			Level: 0,
			Exp:   0,
		}
	} else {
		if enchant.Id == req.Id {
			res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
			ctx.Send(res)
			return
		}
		enchant.Id = req.Id
		pl.Equip.Enchant[req.Index] = enchant
	}

	pushEnchantChange(ctx, pl, req.Index)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 升级
func ReqUpLevelEnchant(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SEnchantLevel) {
	res := &proto_equip.S2CEnchantLevel{}
	//获取道具类型
	itemconf := config.Item.All()[int64(req.Id)]
	if itemconf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	//先判断类型
	enLimitconf := config.Global.Get().EnchantLimit
	enlimittyps := strings.Split(enLimitconf, "|")
	maps := make(map[int32][]int32)
	for i := 0; i < len(enlimittyps); i++ {
		typss := strings.Split(enlimittyps[i], "=")
		values := strings.Split(typss[1], ",")
		typInt := utils.MustParseInt64(typss[0])
		arr := make([]int32, 0)
		for j := 0; j < len(values); j++ {
			arr = append(arr, int32(utils.MustParseInt64(values[j])))
		}
		maps[int32(typInt)] = arr
	}
	enlimits := maps[req.Index]
	typ := define.EnchantTypeById[itemconf.Id]
	islimit := true
	for i := 0; i < len(enlimits); i++ {
		if enlimits[i] == int32(typ) {
			islimit = false
			break
		}
	}

	//限制类型
	if islimit {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if pl.Equip.Enchant == nil {
		pl.Equip.Enchant = make(map[int32]*model.EnchantOption)
	}

	enchant := &model.EnchantOption{}
	//判断是不是自己
	if _, ok := pl.Equip.Enchant[req.Index]; ok {
		enchant = pl.Equip.Enchant[req.Index]
	}

	if enchant.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	if enchant.Id != req.Id {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_PARAMERROR
		ctx.Send(res)
		return
	}

	//判断道具够不够
	costs := make(map[int32]int32)
	costs[req.Id] = req.Count

	//判断道具是否足够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOENGTHERROR
		ctx.Send(res)
		return
	}

	confs := config.Enchant.All()
	conf := conf.Enchant{}
	for _, v := range confs {
		if v.Level == enchant.Level {
			conf = v
			break
		}
	}

	//没有配置表
	if conf.Id <= 0 {
		res.Code = proto_equip.ERRORCODEEQUIP_ERROR_NOCONFIG
		ctx.Send(res)
		return
	}

	if enchant.Exp+req.Count >= conf.CostNum {
		enchant.Level += 1
		enchant.Exp = (enchant.Exp + req.Count) - conf.CostNum
	} else {
		enchant.Exp += req.Count
	}

	//扣除道具
	internal.SubItems(ctx, pl, costs)

	pl.Equip.Enchant[req.Index] = enchant

	pushEnchantChange(ctx, pl, req.Index)
	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 一键升级
func ReqOneKeyUpLevelEnchant(ctx global.IPlayer, pl *model.Player, req *proto_equip.C2SOneKeyEnchantLevel) {
	res := &proto_equip.S2COneKeyEnchantLevel{}

	for i, v := range pl.Equip.Enchant {

		enchant := v
		if enchant.Id <= 0 {
			continue
		}

		num := pl.Bag.Items[v.Id]
		if num <= 0 {
			continue
		}

		confs := config.Enchant.All()
		costs := make(map[int32]int32)

		loopLevel := func() bool {
			conf := conf.Enchant{}
			for _, v := range confs {
				if v.Level == enchant.Level {
					conf = v
					break
				}
			}

			//没有配置表
			if conf.Id <= 0 {
				return true
			}

			if _, ok := costs[enchant.Id]; !ok {
				costs[enchant.Id] = 0
			}

			if enchant.Exp+num >= conf.CostNum {
				enchant.Level += 1
				costs[enchant.Id] += conf.CostNum - enchant.Exp
				num -= (conf.CostNum - enchant.Exp)
				enchant.Exp = 0
				return false
			} else {
				enchant.Exp += num
				costs[enchant.Id] += num
				num -= num
				return true
			}
		}

		for index := enchant.Level; index <= 50; index++ {
			if loopLevel() {
				break
			}
		}

		//扣除道具
		internal.SubItems(ctx, pl, costs)

		pl.Equip.Enchant[i] = enchant
		pushEnchantChange(ctx, pl, i)
	}

	res.Code = proto_equip.ERRORCODEEQUIP_ERR_Ok
	ctx.Send(res)
}

// 同步附魔变化
func pushEnchantChange(ctx global.IPlayer, pl *model.Player, index int32) {
	enchants := make(map[int32]*proto_equip.EquipEnchantOption)
	enchants[index] = &proto_equip.EquipEnchantOption{
		Id:    pl.Equip.Enchant[index].Id,
		Level: pl.Equip.Enchant[index].Level,
		Exp:   pl.Equip.Enchant[index].Exp,
	}

	res := &proto_equip.PushEnchantChange{
		Enchants: enchants,
	}
	ctx.Send(res)
}
