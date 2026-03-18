package activity

import (
	"encoding/json"
	"fmt"
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
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_game"
	"xfx/proto/proto_public"
)

// ReqActivityArenaRefreshBattlePlayer 竞技场刷新更换敌人
func ReqActivityArenaRefreshBattlePlayer(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SArenaRefreshBattlePlayer) {
	res := new(proto_activity.S2CArenaRefreshBattlePlayer)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeArena)
	if err != nil {
		log.Error("ReqActivityArenaRefreshBattlePlayer invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityArenaRefreshBattlePlayer reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":    "arena_refreshbattleplayer",
		"req":    req,
		"player": pl.ToContext(),
	})
}

// ReqActivityTheArenaSetLineUp 竞技场阵容调整
func ReqActivityTheArenaSetLineUp(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SArenaSetLineUp) {
	res := new(proto_activity.S2CArenaSetLineUp)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeArena)
	if err != nil {
		log.Error("ReqActivityTheArenaSetLineUp invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityTheArenaSetLineUp reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	//判断阵容对不对
	lineup, ok := pl.Lineup.LineUps[define.LINEUP_ARENA]
	if !ok {
		log.Error("ReqActivityTheArenaSetLineUp lineup is nil")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if len(lineup.HeroId) <= 0 {
		log.Error("ReqActivityTheArenaSetLineUp lineup is nil")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	for _, v := range req.HeroIds {
		if !utils.ContainsInt32(lineup.HeroId, v) {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}
	}

	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":    "arena_lineup",
		"req":    req,
		"player": pl.ToContext(),
	})
}

// C2SArenaGetPlayerLineUp 竞技场获取布阵
func ReqActivityArenaGetPlayerLineUp(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SArenaGetPlayerLineUp) {
	res := new(proto_activity.S2CArenaGetPlayerLineUp)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeArena)
	if err != nil {
		log.Error("ReqActivityArenaGetPlayerLineUp invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityArenaGetPlayerLineUp reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	_reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActId, req)
	if err != nil {
		log.Error("ReqActivityArenaGetPlayerLineUp invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	res = _reply.(*proto_activity.S2CArenaGetPlayerLineUp)
	ctx.Send(res)
}

// C2SArenaBattle 竞技场挑战
func ReqActivityArenaBattle(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SArenaBattle) {
	res := new(proto_activity.S2CArenaBattle)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeArena)
	if err != nil {
		log.Error("ReqActivityArenaBattle invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityArenaBattle reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	//判断有没有布阵
	if _, ok := pl.Lineup.LineUps[define.LINEUP_ARENA]; !ok {
		log.Debug("load no lineup")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	lineUp := pl.Lineup.LineUps[define.LINEUP_ARENA]
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

	//复仇
	var cost []conf.ItemE
	if req.IsFuchou {
		//判断道具够不够
		cost = config.Global.Get().ArenaFuchouCost
		costMap := global.MergeItemE(cost)
		if !internal.CheckItemsEnough(pl, costMap) {
			log.Debug("fuchou cost is not enough")
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}
	}

	_reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActId, req)
	if err != nil {
		log.Error("ReqActivityArenaBattle invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	res = _reply.(*proto_activity.S2CArenaBattle)
	if res.Code != proto_public.CommonErrorCode_ERR_OK {
		ctx.Send(res)
		return
	}

	battleId, err := invoke.BattleClient(ctx).BattleArena(pl.ToContext(), req.Id, req.ActId, reply.ConfigId, req.IsFuchou)
	if err != nil {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		log.Debug("load no resp: %v", err)
		ctx.Send(res)
		return
	}

	//复仇
	if req.IsFuchou && len(cost) > 0 {
		//扣除道具
		costMap := global.MergeItemE(cost)
		internal.SubItems(ctx, pl, costMap)
	}

	res.BattleId = battleId
	res.Id = req.Id
	res.Data = make(map[int64]*proto_public.BattleHeroData)

	//人机
	rankKey := fmt.Sprintf("%s:%d", define.RankTypeArenaKey, req.ActId)
	otherRankItem := rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, req.Id)
	var _stagelineup []*proto_public.CommonPlayerLineUpItemInfo
	if req.Id < define.PlayerIdBase {
		batData := internal.GetRobotBattleData(req.Id)

		batData.PlayerInfo = global.ToCommonPlayerByRobot(req.Id)
		batData.PlayerInfo.Score = otherRankItem.Rank
		batData.PlayerInfo.ServerId = int32(pl.GetProp(define.PlayerPropServerId))

		res.Data[req.Id] = batData
		_stagelineup = internal.GetRobotLineUp(req.GetId())
	} else {
		//获取战斗数据-他人
		otherData := global.GetPlayerInfo(req.Id)
		//获取布阵
		_lineup := global.GetPlayerLineUpInfo(req.Id)
		_stagelineup = _lineup[define.LINEUP_ARENA].HeroId
		batData := internal.GetBattleOtherPlayerData(otherData, _stagelineup)

		batData.PlayerInfo = otherData.ToCommonPlayer()
		batData.PlayerInfo.Score = otherRankItem.Rank
		batData.PlayerInfo.ServerId = otherData.ServerId

		res.Data[req.Id] = batData
	}

	//获取战斗数据自己
	selfBattleData := internal.GetBattleSelfPlayerData(pl, lineUp.HeroId)

	//流派克制
	selfBattleData = global.GetLiupaiRestrain(_stagelineup, lineUp.HeroId, selfBattleData)

	selfRankItem := rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, pl.Id)

	selfBattleData.PlayerInfo = global.GetPlayerInfo(pl.Id).ToCommonPlayer()
	selfBattleData.PlayerInfo.Score = selfRankItem.Score
	selfBattleData.PlayerInfo.ServerId = int32(pl.GetProp(define.PlayerPropServerId))

	res.Data[pl.Id] = selfBattleData
	res.Code = proto_public.CommonErrorCode_ERR_OK

	//任务
	task.Dispatch(ctx, pl, define.TaskJingjichangChallengeTime, 1, 0, true)

	ctx.Send(res)
}

// 战斗回调
func BattleBack_Arena(ctx global.IPlayer, pl *model.Player, data interface{}) {
	Imodel := data.(model.BattleReportBack_Arena)
	resq := Imodel.Data.(*proto_game.C2SChallengeBattleReport)

	//如果赢了 交换排名
	confs := config.ActArena.All()
	conf := conf.ActArena{}
	for _, v := range confs {
		conf = v
		break
	}
	log.Debug("战斗回调:%v", Imodel.ActCId)
	if conf.Id <= 0 {
		global.BattleBackPlayer(ctx, pl, resq, nil)
		return
	}

	_, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), Imodel.ActId, resq)
	if err != nil {
		log.Debug("load no resp: %v", err)
		global.BattleBackPlayer(ctx, pl, resq, nil)
		return
	}

	isSuc := resq.WinId == pl.Id
	_award := conf.Awards
	if isSuc {
		//掉落奖励
		award := bag.GetDrop(conf.DropAwardId, 0)
		_award = append(_award, award...)

		rankKey := fmt.Sprintf("%s:%d", define.RankTypeArenaKey, Imodel.ActId)
		targetId := Imodel.PlayerId

		selfRankItem := rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, pl.Id)
		selfScore := selfRankItem.Score
		selfInRank := selfRankItem.Rank > 0

		targetRankItem := rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, targetId)
		targetScore := targetRankItem.Score
		targetInRank := targetRankItem.Rank > 0

		if !targetInRank {
			log.Error("对手不在榜上，无法交换排名: targetId=%d", targetId)
			global.BattleBackPlayer(ctx, pl, resq, nil)
			return
		}

		// 获取排名
		var selfRank, targetRank int64
		selfRank = selfRankItem.Rank
		targetRank = targetRankItem.Rank

		if !selfInRank && Imodel.Fuchou {
			// 自己不在榜上，查复仇记录
			isFuchou := MarkFuchouRecord(Imodel.ActId, pl.Id, targetId, targetRank > selfRank)
			if isFuchou {
				log.Debug("复仇成功: 玩家[%d]对[%d]", pl.Id, targetId)
			}
			global.BattleBackPlayer(ctx, pl, resq, nil)
			return
		}

		// 排名检查：对手排名必须比自己高（排名数字更大/更靠后/更弱）
		if targetRank <= selfRank {
			log.Debug("对手排名更好，无法交换: 自己排名%d, 对手排名%d", selfRank, targetRank)
			global.BattleBackPlayer(ctx, pl, resq, nil)
			return
		}

		// 交换分数
		rank.UpdateArenaRank(ctx, Imodel.ActId, pl, int32(targetScore))

		_, err := db.RedisExec("ZADD", rankKey, selfScore, targetId)
		if err != nil {
			log.Error("更新对手分数失败: %v", err)
			global.BattleBackPlayer(ctx, pl, resq, nil)
			return
		}

		// 获取新排名
		selfRankItem = rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, pl.Id)
		targetRankItem = rank.GetSelfRank(pl.Cache.App.GetEnv().ID, rankKey, targetId)
		selfRank = selfRankItem.Rank
		targetRank = targetRankItem.Rank

		// 记录战报
		selfRecordKey := fmt.Sprintf("%s:%d_%d", define.RankTypeArenaRecordKey, Imodel.ActId, pl.Id)
		selfRecord := &model.BattleReportRecord_Arena{
			TargetId: targetId,
			IsAttack: true,
			ActId:    Imodel.ActId,
			Time:     utils.Now().Unix(),
			Rank:     int32(selfRank),
			IsFuchou: false,
		}
		selfRecordJson, _ := json.Marshal(selfRecord)
		db.RedisExec("LPUSH", selfRecordKey, string(selfRecordJson))
		db.RedisExec("LTRIM", selfRecordKey, 0, 99)

		//人机不进库
		if targetId > define.PlayerIdBase {
			targetRecordKey := fmt.Sprintf("%s:%d_%d", define.RankTypeArenaRecordKey, Imodel.ActId, targetId)
			targetRecord := &model.BattleReportRecord_Arena{
				TargetId: pl.Id,
				IsAttack: false,
				ActId:    Imodel.ActId,
				Time:     utils.Now().Unix(),
				Rank:     int32(targetRank),
				IsFuchou: false,
			}
			targetRecordJson, _ := json.Marshal(targetRecord)
			db.RedisExec("LPUSH", targetRecordKey, string(targetRecordJson))
			db.RedisExec("LTRIM", targetRecordKey, 0, 99)
		}
	}

	cons := global.MergeItemE(_award)
	internal.AddItems(ctx, pl, cons, true)

	log.Debug("竞技场，战报回来，推送变化")
	global.BattleBackPlayer(ctx, pl, resq, global.ItemFormat(_award))
}

