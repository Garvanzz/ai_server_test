package model

import "xfx/proto/proto_equip"

// 藏品
type Collection struct {
	Collections     map[int32]*CollectionOption
	CollectionSlots map[int32]map[int32]int32
	Heros           []int32 //布阵列表
}

type CollectionOption struct {
	Id   int32
	Star int32
}

// 藏品
func ToCollectionProto(opt map[int32]*CollectionOption) map[int32]*proto_equip.CollectionOption {
	auras := make(map[int32]*proto_equip.CollectionOption)
	for k, v := range opt {
		auras[k] = &proto_equip.CollectionOption{
			Id:   v.Id,
			Star: v.Star,
		}
	}
	return auras
}

// 藏品槽位
func ToCollectionSlotProto(opt map[int32]map[int32]int32) map[int32]*proto_equip.CollectionSlot {
	auras := make(map[int32]*proto_equip.CollectionSlot)
	for k, v := range opt {
		auras[k] = &proto_equip.CollectionSlot{
			Ids: v,
		}
	}
	return auras
}
