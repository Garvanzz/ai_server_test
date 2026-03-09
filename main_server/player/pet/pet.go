package pet

import (
	"encoding/json"
	"fmt"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_pet"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.Pet = new(model.Pet)
	pl.Pet.Pets = make(map[int32]*model.PetItem)
	pl.Pet.ResetFreeNum = 1
	pl.Pet.ResetFreeTime = utils.Now().Unix()
	pl.Pet.Skills = make(map[int32]*model.PetSkill)

	//默认玉兔
	pl.Pet.Pets[define.DefaultPetId] = &model.PetItem{
		Id:       define.DefaultPetId,
		Level:    1,
		SkillIds: make([]int32, 4),
		Gifts:    make([]int32, 3),
		Equips:   make([]int32, 4),
	}
	pl.Pet.DispatchPets = make(map[int32]int32)

	pl.PetEquip = new(model.PetEquip)
	pl.PetEquip.Equips = make(map[int32]*model.PetEquipItem)

	pl.PetHandbook = new(model.PetHandbook)
	pl.PetHandbook.PetHandbookOption = new(model.PetHandbookOption)
	pl.PetHandbook.HandbookPets = make(map[int32]*model.HandbookPetEquip)

	PetDrawInit(pl)
}

func Save(pl *model.Player, isSync bool) {
	//宠物
	j, err := json.Marshal(pl.Pet)
	if err != nil {
		log.Error("player[%v],save Pet marshal error:%v", pl.Id, err)
		return
	}

	//装备
	je, err := json.Marshal(pl.PetEquip)
	if err != nil {
		log.Error("player[%v],save PetEquip marshal error:%v", pl.Id, err)
		return
	}

	//图鉴
	jh, err := json.Marshal(pl.PetHandbook)
	if err != nil {
		log.Error("player[%v],save PetHandbookOption marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerPet, pl.Id), j)
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerPetEquip, pl.Id), je)
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerPetHandbook, pl.Id), jh)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}

	PetDrawSave(pl, isSync)
}

func Load(pl *model.Player) {

	//宠物
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerPet, pl.Id))
	if err != nil {
		log.Error("player[%v],load pet error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Pet)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load pet unmarshal error:%v", pl.Id, err)
	}

	pl.Pet = m

	//装备
	ereply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerPetEquip, pl.Id))
	if err != nil {
		log.Error("player[%v],load pet error:%v", pl.Id, err)
		return
	}

	em := new(model.PetEquip)
	err = json.Unmarshal(ereply.([]byte), &em)
	if err != nil {
		log.Error("player[%v],load pet unmarshal error:%v", pl.Id, err)
	}

	pl.PetEquip = em

	//图鉴
	hreply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerPetHandbook, pl.Id))
	if err != nil {
		log.Error("player[%v],load pet error:%v", pl.Id, err)
		return
	}

	hm := new(model.PetHandbook)
	err = json.Unmarshal(hreply.([]byte), &hm)
	if err != nil {
		log.Error("player[%v],load pet unmarshal error:%v", pl.Id, err)
	}

	pl.PetHandbook = hm

	PetDrawLoad(pl)
}

// ReqInitPet 请求宠物数据
func ReqInitPet(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SInitPet) {
	resp := new(proto_pet.S2CInitPet)
	resp.PetItems = model.ToPetItemProto(pl.Pet.Pets)
	resp.PetEquipOptions = model.ToPetEquipItemProto(pl.PetEquip.Equips)
	resp.DispatchPets = pl.Pet.DispatchPets
	resp.PetEquipHandbookOption = model.ToPetEquipHandbookProto(pl.PetHandbook)
	resp.PetSkills = model.ToPetSkillItemProto(pl.Pet.Skills)
	resp.Ids = model.ToPetEquipHandbookListProto(pl.PetHandbook)

	if !utils.IsSameWeekBySec(utils.Now().Unix(), pl.Pet.ResetFreeTime) {
		pl.Pet.ResetFreeTime = utils.Now().Unix()
		pl.Pet.ResetFreeNum = 1
	}
	resp.ResetFreeNum = pl.Pet.ResetFreeNum
	ctx.Send(resp)
}

