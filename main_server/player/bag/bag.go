package bag

import (
	"encoding/json"
	"fmt"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_item"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.Bag = new(model.Bag)
	pl.Bag.Items = make(map[int32]int32, 0)

	pl.Bag.Items[104] = 5000
	pl.Bag.Items[105] = 5000
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Bag)
	if err != nil {
		log.Error("player[%v],save bag marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerBag, pl.Id), j)
	} else {
		db.RedisAsyncExec(pl.Cache.Self, define.RedisRetNone, nil, "SET", fmt.Sprintf("%s:%d", define.PlayerBag, pl.Id), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerBag, pl.Id))
	if err != nil {
		log.Error("player[%v],load bag error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Bag)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load bag unmarshal error:%v", pl.Id, err)
	}

	pl.Bag = m
}

// AddAward 添加奖励
func AddAward(ctx global.IPlayer, pl *model.Player, awards []conf2.ItemE, isPush bool) {
	if len(awards) == 0 {
		log.Debug("add nil award, player id : %v", pl.Id)
		return
	}

	items := make(map[int32]int32)
	for _, award := range awards {
		switch award.ItemType {
		case define.ItemTypeItem:
			items[award.ItemId] += award.ItemNum
			//对道具的判断
			confItem := config.Item.All()[int64(award.ItemId)]
			//藏品碎片
			if confItem.Id > 0 && confItem.Type == define.BagItemTypeCollectionPiece {
				if _, ok := pl.Collection.Collections[confItem.CompositeItem]; ok {
					continue
				}

				//自动激活
				conf := config.Collection.All()[int64(confItem.CompositeItem)]
				if conf.Id <= 0 {
					continue
				}

				if pl.Bag.Items[confItem.Id]+award.ItemNum >= conf.ActiveFragNum {
					internal.AddCollection(ctx, pl, conf.Id, 1)
				}
			}
		case define.ItemTypeHero:
			conf := config.Hero.All()[int64(award.ItemId)]
			if _, ok := pl.Hero.Hero[award.ItemId]; ok { // 如果有转换成碎片
				confItem := config.Item.All()[int64(conf.Fragment)]
				items[conf.Fragment] += award.ItemNum * confItem.CompositeNeed
			} else {
				pl.Hero.Hero[award.ItemId] = &model.HeroOption{
					Id:          award.ItemId,
					Star:        0,
					Level:       1,
					Exp:         0,
					Stage:       0,
					Skin:        "Default",
					Cultivation: make(map[int32]int32),
				}

				//更新技能
				internal.CheckHeroSkill(ctx, pl, award.ItemId, 1)

				//增加图鉴
				confs := config.Handbook.All()
				conf := conf2.HandBook{}
				for _, v := range confs {
					if v.TargetId == award.ItemId {
						conf = v
						break
					}
				}
				if conf.Id <= 0 {
					log.Error("没有找到图鉴:%v", award.ItemId)
					continue
				}

				if _, ok = pl.Handbook.Handbooks[conf.Id]; !ok {
					hand := new(model.HandbookHero)
					hand.Id = conf.Id
					hand.IsGetExp = true
					hand.GetExp = conf.Exp
					pl.Handbook.Handbooks[conf.Id] = hand
				}

				//同步
				internal.SyncHeroChange(ctx, pl, award.ItemId)
			}
		case define.ItemTypeSkin:
		case define.ItemTypeEquip:
			//要对装备赋予等级
			heroId := int32(pl.GetProp(define.PlayerPropHeroId))
			mainHero := pl.Hero.Hero[heroId]
			minLevel := mainHero.Level - 5
			if minLevel <= 1 {
				minLevel = 1
			}
			maxLevel := mainHero.Level + 5
			rangeLevel := utils.RandInt(minLevel, maxLevel)
			//添加装备
			internal.AddEquips(ctx, pl, award.ItemId, award.ItemNum, rangeLevel)
		case define.ItemTypeMagic:
			internal.AddMagic(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypeWeaponry:
			internal.AddWeaponry(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypeMount:
			internal.AddMount(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypeCollect:
			internal.AddCollection(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypeBraces:
			internal.AddBrace(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypeLearning:
			internal.AddLearning(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypePet:
			if _, ok := pl.Pet.Pets[award.ItemId]; ok { // 如果有转换成碎片
				conf := config.Pet.All()[int64(award.ItemId)]
				confItem := config.Item.All()[int64(conf.Fragment)]
				items[conf.Fragment] += award.ItemNum * confItem.CompositeNeed
			} else {
				pl.Pet.Pets[award.ItemId] = &model.PetItem{
					Id:       award.ItemId,
					Star:     0,
					Level:    1,
					Stage:    0,
					Gifts:    make([]int32, 3),
					Equips:   make([]int32, 4),
					SkillIds: make([]int32, 4),
				}

				//同步
				internal.SyncPetChange(ctx, pl, award.ItemId)

				//推送恭喜获得
				newPet := []conf2.ItemE{
					conf2.ItemE{
						ItemId:   award.ItemId,
						ItemNum:  1,
						ItemType: define.ItemTypePet,
					},
				}
				internal.PushPopReward(ctx, global.ItemFormat(newPet))
			}
		case define.ItemTypePetEquip:
			internal.AddPetEquip(ctx, pl, award.ItemId, award.ItemNum)

			//增加图鉴
			//confs := config.CfgMgr.AllJson()["PetEquipHandbook"].(map[int64]conf2.PetEquipHandbook)
			//conf := conf2.PetEquipHandbook{}
			//for _, v := range confs {
			//	if v.TargetId == award.ItemId {
			//		conf = v
			//		break
			//	}
			//}
			//if conf.Id <= 0 {
			//	log.Error("没有找到宠物装备图鉴:%v", award.ItemId)
			//	continue
			//}
			//
			//if _, ok = pl.PetHandbook.HandbookPets[conf.Id]; !ok {
			//	hand := new(model.HandbookPetEquip)
			//	hand.Id = conf.Id
			//	hand.IsGetExp = true
			//	hand.GetExp = conf.Exp
			//	pl.PetHandbook.HandbookPets[conf.Id] = hand
			//}
			//
			////同步
			//global.SyncPetEquipHandbookChange(ctx, pl, award.ItemId)
		case define.ItemTypePetSkill:
		case define.ItemTypeHeadFrame:
			internal.AddPlayerPropHeadFrame(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypeTitle:
			internal.AddPlayerPropTitle(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypeBubble:
			internal.AddPlayerPropBubble(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemGuildMaterial:
			internal.AddGuildMaterical(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemGuildElement:
		case define.ItemTypeFish:
		case define.ItemTypeHeadWear:
			internal.AddHeadWear(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypeFashion:
			internal.AddFashion(ctx, pl, award.ItemId, award.ItemNum)
		case define.ItemTypePassportScore:
			internal.AddPassportScore(ctx, pl, award.ItemNum)
			score := make(map[int32]int32)
			score[award.ItemId] = award.ItemNum
			internal.PushResPassportScoreChange(ctx, pl, score)
		}
	}

	if len(items) > 0 {
		internal.AddItems(ctx, pl, items, isPush)
	}
}

// ReqBag 请求背包数据
func ReqBag(ctx global.IPlayer, pl *model.Player, req *proto_item.C2SBag) {
	resp := new(proto_item.S2CBag)
	resp.Bag = pl.Bag.Items
	ctx.Send(resp)
}

// ReqUseItem 使用道具
func ReqUseItem(ctx global.IPlayer, pl *model.Player, req *proto_item.C2SUseItem) {
	var result = &proto_item.S2CUseItem{}
	if req.ItemNum <= 0 {
		result.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		log.Error("ReqUseItem item num error:%v", req.ItemNum)
		ctx.Send(result)
		return
	}

	consume := map[int32]int32{req.ItemId: req.ItemNum}
	if !internal.CheckItemsEnough(pl, consume) {
		log.Error("道具数量不足")
		result.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(result)
		return
	}

	if req.ItemNum > config.Global.Get().UseItemMaxCount {
		result.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(result)
		return
	}

	// 判断道具类型
	itemConfig := config.Item.All()[int64(req.ItemId)]
	if itemConfig.Type == define.BagItemTypeDropBox && itemConfig.UseValue <= 0 {
		log.Error("道具不可使用:%v", req.ItemId)
		result.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(result)
		return
	}

	// 扣除消耗
	internal.SubItems(ctx, pl, consume)

	awards := make([]conf2.ItemE, 0)
	switch itemConfig.Type {
	case define.BagItemTypeDropBox: // 开宝箱
		for i := int32(0); i < req.ItemNum; i++ {
			awards = append(awards, GetDrop(itemConfig.UseValue, req.GetAppointId())...)
		}
	default:
		log.Error("use item type error:%v", itemConfig.Type)
		result.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(result)
		return
	}

	// 添加奖励
	if len(awards) > 0 {
		AddAward(ctx, pl, awards, true)
	}
	log.Debug("award: %v", awards)

	ctx.Send(result)
}

// ReqCompositionItem 合成道具
func ReqCompositionItem(ctx global.IPlayer, pl *model.Player, req *proto_item.C2SCompositionItem) {
	result := new(proto_item.S2CCompositionItem)
	if req.ItemId <= 0 {
		log.Error("ReqCompositionItem item num error:%v", req.GetItemId())
		result.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(result)
		return
	}

	itemConfig := config.Item.All()[int64(req.ItemId)]
	if itemConfig.IsComposite == false {
		log.Error("道具不可合成")
		result.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(result)
		return
	}

	consume := map[int32]int32{req.ItemId: itemConfig.CompositeNeed}
	if !internal.CheckItemsEnough(pl, consume) {
		log.Error("道具数量不足")
		result.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(result)
		return
	}

	//判断是否存在
	switch itemConfig.Type {
	case define.BagItemTypeHeroPiece: //英雄碎片
		if _, ok := pl.Hero.Hero[itemConfig.CompositeItem]; ok {
			result.Code = proto_public.CommonErrorCode_ERR_ALHASPATERNER
			ctx.Send(result)
			return
		}
	}

	internal.SubItems(ctx, pl, consume)

	switch itemConfig.Type {
	case define.BagItemTypeHeroPiece: //英雄碎片
		award := []conf2.ItemE{{
			ItemType: define.ItemTypeHero,
			ItemId:   itemConfig.CompositeItem,
			ItemNum:  1,
		}}

		AddAward(ctx, pl, award, true)
	default:
		result.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(result)
		return
	}

	result.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(result)
}

// ReqSellItem 售卖道具
func ReqSellItem(ctx global.IPlayer, pl *model.Player, req *proto_item.C2SSellItem) {
	result := new(proto_item.S2CSellItem)
	if req.ItemId <= 0 {
		result.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(result)
		log.Error("ReqSellItem item num error:%v", req.GetItemId())
		return
	}

	itemConfig := config.Item.All()[int64(req.ItemId)]
	if itemConfig.IsSell == false {
		result.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(result)
		log.Error("道具不可售卖")
		return
	}

	consume := map[int32]int32{req.ItemId: 1}
	if !internal.CheckItemsEnough(pl, consume) {
		result.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(result)
		log.Error("道具数量不足")
		return
	}

	internal.SubItems(ctx, pl, consume)

	award := []conf2.ItemE{{
		ItemType: define.ItemTypeItem,
		ItemId:   define.ItemIdMoney,
		ItemNum:  itemConfig.SellValue,
	}}
	AddAward(ctx, pl, award, true)
}
