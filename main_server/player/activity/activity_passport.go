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

// ReqPassportGetAward 领取通行证奖励
func ReqPassportGetAward(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SPassportGetAward) {
	resp := &proto_activity.S2CPassportGetAward{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 参数校验
	if req.ActId == 0 || len(req.Ids) == 0 {
		log.Error("ReqPassportGetAward param error: actId=%d, ids=%v", req.ActId, req.Ids)
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 调用活动系统处理领奖逻辑
	reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), int64(req.ActId), req)
	if err != nil {
		log.Error("ReqPassportGetAward invoke activity error: %v", err)
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	if reply == nil {
		log.Error("ReqPassportGetAward reply is nil")
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 获取奖励结构
	awards := reply.([]conf.ItemE)

	// 发放奖励到背包
	if len(awards) > 0 {
		bag.AddAward(ctx, pl, awards, true)
		log.Debug("ReqPassportGetAward success: playerId=%d, actId=%d, ids=%v, awards=%d", pl.Id, req.ActId, req.Ids, len(awards))
	}

	ctx.Send(resp)
}
