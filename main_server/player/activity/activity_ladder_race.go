package activity

import (
	"encoding/json"
	"fmt"
	"time"
	"xfx/core/common"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/rank"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_game"
	"xfx/proto/proto_public"
)

// ReqActivityLadderRaceSetLineUp 天梯阵容调整
func ReqActivityLadderRaceSetLineUp(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SLadderRaceSetLineUp) {
	res := new(proto_activity.S2CLadderRaceSetLineUp)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeLadderRace)
	if err != nil {
		log.Error("ReqActivityLadderRaceSetLineUp invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityLadderRaceSetLineUp reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	//判断阵容对不对
	lineup, ok := pl.Lineup.LineUps[define.LINEUP_Tianti]
	if !ok {
		log.Error("ReqActivityLadderRaceSetLineUp lineup is nil")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if len(lineup.HeroId) <= 0 {
		log.Error("ReqActivityLadderRaceSetLineUp lineup is nil")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	for _, v := range req.HeroIds {
		if !common.IsHaveValueIntArray(lineup.HeroId, v) {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}
	}

	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":    "tianti_lineup",
		"req":    req,
		"player": pl.ToContext(),
	})
}

// C2SLadderRaceGetPlayerLineUp 天梯获取布阵
func ReqActivityLadderRaceGetPlayerLineUp(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SLadderRaceGetPlayerLineUp) {
	res := new(proto_activity.S2CLadderRaceGetPlayerLineUp)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeLadderRace)
	if err != nil {
		log.Error("ReqActivityLadderRaceGetPlayerLineUp invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityLadderRaceGetPlayerLineUp reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	_reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActId, req)
	if err != nil {
		log.Error("ReqActivityLadderRaceGetPlayerLineUp invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	res = _reply.(*proto_activity.S2CLadderRaceGetPlayerLineUp)
	ctx.Send(res)
}

// C2STiantiBattle 天梯挑战
func ReqActivityLadderRaceBattle(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SLadderRaceBattle) {
	res := new(proto_activity.S2CLadderRaceBattle)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeLadderRace)
	if err != nil {
		log.Error("ReqActivityLadderRaceBattle invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityLadderRaceBattle reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	//判断有没有布阵
	if _, ok := pl.Lineup.LineUps[define.LINEUP_Tianti]; !ok {
		log.Debug("load no lineup")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	lineUp := pl.Lineup.LineUps[define.LINEUP_Tianti]
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

	_reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActId, req)
	if err != nil {
		log.Error("ReqActivityLadderRaceBattle invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	res = _reply.(*proto_activity.S2CLadderRaceBattle)
	if res.Code != proto_public.CommonErrorCode_ERR_OK {
		ctx.Send(res)
		return
	}

	battleId, err := invoke.BattleClient(ctx).BattleTianti(pl.ToContext(), res.Id, reply.ConfigId, req.ActId)
	if err != nil {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		log.Debug("load no resp: %v", err)
		ctx.Send(res)
		return
	}

	res.BattleId = battleId
	res.Data = make(map[int64]*proto_public.BattleHeroData)

	//人机
	rankKey := fmt.Sprintf("%s:%d", define.RankTypeTiantiKey, req.ActId)
	otherRankItem := rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, res.Id)
	var _stagelineup []*proto_public.CommonPlayerLineUpItemInfo
	if res.Id < define.PlayerIdBase {
		batData := internal.GetRobotBattleData(res.Id)

		batData.PlayerInfo = global.ToCommonPlayerByRobot(res.Id)
		batData.PlayerInfo.Score = otherRankItem.Rank
		batData.PlayerInfo.ServerId = int32(pl.Id / define.PlayerIdBase)

		res.Data[res.Id] = batData
		_stagelineup = internal.GetRobotLineUp(res.GetId())
	} else {
		//获取战斗数据-他人
		otherData := global.GetPlayerInfo(res.Id)
		//获取布阵
		_lineup := global.GetPlayerLineUpInfo(res.Id)
		_stagelineup = _lineup[define.LINEUP_Tianti].HeroId
		batData := internal.GetBattleOtherPlayerData(otherData, _stagelineup)

		batData.PlayerInfo = otherData.ToCommonPlayer()
		batData.PlayerInfo.Score = otherRankItem.Rank
		batData.PlayerInfo.ServerId = int32(pl.Id / define.PlayerIdBase)

		res.Data[res.Id] = batData
	}

	//获取战斗数据自己
	selfBattleData := internal.GetBattleSelfPlayerData(pl, lineUp.HeroId)

	//流派克制
	selfBattleData = global.GetLiupaiRestrain(_stagelineup, lineUp.HeroId, selfBattleData)

	selfRankItem := rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, pl.Id)
	selfBattleData.PlayerInfo = global.GetPlayerInfo(pl.Id).ToCommonPlayer()
	selfBattleData.PlayerInfo.Score = selfRankItem.Score
	selfBattleData.PlayerInfo.ServerId = int32(pl.Id / define.PlayerIdBase)

	res.Data[pl.Id] = selfBattleData
	res.Code = proto_public.CommonErrorCode_ERR_OK

	ctx.Send(res)
}

// 战斗回调
func BattleBack_Tianti(ctx global.IPlayer, pl *model.Player, data interface{}) {
	Imodel := data.(model.BattleReportBack_Tianti)
	resq := Imodel.Data.(*proto_game.C2SChallengeBattleReport)

	isSuc := resq.WinId == pl.Id
	_, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), Imodel.ActId, resq)
	if err != nil {
		log.Debug("load no resp: %v", err)
		global.BattleBackPlayer(ctx, pl, resq, nil)
		return
	}

	confs := config.ActLadderRace.All()
	conf := conf.ActLadderRace{}
	for _, v := range confs {
		conf = v
		break
	}

	log.Debug("战斗回调:%v", Imodel.ActCId)
	if conf.Id <= 0 {
		global.BattleBackPlayer(ctx, pl, resq, nil)
		return
	}
	_award := conf.Awards
	//上榜数据
	if isSuc {
		//掉落奖励
		award := bag.GetDrop(conf.DropAwardId, 0)
		_award = append(_award, award...)

		rankKey := fmt.Sprintf("%s:%d", define.RankTypeTiantiKey, Imodel.ActId)
		selfRankItem := rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, pl.Id)
		targetRankItem := rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, Imodel.PlayerId)

		// 获取排名
		var selfRank, targetRank int64
		selfRank = selfRankItem.Rank
		targetRank = targetRankItem.Rank

		rdb, _ := db.GetEngine(pl.Cache.App.GetEnv().ID)
		// 记录战报
		selfRecordKey := fmt.Sprintf("%s:%d_%d", define.RankTypeTiantiRecordKey, Imodel.ActId, pl.Id)
		selfRecord := &model.BattleReportRecord_LadderRace{
			TargetId: Imodel.PlayerId,
			IsAttack: true,
			ActId:    Imodel.ActId,
			Time:     time.Now().Unix(),
			Rank:     int32(selfRank),
		}
		selfRecordJson, _ := json.Marshal(selfRecord)
		rdb.RedisExec("LPUSH", selfRecordKey, string(selfRecordJson))
		rdb.RedisExec("LTRIM", selfRecordKey, 0, 99)

		//人机不进库
		if Imodel.PlayerId > define.PlayerIdBase {
			targetRecordKey := fmt.Sprintf("%s:%d_%d", define.RankTypeTiantiRecordKey, Imodel.ActId, Imodel.PlayerId)
			targetRecord := &model.BattleReportRecord_LadderRace{
				TargetId: pl.Id,
				IsAttack: false,
				ActId:    Imodel.ActId,
				Time:     time.Now().Unix(),
				Rank:     int32(targetRank),
			}
			targetRecordJson, _ := json.Marshal(targetRecord)
			rdb.RedisExec("LPUSH", targetRecordKey, string(targetRecordJson))
			rdb.RedisExec("LTRIM", targetRecordKey, 0, 99)
		}
	}

	//积分
	actData, err := invoke.ActivityClient(ctx).GetActivityData(pl.ToContext(), Imodel.ActId)
	if err != nil {
		log.Error("处理积分，活动数据有报错: %v", err)
	} else {
		pl.SetProp(define.PlayerPropRank, int64(actData.LadderRace.Score), false)
	}

	cons := global.MergeItemE(_award)
	internal.AddItems(ctx, pl, cons, true)

	log.Debug("天梯，战报回来，推送变化")
	global.BattleBackPlayer(ctx, pl, resq, global.ItemFormat(_award))
}

// C2STiantiBattleRecord 天梯记录
func ReqActivityTiantiBattleRecord(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SLadderRaceReqRecord) {
	res := new(proto_activity.S2CLadderRaceReqRecord)
	res.Options = make([]*proto_activity.LadderRaceRecordOption, 0)

	// 获取 Redis 连接
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("ReqActivityArenaBattleRecord get redis error:%v", err)
		ctx.Send(res)
		return
	}

	// 构建 Redis Key
	rankRecordKey := fmt.Sprintf("%s:%d_%d", define.RankTypeTiantiRecordKey, req.ActId, pl.Id)

	// 从 Redis 读取战报记录 (LRANGE 0 -1 获取所有记录)
	reply, err := rdb.RedisExec("LRANGE", rankRecordKey, 0, -1)
	if err != nil {
		log.Error("ReqActivityArenaBattleRecord load record error:%v", err)
		ctx.Send(res)
		return
	}

	// 处理 Redis 返回为空的情况
	if reply == nil {
		log.Debug("竞技战报记录为空: playerId=%d, actId=%d", pl.Id, req.ActId)
		ctx.Send(res)
		return
	}

	// 反序列化 JSON 数据
	recordItems := reply.([]interface{})
	for _, item := range recordItems {
		// 序列化的战报数据
		battleRecord := new(model.BattleReportRecord_LadderRace)
		err := json.Unmarshal(item.([]byte), battleRecord)
		if err != nil {
			log.Error("unmarshal battle record error:%v", err)
			continue
		}

		var targetPlayerInfo *proto_public.CommonPlayerInfo
		//人机
		if battleRecord.TargetId < define.PlayerIdBase {
			targetPlayerInfo = global.ToCommonPlayerByRobot(battleRecord.TargetId)
		} else {
			// 获取对方玩家信息
			_targetPlayerInfo := global.GetPlayerInfo(battleRecord.TargetId)
			if _targetPlayerInfo == nil {
				log.Error("get target player info error, playerId=%d", battleRecord.TargetId)
				continue
			}
			targetPlayerInfo = _targetPlayerInfo.ToCommonPlayer()
		}

		// 组装 ArenaRecordOption
		option := new(proto_activity.LadderRaceRecordOption)
		option.Id = battleRecord.TargetId
		option.IsAttack = battleRecord.IsAttack
		option.PlayerInfo = targetPlayerInfo
		option.Rank = battleRecord.Rank
		option.Time = battleRecord.Time
		option.ActId = battleRecord.ActId
		res.Options = append(res.Options, option)
	}

	ctx.Send(res)
}
