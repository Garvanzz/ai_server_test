package pet

import (
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_pet"
	"xfx/proto/proto_public"
)

// ReqPetUnderstandGift 请求宠物天赋继承
func ReqPetUnderstandGift(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SUnderstandGift) {
	res := new(proto_pet.S2CUnderstandGift)

	//传入的是宠物本体，消耗的是碎片
	if req.CostPetId <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[req.CostPetId]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[req.Pet]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	petConf := config.Pet.All()[int64(req.CostPetId)]
	if petConf.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	//判断数量够不够
	itemConf := config.Item.All()[int64(petConf.Fragment)]
	if itemConf.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	curPetConf := config.Pet.All()[int64(req.Pet)]
	//数量
	costConfs := config.PetGiftCost.All()
	costNum := int32(0)
	for _, v := range costConfs {
		if v.Rate != curPetConf.Rate {
			continue
		}
		for k := 0; k < len(v.NormalCost); k++ {
			if v.NormalCost[k][0] == petConf.Rate {
				costNum = v.NormalCost[k][1]
				break
			}
		}
	}

	if costNum <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	cost := make(map[int32]int32)
	cost[itemConf.Id] = costNum
	if !internal.CheckItemsEnough(pl, cost) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	pet := pl.Pet.Pets[req.CostPetId]
	curPet := pl.Pet.Pets[req.Pet]
	//获取天赋
	gitfs := pet.Gifts

	//判断有没有
	sum := 0
	for h := 0; h < len(gitfs); h++ {
		if gitfs[h] <= 0 {
			continue
		}
		sum += 1
	}

	if sum <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	//判断解锁了几个天赋数
	stage := curPet.Stage
	stages := curPetConf.GiftUnlock
	unlockNum := int32(0)
	for i := 0; i < len(stages); i++ {
		s := stages[i]
		if s[0] == stage {
			unlockNum = s[1]
			break
		}
	}

	curPet.Gifts = make([]int32, 3)
	for k := int32(0); k < unlockNum; k++ {
		if k+1 <= int32(len(gitfs)) {
			if gitfs[k] <= 0 {
				continue
			}
			curPet.Gifts[k] = gitfs[k]
		}
	}

	pl.Pet.Pets[req.Pet] = curPet

	//扣除道具
	internal.SubItems(ctx, pl, cost)

	internal.SyncPetChange(ctx, pl, req.Pet)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// ReqPetUnderstandPointGift 请求宠物天赋定向继承
func ReqPetUnderstandPointGift(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPointUnderstandGift) {
	res := new(proto_pet.S2CPointUnderstandGift)

	if req.GiftIndex <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	//传入的是宠物本体，消耗的是碎片
	if req.CostPetId <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[req.CostPetId]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[req.Pet]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	petConf := config.Pet.All()[int64(req.CostPetId)]
	if petConf.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	//判断数量够不够
	itemConf := config.Item.All()[int64(petConf.Fragment)]
	if itemConf.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	curPetConf := config.Pet.All()[int64(req.Pet)]
	//数量
	costConfs := config.PetGiftCost.All()
	costNum := int32(0)
	costItem := []conf2.ItemE{}
	for _, v := range costConfs {
		if v.Rate != curPetConf.Rate {
			continue
		}
		for k := 0; k < len(v.NormalCost); k++ {
			if v.NormalCost[k][0] == petConf.Rate {
				costNum = v.NormalCost[k][1]
				costItem = v.Cost
				break
			}
		}
	}

	if costNum <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	cost := make(map[int32]int32)
	cost[itemConf.Id] = costNum

	cost[costItem[0].ItemId] = costItem[0].ItemNum
	if !internal.CheckItemsEnough(pl, cost) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	pet := pl.Pet.Pets[req.CostPetId]
	curPet := pl.Pet.Pets[req.Pet]
	//获取天赋
	gitfs := pet.Gifts

	//判断解锁了几个天赋数
	stage := curPet.Stage
	stages := curPetConf.GiftUnlock
	unlockNum := int32(0)
	for i := 0; i < len(stages); i++ {
		s := stages[i]
		if s[0] == stage {
			unlockNum = s[1]
			break
		}
	}

	if req.GiftIndex > unlockNum {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	//判断天赋存不存在
	has := false
	if req.GiftId > 0 {
		for l := 0; l < len(curPet.Gifts); l++ {
			if curPet.Gifts[l] == req.GiftId {
				has = true
				break
			}
		}

		if !has {
			res.Code = proto_public.CommonErrorCode_ERR_NoPet
			ctx.Send(res)
			return
		}
	}

	for l := 0; l < len(gitfs); l++ {
		if gitfs[l] == req.TargetGiftId {
			has = true
			break
		}
	}

	if !has {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	//继承
	curPet.Gifts[req.GiftIndex-1] = req.TargetGiftId
	pl.Pet.Pets[req.Pet] = curPet

	//扣除道具
	internal.SubItems(ctx, pl, cost)

	internal.SyncPetChange(ctx, pl, req.Pet)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 洗练
// ReqPetXilianGift 请求宠物天赋洗练
func ReqPetXilianGift(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPetXilianGift) {
	res := new(proto_pet.S2CPetXilianGift)

	if _, ok := pl.Pet.Pets[req.Pet]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	petConf := config.Pet.All()[int64(req.Pet)]
	if petConf.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	pet := pl.Pet.Pets[req.Pet]
	num := 0
	for _, v := range pet.CacheXilian {
		if v > 0 {
			num += 1
		}
	}
	if num > 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	cost := make(map[int32]int32)
	//判断材料够不够
	petXili := config.Global.Get().PetXilianCost
	cost[petXili[0].ItemId] = petXili[0].ItemNum
	if !internal.CheckItemsEnough(pl, cost) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	if model.GiftPoolConf == nil {
		model.GiftPoolConf = make(map[int32]map[int32]int32)
		conf := config.PetGift.All()
		for _, v := range conf {
			if _, ok := model.GiftPoolConf[v.Rate]; !ok {
				model.GiftPoolConf[v.Rate] = make(map[int32]int32)
			}

			model.GiftPoolConf[v.Rate][v.Id] = v.Rare
		}
	}

	//获取解锁的天赋数
	stage := pet.Stage
	curPetConf := config.Pet.All()[int64(req.Pet)]
	stages := curPetConf.GiftUnlock
	unlockNum := int32(0)
	for i := 0; i < len(stages); i++ {
		s := stages[i]
		if s[0] == stage {
			unlockNum = s[1]
			break
		}
	}

	ids := make([]int32, unlockNum)
	for i := int32(0); i < unlockNum; i++ {
		id := internal.PetGiftDrawCard()
		log.Debug("宠物天赋%d:", id)
		ids[i] = id
	}

	pet.CacheXilian = ids
	pl.Pet.Pets[req.Pet] = pet

	//扣除材料
	internal.SubItems(ctx, pl, cost)

	internal.SyncPetChange(ctx, pl, req.Pet)
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 确定洗练
// ReqPetSureXilianGift 请求宠物天赋洗练
func ReqPetSureXilianGift(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPetSureXilianGift) {
	res := new(proto_pet.S2CPetSureXilianGift)

	if _, ok := pl.Pet.Pets[req.PetId]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	petConf := config.Pet.All()[int64(req.PetId)]
	if petConf.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	pet := pl.Pet.Pets[req.PetId]
	num := 0
	for _, v := range pet.CacheXilian {
		if v > 0 {
			num += 1
		}
	}

	if num <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if !req.IsOk {
		pet.CacheXilian = make([]int32, 0)
	} else {
		pet.Gifts = pet.CacheXilian
		pet.CacheXilian = make([]int32, 0)
	}
	pl.Pet.Pets[req.PetId] = pet

	internal.SyncPetChange(ctx, pl, req.PetId)
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}
