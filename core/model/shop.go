package model

import (
	"xfx/proto/proto_shop"
)

type PlayerShop struct {
	Shops map[int32]*ShopType
}

type ShopType struct {
	LastTime  int64
	ShopItems map[int]*ShopItem
}

type ShopItem struct {
	Id  int32
	Num int32
}

func ToShopTypeProto(shoptyps map[int32]*ShopType) map[int32]*proto_shop.TypeShopItem {
	shops := make(map[int32]*proto_shop.TypeShopItem)
	for typs, vals := range shoptyps {
		shopitems := make(map[int32]*proto_shop.ShopItem)
		for k, b := range vals.ShopItems {
			shopitems[int32(k)] = &proto_shop.ShopItem{
				Id:  b.Id,
				Num: b.Num,
			}
		}
		shops[typs] = &proto_shop.TypeShopItem{
			ShopItem: shopitems,
			LastTime: vals.LastTime,
		}
	}

	return shops
}
