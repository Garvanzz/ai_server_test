package model

import "xfx/proto/proto_fashion"

type Fashion struct {
	FashionItems        map[int32]*FashionItem
	FashionHandbookExp  int32
	FashionHandbookIds  []int32
	HeadWear            map[int32]*HeadWearItem
	HeadWearHandbookExp int32
	HeadWearHandbookIds []int32
}
type HeadWearItem struct {
	Id  int32
	Use bool
	Num int32
}

type FashionItem struct {
	Id  int32
	Use bool
	Num int32
}

func ToFashionProtoByFashion(item map[int32]*FashionItem) map[int32]*proto_fashion.FashionOption {
	opts := make(map[int32]*proto_fashion.FashionOption)
	for _, v := range item {
		opts[v.Id] = &proto_fashion.FashionOption{
			Id:    v.Id,
			IsUse: v.Use,
		}
	}
	return opts
}

func ToHeadWearProtoByHeadWear(item map[int32]*HeadWearItem) map[int32]*proto_fashion.HeadWearOption {
	opts := make(map[int32]*proto_fashion.HeadWearOption)
	for _, v := range item {
		opts[v.Id] = &proto_fashion.HeadWearOption{
			Id:    v.Id,
			IsUse: v.Use,
		}
	}
	return opts
}
