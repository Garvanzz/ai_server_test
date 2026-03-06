package internal

import (
	"xfx/core/common"
	"xfx/core/config"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/proto/proto_equip"
)

// AddCollection 添加藏品
func AddCollection(ctx global.IPlayer, pl *model.Player, Id, Num int32) {
	if _, ok := pl.Collection.Collections[Id]; !ok {
		pl.Collection.Collections[Id] = new(model.CollectionOption)
		pl.Collection.Collections[Id].Id = Id
	} else {
		//转成碎片
		items := make(map[int32]int32)
		conf := config.Collection.All()[int64(Id)]
		confItem := config.Item.All()[int64(conf.Fragment)]
		items[conf.Fragment] += Num * confItem.CompositeNeed
		if len(items) > 0 {
			AddItems(ctx, pl, items, true)
		}
		return
	}

	collection := pl.Collection.Collections[Id]
	pl.Collection.Collections[Id] = collection

	pushCol := make(map[int32]*proto_equip.CollectionOption)
	pushCol[collection.Id] = &proto_equip.CollectionOption{
		Id:   collection.Id,
		Star: 0,
	}
	//推送
	pushRes := &proto_equip.PushCollectionChange{
		Collection: pushCol,
	}
	ctx.Send(pushRes)
}

func CollectionSlotHero(ctx global.IPlayer, pl *model.Player) {
	//获取布阵
	lineup := pl.Lineup.LineUps[define.LINEUP_STAGE]
	for j := 0; j < len(pl.Collection.Heros); j++ {
		heroId := pl.Collection.Heros[j]
		if heroId == 0 {
			continue
		}

		if !common.IsHaveValueIntArray(lineup.HeroId, heroId) {
			pl.Collection.Heros[j] = 0
		}
	}

	//布阵
	for k := 0; k < len(lineup.HeroId); k++ {
		if lineup.HeroId[k] <= 0 {
			continue
		}

		if lineup.HeroId[k] >= 3001 && lineup.HeroId[k] <= 3004 {
			continue
		}

		if !common.IsHaveValueIntArray(pl.Collection.Heros, lineup.HeroId[k]) {
			for l := 0; l < len(pl.Collection.Heros); l++ {
				if pl.Collection.Heros[l] == 0 {
					pl.Collection.Heros[l] = lineup.HeroId[k]
					break
				}
			}
		}
	}
}

// CollectionSlotHeroChange 布阵变化
func CollectionSlotHeroChange(ctx global.IPlayer, pl *model.Player) {
	CollectionSlotHero(ctx, pl)
	//推送
	pushRes := &proto_equip.PushCollectionLineUp{
		HeroIds: pl.Collection.Heros,
	}
	ctx.Send(pushRes)
}
