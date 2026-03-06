package stage

import (
	"encoding/json"
	"fmt"
	"strconv"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/messages"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/proto/proto_game"
	"xfx/proto/proto_public"
	"xfx/proto/proto_stage"
)

func Init(pl *model.Player) {
	pl.Stage = new(model.Stage)
	pl.Stage.Stage = make(map[int32]map[int32]*model.ChapterOpt)

	//初始关卡
	pl.Stage.CurStage = 10001
	pl.Stage.CurChapter = 1
	pl.Stage.CurCycle = 1
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Stage)
	if err != nil {
		log.Error("player[%v],save stage marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save stage error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerStage, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save stage error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerStage, pl.Id))
	if err != nil {
		log.Error("player[%v],load stage error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Stage)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load stage unmarshal error:%v", pl.Id, err)
	}

	pl.Stage = m
}

// ReqStageList 请求关卡列表
func ReqStageList(ctx global.IPlayer, pl *model.Player, req *proto_stage.C2SInitStage) {
	maps := model.ToStageProto(pl.Stage.Stage)
	ctx.Send(&proto_stage.S2CInitStage{
		Id:         maps,
		CurChapter: pl.Stage.CurChapter,
		CurStageId: pl.Stage.CurStage,
		CurCycle:   pl.Stage.CurCycle,
	})
}

// 击杀
func ReqStageKillEnemy(ctx global.IPlayer, pl *model.Player, req *proto_stage.C2SKillEnemy) {
	res := &proto_stage.S2CKillEnemy{}

	if req.CurCycle > pl.Stage.CurCycle {
		res.Code = proto_stage.ERRORSTAGECODE_ERROR_NoMatchStageErr
		ctx.Send(res)
		return
	}

	confs := config.Stage.All()
	maxChapter, maxStage := pl.Stage.GetMaxChapterStage(req.CurCycle)
	nextNewChapter := false
	//判断是不是最后一关，要自动进去下一关
	for _, v := range confs {
		if v.Chapter == maxChapter && v.Id == maxStage && v.Next == 0 {
			isPass := pl.Stage.GetIsPass(req.CurCycle, maxChapter, maxStage)
			if isPass {
				nextNewChapter = true
			}
			break
		}
	}

	if !nextNewChapter {
		if req.CurChapter > maxChapter {
			res.Code = proto_stage.ERRORSTAGECODE_ERROR_NoMatchStageErr
			ctx.Send(res)
			return
		}
	}

	conf := confs[int64(req.StageId)]
	//判断前置是否完成
	if conf.Front > 0 {
		isPass := pl.Stage.GetIsPass(req.CurCycle, req.CurChapter, conf.Front)
		if !isPass {
			res.Code = proto_stage.ERRORSTAGECODE_ERROR_NoMatchStageErr
			ctx.Send(res)
			return
		}
	}

	stageMonsterConfs := config.StageMonsterGroup.All()
	chapterConfs := config.Chapter.All()
	chapterConf := chapterConfs[int64(req.CurChapter)]

	if _, ok := pl.Stage.Stage[req.CurCycle]; !ok {
		pl.Stage.Stage[req.CurCycle] = make(map[int32]*model.ChapterOpt)
	}

	if _, ok := pl.Stage.Stage[req.CurCycle][req.CurChapter]; !ok {
		pl.Stage.Stage[req.CurCycle][req.CurChapter] = new(model.ChapterOpt)
	}

	if pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages == nil {
		pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages = make(map[int32]*model.StageOpt)
	}

	if _, ok := pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.StageId]; !ok {
		pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.StageId] = new(model.StageOpt)
		pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.StageId].Id = req.StageId
	}

	stageData := pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.StageId]
	awards := make([]conf2.ItemE, 0)
	monsterConf := stageMonsterConfs[int64(chapterConf.Monster)]
	pl.Stage.CurStage = req.StageId
	pl.Stage.CurChapter = req.CurChapter
	//通关了 要给奖励
	if !stageData.Pass {
		//满经验
		if stageData.Exp >= conf.StageExp {
			if conf.Boss > 0 {
				return
			} else {
				stageData.Pass = true
				pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.StageId] = stageData

				if conf.Next == 0 {
					if chapterConf.Next > 0 {
						pl.Stage.CurChapter = chapterConf.Next
						//再次获取配置
						_chapterConf := chapterConfs[int64(chapterConf.Next)]
						pl.Stage.CurStage = _chapterConf.StartStage
					}

					//周目
					if chapterConf.Next <= 0 && len(chapterConfs) <= len(pl.Stage.Stage[pl.Stage.CurCycle]) {
						pl.Stage.CurCycle += 1
					}
				} else {
					pl.Stage.CurStage = conf.Next
				}

				//任务
				task.Dispatch(ctx, pl, define.TaskMainLinePassStage, pl.Stage.CurStage, 0, false)

				//通关奖励
				if len(conf.StageAward) > 0 {
					cons := global.MergeItemE(conf.StageAward)
					internal.AddItems(ctx, pl, cons, false)
					awards = conf.StageAward
				}
			}
		} else {
			//经验
			getExp := monsterConf.KillExp[req.Id]
			stageData.Exp += getExp
			pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.StageId] = stageData

			if stageData.Exp >= conf.StageExp {
				stageData.Exp = conf.StageExp
				if conf.Boss <= 0 {
					stageData.Pass = true
					pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.StageId] = stageData

					if conf.Next == 0 {
						if chapterConf.Next > 0 {
							pl.Stage.CurChapter = chapterConf.Next
							//再次获取配置
							_chapterConf := chapterConfs[int64(chapterConf.Next)]
							pl.Stage.CurStage = _chapterConf.StartStage
						}

						//周目
						if chapterConf.Next <= 0 && len(chapterConfs) <= len(pl.Stage.Stage[pl.Stage.CurCycle]) {
							pl.Stage.CurCycle += 1
						}
					} else {
						pl.Stage.CurStage = conf.Next
					}

					//任务
					task.Dispatch(ctx, pl, define.TaskMainLinePassStage, pl.Stage.CurStage, 0, false)

					//通关奖励
					if len(conf.StageAward) > 0 {
						cons := global.MergeItemE(conf.StageAward)
						internal.AddItems(ctx, pl, cons, false)
						awards = conf.StageAward
					}
				} else {
					pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.StageId] = stageData
				}
			}
		}
	}

	//奖励
	if len(monsterConf.KillAward) > 0 {
		if req.Index > int32(len(monsterConf.KillAward)) {
			StageKillEnemyGetAward(ctx, pl, monsterConf.KillAward[int32(len(monsterConf.KillAward))-1])
		} else {
			StageKillEnemyGetAward(ctx, pl, monsterConf.KillAward[req.Index-1])
		}
	}

	//任务
	task.Dispatch(ctx, pl, define.TaskKillStageMonster, 1, 0, true)

	PushStage(ctx, pl, req.CurCycle, req.CurChapter, req.StageId, stageData, awards, []conf2.ItemE{}, false)
	res.Code = proto_stage.ERRORSTAGECODE_ERROR_Ok
	ctx.Send(res)
}

