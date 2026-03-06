package hero

import (
	"encoding/json"
	"fmt"
	"math"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/proto/proto_hero"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.Hero = new(model.Hero)
	pl.Hero.Hero = make(map[int32]*model.HeroOption)
	pl.Hero.Skin = make(map[int32]*model.SkinOption)

	//初始新角色
	pl.Hero.Hero[define.PlayerMainHeroNull] = &model.HeroOption{
		Id:          define.PlayerMainHeroNull,
		Star:        0,
		Level:       1,
		Exp:         0,
		Stage:       0,
		Skin:        "Default",
		Cultivation: make(map[int32]int32),
	}

	pl.Hero.Hero[define.PlayerMainHeroYao] = &model.HeroOption{
		Id:          define.PlayerMainHeroYao,
		Star:        0,
		Level:       1,
		Exp:         0,
		Stage:       0,
		Skin:        "Default",
		Cultivation: make(map[int32]int32),
	}

	pl.Hero.Hero[define.PlayerMainHeroShen] = &model.HeroOption{
		Id:          define.PlayerMainHeroShen,
		Star:        0,
		Level:       1,
		Exp:         0,
		Stage:       0,
		Skin:        "Default",
		Cultivation: make(map[int32]int32),
	}

	pl.Hero.Hero[define.PlayerMainHeroFo] = &model.HeroOption{
		Id:          define.PlayerMainHeroFo,
		Star:        0,
		Level:       1,
		Exp:         0,
		Stage:       0,
		Skin:        "Default",
		Cultivation: make(map[int32]int32),
	}
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Hero)
	if err != nil {
		log.Error("player[%v],save hero marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("Save hero error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerHero, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load hero error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerHero, pl.Id))
	if err != nil {
		log.Error("player[%v],load hero error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Hero)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load hero unmarshal error:%v", pl.Id, err)
	}

	pl.Hero = m
}

// ReqInitHero 请求背包英雄数据
func ReqInitHero(ctx global.IPlayer, pl *model.Player, req *proto_hero.C2SInitHero) {
	//要过滤过期
	resp := new(proto_hero.S2CInitHero)
	resp.Heros = model.ToBagHeroProto(pl.Hero.Hero)
	ctx.Send(resp)
}

// ReqInitSkin 请求背包皮肤数据
func ReqInitSkin(ctx global.IPlayer, pl *model.Player, req *proto_hero.C2SInitSkin) {
	//要过滤过期
	resp := new(proto_hero.S2CInitSkin)
	resp.Skins = model.ToBagSkinProto(pl.Hero.Skin)
	ctx.Send(resp)
}

