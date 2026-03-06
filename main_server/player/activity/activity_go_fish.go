package activity

import (
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	Proto_Public "xfx/proto/proto_public"
)

// ReqActivityGoFish 钓鱼
func ReqActivityGoFish(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SGoFish) {
	resp := new(proto_activity.S2CGoFish)

	//判断活动是否开启
	reply, err := invoke.ActivityClient(ctx).GetActivityData(pl.ToContext(), req.ActId)
	if err != nil {
		log.Error("ReqActivityGoFish invoke activity error:%v", err)
		resp.Code = Proto_Public.CommonErrorCode_ERR_ACTIVITYCLOSE
		ctx.Send(resp)
		return
	}

	if reply == nil {
		log.Error("ReqActivityGoFish reply is nil")
		resp.Code = Proto_Public.CommonErrorCode_ERR_ACTIVITYCLOSE
		ctx.Send(resp)
		return
	}

	//判断鱼饵
	activityConfs := config.ActGoFish.All()
	var _conf conf.ActGoFish
	for _, v := range activityConfs {
		if v.Type == req.PoolType {
			_conf = v
			break
		}
	}

	if _conf.Id <= 0 {
		resp.Code = Proto_Public.CommonErrorCode_ERR_NoConfig
		ctx.Send(resp)
		return
	}

	//判断最低类型
	if _conf.NeedMinCost > 0 {
		if req.CostId != _conf.NeedMinCost {
			resp.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
			ctx.Send(resp)
			return
		}
	}

	cost := make(map[int32]int32)
	cost[req.CostId] = 1

	//判断材料够不够
	if !internal.CheckItemsEnough(pl, cost) {
		resp.Code = Proto_Public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(resp)
		return
	}

	_reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActId, req)
	if err != nil {
		log.Error("ReqActivityGoFish invoke activity error:%v", err)
		resp.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	if _reply == nil {
		log.Error("ReqActivityGoFish reply is nil")
		resp.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	_resp := _reply.(*model.GoFishBack)
	//没有鱼了
	if _resp.Code == 1 {
		log.Debug("ReqActivityGoFish is code :%v", _resp.Code)
		resp.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	//扣除材料
	internal.SubItems(ctx, pl, cost)
	if _resp.Code == 2 {
		resp.Code = Proto_Public.CommonErrorCode_ERR_OK
		resp.State = false
		ctx.Send(resp)
		return
	}

	resp.Code = Proto_Public.CommonErrorCode_ERR_OK
	resp.State = true

	if len(_resp.Ids) > 0 {
		for _, v := range _resp.Ids {
			award := []conf.ItemE{{
				ItemType: define.ItemTypeFish,
				ItemId:   v,
				ItemNum:  1,
			}}

			resp.Awards = append(resp.Awards, global.ItemFormat(award)...)
		}
	}
	log.Info("钓鱼回调: %v", resp.Awards)
	ctx.Send(resp)
}

// ReqActivityFishSign  活动签到
func ReqActivityFishSign(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SFishSign) {
	resp := &proto_activity.S2CFishSign{}
	if req.ActId == 0 {
		log.Error("ReqActivitySign id error")
		resp.Code = Proto_Public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(resp)
		return
	}

	reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActId, req)
	if err != nil {
		log.Error("ReqActivitySign invoke activity error:%v", err)
		resp.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	if reply == nil {
		log.Error("ReqActivitySign reply is nil")
		resp.Code = Proto_Public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(resp)
		return
	}

	back := reply.(*model.CommonActivityAwardBack)
	if back.Code == 1 {
		resp.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	//奖励
	bag.AddAward(ctx, pl, back.Award, true)

	resp.Code = Proto_Public.CommonErrorCode_ERR_OK
	ctx.Send(resp)
}

// ReqActivityFishLevelAward  钓鱼等级
func ReqActivityFishLevelAward(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SFishLevelAward) {
	resp := &proto_activity.S2CFishLevelAward{}
	if req.ActId == 0 {
		log.Error("ReqActivityFishLevelAward id error")
		resp.Code = Proto_Public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(resp)
		return
	}

	reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActId, req)
	if err != nil {
		log.Error("ReqActivityFishLevelAward invoke activity error:%v", err)
		resp.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	if reply == nil {
		log.Error("ReqActivityFishLevelAward reply is nil")
		resp.Code = Proto_Public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(resp)
		return
	}

	back := reply.(*model.CommonActivityAwardBack)
	if back.Code == 1 || back.Code == 2 {
		resp.Code = Proto_Public.CommonErrorCode_ERR_OK
		ctx.Send(resp)
		return
	}

	if len(back.Award) > 0 {
		//奖励
		bag.AddAward(ctx, pl, back.Award, true)
	}

	resp.Code = Proto_Public.CommonErrorCode_ERR_OK
	ctx.Send(resp)
}