// 击杀敌人获取奖励
func StageKillEnemyGetAward(ctx global.IPlayer, pl *model.Player, award []int32) {
	awards := make([]conf2.ItemE, 0)
	for l := 0; l < len(award); l++ {
		if award[0] == define.ItemTypeItem {
			awards = append(awards, conf2.ItemE{
				ItemId:   award[1],
				ItemNum:  award[2],
				ItemType: award[0],
			})
		}
	}

	if len(awards) > 0 {
		bag.AddAward(ctx, pl, awards, false)
	}
}

// 推送关卡信息
func PushStage(ctx global.IPlayer, pl *model.Player, cycle int32, chapter int32, stage int32, opt *model.StageOpt, award []conf2.ItemE, bossAward []conf2.ItemE, killBoss bool) {
	mapres := model.ToStageSingleProto(cycle, chapter, opt)
	res := &proto_stage.PushChange{
		Id:         mapres,
		CurStageId: pl.Stage.CurStage,
		CurChapter: pl.Stage.CurChapter,
		CurCycle:   pl.Stage.CurCycle,
		Awards:     global.ItemFormat(award),
		KillBoss:   killBoss,
		BossAwards: global.ItemFormat(bossAward),
	}
	ctx.Send(res)
}

// 请求挑战关卡boss
func ReqStageBossBattleChallenge(ctx global.IPlayer, pl *model.Player, req *proto_stage.C2SChallengeStageBossBattle) {
	res := new(proto_stage.S2CChallengeStageBattle)

	if req.CurCycle > pl.Stage.CurCycle {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	confs := config.Stage.All()
	maxChapter, maxStage := pl.Stage.GetMaxChapterStage(req.CurCycle)
	nextNewChapter := false
	//判断是不是最后一关，要自动进去下一关
	for _, v := range confs {
		if v.Chapter == maxChapter && v.Id == maxStage && v.Next == 0 {
			isPass := pl.Stage.GetIsPass(req.CurCycle, maxChapter, maxStage)
			if isPass {
				nextNewChapter = true
			}
			break
		}
	}

	if !nextNewChapter {
		if req.CurChapter > maxChapter {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}
	}

	conf := confs[int64(req.CurStageId)]
	//判断前置是否完成
	if conf.Front > 0 {
		isPass := pl.Stage.GetIsPass(req.CurCycle, req.CurChapter, conf.Front)
		if !isPass {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}
	}

	if _, ok := pl.Stage.Stage[req.CurCycle]; !ok {
		log.Debug("该关卡没有数据")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Stage.Stage[req.CurCycle][req.CurChapter]; !ok {
		log.Debug("该关卡没有数据章节")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages == nil {
		pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages = make(map[int32]*model.StageOpt)
	}

	if _, ok := pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.CurStageId]; !ok {
		pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.CurStageId] = new(model.StageOpt)
		pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.CurStageId].Id = req.CurStageId
	}

	//判断经验
	stageData := pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages[req.CurStageId]
	//没有通关
	if stageData.Pass == true {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	//满经验
	if stageData.Exp < conf.StageExp {
		log.Debug("经验不足:%v", stageData.Id)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if conf.Boss <= 0 {
		log.Debug("该关卡没有boss:%v", stageData.Id)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
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

	battleId, err := invoke.BattleClient(ctx).BattleStageBoss(pl.ToContext(), req.CurCycle, req.CurStageId, req.CurChapter)
	if err != nil {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		log.Debug("load battle stageBoss no resp: %v", err)
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
	res.CurChapter = req.CurChapter
	res.CurStageId = req.CurStageId
	res.CurCycle = req.CurCycle

	//获取战斗数据
	batData := internal.GetBattleSelfPlayerData(pl, lineUp.HeroId)
	res.Data = batData
	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

// 战斗回调
func BattleBack_StageBoss(ctx global.IPlayer, pl *model.Player, data interface{}) {
	Imodel := data.(model.BattleReportBack_StageBoss)
	resq := Imodel.Data.(*proto_game.C2SChallengeBattleReport)

	confs := config.Stage.All()
	conf := confs[int64(Imodel.Stage)]
	if conf.Boss <= 0 {
		return
	}

	if _, ok := pl.Stage.Stage[Imodel.Cycle]; !ok {
		return
	}

	if _, ok := pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter]; !ok {
		return
	}

	if pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter].Stages == nil {
		pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter].Stages = make(map[int32]*model.StageOpt)
	}

	if _, ok := pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter].Stages[Imodel.Stage]; !ok {
		pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter].Stages[Imodel.Stage] = new(model.StageOpt)
		pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter].Stages[Imodel.Stage].Id = Imodel.Stage
	}

	stageData := pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter].Stages[Imodel.Stage]
	awards := make([]conf2.ItemE, 0)
	bossAwards := make([]conf2.ItemE, 0)
	chapterConfs := config.Chapter.All()
	chapterConf := chapterConfs[int64(Imodel.Chapter)]
	if !stageData.Pass {
		if resq.WinId != pl.Id {
			stageData.PassState = 1
			log.Error("主关卡boss战斗失败")
			stageData.Pass = false
			pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter].Stages[Imodel.Stage] = stageData
			//推送变化
			PushStage(ctx, pl, Imodel.Cycle, Imodel.Chapter, Imodel.Stage, stageData, []conf2.ItemE{}, nil, true)
			return
		}

		stageData.Pass = true
		stageData.PassState = 0
		pl.Stage.Stage[Imodel.Cycle][Imodel.Chapter].Stages[Imodel.Stage] = stageData

		if conf.Next == 0 {
			if chapterConf.Next > 0 {
				pl.Stage.CurChapter = chapterConf.Next
				//再次获取配置
				_chapterConf := chapterConfs[int64(chapterConf.Next)]
				pl.Stage.CurStage = _chapterConf.StartStage
			}

			//周目
			if chapterConf.Next <= 0 && len(chapterConfs) <= len(pl.Stage.Stage[pl.Stage.CurCycle]) {
				pl.Stage.CurCycle += 1
			}
		} else {
			pl.Stage.CurStage = conf.Next
		}

		//任务
		task.Dispatch(ctx, pl, define.TaskMainLinePassStage, pl.Stage.CurStage, 0, false)

		log.Debug("主关卡boss，战报回来，推送变化")
		//通关奖励
		if len(conf.StageAward) > 0 {
			_award := conf.StageAward
			cons := global.MergeItemE(_award)
			internal.AddItems(ctx, pl, cons, false)
			awards = _award
		}

		//掉落奖励
		if conf.BossDropAward > 0 {
			_award := bag.GetDrop(conf.BossDropAward, 0)
			cons := global.MergeItemE(_award)
			internal.AddItems(ctx, pl, cons, false)
			awards = append(awards, _award...)
		}

		//boss奖励
		if len(conf.BossAward) > 0 {
			_award := conf.BossAward
			cons := global.MergeItemE(_award)
			internal.AddItems(ctx, pl, cons, false)
			bossAwards = append(bossAwards, _award...)
		}
	} else {
		if resq.WinId != pl.Id {
			return
		}

		if conf.Next == 0 {
			if chapterConf.Next > 0 {
				pl.Stage.CurChapter = chapterConf.Next
				//再次获取配置
				_chapterConf := chapterConfs[int64(chapterConf.Next)]
				pl.Stage.CurStage = _chapterConf.StartStage
			}

			//周目
			if chapterConf.Next <= 0 && len(chapterConfs) <= len(pl.Stage.Stage[pl.Stage.CurCycle]) {
				pl.Stage.CurCycle += 1
			}
		} else {
			pl.Stage.CurStage = conf.Next
		}

		//任务
		task.Dispatch(ctx, pl, define.TaskMainLinePassStage, pl.Stage.CurStage, 0, false)

		log.Debug("主关卡boss，战报回来，推送变化")
		//掉落奖励
		if conf.BossDropAward > 0 {
			_award := bag.GetDrop(conf.BossDropAward, 0)
			cons := global.MergeItemE(_award)
			internal.AddItems(ctx, pl, cons, false)
			bossAwards = append(bossAwards, _award...)
		}
	}
	//推送变化
	PushStage(ctx, pl, Imodel.Cycle, Imodel.Chapter, Imodel.Stage, stageData, awards, bossAwards, true)
	//通告相关
	internal.SyncNotice_CharperChange(ctx, pl, chapterConf.Id)
}

