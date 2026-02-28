package internal

import (
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_item"
	"xfx/proto/proto_public"
)

// AddItems 添加道具
func AddItems(ctx global.IPlayer, pl *model.Player, items map[int32]int32, isPush bool) {
	if len(items) == 0 {
		log.Debug("add nil items, player id : %v", pl.Id)
		return
	}

	bag := pl.Bag.Items
	changedItems := make(map[int32]int32)

	bagLimit := 999999999

	for k, v := range items {
		if bagLimit == 0 {
			bag[k] += v
			changedItems[k] = bag[k]
		} else if bagLimit > 0 {
			bag[k] += v
			if bag[k] > int32(bagLimit) {
				bag[k] = int32(bagLimit)
			}
			changedItems[k] = bag[k]
		}
	}

	if len(changedItems) > 0 {
		ctx.Send(&proto_item.PushChange{Change: changedItems})
		if isPush {
			ctx.Send(&proto_item.PushPopReward{Items: global.ItemFormatWithMap(items)})
		}
	}
}

// CheckItemsEnough 检查道具是否足够
func CheckItemsEnough(pl *model.Player, items map[int32]int32) bool {
	bag := pl.Bag.Items
	for k, v := range items {
		if bag[k] < v {
			return false
		}
	}
	return true
}

// SubItems 删除道具
func SubItems(ctx global.IPlayer, pl *model.Player, items map[int32]int32) {
	if len(items) == 0 {
		log.Debug("error : subItems sub nil items, player id : %v", pl.Id)
		return
	}

	bag := pl.Bag.Items
	changedItems := make(map[int32]int32)
	for k, v := range items {
		switch k {
		default:
			if _, ok := bag[k]; !ok {
				return
			}

			if bag[k] < v {
				return
			} else {
				bag[k] -= v
			}
			changedItems[k] = bag[k]
		}
	}

	if len(changedItems) > 0 {
		ctx.Send(&proto_item.PushChange{Change: changedItems})
	}
}

// 推送恭喜获得
func PushPopReward(ctx global.IPlayer, award []*proto_public.Item) {
	ctx.Send(&proto_item.PushPopReward{Items: award})
}