// ReqLevelUpPet 请求宠物升级
func ReqPetLevelUp(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SUpLevelPet) {
	res := new(proto_pet.S2CUpLevelPet)
	if pl.Bag.Items[define.ItemIdXiantao] <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[req.Id]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	pet := pl.Pet.Pets[req.Id]
	if pet.Level >= define.LevelPetMaxLimit {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	costs := make(map[int32]int32)

	levelRange := int32(10)
	//升阶
	if pet.Level%levelRange == 0 && pet.Stage < pet.Level/levelRange {
		res.Code = proto_public.CommonErrorCode_ERR_NoUpLevel
		ctx.Send(res)
		return
	} else {
		levelConf, _ := config.PetUpLevel.Find(int64(req.Id))
		offsetLevel := pet.Level - pet.Stage*levelRange

		//道具数量
		itemCount := int32(0)
		if _, ok := pl.Bag.Items[define.ItemIdXiantao]; ok {
			itemCount = pl.Bag.Items[define.ItemIdXiantao]
		}

		MaxLevel := calculateMaxLevel(itemCount, levelConf.UpLevelCondition[pet.Stage], offsetLevel)
		log.Debug("升级的最大等级:%v", MaxLevel)
		if MaxLevel-pet.Level < req.Count {
			res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
			ctx.Send(res)
			return
		}

		log.Debug("宠物升级的消耗:%v", (levelConf.UpLevelCondition[pet.Stage]/2)*((offsetLevel+req.Count)*(offsetLevel+req.Count+1)-(offsetLevel)*(offsetLevel+1)))
		costs[define.ItemIdMoney] = (levelConf.UpLevelCondition[pet.Stage] / 2) * ((offsetLevel+req.Count)*(offsetLevel+req.Count+1) - (offsetLevel)*(offsetLevel+1))
	}

	//判断道具是否足够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	pet.Level += req.Count
	pl.Pet.Pets[req.Id] = pet

	//扣除道具
	internal.SubItems(ctx, pl, costs)

	internal.SyncPetChange(ctx, pl, req.Id)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 计算能升多少次
func calculateMaxLevel(totalCoins, baseCoins, level int32) int32 {
	n := level
	requiredCoins := int32(0)

	// 循环计算每一级所需金币，直到超过总金币
	for {
		n++
		requiredCoins = (baseCoins / 2) * n * (n + 1)

		if requiredCoins-((baseCoins/2)*(level-1)*level) > totalCoins {
			// 如果当前等级的金币消耗超过总金币，返回上一级
			return n - 1
		}
	}
}

// ReqPetStageUp 请求宠物升阶
func ReqPetStageUp(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SUpStagePet) {
	res := new(proto_pet.S2CUpStagePet)
	if pl.Bag.Items[define.ItemIdXiantao] <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[req.Id]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	pet := pl.Pet.Pets[req.Id]
	levelRange := int32(10)

	if pet.Stage >= define.LevelMaxLimit/levelRange {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	offse := pet.Level % levelRange
	//突破
	if offse != 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoAdvance
		ctx.Send(res)
		return
	}

	if pet.Stage*levelRange >= pet.Level {
		res.Code = proto_public.CommonErrorCode_ERR_NoAdvance
		ctx.Send(res)
		return
	}

	stageConf, _ := config.PetUpStage.Find(int64(req.Id))
	costs := make(map[int32]int32)
	costs[define.ItemIdXiantao] = stageConf.UpStageCondition[pet.Stage]

	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	pet.Stage += 1
	pl.Pet.Pets[req.Id] = pet

	//扣除道具
	internal.SubItems(ctx, pl, costs)

	internal.SyncPetChange(ctx, pl, req.Id)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// ReqPetStarUp 请求宠物升星
func ReqPetStarUp(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SUpStarPet) {
	res := new(proto_pet.S2CUpStarPet)
	if _, ok := pl.Pet.Pets[req.Id]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	pet := pl.Pet.Pets[req.Id]
	costs := make(map[int32]int32)
	UpStarConf, _ := config.PetUpStar.Find(int64(req.Id))
	log.Debug("需要消耗碎片:%v", UpStarConf.NeedCostNum[pet.Star])
	costs[UpStarConf.NeedId] = UpStarConf.NeedCostNum[pet.Star]

	//判断够不够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	if pet.Star >= 6 {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	pet.Star += 1
	pl.Pet.Pets[req.Id] = pet

	//扣除道具
	internal.SubItems(ctx, pl, costs)

	internal.SyncPetChange(ctx, pl, req.Id)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// ReqPetReset 请求宠物重置
func ReqPetReset(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SResetPet) {
	res := new(proto_pet.S2CResetPet)
	if _, ok := pl.Pet.Pets[req.Id]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPet
		ctx.Send(res)
		return
	}

	costs := make(map[int32]int32, 0)
	if pl.Pet.ResetFreeNum <= 0 {
		costs[define.ItemIdPetResetStore] = 1
		if !internal.CheckItemsEnough(pl, costs) {
			res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
			ctx.Send(res)
		}
	} else {
		pl.Pet.ResetFreeNum -= 1
		pl.Pet.ResetFreeTime = utils.Now().Unix()
	}

	pet := pl.Pet.Pets[req.Id]
	costs[define.ItemIdXiantao] = 0

	//等级
	levelConf, _ := config.PetUpLevel.Find(int64(req.Id))
	sumCoins := calculateSumCostNumByLevel(levelConf.UpLevelCondition[pet.Stage], pet.Level)
	costs[define.ItemIdXiantao] += sumCoins
	pl.Pet.Pets[req.Id].Level = 1

	//阶数
	stageConf, _ := config.PetUpStage.Find(int64(req.Id))
	sumStors := int32(0)
	for k := int32(0); k < pet.Stage; k++ {
		sumStors += stageConf.UpStageCondition[k]
	}
	costs[define.ItemIdXiantao] += sumStors
	pl.Pet.Pets[req.Id].Stage = 0

	//星级
	starConf, _ := config.PetUpStar.Find(int64(req.Id))
	sumStar := int32(0)
	for k := int32(0); k < pet.Star; k++ {
		sumStar += starConf.NeedCostNum[k]
	}
	costs[starConf.NeedId] += sumStar
	pl.Pet.Pets[req.Id].Star = 0

	items := make([]conf2.ItemE, 0)
	for k, v := range costs {
		if v <= 0 {
			continue
		}
		items = append(items, conf2.ItemE{
			ItemId:   k,
			ItemNum:  v,
			ItemType: define.ItemTypeItem,
		})
	}
	//添加背包
	bag.AddAward(ctx, pl, items, true)

	//宠物变化
	internal.SyncPetChange(ctx, pl, req.Id)
	res.Code = proto_public.CommonErrorCode_ERR_OK
	res.ResetFreeNum = pl.Pet.ResetFreeNum
	ctx.Send(res)
	return
}

// 计算升到等级需要花费的总金币数量
func calculateSumCostNumByLevel(baseCoins, level int32) int32 {
	totalCoins := (baseCoins / 2) * level * (level + 1)
	return totalCoins
}

// 派遣
func ReqPetDispatchPet(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SDispatchPet) {
	res := new(proto_pet.S2CDispatchPet)

	//助战
	if req.Scene == 4 {
		if req.IsLineUp {
			pl.Pet.DispatchPets[req.Scene] = req.Id
		} else {
			delete(pl.Pet.DispatchPets, req.Scene)
		}
	} else {
		if req.IsLineUp {
			has := false
			for scene, id := range pl.Pet.DispatchPets {
				if scene == 4 {
					continue
				}

				if id == req.Id {
					has = true
					break
				}
			}

			if has {
				res.DispatchPets = pl.Pet.DispatchPets
				ctx.Send(res)
				return
			}
			pl.Pet.DispatchPets[req.Scene] = req.Id
		} else {
			delete(pl.Pet.DispatchPets, req.Scene)
		}
	}
	res.DispatchPets = pl.Pet.DispatchPets
	ctx.Send(res)
}

// ReqCatchPet 请求宠物捕捉
func ReqPetCatch(ctx global.IPlayer, pl *model.Player, req *proto_pet.C2SCatchPetBattle) {
	res := new(proto_pet.S2CCatchPetBattle)

	//检验是否战斗中

	//是否配置中
	curMonsterConf, _ := config.Monster.Find(int64(req.Id))
	if curMonsterConf.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	if !curMonsterConf.CanCatch {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if curMonsterConf.CatchPetId <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Pet.Pets[curMonsterConf.CatchPetId]; ok {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	awards := make([]conf2.ItemE, 0)
	awards = append(awards, conf2.ItemE{
		ItemId:   curMonsterConf.CatchPetId,
		ItemNum:  1,
		ItemType: define.ItemTypePet,
	})
	bag.AddAward(ctx, pl, awards, true)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}