// C2SArenaBattleRecord 竞技记录
func ReqActivityArenaBattleRecord(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SArenaReqRecord) {
	res := new(proto_activity.S2CArenaReqRecord)
	res.Options = make([]*proto_activity.ArenaRecordOption, 0)

	// 构建 Redis Key
	rankRecordKey := fmt.Sprintf("%s:%d_%d", define.RankTypeArenaRecordKey, req.ActId, pl.Id)

	// 从 Redis 读取战报记录 (LRANGE 0 -1 获取所有记录)
	reply, err := db.RedisExec("LRANGE", rankRecordKey, 0, -1)
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
		battleRecord := new(model.BattleReportRecord_Arena)
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
		option := new(proto_activity.ArenaRecordOption)
		option.Id = battleRecord.TargetId
		option.IsAttack = battleRecord.IsAttack
		option.PlayerInfo = targetPlayerInfo
		option.Rank = battleRecord.Rank
		option.Time = battleRecord.Time
		option.ActId = battleRecord.ActId
		option.IsFuchou = battleRecord.IsFuchou

		res.Options = append(res.Options, option)
	}

	ctx.Send(res)
}

// MarkFuchouRecord 查询战报记录中是否存在对定目标的记录，并标记为复仇
// 返回值: true 为找到并标记成功， false 为未找到
func MarkFuchouRecord(actId int64, playerId int64, targetId int64, isExchange bool) bool {
	if !isExchange {
		return false
	}

	// 构建 Redis Key: rank_arena_record:活动ID_玩家ID
	rankRecordKey := fmt.Sprintf("%s:%d_%d", define.RankTypeArenaRecordKey, actId, playerId)

	// 从 Redis 读取战报记录 (LRANGE 0 -1 获取所有记录)
	reply, err := db.RedisExec("LRANGE", rankRecordKey, 0, -1)
	if err != nil {
		log.Error("MarkFuchouRecord load record error:%v", err)
		return false
	}

	// 处理 Redis 返回为空的情况
	if reply == nil {
		log.Debug("样战记录户为空: playerId=%d, actId=%d", playerId, actId)
		return false
	}

	// 反序列化 JSON 数据，查找是否存在对该目标的战报
	recordItems := reply.([]interface{})
	found := false
	var foundIndex int

	for i, item := range recordItems {
		battleRecord := new(model.BattleReportRecord_Arena)
		err := json.Unmarshal(item.([]byte), battleRecord)
		if err != nil {
			log.Error("unmarshal battle record error:%v", err)
			continue
		}

		// 查找是否是被该目标攻击的记录 (AttackerId == targetId && IsAttack == false)
		if battleRecord.TargetId == targetId && !battleRecord.IsAttack {
			found = true
			foundIndex = i
			log.Debug("找到样战记录: %v", battleRecord)
			break
		}
	}

	if !found {
		log.Debug("未找到对手[%d]的战报记录, playerId=%d", targetId, playerId)
		return false
	}

	// 找到了，需要将该记录的 Fuchou 标记为 true
	// 1. 获取存储的战报
	item := recordItems[foundIndex].([]byte)
	battleRecord := new(model.BattleReportRecord_Arena)
	err = json.Unmarshal(item, battleRecord)
	if err != nil {
		log.Error("unmarshal battle record for update error:%v", err)
		return false
	}

	// 2. 标记为复仇
	battleRecord.IsFuchou = true

	// 3. 重新序列化
	updatedJson, err := json.Marshal(battleRecord)
	if err != nil {
		log.Error("marshal battle record error:%v", err)
		return false
	}

	// 4. 更新 Redis中的记录
	// 使用 LSET 根据索引过程中不改变 List 的位置
	_, err = db.RedisExec("LSET", rankRecordKey, foundIndex, string(updatedJson))
	if err != nil {
		log.Error("update battle record in redis error:%v", err)
		return false
	}

	log.Debug("复仇标记成功: playerId=%d, targetId=%d, recordIndex=%d", playerId, targetId, foundIndex)
	return true
}