// 升级
func ReqHeroUpLevel(ctx global.IPlayer, pl *model.Player, req *proto_hero.C2SHeroUpLevel) {
	res := &proto_hero.S2CHeroUpLevel{}
	if req.Type == 1 {
		if pl.Bag.Items[define.ItemIdMoney] <= 0 {
			res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
			ctx.Send(res)
			return
		}
	} else if req.Type == 2 {
		heroId := int32(pl.GetProp(define.PlayerPropHeroId))
		if heroId != req.Id {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}
	}

	if _, ok := pl.Hero.Hero[req.Id]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		ctx.Send(res)
		return
	}

	hero := pl.Hero.Hero[req.Id]
	if hero.Level >= define.LevelMaxLimit {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	levelRange := config.Global.Get().HeroUpLevelRange
	costs := make(map[int32]int32)
	//主角，600级前不需要考虑其他条件
	heroId := int32(pl.GetProp(define.PlayerPropHeroId))
	if heroId == req.Id {
		levelConf := config.HeroUpLevel.All()[int64(heroId)]
		levelLimit := config.Global.Get().MainHeroLevelLimit
		mainStage := int32(hero.Level / 100)
		offsetLevel := hero.Level - mainStage*levelRange

		//道具数量
		itemCount := int32(0)
		if _, ok := pl.Bag.Items[define.ItemIdMoney]; ok {
			itemCount = pl.Bag.Items[define.ItemIdMoney]
		}

		MaxLevel := calculateMaxLevel(itemCount, levelConf.UpLevelCondition[mainStage], offsetLevel)
		log.Debug("升级主角的等级:%v,%v,%v", MaxLevel, hero.Level, offsetLevel)
		if MaxLevel-hero.Level < req.Count {
			res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
			ctx.Send(res)
			return
		}

		if hero.Level < levelLimit {
			if req.Type == 1 {
				costs[define.ItemIdMoney] = (levelConf.UpLevelCondition[mainStage] / 2) * ((offsetLevel+req.Count)*(offsetLevel+req.Count+1) - (offsetLevel)*(offsetLevel+1))
			} else if req.Type == 2 { //修为

			}

		} else {
			//判断上阵数量
			num := 0
			for _, v := range pl.Lineup.LineUps[define.LINEUP_STAGE].HeroId {
				if v > 0 {
					num += 1
				}
			}

			//上阵人数不足
			if num < 6 {
				res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
				ctx.Send(res)
				return
			}

			//主角
			mhero := pl.Hero.Hero[heroId]

			//等级差
			for _, v := range pl.Lineup.LineUps[define.LINEUP_STAGE].HeroId {
				if v <= 0 || v == heroId {
					continue
				}

				lhero := pl.Hero.Hero[v]
				if math.Abs(float64(lhero.Level-mhero.Level)) >= 200 {
					res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
					ctx.Send(res)
					return
				}
			}

			if req.Type == 1 {
				log.Debug("主角升级的消耗:%v", (levelConf.UpLevelCondition[mainStage]/2)*((offsetLevel+req.Count)*(offsetLevel+req.Count+1)-(offsetLevel)*(offsetLevel+1)))
				costs[define.ItemIdMoney] = (levelConf.UpLevelCondition[mainStage] / 2) * ((offsetLevel+req.Count)*(offsetLevel+req.Count+1) - (offsetLevel)*(offsetLevel+1))
			} else if req.Type == 2 { //修为

			}
		}

		//任务
		task.Dispatch(ctx, pl, define.TaskHeroLevel, hero.Level+req.Count, 0, false)
		task.Dispatch(ctx, pl, define.TaskLevelUpMainHeroTime, req.Count, 0, true)
	} else {
		//不能超过主角等级
		mhero := pl.Hero.Hero[heroId]
		if hero.Level >= mhero.Level {
			res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
			ctx.Send(res)
			return
		}

		//升阶
		if hero.Level%levelRange == 0 && hero.Stage < hero.Level/levelRange {
			res.Code = proto_public.CommonErrorCode_ERR_NoUpLevel
			ctx.Send(res)
			return
		} else {
			levelConf := config.HeroUpLevel.All()[int64(heroId)]
			offsetLevel := hero.Level - hero.Stage*levelRange

			//道具数量
			itemCount := int32(0)
			if _, ok := pl.Bag.Items[define.ItemIdMoney]; ok {
				itemCount = pl.Bag.Items[define.ItemIdMoney]
			}

			MaxLevel := calculateMaxLevel(itemCount, levelConf.UpLevelCondition[hero.Stage], offsetLevel)
			log.Debug("升级的等级:%v", MaxLevel)
			if MaxLevel-hero.Level < req.Count {
				res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
				ctx.Send(res)
				return
			}
			log.Debug("神将升级的消耗:%v", (levelConf.UpLevelCondition[hero.Stage]/2)*((offsetLevel+req.Count)*(offsetLevel+req.Count+1)-(offsetLevel)*(offsetLevel+1)))
			costs[define.ItemIdMoney] = (levelConf.UpLevelCondition[hero.Stage] / 2) * ((offsetLevel+req.Count)*(offsetLevel+req.Count+1) - (offsetLevel)*(offsetLevel+1))
		}

		//任务
		task.Dispatch(ctx, pl, define.TaskLevelUpHeroTime, req.Count, 0, true)
	}

	//判断道具是否足够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	hero.Level += req.Count
	pl.Hero.Hero[req.Id] = hero

	//主角，处理等级变化
	if req.Id == heroId {
		pl.SetProp(define.PlayerPropLevel, int64(hero.Level), false)
	}

	//扣除道具
	internal.SubItems(ctx, pl, costs)

	internal.SyncHeroChange(ctx, pl, req.Id)

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

// 计算升到等级需要花费的总金币数量
func calculateSumCostNumByLevel(baseCoins, level int32) int32 {
	totalCoins := (baseCoins / 2) * level * (level + 1)
	return totalCoins
}

// 升阶
func ReqHeroUpStage(ctx global.IPlayer, pl *model.Player, req *proto_hero.C2SHeroUpStage) {
	res := &proto_hero.S2CHeroUpStage{}

	if _, ok := pl.Hero.Hero[req.Id]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		ctx.Send(res)
		return
	}

	levelRange := config.Global.Get().HeroUpLevelRange
	hero := pl.Hero.Hero[req.Id]
	//主角突破,要判断条件
	heroId := int32(pl.GetProp(define.PlayerPropHeroId))
	if heroId == req.Id {
		mainHero := pl.Hero.Hero[heroId]
		//任务+修为
		taskConf := config.Task.All()
		tasks := make([]conf2.Task, 0)
		for _, v := range taskConf {
			if v.Type == 5 && v.Param[0] == mainHero.Stage {
				tasks = append(tasks, v)
			}
		}

		var finish bool = true
		//检查任务是否完成
		for _, v := range tasks {
			if _, ok := pl.Task.MainTask[v.Id]; !ok {
				finish = false
				break
			} else {
				if pl.Task.MainTask[v.Id].ReceiveAward == false {
					finish = false
					break
				}
			}
		}

		if finish == false {
			res.Code = proto_public.CommonErrorCode_ERR_NoAdvance
			ctx.Send(res)
			return
		}

		//判断修为
		HeroCultivationConf := config.HeroCultivation.All()
		var conf conf2.HeroCultivation
		for _, v := range HeroCultivationConf {
			if v.Job == int32(pl.GetProp(define.PlayerPropJob)) && v.Stage == mainHero.Stage {
				conf = v
				break
			}
		}

		if conf.Id <= 0 {
			res.Code = proto_public.CommonErrorCode_ERR_NoAdvance
			ctx.Send(res)
			return
		}

		all := int32(0)
		for _, v := range mainHero.Cultivation {
			all += v
		}

		if all <= conf.Cultivation*4 {
			res.Code = proto_public.CommonErrorCode_ERR_NoAdvance
			ctx.Send(res)
		}

		//设置技能等级
		for id, level := range conf.SkillLevel {
			internal.UpdateSkill(ctx, pl, id, level)
		}

		mainHero.Stage += 1
		mainHero.Cultivation = make(map[int32]int32)
		pl.Hero.Hero[heroId] = mainHero

		//任务
		task.Dispatch(ctx, pl, define.TaskLevelUpMainHeroStage, mainHero.Stage, 0, false)

		internal.SyncHeroChange(ctx, pl, req.Id)
		res.Code = proto_public.CommonErrorCode_ERR_OK
		ctx.Send(res)

	} else {
		if pl.Bag.Items[define.ItemIdTupoStore] <= 0 {
			res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
			ctx.Send(res)
			return
		}

		if hero.Stage >= define.LevelMaxLimit/10 {
			res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
			ctx.Send(res)
			return
		}

		offse := hero.Level % levelRange
		//突破
		if offse != 0 {
			res.Code = proto_public.CommonErrorCode_ERR_NoAdvance
			ctx.Send(res)
			return
		}

		if hero.Stage*levelRange >= hero.Level {
			res.Code = proto_public.CommonErrorCode_ERR_NoAdvance
			ctx.Send(res)
			return
		}

		stageConf := config.HeroUpStage.All()[int64(hero.Id)]
		costs := make(map[int32]int32)
		costs[define.ItemIdTupoStore] = stageConf.UpStageCondition[hero.Stage]

		if !internal.CheckItemsEnough(pl, costs) {
			res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
			ctx.Send(res)
			return
		}
		hero.Stage += 1
		pl.Hero.Hero[req.Id] = hero

		if len(costs) > 0 {
			//扣除道具
			internal.SubItems(ctx, pl, costs)
		}
		internal.SyncHeroChange(ctx, pl, req.Id)

		//任务
		task.Dispatch(ctx, pl, define.TaskLevelUpHeroStageTime, 1, 0, true)

		res.Code = proto_public.CommonErrorCode_ERR_OK
		ctx.Send(res)
	}
}

