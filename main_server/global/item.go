package global

import (
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/proto/proto_public"
)

func MergeItemE(items []conf.ItemE) map[int32]int32 {
	m := make(map[int32]int32, len(items))

	for _, v := range items {
		if v.ItemNum <= 0 {
			continue
		}
		m[v.ItemId] += v.ItemNum
	}

	return m
}

func MergeItemEWithMap(m map[int32]int32, items []conf.ItemE) map[int32]int32 {
	if m == nil {
		m = make(map[int32]int32, len(items))
	}

	for _, v := range items {
		if v.ItemNum <= 0 {
			continue
		}
		m[v.ItemId] += v.ItemNum
	}

	return m
}

func ItemFormatWithMap(items map[int32]int32) []*proto_public.Item {
	list := make([]*proto_public.Item, 0)
	for id, num := range items {
		list = append(list, &proto_public.Item{
			ItemId:   id,
			ItemNum:  num,
			ItemType: define.ItemTypeItem,
		})
	}
	return list
}

func ItemFormat(items []conf.ItemE) []*proto_public.Item {
	list := make([]*proto_public.Item, 0)
	for _, v := range items {
		list = append(list, &proto_public.Item{
			ItemId:   v.ItemId,
			ItemNum:  v.ItemNum,
			ItemType: v.ItemType,
		})
	}
	return list
}

func PassportScoreFormatWithMap(items map[int32]int32) []*proto_public.Item {
	list := make([]*proto_public.Item, 0)
	for id, num := range items {
		list = append(list, &proto_public.Item{
			ItemId:   id,
			ItemNum:  num,
			ItemType: define.ItemTypePassportScore,
		})
	}
	return list
}
