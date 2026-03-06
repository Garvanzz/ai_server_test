package pet

import (
	"xfx/core/common"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_pet"
	"xfx/proto/proto_public"
)

// 获取宠物装备经验
func ReqPetHandBookGetExp(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SGetPetHandBookExp) {
	resp := new(proto_pet.S2CGetPetHandBookExp)
	if _, ok := pl.PetHandbook.HandbookPets[req.Id]; !ok {
		resp.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(resp)
		return
	}

	if pl.PetHandbook.HandbookPets[req.Id].IsGetExp == false || pl.PetHandbook.HandbookPets[req.Id].GetExp <= 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(resp)
		return
	}

	pl.PetHandbook.PetHandbookOption.Exp += pl.PetHandbook.HandbookPets[req.Id].GetExp
	//判断经验
	confs := config.PetEquipHbAward.All()
	level := int64(pl.PetHandbook.PetHandbookOption.Level)
	exp := pl.PetHandbook.PetHandbookOption.Exp

	curLevel := int32(level)
	for k := level; k < 20; k++ {
		conf := confs[k]
		curLevel = int32(k)
		if exp >= conf.Exp {
			exp -= conf.Exp
		} else {
			break
		}
	}

	pl.PetHandbook.PetHandbookOption.Exp = exp
	pl.PetHandbook.PetHandbookOption.Level = curLevel
	pl.PetHandbook.HandbookPets[req.Id].IsGetExp = false
	pl.PetHandbook.HandbookPets[req.Id].GetExp = 0

	resp.Code = proto_public.CommonErrorCode_ERR_OK
	resp.Ids = model.ToPetEquipHandbookListProto(pl.PetHandbook)
	resp.PetEquipHandbookOption = model.ToPetEquipHandbookProto(pl.PetHandbook)

	ctx.Send(resp)
}

// 领取奖励
func ReqPetHandBookAward(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SGetPetHandBookAward) {
	resp := new(proto_pet.S2CGetPetHandBookAward)

	confs := config.PetEquipHbAward.All()
	for _, v := range req.Id {
		conf := confs[int64(v)]
		if conf.Id <= 0 {
			resp.Code = proto_public.CommonErrorCode_ERR_NoConfig
			ctx.Send(resp)
			return
		}
	}

	for _, k := range pl.PetHandbook.PetHandbookOption.GetId {
		if common.IsHaveValueIntArray(req.Id, k) {
			resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(resp)
			return
		}
	}

	for _, k := range req.Id {
		if k <= pl.PetHandbook.PetHandbookOption.Level {
			pl.PetHandbook.PetHandbookOption.GetId = append(pl.PetHandbook.PetHandbookOption.GetId, k)
		}
	}

	internal.SyncPetEquipHandbookInfoChange(ctx, pl, req.Id)

	resp.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(resp)
}

// 装备强化
func ReqPetEquipStrengthen(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPetForge) {
	resp := new(proto_pet.S2CPetForge)
	if req.Id <= 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	if _, ok := pl.PetEquip.Equips[req.Id]; !ok {
		resp.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(resp)
		return
	}

	equip := pl.PetEquip.Equips[req.Id]
	conf := config.PetEquip.All()[int64(equip.Id)]
	confs := config.PetEquipLevel.All()
	costConf := conf2.PetEquipLevel{}
	for _, v := range confs {
		if v.Level == equip.Level && v.Rate == conf.Rate {
			costConf = v
			break
		}
	}

	if costConf.Id <= 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(resp)
		return
	}

	cost := make(map[int32]int32)
	cost[costConf.NeedCost[0].ItemId] = cost[costConf.NeedCost[0].ItemNum]

	if !internal.CheckItemsEnough(pl, cost) {
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(resp)
		return
	}

	if equip.Num < costConf.NeedNum {
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(resp)
		return
	}

	equip.Num -= costConf.NeedNum
	equip.Level += 1

	internal.SubItems(ctx, pl, cost)
	pl.PetEquip.Equips[req.Id] = equip
	internal.SyncPetEquipChange(ctx, pl, req.Id)

	resp.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(resp)
}

// 装备分解
func ReqPetEquipBreakdown(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPetBreakdown) {
	resp := new(proto_pet.S2CPetBreakdown)
	if req.Id <= 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	log.Debug("ReqPetEquipBreakdown:%v, %v", pl.PetEquip.Equips, req.Id)
	if _, ok := pl.PetEquip.Equips[req.Id]; !ok {
		resp.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(resp)
		return
	}

	equip := pl.PetEquip.Equips[req.Id]
	if equip.Num < 1 {
		resp.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(resp)
		return
	}

	conf := config.PetEquip.All()[int64(equip.Id)]
	if conf.Id <= 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(resp)
		return
	}

	equip.Num -= 1

	bag.AddAward(ctx, pl, conf.BreakDown, true)
	pl.PetEquip.Equips[req.Id] = equip
	internal.SyncPetEquipChange(ctx, pl, req.Id)

	resp.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(resp)
}

// 装备穿戴
func ReqPetEquipWear(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPetWearEquip) {
	resp := new(proto_pet.S2CPetWearEquip)
	if req.Id <= 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	if _, ok := pl.PetEquip.Equips[req.Id]; !ok {
		resp.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(resp)
		return
	}

	//判断宠物
	if _, ok := pl.Pet.Pets[req.PetId]; !ok {
		resp.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(resp)
		return
	}

	//判断是否穿戴
	pet := pl.Pet.Pets[req.PetId]
	if common.IsHaveValueIntArray(pet.Equips, req.Id) {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	equipMap := make(map[int32]int32)
	for _, v := range pet.Equips {
		if v <= 0 {
			continue
		}
		conf := config.PetEquip.All()[int64(v)]
		if conf.Id <= 0 {
			resp.Code = proto_public.CommonErrorCode_ERR_NoConfig
			ctx.Send(resp)
			return
		}

		equipMap[conf.Type] = 1
	}

	confequip := config.PetEquip.All()[int64(req.Id)]
	if _, ok := equipMap[confequip.Type]; ok {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	pet.Equips[confequip.Type-1] = req.Id
	pl.Pet.Pets[req.PetId] = pet

	internal.SyncPetChange(ctx, pl, req.PetId)

	resp.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(resp)
}
