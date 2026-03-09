package gemappraisal

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/logic/activity/impl"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_draw"
	"xfx/proto/proto_equip"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.GemAppraisal = new(model.GemAppraisal)
	pl.GemAppraisal.Reward = make([]int32, 0)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.GemAppraisal)
	if err != nil {
		log.Error("player[%v],save GemAppraisal marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerGemAppraisal, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerGemAppraisal, pl.Id))
	if err != nil {
		log.Error("player[%v],load GemAppraisal error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.GemAppraisal)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load GemAppraisal unmarshal error:%v", pl.Id, err)
	}

	pl.GemAppraisal = m
}

// 请求鉴宝
func ReqInitGemAppraisal(ctx global.IPlayer, pl *model.Player, req *proto_draw.C2SGemAppraisalInit) {
	res := &proto_draw.S2CGemAppraisalInit{}
	//判断一下卡池
	updatePoolState(pl)
	res.Pool = &proto_draw.GemPoolOption{
		PoolId:        int32(pl.GemAppraisal.PoolId),
		PoolStartTime: pl.GemAppraisal.PoolStarTime,
		Num:           pl.GemAppraisal.Num,
		Awards:        pl.GemAppraisal.Reward,
	}
	ctx.Send(res)
}

// 鉴宝
func ReqDrawGemAppraisal(ctx global.IPlayer, pl *model.Player, req *proto_draw.C2SDrawGemAppraisal) {
	res := &proto_draw.S2CDrawGemAppraisal{}
	//判断一下卡池
	updatePoolState(pl)

	if pl.GemAppraisal.PoolId <= 0 {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_ConditionNo
		ctx.Send(res)
		return
	}

	if pl.GemAppraisal.PoolId != int(req.PoolId) {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_ConditionNo
		ctx.Send(res)
		return
	}

	//消耗道具
	costItems := make(map[int32]int32)
	costItems[define.ItemIdGemAppraisal] = req.Count

	if !internal.CheckItemsEnough(pl, costItems) {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NumNotEnough
		ctx.Send(res)
		return
	}

	resp, err := ctx.Invoke("Recruit", "RecruitGemAppraisal", int32(define.CARDPOOL_GEMAPPRAISAL), int32(pl.GemAppraisal.Title), req.Count, pl.GemAppraisal.Num)
	if err != nil {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_NOConfig
		ctx.Send(res)
	}

	//扣道具
	internal.SubItems(ctx, pl, costItems)
	rect := resp.(*model.RecruitResp)

	//增加次数
	pl.GemAppraisal.Num += req.Count

	var items []conf2.ItemE
	resItems := make([]*proto_public.Item, 0)
	for _, v := range rect.Ids {
		conf := config.TreasurePool.All()[int64(v)]

		if conf.Type == 3 || conf.Type == 4 {
			items = append(items, conf2.ItemE{
				ItemType: define.ItemTypeItem,
				ItemId:   conf.Value,
				ItemNum:  conf.Num,
			})

			resItems = append(resItems, &proto_public.Item{
				ItemId:   conf.Value,
				ItemNum:  conf.Num,
				ItemType: define.ItemTypeItem,
			})
		} else if conf.Type == 5 {
			if _, ok := pl.Collection.Collections[conf.Value]; !ok {
				pl.Collection.Collections[conf.Value] = new(model.CollectionOption)
				pl.Collection.Collections[conf.Value].Id = conf.Value
			} else {
				//转成碎片
				items := make(map[int32]int32)
				conf1 := config.Collection.All()[int64(conf.Value)]
				confItem := config.Item.All()[int64(conf1.Fragment)]
				items[conf1.Fragment] += conf.Num * confItem.CompositeNeed
				if len(items) > 0 {
					internal.AddItems(ctx, pl, items, false)
					resItems = append(resItems, &proto_public.Item{
						ItemId:   conf1.Fragment,
						ItemNum:  conf.Num * confItem.CompositeNeed,
						ItemType: define.ItemTypeItem,
					})
				}
				continue
			}

			collection := pl.Collection.Collections[conf.Value]
			pl.Collection.Collections[conf.Value] = collection

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

			resItems = append(resItems, &proto_public.Item{
				ItemId:   conf.Value,
				ItemNum:  conf.Num,
				ItemType: define.ItemTypeCollect,
			})
			//通告相关
			internal.SyncNotice_DrawCardGem(ctx, pl, define.CARDPOOL_GEMAPPRAISAL, conf.Type, conf.Value)
		}
	}

	//添加道具
	bag.AddAward(ctx, pl, items, false)

	res.Pool = &proto_draw.GemPoolOption{
		PoolId:        int32(pl.GemAppraisal.PoolId),
		PoolStartTime: pl.GemAppraisal.PoolStarTime,
		Num:           pl.GemAppraisal.Num,
		Awards:        pl.GemAppraisal.Reward,
	}

	res.Ids = resItems
	res.Code = proto_draw.ERRORCODEDRAW_ERR_Ok
	ctx.Send(res)
}

// 刷新卡池状态
func updatePoolState(pl *model.Player) {
	//判断当前卡池
	poolConfs := config.DrawPool.All()
	if pl.GemAppraisal.PoolId >= 0 {
		//时间
		poolConf := poolConfs[int64(pl.GemAppraisal.PoolId)]
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(poolConf.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			return
		}
		if utils.Now().Unix() < endTime.Unix() {
			return
		}

		//清除数据
		pl.GemAppraisal = new(model.GemAppraisal)
		pl.GemAppraisal.Reward = make([]int32, 0)
	}

	for _, v := range poolConfs {
		if v.Type == define.CARDPOOL_GEMAPPRAISAL {
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

			if utils.Now().Unix() >= startTime.Unix() && utils.Now().Unix() < endTime.Unix() {
				pl.GemAppraisal.PoolId = int(v.Id)
				pl.GemAppraisal.PoolStarTime = startTime.Unix()
				pl.GemAppraisal.PoolEndTime = endTime.Unix()
				pl.GemAppraisal.Title = int(v.Param)
				break
			}
		}
	}
}

// ReqGemStageAward 请求鉴宝奖励
func ReqGemAppraisalStageAward(ctx global.IPlayer, pl *model.Player, req *proto_draw.C2SGetGemAppraisalStageAward) {
	res := &proto_draw.S2CGetGemAppraisalStageAward{}

	if pl.GemAppraisal.PoolId <= 0 {
		res.Code = proto_draw.ERRORCODEDRAW_ERROR_ConditionNo
		ctx.Send(res)
		return
	}

	conf := config.DrawPool.All()[int64(pl.GemAppraisal.PoolId)]
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

	if utils.ContainsInt32(pl.GemAppraisal.Reward, req.Progress) {
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
	pl.GemAppraisal.Reward = append(pl.GemAppraisal.Reward, req.Progress)

	res.Pool = &proto_draw.GemPoolOption{
		PoolId:        int32(pl.GemAppraisal.PoolId),
		PoolStartTime: pl.GemAppraisal.PoolStarTime,
		Num:           pl.GemAppraisal.Num,
		Awards:        pl.GemAppraisal.Reward,
	}

	res.Code = proto_draw.ERRORCODEDRAW_ERR_Ok
	ctx.Send(res)
}