// 升星
func ReqHeroUpStar(ctx global.IPlayer, pl *model.Player, req *proto_hero.C2SHeroUpStar) {
	res := &proto_hero.S2CHeroUpStar{}

	if _, ok := pl.Hero.Hero[req.Id]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		ctx.Send(res)
		return
	}

	hero := pl.Hero.Hero[req.Id]
	heroUpStarConf := config.HeroUpStar.All()[(int64(hero.Id))]
	costs := make(map[int32]int32)
	log.Debug("vvvv:%v", heroUpStarConf.NeedCostNum[hero.Star])
	costs[heroUpStarConf.NeedId] = heroUpStarConf.NeedCostNum[hero.Star]

	//判断够不够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	if hero.Star >= 6 {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	hero.Star += 1

	pl.Hero.Hero[req.Id] = hero

	heroId := int32(pl.GetProp(define.PlayerPropHeroId))
	//任务
	if req.Id == heroId {
		task.Dispatch(ctx, pl, define.TaskLevelUpHeroStarTime, 1, 0, true)
	}

	//扣除道具
	internal.SubItems(ctx, pl, costs)

	internal.SyncHeroChange(ctx, pl, req.Id)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 提升修为
func ReqHeroUpCultivation(ctx global.IPlayer, pl *model.Player, req *proto_hero.C2SHeroUpCultivation) {
	res := &proto_hero.S2CHeroUpCultivation{}

	if _, ok := pl.Hero.Hero[int32(pl.GetProp(define.PlayerPropHeroId))]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		ctx.Send(res)
		return
	}

	HeroId := int32(pl.GetProp(define.PlayerPropHeroId))
	hero := pl.Hero.Hero[HeroId]
	heroConf := config.Hero.All()[int64(HeroId)]
	cultivationConfs := config.HeroCultivation.All()
	con := conf2.HeroCultivation{}
	for _, v := range cultivationConfs {
		if v.Job == heroConf.Job && v.Stage == hero.Stage {
			con = v
			break
		}
	}

	num := hero.Cultivation[req.Type]
	if num >= con.Cultivation {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	costs := make(map[int32]int32)
	costNum := con.CostCult + con.CostCult*(int32)(math.Pow(float64(num/con.Cultivation), float64(con.CostRatio)))
	costs[define.ItemIdCultivation] = costNum

	//判断够不够
	if !internal.CheckItemsEnough(pl, costs) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	hero.Cultivation[req.Type] = num + 1
	pl.Hero.Hero[HeroId] = hero

	//任务
	sum := int32(0)
	for _, v := range hero.Cultivation {
		sum += v
	}
	task.Dispatch(ctx, pl, define.TaskLevelUpMainHeroXiuweiLevel, sum, 0, false)
	task.Dispatch(ctx, pl, define.TaskLevelUpMainHeroXiuweiLevel, 1, 0, true)

	internal.SyncHeroChange(ctx, pl, HeroId)

	//扣除道具
	internal.SubItems(ctx, pl, costs)

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 重生
func ReqReSetHero(ctx global.IPlayer, pl *model.Player, req *proto_hero.C2SReSetHero) {
	res := &proto_hero.S2CReSetHero{}
	res.Heros = make(map[int32]*proto_hero.HeroOption)

	for _, v := range req.Id {
		if _, ok := pl.Hero.Hero[v]; !ok {
			ctx.Send(res)
			return
		}
	}

	costs := make(map[int32]int32, 0)
	costs[define.ItemIdMoney] = 0
	costs[define.ItemIdTupoStore] = 0
	for _, v := range req.Id {
		//等级
		levelConf := config.HeroUpLevel.All()[int64(v)]
		heroData := pl.Hero.Hero[v]
		sumCoins := calculateSumCostNumByLevel(levelConf.UpLevelCondition[heroData.Stage], heroData.Level)
		costs[define.ItemIdMoney] += sumCoins
		pl.Hero.Hero[v].Level = 1

		//阶数
		stageConf := config.HeroUpStage.All()[int64(v)]
		sumStors := int32(0)
		for k := int32(0); k < heroData.Stage; k++ {
			sumStors += stageConf.UpStageCondition[k]
		}
		costs[define.ItemIdTupoStore] += sumStors
		pl.Hero.Hero[v].Stage = 0

		res.Heros[v] = &proto_hero.HeroOption{
			Id:          pl.Hero.Hero[v].Id,
			Star:        pl.Hero.Hero[v].Star,
			Level:       pl.Hero.Hero[v].Level,
			Exp:         pl.Hero.Hero[v].Exp,
			Stage:       pl.Hero.Hero[v].Stage,
			Cultivation: pl.Hero.Hero[v].Cultivation,
		}
	}

	if costs[define.ItemIdMoney] <= 0 && costs[define.ItemIdTupoStore] <= 0 {
		ctx.Send(res)
		return
	}

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
	ctx.Send(res)
	return
}