// SettleStageGame
func SettleStageGame(ctx global.IPlayer, pl *model.Player, msg *messages.StageSettle) {
	//奖励
	//confs := config.Stage.All()
	//conf := confs[int64(msg.StageId)]
	//items := make([]conf2.ItemE, 0)
	//settleItems := make([]*proto_stage.Items, 0)
	//
	////胜利
	//if msg.IsWin {
	//	//任务
	//	if msg.KillNum > 0 {
	//		task.Dispatch(ctx, pl, define.TaskKillXTimes, msg.KillNum, 0, true)
	//	}
	//
	//	if msg.Damage > 0 {
	//		task.Dispatch(ctx, pl, define.TaskTotalDamageToX, int32(msg.Damage), 0, true)
	//	}
	//
	//	task.Dispatch(ctx, pl, define.TaskWinGamesXTimes, 1, 0, true)
	//
	//	if _, ok := pl.Stage.Stage[msg.StageId]; !ok {
	//		//首通
	//		pl.Stage.Stage[msg.StageId] = &model.StageOpt{
	//			Id: msg.StageId,
	//		}
	//		items = append(items, conf.PassReward...)
	//
	//		for k := 0; k < len(conf.PassReward); k++ {
	//			settleItems = append(settleItems, &proto_stage.Items{
	//				Id:   conf.PassReward[k].ItemId,
	//				Num:  conf.PassReward[k].ItemNum,
	//				Type: conf.PassReward[k].ItemType,
	//			})
	//		}
	//
	//		//玩家经验
	//		global.PlayerUpLevel(ctx, pl, int64(conf.PlayerExp))
	//	} else {
	//		dropItem := bag.GetDrop(conf.DropReawrd)
	//		items = append(items, dropItem...)
	//
	//		for k := 0; k < len(dropItem); k++ {
	//			settleItems = append(settleItems, &proto_stage.Items{
	//				Id:   dropItem[k].ItemId,
	//				Num:  dropItem[k].ItemNum,
	//				Type: dropItem[k].ItemType,
	//			})
	//		}
	//	}
	//
	//	//添加道具
	//	bag.AddAward(ctx, pl, items)
	//}
	//
	//res := &proto_stage.S2CStageSettle{
	//	StageId:  msg.StageId,
	//	PlayerId: pl.Id,
	//	IsWin:    msg.IsWin,
	//	PvePlayer: &proto_stage.PvePlayer{
	//		Damage:          msg.Damage,
	//		WithStandDamage: msg.WithStandDamage,
	//		Heal:            msg.Heal,
	//		KillNum:         msg.KillNum,
	//	},
	//	Items: settleItems,
	//}
	//ctx.Send(res)
	//
	//if msg.IsWin {
	//	maps := make(map[int32]int32)
	//	maps[msg.StageId] = msg.StageId
	//	//更新关卡
	//	ctx.Send(&proto_stage.PushChange{
	//		Id: maps,
	//	})
	//}
}

