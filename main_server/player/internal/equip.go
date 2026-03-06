package internal

import (
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_equip"
)

// AddEquips 添加装备 key = id, value = 等级
func AddEquips(ctx global.IPlayer, pl *model.Player, Id, Num, Level int32) {
	equip := pl.Equip.Equips
	changedItems := make(map[int32]*proto_equip.EquipOption)

	bagLimit := 999999999
	has := false
	eop := new(model.EquipOption)
	for _, v := range equip {
		if v.CId == Id && v.Level == Level {
			has = true
			v.Num += 1
			if v.Num > int32(bagLimit) {
				v.Num = int32(bagLimit)
			}
			eop = v
			break
		}
	}

	if !has {
		rdb, _ := db.GetEngineByPlayerId(pl.Id)
		equipId, _ := rdb.GetEquipId()
		eop = &model.EquipOption{
			Id:    int32(equipId),
			Level: Level,
			Index: 0,
			Num:   Num,
			CId:   Id,
		}

		equip = append(equip, eop)
	}

	pl.Equip.Equips = equip

	changedItems[eop.Id] = &proto_equip.EquipOption{
		Id:    eop.Id,
		Level: eop.Level,
		Num:   eop.Num,
		Index: eop.Index,
		CId:   Id,
		IsUse: eop.IsUse,
	}

	if len(changedItems) > 0 {
		ctx.Send(&proto_equip.PushEquipChange{EquipOption: changedItems})
	}
}

// AddWeaponry 添加神兵 key = id, value = 等级
func AddWeaponry(ctx global.IPlayer, pl *model.Player, Id, Num int32) {
	log.Debug("领取神器:%v", Id, Num)
	if pl.Equip.Weaponry.WeaponryItems == nil {
		pl.Equip.Weaponry.WeaponryItems = make(map[int32]*model.WeaponryItem)
	}
	if _, ok := pl.Equip.Weaponry.WeaponryItems[Id]; !ok {
		pl.Equip.Weaponry.WeaponryItems[Id] = new(model.WeaponryItem)
		pl.Equip.Weaponry.WeaponryItems[Id].Id = Id
		pl.Equip.Weaponry.WeaponryItems[Id].Level = 1

		confs := config.Weaponry.All()
		conf := confs[int64(Id)]
		pl.Equip.Weaponry.HandbookExp += conf.HandBookExp
		log.Debug("领取新神器:%v", Id)
		//推送恭喜获得
		newTemp := []conf2.ItemE{
			conf2.ItemE{
				ItemId:   Id,
				ItemNum:  1,
				ItemType: define.ItemTypeWeaponry,
			},
		}
		PushPopReward(ctx, global.ItemFormat(newTemp))
	}

	weaponry := pl.Equip.Weaponry.WeaponryItems[Id]
	weaponry.Num += Num
	pl.Equip.Weaponry.WeaponryItems[Id] = weaponry

	SyncWeaponryChange(ctx, pl)
}

// SyncWeaponryChange 推送神兵变化
func SyncWeaponryChange(ctx global.IPlayer, pl *model.Player) {
	pushRes := &proto_equip.PushWeaponryChange{
		WeaponryOption: model.ToWeaponryProto(pl.Equip.Weaponry),
	}
	ctx.Send(pushRes)
}

// AddMount 添加坐骑
func AddMount(ctx global.IPlayer, pl *model.Player, Id, Num int32) {
	if _, ok := pl.Equip.Mount.Mount[Id]; !ok {
		pl.Equip.Mount.Mount[Id] = new(model.MountItemOption)
		pl.Equip.Mount.Mount[Id].Id = Id
		pl.Equip.Mount.Mount[Id].Level = 1

		SyncNotice_AddMount(ctx, pl, Id)

		//推送恭喜获得
		newTemp := []conf2.ItemE{
			conf2.ItemE{
				ItemId:   Id,
				ItemNum:  1,
				ItemType: define.ItemTypeMount,
			},
		}
		PushPopReward(ctx, global.ItemFormat(newTemp))
	}

	mount := pl.Equip.Mount.Mount[Id]
	mount.Num += Num
	pl.Equip.Mount.Mount[Id] = mount

	SyncMountChange(ctx, pl)
}

// SyncMountChange 推送坐骑变化
func SyncMountChange(ctx global.IPlayer, pl *model.Player) {
	pushRes := &proto_equip.PushMountChange{
		MountOption: model.ToMountProto(pl.Equip.Mount),
	}
	ctx.Send(pushRes)
}

// AddBrace 添加背饰
func AddBrace(ctx global.IPlayer, pl *model.Player, Id, Num int32) {
	if _, ok := pl.Equip.Mount.Mount[Id]; !ok {
		pl.Equip.Brace.BraceItems[Id] = new(model.BraceItem)
		pl.Equip.Brace.BraceItems[Id].Id = Id
		pl.Equip.Brace.BraceItems[Id].Level = 1

		//图鉴经验
		confs := config.Braces.All()
		conf := confs[int64(Id)]
		pl.Equip.Brace.HandbookExp += conf.HandBookExp

		//推送恭喜获得
		newTemp := []conf2.ItemE{
			conf2.ItemE{
				ItemId:   Id,
				ItemNum:  1,
				ItemType: define.ItemTypeBraces,
			},
		}
		PushPopReward(ctx, global.ItemFormat(newTemp))
	}

	brace := pl.Equip.Brace.BraceItems[Id]
	brace.Num += Num
	pl.Equip.Brace.BraceItems[Id] = brace
}

// AddLearning 添加心得
func AddLearning(ctx global.IPlayer, pl *model.Player, Id, Num int32) {
	if _, ok := pl.Divine.Learning[Id]; !ok {
		pl.Divine.Learning[Id] = new(model.LearningOption)
		pl.Divine.Learning[Id].Id = Id
	}

	pl.Divine.Learning[Id].Num += Num
	item := make(map[int32]*proto_equip.LearningOption)
	item[Id] = &proto_equip.LearningOption{
		Id:  Id,
		Num: pl.Divine.Learning[Id].Num,
	}

	pushRes := &proto_equip.PushLearningOptionChange{
		LearningOptions: item,
	}

	PushLearningReawardPop(ctx, pl, []int32{Id})

	ctx.Send(pushRes)
}

// PushLearningReawardPop 推送心得奖励弹窗
func PushLearningReawardPop(ctx global.IPlayer, pl *model.Player, ids []int32) {
	learning := make(map[int32]int32)
	for _, v := range ids {
		if _, ok := pl.Divine.Learning[v]; ok {
			learning[v] = pl.Divine.Learning[v].Num
		}
	}
	ctx.Send(&proto_equip.PushLearningPopReward{Rewards: learning})
}
