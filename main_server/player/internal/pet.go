package internal

import (
	"sort"
	"xfx/core/config"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/utils"
	"xfx/proto/proto_pet"
)

// SyncPetChange 同步宠物
func SyncPetChange(ctx global.IPlayer, pl *model.Player, id int32) {
	res := &proto_pet.PushChangePetItem{}
	if _, ok := pl.Pet.Pets[id]; !ok {
		return
	}

	maps := make(map[int32]*model.PetItem)
	maps[id] = pl.Pet.Pets[id]
	pet := model.ToPetItemProto(maps)
	res.PetItems = pet
	ctx.Send(res)
}

// 同步全部宠物
func SyneAllPet(ctx global.IPlayer, pl *model.Player) {
	res := &proto_pet.PushChangePetItem{}
	maps := make(map[int32]*model.PetItem)
	for _, v := range pl.Pet.Pets {
		maps[v.Id] = pl.Pet.Pets[v.Id]
	}
	pet := model.ToPetItemProto(maps)
	res.PetItems = pet
	ctx.Send(res)
}

// 增加宠物装备
func AddPetEquip(ctx global.IPlayer, pl *model.Player, id int32, num int32) {
	if _, ok := pl.PetEquip.Equips[id]; ok {
		pl.PetEquip.Equips[id].Num += num
	} else {
		pl.PetEquip.Equips[id] = &model.PetEquipItem{
			Id:    id,
			Level: 1,
			Num:   num,
			Exp:   0,
		}
	}

	//同步变化
	SyncPetEquipChange(ctx, pl, id)
}

// SyncPetEquipChange 同步宠物装备
func SyncPetEquipChange(ctx global.IPlayer, pl *model.Player, id int32) {
	res := &proto_pet.PushChangePetEquip{}
	if _, ok := pl.PetEquip.Equips[id]; !ok {
		return
	}

	equip := pl.PetEquip.Equips[id]
	mapequip := make(map[int32]*model.PetEquipItem)
	mapequip[id] = equip
	res.PetEquipOptions = model.ToPetEquipItemProto(mapequip)
	ctx.Send(res)
}

// SyncPetEquipHandbookChange 同步宠物装备图鉴
func SyncPetEquipHandbookChange(ctx global.IPlayer, pl *model.Player, ids []int32) {
	res := &proto_pet.PushPetHandBookChange{}
	for _, v := range ids {
		if _, ok := pl.PetHandbook.HandbookPets[v]; !ok {
			return
		}
	}
	res.Ids = model.ToPetEquipHandbookItemProto(pl.PetHandbook, ids)
	ctx.Send(res)
}

// SyncPetEquipHandbookInfoChange 同步宠物装备图鉴详情
func SyncPetEquipHandbookInfoChange(ctx global.IPlayer, pl *model.Player, ids []int32) {
	res := &proto_pet.PushPetHandBookChange{}
	res.PetEquipHandbookOption = model.ToPetEquipHandbookProto(pl.PetHandbook)
	ctx.Send(res)
}

// 宠物天赋抽卡
func PetGiftDrawCard() int32 {
	var poolMap map[int32]int32
	//判断等级
	confs := config.DrawPool.All()
	var weights []int32
	for _, v := range confs {
		if v.Type == define.CARDPOOL_PETGIFT {
			weights = v.Weight
			break
		}
	}
	rateIndex := utils.WeightIndex(weights)
	//起始的品质是3
	rateIndex += 2
	if _, ok := model.GiftPoolConf[int32(rateIndex+1)]; ok {
		poolMap = model.GiftPoolConf[int32(rateIndex+1)]
	}

	//排序
	ids := make([]int32, 0)
	for id, _ := range poolMap {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(x, y int) bool {
		return ids[x] < ids[y]
	})

	weightArr := make([]int32, 0)
	for k := 0; k < len(ids); k++ {
		weightArr = append(weightArr, poolMap[ids[k]])
	}

	//随机权重
	if len(weightArr) <= 0 {
		return 0
	}

	weightIndex := utils.WeightIndex(weightArr)
	id := ids[weightIndex]
	return id
}

// SyncPetEquipHandbookChange 同步宠物技能变化
func SyncPetSkillItemChange(ctx global.IPlayer, pl *model.Player, ids []int32) {
	res := &proto_pet.PushChangePetSkillChange{}
	for _, v := range ids {
		if _, ok := pl.Pet.Skills[v]; !ok {
			return
		}
	}

	petSkillItems := make(map[int32]*proto_pet.PetSkillItem)
	for _, v := range ids {
		petSkillItems[v] = &proto_pet.PetSkillItem{
			Id:  pl.Pet.Skills[v].Id,
			Num: pl.Pet.Skills[v].Num,
		}
	}
	res.PetSkillItem = petSkillItems
	ctx.Send(res)
}
