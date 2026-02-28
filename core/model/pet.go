package model

import (
	"xfx/proto/proto_pet"
)

// 宠物
type Pet struct {
	Pets          map[int32]*PetItem
	DispatchPets  map[int32]int32
	ResetFreeNum  int32
	ResetFreeTime int64
	Skills        map[int32]*PetSkill
}

// 宠物装备
type PetEquip struct {
	Equips map[int32]*PetEquipItem
}

type PetItem struct {
	Id          int32
	Level       int32
	Stage       int32
	Star        int32
	Name        string
	Gifts       []int32
	Equips      []int32
	SkillIds    []int32
	CacheXilian []int32
	IsUse       bool
}

type PetEquipItem struct {
	Id    int32
	Level int32
	Exp   int32
	Num   int32
}

var GiftPoolConf map[int32]map[int32]int32

// 技能
type PetSkill struct {
	Id  int32
	Num int32
}

// 图鉴
type PetHandbook struct {
	HandbookPets      map[int32]*HandbookPetEquip
	PetHandbookOption *PetHandbookOption
}

// 图鉴
type PetHandbookOption struct {
	Level int32
	Exp   int32
	GetId []int32
}

type HandbookPetEquip struct {
	Id       int32
	IsGetExp bool
	GetExp   int32
}

// 抽卡
type PetDraw struct {
	BdCount int32
	Score   int32
	Awards  []int32
	Pools   map[int32]*PetDrawPool
}

type PetDrawPool struct {
	PoolId   int32
	StarTime int64
	DrawNum  int32
	Recores  []*DrawPetRecord
}

type DrawPetRecord struct {
	Id   int32
	Num  int32
	Type int32
}

// 序列化
func ToPetItemProto(petMap map[int32]*PetItem) map[int32]*proto_pet.PetItem {
	petS := make(map[int32]*proto_pet.PetItem, 0)
	for _, v := range petMap {
		petS[v.Id] = &proto_pet.PetItem{
			Id:        v.Id,
			Level:     v.Level,
			Star:      v.Star,
			Name:      v.Name,
			Stage:     v.Stage,
			Gift:      v.Gifts,
			Equip:     v.Equips,
			SkillId:   v.SkillIds,
			CacheGift: v.CacheXilian,
		}
	}

	return petS
}

// 序列化
func ToPetEquipItemProto(petMap map[int32]*PetEquipItem) map[int32]*proto_pet.PetEquipOption {
	petS := make(map[int32]*proto_pet.PetEquipOption, 0)
	for _, v := range petMap {
		petS[v.Id] = &proto_pet.PetEquipOption{
			Id:    v.Id,
			Level: v.Level,
			Num:   v.Num,
		}
	}

	return petS
}

// 序列化
func ToPetEquipHandbookProto(handbook *PetHandbook) *proto_pet.PetEquipHandbookOption {
	return &proto_pet.PetEquipHandbookOption{
		Level: handbook.PetHandbookOption.Level,
		Exp:   handbook.PetHandbookOption.Exp,
		GetId: handbook.PetHandbookOption.GetId,
	}
}

// 序列化
func ToPetSkillItemProto(skills map[int32]*PetSkill) map[int32]*proto_pet.PetSkillItem {
	petSkillItems := make(map[int32]*proto_pet.PetSkillItem)
	for _, v := range skills {
		petSkillItems[v.Id] = &proto_pet.PetSkillItem{
			Id:  v.Id,
			Num: v.Num,
		}
	}

	return petSkillItems
}

func ToPetEquipHandbookListProto(handbook *PetHandbook) map[int32]*proto_pet.PetEquipHandbook {
	mapHandbook := make(map[int32]*proto_pet.PetEquipHandbook)
	for _, v := range handbook.HandbookPets {
		mapHandbook[v.Id] = &proto_pet.PetEquipHandbook{
			Id:       v.Id,
			IsGetExp: v.IsGetExp,
			GetExp:   v.GetExp,
		}
	}
	return mapHandbook
}

func ToPetEquipHandbookItemProto(handbook *PetHandbook, ids []int32) map[int32]*proto_pet.PetEquipHandbook {
	mapHandbook := make(map[int32]*proto_pet.PetEquipHandbook)
	for _, id := range ids {
		v := handbook.HandbookPets[id]
		mapHandbook[v.Id] = &proto_pet.PetEquipHandbook{
			Id:       v.Id,
			IsGetExp: v.IsGetExp,
			GetExp:   v.GetExp,
		}
	}
	return mapHandbook
}

// 序列化
func ToPetDrawProto(petDraw *PetDraw) *proto_pet.PetDrawOption {
	pools := make(map[int32]*proto_pet.PetDrawPoolOption)
	for _, v := range petDraw.Pools {
		pools[v.PoolId] = &proto_pet.PetDrawPoolOption{
			PoolId:        v.PoolId,
			PoolStartTime: v.StarTime,
			DrawNum:       v.DrawNum,
		}
	}
	return &proto_pet.PetDrawOption{
		BdCount: petDraw.BdCount,
		Score:   petDraw.Score,
		Pools:   pools,
		Rewards: petDraw.Awards,
	}
}
