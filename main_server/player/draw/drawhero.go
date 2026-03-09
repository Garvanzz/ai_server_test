package drawhero

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/logic/activity/impl"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_draw"
)

func Init(pl *model.Player) {
	pl.Draw = new(model.DrawHero)
	pl.Draw.ToDayIsFree = true
	pl.Draw.LastTime = utils.Now().Unix()
	pl.Draw.Pools = make(map[int32]*model.DrawPool)
	pl.Draw.BigCard = make([]int32, 0)

}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Draw)
	if err != nil {
		log.Error("player[%v],save draw marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerDraw, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerDraw, pl.Id))
	if err != nil {
		log.Error("player[%v],load hero error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.DrawHero)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load draw unmarshal error:%v", pl.Id, err)
	}

	//判断每日免费一次这个
	if !utils.CheckIsSameDayBySec(utils.Now().Unix(), m.LastTime, 0) {
		m.ToDayIsFree = true
		m.LastTime = utils.Now().Unix()
	}
	pl.Draw = m
}

// ReqInitDraw 请求抽卡数据
func ReqInitDraw(ctx global.IPlayer, pl *model.Player, req *proto_draw.C2SDrawInit) {
	res := &proto_draw.S2CDrawInit{}
	drawPoolRefreshState(pl)
	res.Drawinfo = model.ToDrawCardProto(pl.Draw)
	ctx.Send(res)
}

// 刷新卡池的状态
func drawPoolRefreshState(pl *model.Player) {
	confs := config.DrawPool.All()
	for k, v := range pl.Draw.Pools {
		conf := confs[int64(v.PoolId)]

		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(conf.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			continue
		}

		if utils.Now().Unix() >= endTime.Unix() && conf.ActivityType == 2 && conf.Type == 1 {
			delete(pl.Draw.Pools, k)
		}
	}

	//判断有没有新加的
	for _, v := range confs {
		if v.Type != 1 {
			continue
		}

		if _, ok := pl.Draw.Pools[v.Id]; ok {
			continue
		}

		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.StartTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse startTime err:%v", err)
			continue
		}

		if v.ActivityType == 2 {

			endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(v.EndTime), time.Local)
			if err != nil {
				log.Error("checkCfg parse endTime err:%v", err)
				continue
			}

			log.Error("卡池:%v, %v, %v, %v, %v", utils.Now().Unix(), startTime.Unix(), endTime.Unix(), utils.Now().Unix() < startTime.Unix(), utils.Now().Unix() >= endTime.Unix())
			if utils.Now().Unix() < startTime.Unix() || utils.Now().Unix() >= endTime.Unix() {
				continue
			}
		}

		//添加新的活动
		pl.Draw.Pools[v.Id] = &model.DrawPool{
			PoolId:      v.Id,
			StarTime:    startTime.Unix(),
			StageAwards: make([]int32, 0),
		}

		if utils.ContainsInt32(pl.Draw.BigCard, v.HeroId) == false {
			pl.Draw.BigCard = append(pl.Draw.BigCard, v.HeroId)
		}
	}
}

