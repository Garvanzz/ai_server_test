package danaotiangong

import (
	"encoding/json"
	"fmt"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_danaotiangong"
	"xfx/proto/proto_game"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.Danaotiangong = new(model.Danaotiangong)
	pl.Danaotiangong.Stage = 1
	pl.Danaotiangong.Frequency = 1 //默认一周目
	pl.Danaotiangong.DaychallengeNum = 0
	pl.Danaotiangong.DaychallengeTime = utils.Now().Unix()
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Danaotiangong)
	if err != nil {
		log.Error("player[%v],save Danaotiangong marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save Danaotiangong error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.Danaotiangong, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load Danaotiangong error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.Danaotiangong, pl.Id))
	if err != nil {
		log.Error("player[%v],load Danaotiangong error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Danaotiangong)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Danaotiangong unmarshal error:%v", pl.Id, err)
	}

	pl.Danaotiangong = m
}

// 保存战斗记录
func saveBattleRecord(pl *model.Player, stageId int32, record *model.BattleRecord_Dabaotiangong, isSync bool) {
	j, err := json.Marshal(record)
	if err != nil {
		log.Error("player[%v],save Danaotiangong record marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save Danaotiangong record error, no this server:%v", err)
			return
		}
		rdb.RedisExec("HSET", fmt.Sprintf("%s:%d", define.BRDanaotiangong, pl.Id), stageId, j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

// 取战斗记录
func loadBattleRecord(pl *model.Player, stage int32) *model.BattleRecord_Dabaotiangong {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load Danaotiangong Record error, no this server:%v", err)
		return nil
	}
	reply, err := rdb.RedisExec("HGET", fmt.Sprintf("%s:%d", define.BRDanaotiangong, pl.Id), stage)
	if err != nil {
		log.Error("player[%v],load Danaotiangong Record error:%v", pl.Id, err)
		return nil
	}

	if reply == nil {
		return nil
	}

	m := new(model.BattleRecord_Dabaotiangong)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Danaotiangong Record unmarshal error:%v", pl.Id, err)
	}

	return m
}

// 请求初始大闹天宫
func ReqInitDanaotiangong(ctx global.IPlayer, pl *model.Player, req *proto_danaotiangong.C2SReqTiangongData) {
	res := &proto_danaotiangong.S2CRespTiangongData{}
	res.Stage = pl.Danaotiangong.Stage
	res.Frequency = pl.Danaotiangong.Frequency
	if !utils.CheckIsSameDayBySec(pl.Danaotiangong.DaychallengeTime, utils.Now().Unix(), 0) {
		pl.Danaotiangong.DaychallengeNum = 0
		pl.Danaotiangong.DaychallengeTime = utils.Now().Unix()
	}
	res.DaychallengeNum = pl.Danaotiangong.DaychallengeNum
	ctx.Send(res)
}

// 请求挑战
func ReqDntgBattleChallenge(ctx global.IPlayer, pl *model.Player, req *proto_danaotiangong.C2SChallengeDanaoTiangongBattle) {
	res := new(proto_danaotiangong.S2CChallengeDanaoTiangongBattle)

	//判断次数
	state, conf := internal.GetNormalMonthCard(ctx, pl)
	addCount := int32(0)
	if state {
		addCount = conf.DanaotiangongCount
	}
	if pl.Danaotiangong.DaychallengeNum > 5+addCount {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	//判断布阵
	if _, ok := pl.Lineup.LineUps[define.LINEUP_DANAOTIANGONG]; !ok {
		log.Debug("load no lineup")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	lineUp := pl.Lineup.LineUps[define.LINEUP_DANAOTIANGONG]
	isNull := true
	for _, v := range lineUp.HeroId {
		if v > 0 {
			isNull = false
			break
		}
	}

	if isNull {
		log.Debug("load no lineup")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	stageId := pl.Danaotiangong.Stage
	battleId, err := invoke.BattleClient(ctx).BattleDanaotiangong(pl.ToContext(), stageId)
	if err != nil {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		log.Debug("load no resp: %v", err)
		ctx.Send(res)
		return
	}

	if battleId == 0 {
		log.Debug("load battle danao err : %v", battleId)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	res.BattleId = battleId
	res.StageId = pl.Danaotiangong.Stage

	//获取战斗数据
	batData := internal.GetBattleSelfPlayerData(pl, lineUp.HeroId)
	res.Data = batData
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 请求记录
func ReqDntgBattleRecord(ctx global.IPlayer, pl *model.Player, req *proto_danaotiangong.C2SReqTiangongRecord) {
	res := new(proto_danaotiangong.S2CRespTiangongRecord)
	res.Records = make([]*proto_danaotiangong.TiangongRecordItem, 0)

	record := loadBattleRecord(pl, req.Stage)
	log.Debug("请求记录:%v", record)
	if record != nil {
		for _, v := range record.Records {
			res.Records = append(res.Records, &proto_danaotiangong.TiangongRecordItem{
				Id:    v.Id,
				Time:  v.Time,
				IsWin: v.IsWin,
			})
		}
		log.Debug("请求记录:%v", record.Records)
	}
	ctx.Send(res)
}

// 战斗回调
func BattleBack_Danaotiangong(ctx global.IPlayer, pl *model.Player, data interface{}) {
	Imodel := data.(model.BattleReportBack_Danaotiangong)
	resq := Imodel.Data.(*proto_game.C2SChallengeBattleReport)

	pl.Danaotiangong.DaychallengeNum += 1
	resp := &proto_danaotiangong.S2CTiangongSettle{
		IsWin:           resq.WinId == pl.Id,
		Stage:           pl.Danaotiangong.Stage,
		DaychallengeNum: pl.Danaotiangong.DaychallengeNum,
	}

	//战斗记录
	record := loadBattleRecord(pl, pl.Danaotiangong.Stage)
	if record == nil {
		record = new(model.BattleRecord_Dabaotiangong)
		record.Records = make([]*model.BattleRecord_DabaotiangongOpt, 0)
	}

	confUproar := config.Uproar.All()[int64(pl.Danaotiangong.Stage)]

	record.Records = append(record.Records, &model.BattleRecord_DabaotiangongOpt{
		Id:         confUproar.BossId,
		Time:       resq.Time,
		IsWin:      resq.WinId == pl.Id,
		CreateTime: utils.Now().Unix(),
	})

	//保存战斗记录
	saveBattleRecord(pl, pl.Danaotiangong.Stage, record, true)

	//任务
	task.Dispatch(ctx, pl, define.TaskDanaotiangongChallengeTime, 1, 0, true)

	//自己赢
	if resq.WinId == pl.Id {
		//获取奖励
		conf := getConfigUproarFrequency(pl.Danaotiangong.Frequency, pl.Danaotiangong.Stage)

		pl.Danaotiangong.Stage += 1
		if pl.Danaotiangong.Stage > 9 && pl.Danaotiangong.Frequency < 10 {
			pl.Danaotiangong.Stage = 1
			pl.Danaotiangong.Frequency += 1
		}
		if conf != nil {
			resp.Awards = global.ItemFormat(conf)
		}
	}
	ctx.Send(resp)
}

// 获取奖励通过周目
func getConfigUproarFrequency(frequency int32, stage int32) []conf.ItemE {
	confs := config.UproarFrequency.All()
	for _, v := range confs {
		if v.Frequency == frequency {
			switch stage {
			case 1:
				return v.OneLevelAward
			case 2:
				return v.TwoLevelAward
			case 3:
				return v.ThreeLevelAward
			case 4:
				return v.FourLevelAward
			case 5:
				return v.FiveLevelAward
			case 6:
				return v.SexLevelAward
			case 7:
				return v.SevenLevelAward
			case 8:
				return v.EightLevelAward
			case 9:
				return v.NineLevelAward
			}
		}
	}

	return nil
}