// 获取自由玩家数据
func GetStageFreePlayer(ctx global.IPlayer, pl *model.Player, msg *proto_stage.C2SGetFreePlayer) {
	res := &proto_stage.S2CGetFreePlayer{}
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save stage error, no this server:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_MYSQLERROR
		ctx.Send(res)
		return
	}

	count := msg.Count
	if count <= 0 {
		count = 1
	} else if count >= 10 {
		count = 10
	}
	log.Debug("请求自由玩家:%v", count)
	//获取战力排行榜
	rdb.RedisAsyncExec(ctx.Self(), define.RedisRetStage, []int64{1, int64(count)}, "zrevrange", fmt.Sprintf("%s", define.RankTypePowerKey), 0, define.RankTop-1, "WITHSCORES")
}

// 解锁隐藏剧情
func UnlockHiddenStory(ctx global.IPlayer, pl *model.Player, req *proto_stage.C2SUnLockHiddenStory) {
	res := &proto_stage.S2CUnLockHiddenStory{}

	//判断当前是否通关
	if req.CurCycle > pl.Stage.CurCycle {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(res)
		return
	}

	//当周目相等时
	if req.CurCycle == pl.Stage.CurCycle {
		confs := config.Stage.All()
		maxChapter, maxStage := pl.Stage.GetMaxChapterStage(req.CurCycle)
		nextNewChapter := false
		//判断是不是最后一关，要自动进去下一关
		for _, v := range confs {
			if v.Chapter == maxChapter && v.Id == maxStage && v.Next == 0 {
				isPass := pl.Stage.GetIsPass(req.CurCycle, maxChapter, maxStage)
				if isPass {
					nextNewChapter = true
				}
				break
			}
		}

		if !nextNewChapter {
			if req.CurChapter > maxChapter {
				res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
				ctx.Send(res)
				return
			}
		}

		if req.CurStageId > maxStage {
			res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
			ctx.Send(res)
			return
		}
	}

	//判断隐藏剧情
	chapterConfs := config.Chapter.All()
	chapterConf := chapterConfs[int64(req.CurChapter)]

	if chapterConf.UnlockStoryType == define.UnLockHiddleStory_None {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Stage.Stage[req.CurCycle]; !ok {
		log.Debug("该关卡没有数据")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if _, ok := pl.Stage.Stage[req.CurCycle][req.CurChapter]; !ok {
		log.Debug("该关卡没有数据章节")
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	if pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages == nil {
		pl.Stage.Stage[req.CurCycle][req.CurChapter].Stages = make(map[int32]*model.StageOpt)
	}

	if pl.Stage.Stage[req.CurCycle][req.CurChapter].HiddenStory == nil {
		pl.Stage.Stage[req.CurCycle][req.CurChapter].HiddenStory = new(model.HiddenStory)
	}

	data := pl.Stage.Stage[req.CurCycle][req.CurChapter]
	data.HiddenStory.UnlockStory = true

	if chapterConf.UnlockStoryType == define.UnLockHiddleStory_Click { //点击
		data.HiddenStory.FinishStory = true
	}

	pl.Stage.Stage[req.CurCycle][req.CurChapter] = data

	res.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(res)
}

func OnRetStageData(ctx global.IPlayer, pl *model.Player, ret *db.RedisRet) {
	opt := int32(ret.Params[0])
	log.Debug("请求自由玩家redis回调")
	switch opt {
	case 1:
		resp := new(proto_stage.S2CGetFreePlayer)
		count := int32(ret.Params[1])
		res, _ := ret.Reply.([]interface{})
		var ids []int64
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			//score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			if key == pl.Id {
				continue
			}
			ids = append(ids, key)
			if int32(len(ids)) >= count {
				break
			}
		}
		log.Debug("请求自由玩家数据:%v", ids)
		if len(ids) <= 0 {
			resp.Code = proto_public.CommonErrorCode_ERR_MYSQLERROR
			ctx.Send(resp)
			return
		}

		resp.HeroInfos = make([]*proto_public.BattleHeroData, 0)
		for _, id := range ids {
			if id == pl.Id {
				continue
			}

			plCtx := global.GetPlayerInfo(id)
			//获取布阵
			lineup := global.GetPlayerLineUpInfo(id)
			if _, ok := lineup[define.LINEUP_STAGE]; !ok {
				continue
			}
			_lineup := lineup[define.LINEUP_STAGE]
			_tempLineup := make([]*proto_public.CommonPlayerLineUpItemInfo, len(_lineup.HeroId))
			for i := 0; i < len(_lineup.HeroId); i++ {
				if _lineup.HeroId[i].Id >= 3001 && _lineup.HeroId[i].Id <= 3004 {
					_tempLineup[i] = _lineup.HeroId[i]
				} else {
					_tempLineup[i] = new(proto_public.CommonPlayerLineUpItemInfo)
				}
			}
			data := global.GetBattlePlayerData(plCtx.ToToContext(), _lineup.HeroId)
			data.PlayerInfo = plCtx.ToCommonPlayer()
			resp.HeroInfos = append(resp.HeroInfos, data)
		}
		resp.Code = proto_public.CommonErrorCode_ERR_OK
		ctx.Send(resp)
	default:
	}
}
