package shenjidraw

import (
	"encoding/json"
	"fmt"
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
	"xfx/proto/proto_destiny"
)

func Init(pl *model.Player) {
	pl.ShenjiDraw = new(model.ShenjiDraw)
	pl.ShenjiDraw.Pools = make(map[int32]*model.ShenjiDrawPool)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.ShenjiDraw)
	if err != nil {
		log.Error("player[%v],save ShenjiDraw marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save ShenjiDraw error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerShenjiDraw, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load ShenjiDraw error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerShenjiDraw, pl.Id))
	if err != nil {
		log.Error("player[%v],load bag error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.ShenjiDraw)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Equip unmarshal error:%v", pl.Id, err)
	}

	pl.ShenjiDraw = m
}

// 请求神机
func ReqInitShenjiDraw(ctx global.IPlayer, pl *model.Player, req *proto_destiny.C2SShenjiBoxInit) {
	res := &proto_destiny.S2CShenjiBoxInit{}
	drawPoolRefreshState(pl)
	//纪录
	for i, v := range pl.ShenjiDraw.Pools {
		if v.LastRecordTime != 0 {
			if utils.IsSameWeekBySec(v.LastRecordTime, utils.Now().Unix()) == false {
				pl.ShenjiDraw.Pools[i].LastRecordTime = utils.Now().Unix()
				pl.ShenjiDraw.Pools[i].ShenjiRecords = make([]*model.ShenjiRecord, 0)
			}
		}
	}

	res.Pools = model.ToDestinyShenjiProto(pl.ShenjiDraw.Pools)
	ctx.Send(res)
}

// 刷新卡池的状态
func drawPoolRefreshState(pl *model.Player) {
	confs := config.DrawPool.All()
	for k, v := range pl.ShenjiDraw.Pools {
		conf := confs[int64(v.PoolId)]
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", impl.Trim(conf.EndTime), time.Local)
		if err != nil {
			log.Error("checkCfg parse endTime err:%v", err)
			continue
		}
		if utils.Now().Unix() >= endTime.Unix() && conf.ActivityType == 2 && conf.Type == 2 {
			delete(pl.ShenjiDraw.Pools, k)
		}
	}

	//判断有没有新加的
	for _, v := range confs {
		if v.Type != 2 {
			continue
		}
		if _, ok := pl.ShenjiDraw.Pools[v.Id]; ok {
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
			if utils.Now().Unix() < startTime.Unix() || utils.Now().Unix() >= endTime.Unix() {
				continue
			}
		}

		//添加新的活动
		pl.ShenjiDraw.Pools[v.Id] = &model.ShenjiDrawPool{
			PoolId:        v.Id,
			PoolStartTime: startTime.Unix(),
		}
	}
}

// 神机信息
func ReqShenjiDraw(ctx global.IPlayer, pl *model.Player, req *proto_destiny.C2SDrawShenji) {
	res := &proto_destiny.S2CDrawShenji{}

	//获取卡池
	conf := config.DrawPool.All()
	if _, ok := conf[int64(req.PoolId)]; !ok {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_NOConfig
		ctx.Send(res)
		return
	}

	if _, ok := pl.ShenjiDraw.Pools[req.PoolId]; !ok {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_ConditionNo
		ctx.Send(res)
		return
	}

	//消耗道具
	costItems := make(map[int32]int32)
	costItems[define.ItemIdShenji] = req.Count

	if !internal.CheckItemsEnough(pl, costItems) {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_NumNotEnough
		ctx.Send(res)
		return
	}

	pool := pl.ShenjiDraw.Pools[req.PoolId]
	resp, err := ctx.Invoke("Recruit", "RecruitShenji", int32(define.CARDPOOL_SHENJI), req.Count, pool.BdNum)
	if err != nil {
		res.Code = proto_destiny.ERRORCODEDESTINY_ERROR_NOConfig
		ctx.Send(res)
	}

	//扣道具
	internal.SubItems(ctx, pl, costItems)
	rect := resp.(*model.RecruitResp)
	//保底次数
	pool.BdNum = rect.BdNum

	//增加次数
	pool.Num += req.Count

	var items []conf2.ItemE
	for _, v := range rect.Ids {
		confHero := config.ShenjiPool.All()[int64(v)]

		if confHero.Type == 3 {
			items = append(items, conf2.ItemE{
				ItemType: define.ItemTypeItem,
				ItemId:   confHero.Value,
				ItemNum:  confHero.Num,
			})

			//记录
			if pool.ShenjiRecords == nil {
				pool.ShenjiRecords = make([]*model.ShenjiRecord, 0)
				pool.LastRecordTime = utils.Now().Unix()
			}

			//跨周
			if !utils.IsSameWeekBySec(utils.Now().Unix(), pool.LastRecordTime) {
				pool.ShenjiRecords = make([]*model.ShenjiRecord, 0)
				pool.LastRecordTime = utils.Now().Unix()
			}

			pool.ShenjiRecords = append(pool.ShenjiRecords, &model.ShenjiRecord{
				Id:  confHero.Value,
				Num: confHero.Num,
			})
		}
	}

	//添加道具
	bag.AddAward(ctx, pl, items, true)
	pl.ShenjiDraw.Pools[req.PoolId] = pool

	rec := make([]*proto_destiny.ShenjiAwardRecord, 0)
	for i := 0; i < len(pool.ShenjiRecords); i++ {
		rec = append(rec, &proto_destiny.ShenjiAwardRecord{
			Id:  pool.ShenjiRecords[i].Id,
			Num: pool.ShenjiRecords[i].Num,
		})
	}

	res.ShenjiPoolInfo = &proto_destiny.ShenjiPoolOption{
		PoolId:        pool.PoolId,
		PoolStartTime: pool.PoolStartTime,
		Num:           pool.Num,
	}
	res.Code = proto_destiny.ERRORCODEDESTINY_ERR_Ok
	ctx.Send(res)
}

// 神机记录
func ReqShenjiDrawRecord(ctx global.IPlayer, pl *model.Player, req *proto_destiny.C2SReqGetDrawRecord) {
	res := &proto_destiny.S2CReqGetDrawRecord{}

	//获取卡池
	conf := config.DrawPool.All()
	if _, ok := conf[int64(req.PoolId)]; !ok {
		ctx.Send(res)
		return
	}

	if _, ok := pl.ShenjiDraw.Pools[req.PoolId]; !ok {
		ctx.Send(res)
		return
	}

	if pl.ShenjiDraw.Pools == nil {
		ctx.Send(res)
		return
	}

	arr := make([]*proto_destiny.ShenjiAwardRecord, 0)
	for i := 0; i < len(pl.ShenjiDraw.Pools[req.PoolId].ShenjiRecords); i++ {
		arr = append(arr, &proto_destiny.ShenjiAwardRecord{
			Id:  pl.ShenjiDraw.Pools[req.PoolId].ShenjiRecords[i].Id,
			Num: pl.ShenjiDraw.Pools[req.PoolId].ShenjiRecords[i].Num,
		})
	}

	res.Records = arr
	ctx.Send(res)
}
