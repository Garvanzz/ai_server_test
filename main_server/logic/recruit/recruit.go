package recruit

import (
	"sort"
	"time"
	"xfx/core/common"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/logic/activity/impl"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/pkg/utils"
)

var Module = func() module.Module {
	recruit := new(Recruit)
	recruit.cardPoolConf = make(map[int32]map[int32]map[int32]int32)
	recruit.shenjiPoolConf = make(map[int32]map[int32]map[int32]int32)
	recruit.gemAppraisalPoolConf = make(map[int32]map[int32]map[int32]int32)
	recruit.petPoolConf = make(map[int32]map[int32]map[int32]int32)
	recruit.heroCardPoolLevel = -1
	return recruit
}

type Recruit struct {
	modules.BaseModule
	cardPoolConf         map[int32]map[int32]map[int32]int32 //卡池配置  类型 品质  id 权重
	shenjiPoolConf       map[int32]map[int32]map[int32]int32 //卡池配置  类型 品质  id 权重
	gemAppraisalPoolConf map[int32]map[int32]map[int32]int32 //卡池配置  类型 品质  id 权重
	petPoolConf          map[int32]map[int32]map[int32]int32 //卡池配置  类型 品质  id 权重
	heroCardPoolLevel    int32
}

func (l *Recruit) OnInit(app module.App) {
	l.BaseModule.OnInit(app)
	l.Register("RecruitHero", l.OnRecruitHero)
	l.Register("RecruitShenji", l.OnRecruitShenji)
	l.Register("RecruitGemAppraisal", l.OnRecruitGemAppraisal)
	l.Register("RecruitPet", l.OnRecruitPet)
}

func (l *Recruit) GetType() string { return "Recruit" }

func (l *Recruit) OnTick(delta time.Duration) {}

func (l *Recruit) OnMessage(msg any) any {
	log.Debug("* Recruit message %v", msg)
	return nil
}

// 抽角色
func (mgr *Recruit) OnRecruitHero(typ int32, count int32, miniNum int32, level int32, heroIds []int32, typeId int32) (*model.RecruitResp, error) {
	resIds := make([]int32, 0)
	//获取抽到的ID
	if mgr.heroCardPoolLevel != level || mgr.cardPoolConf == nil {
		mgr.cardPoolConf = make(map[int32]map[int32]map[int32]int32)
		mgr.heroCardPoolLevel = level
	}

	if _, ok := mgr.cardPoolConf[typ]; !ok {
		mgr.cardPoolConf[typ] = make(map[int32]map[int32]int32)
		if typ == define.CARDPOOL_HERO {
			conf := config.HeroPool.All()
			for _, v := range conf {
				if v.PoolType != typeId {
					continue
				}

				if v.Level != level {
					continue
				}
				if _, ok = mgr.cardPoolConf[typ][v.Rate]; !ok {
					mgr.cardPoolConf[typ][v.Rate] = make(map[int32]int32)
				}

				//神话+
				if v.Rate >= 6 {
					//1：角色 2:碎片
					if v.Type == 1 {
						if common.IsHaveValueIntArray(heroIds, v.Value) == false {
							continue
						}
					} else if v.Type == 2 {
						conf_item := config.Item.All()[int64(v.Value)]
						if common.IsHaveValueIntArray(heroIds, conf_item.CompositeItem) == false {
							continue
						}
					}
				}
				mgr.cardPoolConf[typ][v.Rate][v.Id] = v.Weight
			}
		}
	}

	pool := mgr.cardPoolConf[typ]
	drawConfS := config.DrawPool.All()
	var drawConf conf2.DrawPool
	for _, v := range drawConfS {
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			continue
		}

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			continue
		}

		if v.Type == typ && time.Now().Unix() >= startTime.Unix() && time.Now().Unix() < endTime.Unix() {
			drawConf = v
			break
		}
	}

	weightFunc := func(mini bool) int32 {
		var poolMap map[int32]int32
		//保底
		if mini {
			poolMap = pool[drawConf.MiniValue]
		} else {
			//随机权重
			if typ == define.CARDPOOL_HERO {
				//判断等级
				confs := config.RecruitLvAward.All()
				var weights []int32
				for _, v := range confs {
					if v.Level == level {
						weights = v.Weight
						break
					}
				}
				rateIndex := utils.WeightIndex(weights)
				if _, ok := pool[int32(rateIndex+1)]; ok {
					poolMap = pool[int32(rateIndex+1)]
				}
			} else {
				rateIndex := utils.WeightIndex(drawConf.Weight)
				if _, ok := pool[int32(rateIndex+1)]; ok {
					poolMap = pool[int32(rateIndex+1)]
				}
			}
		}

		//排序
		ids := make([]int32, 0)
		for id, _ := range poolMap {
			ids = append(ids, id)
		}

		sort.Slice(ids, func(x, y int) bool {
			return ids[x] < ids[y]
		})

		weightArr := make([]int32, 0)
		for k := 0; k < len(ids); k++ {
			weightArr = append(weightArr, poolMap[ids[k]])
		}

		//随机权重
		if len(weightArr) <= 0 {
			return 0
		}

		weightIndex := utils.WeightIndex(weightArr)
		id := ids[weightIndex]
		return id
	}

	for i := int32(0); i < count; i++ {
		//判断保底
		id := int32(0)
		if miniNum+1 >= drawConf.MiniNum {
			//暂时不要这个保底
			id = weightFunc(false)
			miniNum = 0
		} else {
			id = weightFunc(false)
		}
		log.Debug("抽取的卡池:%v, %v", typ, id)

		if id <= 0 {
			continue
		}
		resIds = append(resIds, id)
		miniNum++
	}

	return &model.RecruitResp{
		BdNum: miniNum,
		Ids:   resIds,
	}, nil
}

