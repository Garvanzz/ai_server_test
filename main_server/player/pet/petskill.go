package pet

import (
	"xfx/core/config"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_pet"
	"xfx/proto/proto_public"
)

// ReqPetUnderstandSkill 请求宠物技能领悟
func ReqPetUnderstandSkill(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SUnderstandSkill) {
	res := new(proto_pet.S2CUnderstandSkill)
	log.Debug("技能领悟：%v, %v", pl.Pet.Pets, req)
	if req.SkillId <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[req.PetId]; !ok {
		log.Debug("&&&&&&&&&^^^^^^^^")
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Skills[req.SkillId]; !ok {
		log.Debug("&&&&&&&&&^^^^^^^^66666666")
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	skillConf := config.PetSkill.All()[int64(req.SkillId)]
	if skillConf.Id <= 0 {
		log.Debug("&&&&&&&&&^^^^^^^^89999999999")
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	//是否是专属
	if skillConf.PetId > 0 && skillConf.PetId != req.PetId {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		log.Debug("&&&&&&&&&^^^^^^^^000088888")
		return
	}

	skills := pl.Pet.Skills[req.SkillId]
	if skills.Num <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		log.Debug("&&&&&&&&&^^^^^^^^111111111")
		return
	}

	pet := pl.Pet.Pets[req.PetId]
	//判断解锁了几个技能数
	star := pet.Star
	petConf := config.Pet.All()[int64(req.PetId)]
	skillUnlocks := petConf.SkillUnlock
	unlockNum := int32(0)
	for i := 0; i < len(skillUnlocks); i++ {
		s := skillUnlocks[i]
		if s[0] == star {
			unlockNum = s[1]
			break
		}
	}

	//获取解锁的槽位
	index := utils.RandInt(1, unlockNum)
	pet.SkillIds[index-1] = req.SkillId
	skills.Num -= 1
	pl.Pet.Skills[req.SkillId] = skills
	pl.Pet.Pets[req.PetId] = pet

	internal.SyncPetChange(ctx, pl, req.PetId)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// ReqPetUnderstandPointSkill 请求宠物技能定向继承
func ReqPetUnderstandPointSkill(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SPointUnderstandSkill) {
	res := new(proto_pet.S2CPointUnderstandSkill)

	if req.Index <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if req.PetId <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	log.Debug("&&&&&&&&123:%v, %v", pl.Pet.Pets, req.PetId)
	if _, ok := pl.Pet.Pets[req.PetId]; !ok {
		log.Debug("&&&&&&&&")
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	if req.SkillId <= 0 {
		log.Debug("&&&&&&&&2222222")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Skills[req.SkillId]; !ok {
		log.Debug("&&&&&&&&5555555")
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	skillConf := config.PetSkill.All()[int64(req.SkillId)]
	if skillConf.Id <= 0 {
		log.Debug("&&&&&&&&6666666")
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	//是否是专属
	if skillConf.PetId > 0 && skillConf.PetId != req.PetId {
		log.Debug("&&&&&&&&77777777")
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	skills := pl.Pet.Skills[req.SkillId]
	if skills.Num <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		log.Debug("&&&&&&&&888888")
		return
	}

	//材料够不够
	cost := make(map[int32]int32)
	costskill := config.Global.Get().PetSkillCost
	cost[costskill[0].ItemId] = costskill[0].ItemNum
	if !internal.CheckItemsEnough(pl, cost) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		log.Debug("&&&&&&&&999999999")
		return
	}

	pet := pl.Pet.Pets[req.PetId]
	//判断解锁了几个技能数
	star := pet.Star
	petConf := config.Pet.All()[int64(req.PetId)]
	skillUnlocks := petConf.SkillUnlock
	unlockNum := int32(0)
	for i := 0; i < len(skillUnlocks); i++ {
		s := skillUnlocks[i]
		if s[0] == star {
			unlockNum = s[1]
			break
		}
	}

	if req.Index > unlockNum {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		log.Debug("&&&&&&&&0000000")
		return
	}

	pet.SkillIds[req.Index-1] = req.SkillId
	skills.Num -= 1
	pl.Pet.Skills[req.SkillId] = skills
	pl.Pet.Pets[req.PetId] = pet

	//扣除材料
	internal.SubItems(ctx, pl, cost)
	internal.SyncPetChange(ctx, pl, req.PetId)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// ReqPetRemovePointSkill 请求宠物技能移除
func ReqPetRemovePointSkill(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SForgetSkill) {
	res := new(proto_pet.S2CForgetSkill)

	if req.SkillIndex <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if req.PetId <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[req.PetId]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	pet := pl.Pet.Pets[req.PetId]
	if req.SkillIndex > int32(len(pet.SkillIds)) {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	if pet.SkillIds[req.SkillIndex-1] <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	//材料够不够
	cost := make(map[int32]int32)
	costskill := config.Global.Get().PetSkillRemoveCost
	cost[costskill[0].ItemId] = costskill[0].ItemNum
	if !internal.CheckItemsEnough(pl, cost) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	pet.SkillIds[req.SkillIndex-1] = 0
	pl.Pet.Pets[req.PetId] = pet

	//扣除材料
	internal.SubItems(ctx, pl, cost)
	internal.SyncPetChange(ctx, pl, req.PetId)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}
