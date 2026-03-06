package lineup

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_lineup"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.Lineup = new(model.LineUp)
	pl.Lineup.LineUps = make(map[int32]*model.LineUpOption)

	pl.Lineup.LineUps[define.LINEUP_STAGE] = &model.LineUpOption{
		Type:   define.LINEUP_STAGE,
		HeroId: []int32{0, 0, 0, 0, 3001, 0, 0, 0, 0},
	}
	pl.Lineup.LineUps[define.LINEUP_DANAOTIANGONG] = &model.LineUpOption{
		Type:   define.LINEUP_DANAOTIANGONG,
		HeroId: []int32{0, 0, 0, 0, 3001, 0, 0, 0, 0},
	}
	pl.Lineup.LineUps[define.LINEUP_CLIMBTOWER] = &model.LineUpOption{
		Type:   define.LINEUP_CLIMBTOWER,
		HeroId: []int32{0, 0, 0, 0, 3001, 0, 0, 0, 0},
	}
	pl.Lineup.LineUps[define.LINEUP_TheCompetition] = &model.LineUpOption{
		Type:   define.LINEUP_TheCompetition,
		HeroId: []int32{0, 0, 0, 0, 3001, 0, 0, 0, 0},
	}

	pl.Lineup.LineUps[define.LINEUP_ARENA] = &model.LineUpOption{
		Type:   define.LINEUP_ARENA,
		HeroId: []int32{0, 0, 0, 0, 3001, 0, 0, 0, 0},
	}
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Lineup)
	if err != nil {
		log.Error("player[%v],save card marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("Load lineup error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerLineUp, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load lineup error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerLineUp, pl.Id))
	if err != nil {
		log.Error("player[%v],load card error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.LineUp)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load lineup unmarshal error:%v", pl.Id, err)
	}

	pl.Lineup = m
}

// 设置布阵
func ReqSetLineUp(ctx global.IPlayer, pl *model.Player, req *proto_lineup.C2SSetLineUp) {
	res := new(proto_lineup.S2CSetLineUp)
	//判读是不是只是换了位置，还是角色都变了
	var isPositionOnly bool
	if oldLineUp, ok := pl.Lineup.LineUps[req.Type]; ok {
		// 旧布阵存在，比较角色集合是否相同
		isPositionOnly = compareLineUpHeros(oldLineUp.HeroId, req.HeroId)
	} else {
		// 旧布阵不存在，当前是新增布阵
		isPositionOnly = false
	}

	suc := internal.UpdateLineUp(ctx, pl, req.Type, req.HeroId)
	if !suc {
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		return
	}

	// 竞技场活动运行中，检查是否是竞技场布阵
	if req.Type == define.LINEUP_ARENA && !isPositionOnly {
		//告知活动，布阵调整
		// 获取竞技场活动状态
		arenaActivity, errArena := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeArena)
		if errArena != nil {
			log.Error("ReqSetLineUp get arena activity error:%v", errArena)
		} else if arenaActivity != nil && arenaActivity.ActivityId > 0 {
			_, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), arenaActivity.ActivityId, req)
			if err != nil {
				log.Error("ReqSetLineUp notify arena activity error:%v", err)
			}
		}
	}

	// 天梯活动运行中，检查是否是天梯布阵
	if req.Type == define.LINEUP_Tianti && !isPositionOnly {
		// 获取天梯活动状态
		ladderActivity, errLadder := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeLadderRace)
		if errLadder != nil {
			log.Error("ReqSetLineUp get ladder activity error:%v", errLadder)
		} else if ladderActivity != nil && ladderActivity.ActivityId > 0 {
			_, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), ladderActivity.ActivityId, req)
			if err != nil {
				log.Error("ReqSetLineUp notify ladder activity error:%v", err)
			}
		}
	}

	res.Code = proto_public.CommonErrorCode_ERR_OK
	res.Cards = model.ToLineUpProto(pl.Lineup.LineUps)
	ctx.Send(res)
}

