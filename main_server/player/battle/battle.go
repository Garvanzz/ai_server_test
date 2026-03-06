package battle

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/activity"
	"xfx/main_server/player/danaotiangong"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/mission"
	"xfx/main_server/player/stage"
	"xfx/pkg/log"
	"xfx/proto/proto_game"
	"xfx/proto/proto_public"
)

// 战斗战报
func ReqChallengeBattleReport(ctx global.IPlayer, pl *model.Player, req *proto_game.C2SChallengeBattleReport) {
	res := new(proto_game.S2CChallengeBattleReport)
	//验证战报
	data, err := invoke.BattleClient(ctx).ReqChallengeBattleReport(pl.ToContext(), req)
	if err != nil {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	switch data.Scene {
	case define.BattleScene_Danaotiangong: //大闹天宫
		danaotiangong.BattleBack_Danaotiangong(ctx, pl, data.Data)
		break
	case define.BattleScene_Mission: //副本
		mission.BattleBack_Mission(ctx, pl, data.Data)
		break
	case define.BattleScene_StageBoss: //主关卡Boss
		stage.BattleBack_StageBoss(ctx, pl, data.Data)
		break
	case define.BattleScene_Player: //玩家
		BattleBack_Player(ctx, pl, data.Data)
	case define.BattleScene_Arena: //竞技场
		activity.BattleBack_Arena(ctx, pl, data.Data)
	case define.BattleScene_Tianti: //天梯
		activity.BattleBack_Tianti(ctx, pl, data.Data)
		break
	}
	log.Debug("战报回调:%v", data.Scene)
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// PVP战斗
func ReqChallengePlayerBattle(ctx global.IPlayer, pl *model.Player, req *proto_game.C2SChallengePlayerBattle) {
	res := new(proto_game.S2CChallengePlayerBattle)

	if req.PlayerId <= 0 {
		log.Debug("load no req playerId")
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		ctx.Send(res)
		return
	}

	//判断布阵
	if _, ok := pl.Lineup.LineUps[define.LINEUP_STAGE]; !ok {
		log.Debug("load no lineup")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	lineUp := pl.Lineup.LineUps[define.LINEUP_STAGE]
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

	battleId, err := invoke.BattleClient(ctx).BattlePlayer(pl.ToContext(), req.PlayerId)
	if err != nil {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		log.Debug("load no resp: %v", err)
		ctx.Send(res)
		return
	}

	if battleId == 0 {
		log.Debug("load battle player err : %v", battleId)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	res.BattleId = battleId
	res.PlayerId = req.PlayerId
	res.Data = make(map[int64]*proto_public.BattleHeroData)

	//获取战斗数据-他人
	otherData := global.GetPlayerInfo(req.PlayerId)
	//获取布阵
	_lineup := global.GetPlayerLineUpInfo(req.PlayerId)
	_stagelineup := _lineup[define.LINEUP_STAGE].HeroId
	batData := internal.GetBattleOtherPlayerData(otherData, _stagelineup)
	res.Data[req.PlayerId] = batData

	//获取战斗数据自己
	selfBattleData := internal.GetBattleSelfPlayerData(pl, lineUp.HeroId)

	//流派克制
	selfBattleData = global.GetLiupaiRestrain(_stagelineup, lineUp.HeroId, selfBattleData)

	res.Data[pl.Id] = selfBattleData
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 战斗回调
func BattleBack_Player(ctx global.IPlayer, pl *model.Player, data interface{}) {
	Imodel := data.(model.BattleReportBack_Player)
	resq := Imodel.Data.(*proto_game.C2SChallengeBattleReport)

	log.Debug("PVP玩家，战报回来，推送")

	global.BattleBackPlayer(ctx, pl, resq, nil)
}