func (mgr *Recruit) OnRecruitShenji(typ int32, count int32, miniNum int32) (*model.RecruitResp, error) {
	resIds := make([]int32, 0)

	//获取抽到的ID
	if mgr.shenjiPoolConf == nil {
		mgr.shenjiPoolConf = make(map[int32]map[int32]map[int32]int32)
	}

	if _, ok := mgr.shenjiPoolConf[typ]; !ok {
		mgr.shenjiPoolConf[typ] = make(map[int32]map[int32]int32)
		if typ == define.CARDPOOL_SHENJI {
			conf := config.ShenjiPool.All()
			for _, v := range conf {
				if _, ok = mgr.shenjiPoolConf[typ][v.Rate]; !ok {
					mgr.shenjiPoolConf[typ][v.Rate] = make(map[int32]int32)
				}

				mgr.shenjiPoolConf[typ][v.Rate][v.Id] = v.Weight
			}
		}
	}

	pool := mgr.shenjiPoolConf[typ]
	drawConfS := config.DrawPool.All()
	var drawConf conf2.DrawPool
	for _, v := range drawConfS {
		if v.Type == typ {
			drawConf = v
			break
		}
	}

	weightFunc := func(mini bool) int32 {
		var poolMap map[int32]int32
		//保底
		if mini {
			poolMap = pool[drawConf.MiniValue]
		} else {
			rateIndex := utils.WeightIndex(drawConf.Weight)
			if _, ok := pool[int32(rateIndex+1)]; ok {
				poolMap = pool[int32(rateIndex+1)]
			}
		}

		//排序
		ids := make([]int32, 0)
		for id, _ := range poolMap {
			ids = append(ids, id)
		}

		sort.Slice(ids, func(x, y int) bool {
			return ids[x] < ids[y]
		})

		weightArr := make([]int32, 0)
		for k := 0; k < len(ids); k++ {
			weightArr = append(weightArr, poolMap[ids[k]])
		}

		//随机权重
		if len(weightArr) <= 0 {
			return 0
		}

		weightIndex := utils.WeightIndex(weightArr)
		id := ids[weightIndex]
		return id
	}

	for i := int32(0); i < count; i++ {
		//判断保底
		id := int32(0)
		if miniNum+1 >= drawConf.MiniNum {
			id = weightFunc(true)
			miniNum = 0
		} else {
			id = weightFunc(false)
		}
		log.Debug("抽取的卡池:%v, %v", typ, id)

		if id <= 0 {
			continue
		}
		resIds = append(resIds, id)
		miniNum++
	}

	return &model.RecruitResp{
		BdNum: miniNum,
		Ids:   resIds,
	}, nil
}

