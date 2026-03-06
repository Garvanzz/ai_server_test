package activity

import (
	"xfx/core/config/conf"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_public"
)

// ReqGetActivityFundAward  领取基金活动奖励
func ReqGetActivityFundAward(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SGetFundAward) {
	resp := new(proto_activity.S2CGetFundAward)
	if req.Type == 0 {
		log.Error("ReqGetActivityFundAward id error")
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActId, req)
	if err != nil {
		log.Error("ReqGetActivityFundAward invoke activity error:%v", err)
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	if reply == nil {
		log.Error("ReqGetActivityFundAward reply is nil:%v", reply)
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	//获取奖励结构
	awards := reply.([]conf.ItemE)

	if len(awards) > 0 {
		bag.AddAward(ctx, pl, awards, true)
	}

	resp.Code = proto_public.CommonErrorCode_ERR_OK
	ctx.Send(resp)
}