// ReqcardDraw 请求角色
func ReqDrawCard(ctx global.IPlayer, pl *model.Player, req *proto_draw.C2SDrawCard) {
	res := &proto_draw.S2CDrawCard{}

	if req.Type != define.CARDPOOL_HERO {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NOConfig
		ctx.Send(res)
		return
	}

	//获取卡池
	conf := config.DrawPool.All()
	if _, ok := conf[int64(req.PoolId)]; !ok {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NOConfig
		ctx.Send(res)
		return
	}

	if _, ok := pl.Draw.Pools[req.PoolId]; !ok {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_ConditionNo
		ctx.Send(res)
		return
	}

	//消耗道具
	costItems := make(map[int32]int32)
	costItems[define.ItemIdDrawCard] = req.Count

	//判断抽奖次数
	if pl.Draw.ToDayIsFree {
		pl.Draw.ToDayIsFree = false
		costItems[define.ItemIdDrawCard] -= 0
	} else {
		if !utils.CheckIsSameDayBySec(utils.Now().Unix(), pl.Draw.LastTime, 0) {
			pl.Draw.ToDayIsFree = false
			pl.Draw.LastTime = utils.Now().Unix()
			costItems[define.ItemIdDrawCard] -= 0
		}
	}

	if !internal.CheckItemsEnough(pl, costItems) {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NumNotEnough
		ctx.Send(res)
		return
	}

	//筛选神话+的数据
	heroId := make([]int32, 0)
	heroId = append(heroId, conf[int64(req.PoolId)].HeroId)

	//判断之前4星以上
	for _, v := range pl.Draw.BigCard {
		if v == conf[int64(req.PoolId)].HeroId {
			continue
		}

		if _, ok := pl.Hero.Hero[v]; !ok {
			continue
		}

		if pl.Hero.Hero[v].Star >= 4 {
			heroId = append(heroId, v)
		}
	}

	cardPoolConf := conf[int64(req.PoolId)]
	resp, err := ctx.Invoke("Recruit", "RecruitHero", req.Type, req.Count, pl.Draw.BdCount, pl.Draw.Level, heroId, cardPoolConf.Param)
	if err != nil {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NOConfig
		ctx.Send(res)
	}

	//扣道具
	internal.SubItems(ctx, pl, costItems)
	rect := resp.(*model.RecruitResp)
	//保底次数
	pl.Draw.BdCount = rect.BdNum

	//增加次数
	pl.Draw.Pools[req.PoolId].DrawNum += req.Count

	//增加经验
	pl.Draw.Exp += req.Count

	//判断等级
	confs := config.RecruitLvAward.All()
	for _, v := range confs {
		if v.Level == pl.Draw.Level && v.Exp <= pl.Draw.Exp {
			pl.Draw.Level += 1
			pl.Draw.Exp = pl.Draw.Exp - v.Exp
			break
		}
	}

	//活动链路
	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":         "draw_hero",
		"player":      pl.ToContext(),
		"value":       req.Count,
		"playermodel": pl,
	})

	//任务
	task.Dispatch(ctx, pl, define.TaskDrawCard, req.Count, 0, true)

	var items []conf2.ItemE
	//回调
	resmap := make([]*proto_draw.DrawCardResult, 0)
	for _, v := range rect.Ids {
		confHero := config.HeroPool.All()[int64(v)]
		if confHero.PoolType != req.Type {
			continue
		}

		re := new(proto_draw.DrawCardResult)
		if confHero.Type == 1 {
			if _, ok := pl.Hero.Hero[confHero.Value]; ok {
				conf := config.Hero.All()[int64(confHero.Value)]
				confItem := config.Item.All()[int64(conf.Fragment)]
				re.Id = conf.Fragment
				re.Num = confItem.CompositeNeed
				re.IsFragment = false
				re.Type = define.ItemTypeItem

				items = append(items, conf2.ItemE{
					ItemType: define.ItemTypeItem,
					ItemId:   conf.Fragment,
					ItemNum:  confItem.CompositeNeed,
				})
			} else {
				re.Id = confHero.Value
				re.Num = 1
				re.Type = define.ItemTypeHero

				items = append(items, conf2.ItemE{
					ItemType: define.ItemTypeHero,
					ItemId:   confHero.Value,
					ItemNum:  1,
				})
			}
		} else if confHero.Type == 2 {
			re.Id = confHero.Value
			re.Num = 1
			re.IsFragment = true
			re.Type = define.ItemTypeItem

			items = append(items, conf2.ItemE{
				ItemType: define.ItemTypeItem,
				ItemId:   confHero.Value,
				ItemNum:  1,
			})
		}
		//通告相关
		internal.SyncNotice_DrawCardHero(ctx, pl, confHero.PoolType, confHero.Type, confHero.Value)
		resmap = append(resmap, re)
	}

	log.Debug("抽卡结果:%v", resmap)
	//添加道具
	bag.AddAward(ctx, pl, items, false)

	res.Items = resmap
	res.PoolId = req.PoolId
	res.Drawinfo = model.ToDrawCardProto(pl.Draw)
	res.Code = proto_draw.ERRORCODEDRAW_ERR_Ok
	ctx.Send(res)
}