func (mgr *Recruit) OnRecruitGemAppraisal(typ int32, title int32, count int32, miniNum int32) (*model.RecruitResp, error) {
	resIds := make([]int32, 0)

	//获取抽到的ID
	if mgr.gemAppraisalPoolConf == nil {
		mgr.gemAppraisalPoolConf = make(map[int32]map[int32]map[int32]int32)
	}

	if _, ok := mgr.gemAppraisalPoolConf[title]; !ok {
		mgr.gemAppraisalPoolConf[title] = make(map[int32]map[int32]int32)
		if typ == define.CARDPOOL_GEMAPPRAISAL {
			conf := config.TreasurePool.All()
			for _, v := range conf {
				if v.Title != title {
					continue
				}
				if _, ok = mgr.gemAppraisalPoolConf[title][v.Rate]; !ok {
					mgr.gemAppraisalPoolConf[title][v.Rate] = make(map[int32]int32)
				}

				mgr.gemAppraisalPoolConf[title][v.Rate][v.Id] = v.Weight
			}
		}
	}

	pool := mgr.gemAppraisalPoolConf[title]
	drawConfS := config.DrawPool.All()
	var drawConf conf2.DrawPool
	for _, v := range drawConfS {
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			continue
		}

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			continue
		}

		if v.Type == typ && time.Now().Unix() >= startTime.Unix() && time.Now().Unix() < endTime.Unix() {
			drawConf = v
			break
		}
	}

	weightFunc := func(mini bool) int32 {
		var poolMap map[int32]int32
		//保底
		if mini {
			poolMap = pool[drawConf.MiniValue]
		} else {
			rateIndex := utils.WeightIndex(drawConf.Weight)
			if _, ok := pool[int32(rateIndex+1)]; ok {
				poolMap = pool[int32(rateIndex+1)]
			}
		}

		//排序
		ids := make([]int32, 0)
		for id, _ := range poolMap {
			ids = append(ids, id)
		}

		sort.Slice(ids, func(x, y int) bool {
			return ids[x] < ids[y]
		})

		weightArr := make([]int32, 0)
		for k := 0; k < len(ids); k++ {
			weightArr = append(weightArr, poolMap[ids[k]])
		}

		//随机权重
		if len(weightArr) <= 0 {
			return 0
		}

		weightIndex := utils.WeightIndex(weightArr)
		id := ids[weightIndex]
		return id
	}

	for i := int32(0); i < count; i++ {
		//判断保底
		id := int32(0)
		if miniNum+1 >= drawConf.MiniNum {
			id = weightFunc(false)
			miniNum = 0
		} else {
			id = weightFunc(false)
		}
		log.Debug("抽取的鉴宝卡池:%v, %v", typ, id)

		if id <= 0 {
			continue
		}
		resIds = append(resIds, id)
		miniNum++
	}

	return &model.RecruitResp{
		BdNum: miniNum,
		Ids:   resIds,
	}, nil
}

func (mgr *Recruit) OnRecruitPet(typ int32, subType int32, count int32, miniNum int32) (*model.RecruitResp, error) {
	resIds := make([]int32, 0)

	//获取抽到的ID
	if mgr.petPoolConf == nil {
		mgr.petPoolConf = make(map[int32]map[int32]map[int32]int32)
	}

	if _, ok := mgr.petPoolConf[subType]; !ok {
		mgr.petPoolConf[subType] = make(map[int32]map[int32]int32)
		if typ == define.CARDPOOL_PET {
			conf := config.PetDrawPool.All()
			for _, v := range conf {
				if v.PoolType != subType {
					continue
				}
				if _, ok = mgr.petPoolConf[subType][v.Rate]; !ok {
					mgr.petPoolConf[subType][v.Rate] = make(map[int32]int32)
				}

				mgr.petPoolConf[subType][v.Rate][v.Id] = v.Weight
			}
		}
	}

	pool := mgr.petPoolConf[subType]
	drawConfS := config.DrawPool.All()
	var drawConf conf2.DrawPool
	for _, v := range drawConfS {
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			continue
		}

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			continue
		}

		if v.Type == typ && time.Now().Unix() >= startTime.Unix() && time.Now().Unix() < endTime.Unix() && v.Param == subType {
			drawConf = v
			break
		}
	}

	weightFunc := func(mini bool) int32 {
		var poolMap map[int32]int32
		//保底
		if mini {
			poolMap = pool[drawConf.MiniValue]
		} else {
			rateIndex := utils.WeightIndex(drawConf.Weight)
			if _, ok := pool[int32(rateIndex+1)]; ok {
				poolMap = pool[int32(rateIndex+1)]
			}
		}

		//排序
		ids := make([]int32, 0)
		for id, _ := range poolMap {
			ids = append(ids, id)
		}

		sort.Slice(ids, func(x, y int) bool {
			return ids[x] < ids[y]
		})

		weightArr := make([]int32, 0)
		for k := 0; k < len(ids); k++ {
			weightArr = append(weightArr, poolMap[ids[k]])
		}

		//随机权重
		if len(weightArr) <= 0 {
			return 0
		}

		weightIndex := utils.WeightIndex(weightArr)
		id := ids[weightIndex]
		return id
	}

	for i := int32(0); i < count; i++ {
		//判断保底
		id := int32(0)
		if miniNum+1 >= drawConf.MiniNum {
			id = weightFunc(true)
			miniNum = 0
		} else {
			id = weightFunc(false)
		}
		log.Debug("抽取的宠物卡池:%v, %v", typ, id)

		if id <= 0 {
			continue
		}
		resIds = append(resIds, id)
		miniNum++
	}

	return &model.RecruitResp{
		BdNum: miniNum,
		Ids:   resIds,
	}, nil
}