// compareLineUpHeros 比较两个布阵的角色集合是否相同（忽略顺序）
func compareLineUpHeros(oldHeros, newHeros []int32) bool {
	// 创建旧布阵的角色集合
	oldSet := make(map[int32]bool)
	for _, h := range oldHeros {
		if h > 0 {
			oldSet[h] = true
		}
	}

	// 创建新布阵的角色集合
	newSet := make(map[int32]bool)
	for _, h := range newHeros {
		if h > 0 {
			newSet[h] = true
		}
	}

	// 比较两个集合是否相同
	if len(oldSet) != len(newSet) {
		return false
	}

	for h := range oldSet {
		if !newSet[h] {
			return false
		}
	}

	return true
}

// 请求初始布阵
func ReqInitLineUp(ctx global.IPlayer, pl *model.Player, req *proto_lineup.C2SInitLineUp) {
	res := new(proto_lineup.S2CInitLineUp)

	cardmap := make(map[int32]*proto_lineup.CardLineUpMap, len(pl.Lineup.LineUps))
	playerHeroID := int32(pl.GetProp(define.PlayerPropHeroId))

	for k, v := range pl.Lineup.LineUps {
		heroIds := make([]int32, len(v.HeroId))
		copy(heroIds, v.HeroId)

		// 优化循环逻辑
		for i := range heroIds {
			if heroIds[i] >= 3001 && heroIds[i] <= 3004 && heroIds[i] != playerHeroID {
				heroIds[i] = playerHeroID
				break
			}
		}

		cardmap[k] = &proto_lineup.CardLineUpMap{
			Type:   v.Type,
			HeroId: heroIds,
		}

		// 直接更新原始数据
		v.HeroId = heroIds
	}

	res.Cards = cardmap
	ctx.Send(res)
}

// 无损替换设置布阵
func ReqReplaceSetLineUp(ctx global.IPlayer, pl *model.Player, req *proto_lineup.C2SSetLineUpAndReplace) {
	res := new(proto_lineup.S2CSetLineUpAndReplace)

	//判断有没有相同的布阵
	tempIds := make(map[int32]int32)
	for _, v := range req.HeroId {
		if _, ok := tempIds[v]; ok {
			res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(res)
			return
		}

		tempIds[v] = 1
	}

	ishas := false
	//判断有没有角色
	for _, v := range req.HeroId {
		if v == 0 {
			continue
		}

		if v == req.TId {
			ishas = true
		}

		if _, ok := pl.Hero.Hero[v]; !ok {
			res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
			ctx.Send(res)
			return
		}

		//判断角色阵位对不对
		//confs := config.CfgMgr.AllJson["Hero"].(map[int64]conf2.Hero)
		//conf := confs[int64(v)]

		//if common.IsHaveValueIntArray(conf.State, int32(index)) == false {
		//	res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		//	ctx.Send(res)
		//	return
		//}
	}

	if !ishas {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
	}

	if _, ok := pl.Hero.Hero[req.SId]; !ok {
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		ctx.Send(res)
		return
	}

	//替换
	sLevel := pl.Hero.Hero[req.SId].Level
	sStage := pl.Hero.Hero[req.SId].Stage
	pl.Hero.Hero[req.SId].Level = pl.Hero.Hero[req.TId].Level
	pl.Hero.Hero[req.SId].Stage = pl.Hero.Hero[req.TId].Stage
	pl.Hero.Hero[req.TId].Level = sLevel
	pl.Hero.Hero[req.TId].Stage = sStage

	//同步
	internal.SyncHeroChange(ctx, pl, req.SId)
	internal.SyncHeroChange(ctx, pl, req.TId)

	pl.Lineup.LineUps[req.Type] = &model.LineUpOption{
		Type:   req.Type,
		HeroId: req.HeroId,
	}

	//藏品槽位角色要更新下
	if req.Type == define.LINEUP_STAGE {
		internal.CollectionSlotHeroChange(ctx, pl)
	}

	res.Code = proto_public.CommonErrorCode_ERR_OK
	res.Cards = model.ToLineUpProto(pl.Lineup.LineUps)
	ctx.Send(res)
}