// ReqHeroDrawLevelAward 请求招募等级奖励
func ReqHeroDrawLevelAward(ctx global.IPlayer, pl *model.Player, req *proto_draw.C2SGetDrawCardLevelAward) {
	res := &proto_draw.S2CGetDrawCardLevelAward{}

	awards := make([]conf2.ItemE, 0)
	ids := make([]int32, 0)
	for _, v := range req.Id {
		//判断领取没
		if utils.ContainsInt32(pl.Draw.LevelAwards, v) {
			continue
		}

		conf := config.RecruitLvAward.All()[int64(v)]
		if conf.Level > pl.Draw.Level {
			continue
		}
		awards = append(awards, conf.Award...)
		ids = append(ids, v)
	}

	//添加道具
	bag.AddAward(ctx, pl, awards, true)
	pl.Draw.LevelAwards = append(pl.Draw.LevelAwards, ids...)
	res.Awards = pl.Draw.LevelAwards
	res.Code = proto_draw.ERRORCODEDRAW_ERR_Ok
	ctx.Send(res)
}

// ReqHeroDrawStageAward 请求招募阶段奖励
func ReqHeroDrawStageAward(ctx global.IPlayer, pl *model.Player, req *proto_draw.C2SGetDrawCardStageAward) {
	res := &proto_draw.S2CGetDrawCardStageAward{}

	//先获取池子
	if _, ok := pl.Draw.Pools[req.Id]; !ok {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_ConditionNo
		ctx.Send(res)
		return
	}

	conf := config.DrawPool.All()[int64(req.Id)]
	if conf.Id <= 0 {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NOConfig
		ctx.Send(res)
		return
	}

	confStage := config.DrawStageAward.All()[int64(conf.StageId)]
	if confStage.Id <= 0 {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NOConfig
		ctx.Send(res)
		return
	}

	//判断领取没
	if utils.ContainsInt32(confStage.Progress, req.Progress) == false {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NOConfig
		ctx.Send(res)
		return
	}

	if utils.ContainsInt32(pl.Draw.Pools[req.Id].StageAwards, req.Progress) {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_AlGetAward
		ctx.Send(res)
		return
	}

	index := 0
	for k, v := range confStage.Progress {
		if v == req.Progress {
			index = k
			break
		}
	}

	awardstr := confStage.Award[index]
	awards := strings.Split(awardstr, ",")

	var items []conf2.ItemE
	items = append(items, conf2.ItemE{
		ItemType: int32(utils.MustParseInt64(awards[2])),
		ItemId:   int32(utils.MustParseInt64(awards[0])),
		ItemNum:  int32(utils.MustParseInt64(awards[1])),
	})
	//添加道具
	bag.AddAward(ctx, pl, items, true)
	pl.Draw.Pools[req.Id].StageAwards = append(pl.Draw.Pools[req.Id].StageAwards, req.Progress)

	opt := &proto_draw.DrawPoolOption{
		PoolId:        pl.Draw.Pools[req.Id].PoolId,
		PoolStartTime: pl.Draw.Pools[req.Id].StarTime,
		StageAwards:   pl.Draw.Pools[req.Id].StageAwards,
		DrawNum:       pl.Draw.Pools[req.Id].DrawNum,
	}

	res.PoolOption = opt
	res.Code = proto_draw.ERRORCODEDRAW_ERR_Ok
	ctx.Send(res)
}
