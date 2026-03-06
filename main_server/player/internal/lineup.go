package internal

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/proto/proto_lineup"
)

// 布阵更新
func UpdateLineUp(ctx global.IPlayer, pl *model.Player, typ int32, ids []int32) bool {

	//判断有没有角色
	for _, v := range ids {
		if v == 0 {
			continue
		}
		if _, ok := pl.Hero.Hero[v]; !ok {
			return false
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

	//如果是主角 要变化
	for k, v := range ids {
		if v >= 3001 && v <= 3004 {
			ids[k] = int32(pl.GetProp(define.PlayerPropHeroId))
		}
	}

	//判断有没有相同的布阵
	tempIds := make(map[int32]int32)
	for _, v := range ids {
		if v <= 0 {
			continue
		}
		if _, ok := tempIds[v]; ok {
			return false
		}

		tempIds[v] = 1
	}

	pl.Lineup.LineUps[typ] = &model.LineUpOption{
		Type:   typ,
		HeroId: ids,
	}

	//藏品槽位角色要更新下
	if typ == define.LINEUP_STAGE {
		CollectionSlotHeroChange(ctx, pl)
	}

	//更新
	res := &proto_lineup.PushChangeLineUp{}
	res.Cards = make(map[int32]*proto_lineup.CardLineUpMap)
	res.Cards[typ] = model.ToLineUpProto(pl.Lineup.LineUps)[typ]
	ctx.Send(res)
	return true
}
