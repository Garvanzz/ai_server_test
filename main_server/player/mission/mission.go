package mission

import (
	"encoding/json"
	"fmt"
	config2 "xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/rank"
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_game"
	"xfx/proto/proto_mission"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.Mission = new(model.Mission)
	pl.Mission.Box = new(model.MissionItem)
	pl.Mission.Lingyu = new(model.MissionItem)
	pl.Mission.ClimbTower = new(model.MissionItem)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Mission)
	if err != nil {
		log.Error("player[%v],save Mission marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save Mission error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerMission, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save draw Mission, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerMission, pl.Id))
	if err != nil {
		log.Error("player[%v],load hero error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Mission)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load Mission unmarshal error:%v", pl.Id, err)
	}

	pl.Mission = m
}

// ReqInitMission 请求副本数据
func ReqInitMission(ctx global.IPlayer, pl *model.Player, req *proto_mission.C2SInitMissionStageData) {
	res := &proto_mission.S2CInitMissionStageData{}

	resetMission(pl)
	mission := model.ToMissionProtoByMisson(pl.Mission)
	res.Opt = mission
	ctx.Send(res)
}

func resetMission(pl *model.Player) {
	if !utils.CheckIsSameDayBySec(pl.Mission.Box.Time, utils.Now().Unix(), 0) {
		pl.Mission.Box.ChallengeNum = 0
		pl.Mission.Box.Time = utils.Now().Unix()
	}

	if !utils.CheckIsSameDayBySec(pl.Mission.Lingyu.Time, utils.Now().Unix(), 0) {
		pl.Mission.Lingyu.ChallengeNum = 0
		pl.Mission.Lingyu.Time = utils.Now().Unix()
	}
}

// 请求挑战副本
func ReqMissionBattleChallenge(ctx global.IPlayer, pl *model.Player, req *proto_mission.C2SChallengeMissionBattle) {
	res := new(proto_mission.S2CChallengeMissionBattle)

	//获取配置
	confs := config2.Mission.All()
	var config conf.Mission
	typ := int32(0)
	if req.Type == proto_mission.MissionType_Box {
		typ = 1
	} else if req.Type == proto_mission.MissionType_Lingyu {
		typ = 2
	} else if req.Type == proto_mission.MissionType_ClimbTower {
		typ = 4
	}

	for _, v := range confs {
		if v.Type == typ {
			config = v
			break
		}
	}

	if config.Id <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(res)
		return
	}

	//判断是否开启

	var data *model.MissionItem
	if req.Type == proto_mission.MissionType_Box {
		data = pl.Mission.Box
	} else if req.Type == proto_mission.MissionType_Lingyu {
		data = pl.Mission.Lingyu
	} else if req.Type == proto_mission.MissionType_ClimbTower {
		data = pl.Mission.ClimbTower
	}

	//判断次数
	state, conf := internal.GetNormalMonthCard(ctx, pl)
	addCount := int32(0)
	if state {
		if req.Type == proto_mission.MissionType_Box {
			addCount = conf.BoxMissionCount
		} else if req.Type == proto_mission.MissionType_Lingyu {
			addCount = conf.LingyuMissionCount
		} else if req.Type == proto_mission.MissionType_ClimbTower {
			addCount = conf.ClimbeTowerCount
		}
	}
	if config.ChallengeNum > 0 && data.ChallengeNum >= config.ChallengeNum+addCount {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	//判断材料
	consume := global.MergeItemE(config.ChallengeCost)
	if !internal.CheckItemsEnough(pl, consume) {
		res.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(res)
		return
	}

	//判断布阵
	if req.Type == proto_mission.MissionType_Box || req.Type == proto_mission.MissionType_Lingyu {
		if _, ok := pl.Lineup.LineUps[define.LINEUP_STAGE]; !ok {
			log.Debug("load no lineup")
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}
	} else if req.Type == proto_mission.MissionType_ClimbTower {
		if _, ok := pl.Lineup.LineUps[define.LINEUP_CLIMBTOWER]; !ok {
			log.Debug("load no lineup")
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}
	}

	lineUp := pl.Lineup.LineUps[define.LINEUP_STAGE]
	if req.Type == proto_mission.MissionType_ClimbTower {
		lineUp = pl.Lineup.LineUps[define.LINEUP_CLIMBTOWER]
	}
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

	stageId := data.Stage
	battleId, err := invoke.BattleClient(ctx).BattleMission(pl.ToContext(), typ, stageId)
	if err != nil {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		log.Debug("load no resp: %v", err)
		ctx.Send(res)
		return
	}

	if battleId == 0 {
		log.Debug("load battle mission err : %v", battleId)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	res.BattleId = battleId
	res.Type = req.Type

	//消耗材料
	internal.SubItems(ctx, pl, consume)
	if req.Type == proto_mission.MissionType_Box {
		pl.Mission.Box.ChallengeNum += 1
		pl.Mission.Box.Stage += 1
	} else if req.Type == proto_mission.MissionType_Lingyu {
		pl.Mission.Lingyu.ChallengeNum += 1
		pl.Mission.Lingyu.Stage += 1
	} else if req.Type == proto_mission.MissionType_ClimbTower {
		pl.Mission.ClimbTower.ChallengeNum += 1
		pl.Mission.ClimbTower.Stage += 1
	}

	//获取战斗数据
	batData := internal.GetBattleSelfPlayerData(pl, lineUp.HeroId)
	res.Data = batData
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 战斗回调
func BattleBack_Mission(ctx global.IPlayer, pl *model.Player, data interface{}) {
	Imodel := data.(model.BattleReportBack_Mission)
	resq := Imodel.Data.(*proto_game.C2SChallengeBattleReport)

	if resq.WinId != pl.Id {
		log.Error("副本战斗失败")
		return
	}

	//宝箱副本 - 任务
	if Imodel.Typ == 1 {
		task.Dispatch(ctx, pl, define.TaskPassBoxMissionTime, 1, 0, true)
	} else if Imodel.Typ == 2 {
		task.Dispatch(ctx, pl, define.TaskPassLingyuMissionTime, 1, 0, true)
	}

	//发奖proto_mission.MissionType_ClimbTower
	if Imodel.Typ == 4 {
		//更新排名
		rank.UpdateClimbTowerRank(ctx, pl, 1)
		//爬塔发奖
		climbTowerAward(ctx, pl)

		//任务
		flower := pl.Mission.ClimbTower.Stage
		task.Dispatch(ctx, pl, define.TaskClimbTower, flower, 0, false)
		return
	}

	confs := config2.Mission.All()
	var config conf.Mission
	typ := Imodel.Typ
	for _, v := range confs {
		if v.Type == typ {
			config = v
			break
		}
	}

	if config.Id <= 0 {
		log.Error("发放副本奖励错误，配置出错: %v", typ)
		return
	}

	award := config.Award
	for _, v := range award {
		v.ItemNum += v.ItemNum * Imodel.Stage * int32(float32(config.AddRate)/float32(1000))
	}

	bag.AddAward(ctx, pl, award, true)
}

// 爬塔发奖
func climbTowerAward(ctx global.IPlayer, pl *model.Player) {
	flower := pl.Mission.ClimbTower.Stage
	flower += 1
	rate := float32(flower) / float32(50)
	if flower/50 == 0 && flower > 0 {
		flower = 50
	} else {
		flower = flower % 50
	}
	confs := config2.ClimbTower.All()
	var conf conf.ClimbTower
	for _, v := range confs {
		if v.Flower == flower {
			conf = v
			break
		}
	}

	if conf.Id <= 0 {
		log.Error("climbTowerAward error no config")
		return
	}

	award := conf.Award
	for _, v := range award {
		v.ItemNum += v.ItemNum * int32(float32(conf.AwardAddRate)/float32(1000)*float32(rate))
	}
	bag.AddAward(ctx, pl, award, true)
}
