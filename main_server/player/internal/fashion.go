package internal

import (
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/proto/proto_fashion"
)

// AddFashion 添加时装
func AddFashion(ctx global.IPlayer, pl *model.Player, Id, Num int32) {
	if _, ok := pl.Fashion.FashionItems[Id]; !ok {
		pl.Fashion.FashionItems[Id] = new(model.FashionItem)
		pl.Fashion.FashionItems[Id].Id = Id
		if Id == define.DefaultFashionId {
			pl.Fashion.FashionItems[Id].Use = true
		}

		confs := config.Fashion.All()
		conf := confs[int64(Id)]
		pl.Fashion.FashionHandbookExp += conf.HandBookExp
	}

	fashion := pl.Fashion.FashionItems[Id]
	fashion.Num += Num
	pl.Fashion.FashionItems[Id] = fashion

	SyncFashionChange(ctx, pl)
}

// SyncFashionChange 推送时装变化
func SyncFashionChange(ctx global.IPlayer, pl *model.Player) {
	pushRes := &proto_fashion.PushFashionChange{
		HandbookExp: pl.Fashion.FashionHandbookExp,
		HandbookIds: pl.Fashion.FashionHandbookIds,
		Fashions:    model.ToFashionProtoByFashion(pl.Fashion.FashionItems),
	}
	ctx.Send(pushRes)
}

// AddHeadWear 添加头饰
func AddHeadWear(ctx global.IPlayer, pl *model.Player, Id, Num int32) {
	if _, ok := pl.Fashion.HeadWear[Id]; !ok {
		pl.Fashion.HeadWear[Id] = new(model.HeadWearItem)
		pl.Fashion.HeadWear[Id].Id = Id
		if Id == define.DefaultHeadWearId {
			pl.Fashion.HeadWear[Id].Use = true
		}

		confs := config.Headwear.All()
		conf := confs[int64(Id)]
		pl.Fashion.HeadWearHandbookExp += conf.HandBookExp

		//推送恭喜获得
		newTemp := []conf2.ItemE{
			conf2.ItemE{
				ItemId:   Id,
				ItemNum:  1,
				ItemType: define.ItemTypeHeadWear,
			},
		}
		PushPopReward(ctx, global.ItemFormat(newTemp))
	}

	fashion := pl.Fashion.HeadWear[Id]
	fashion.Num += Num
	pl.Fashion.HeadWear[Id] = fashion

	SyncHeadWearChange(ctx, pl)
}

// SyncHeadWearChange 推送头饰变化
func SyncHeadWearChange(ctx global.IPlayer, pl *model.Player) {
	pushRes := &proto_fashion.PushHeadWearChange{
		HandbookExp: pl.Fashion.HeadWearHandbookExp,
		HandbookIds: pl.Fashion.HeadWearHandbookIds,
		Headwear:    model.ToHeadWearProtoByHeadWear(pl.Fashion.HeadWear),
	}
	ctx.Send(pushRes)
}
