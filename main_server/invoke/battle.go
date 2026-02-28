package invoke

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_game"
	"xfx/proto/proto_player"
)

type BattleModClient struct {
	invoke Invoker
	Type   string
}

func BattleClient(invoker Invoker) BattleModClient {
	return BattleModClient{
		invoke: invoker,
		Type:   define.ModuleBattle,
	}
}

// BattleDanaotiangong 大闹天宫
func (m BattleModClient) BattleDanaotiangong(ctx *proto_player.Context, stageId int32) (int64, error) {
	return Int64(m.invoke.Invoke(m.Type, "BattleDanaotiangong", ctx, stageId))
}

// ReqChallengeBattleReport 战报
func (m BattleModClient) ReqChallengeBattleReport(ctx *proto_player.Context, req *proto_game.C2SChallengeBattleReport) (model.ChallengeBattleReportBack, error) {
	result, err := m.invoke.Invoke(m.Type, "ReqChallengeBattleReport", ctx, req)
	if err != nil {
		log.Error("ReqChallengeBattleReport err:%v", err)
		return model.ChallengeBattleReportBack{}, err
	}
	if result == nil {
		return model.ChallengeBattleReportBack{}, nil
	}

	return result.(model.ChallengeBattleReportBack), nil
}

// BattleMission 副本
func (m BattleModClient) BattleMission(ctx *proto_player.Context, typ, stageId int32) (int64, error) {
	return Int64(m.invoke.Invoke(m.Type, "BattleMission", ctx, typ, stageId))
}

// BattleStageBoss 关卡boss
func (m BattleModClient) BattleStageBoss(ctx *proto_player.Context, cycle, stageId, chapter int32) (int64, error) {
	return Int64(m.invoke.Invoke(m.Type, "BattleStageBoss", ctx, cycle, stageId, chapter))
}

// BattlePlayer 玩家
func (m BattleModClient) BattlePlayer(ctx *proto_player.Context, playerId int64) (int64, error) {
	return Int64(m.invoke.Invoke(m.Type, "BattlePlayer", ctx, playerId))
}

// BattleArena 竞技场
func (m BattleModClient) BattleArena(ctx *proto_player.Context, playerId int64, ActId int64, ActCid int64, IsFuchou bool) (int64, error) {
	return Int64(m.invoke.Invoke(m.Type, "BattleArena", ctx, playerId, ActId, ActCid, IsFuchou))
}

// BattleTianti 天梯
func (m BattleModClient) BattleTianti(ctx *proto_player.Context, playerId int64, ActId int64, ActCid int64) (int64, error) {
	return Int64(m.invoke.Invoke(m.Type, "BattleTianti", ctx, playerId, ActId, ActCid))
}
