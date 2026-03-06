package activity

import (
	"xfx/core/config/conf"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
)

// ReqActivityStatus 获取活动状态列表
func ReqActivityStatus(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SActivityStatus) {
	reply, err := invoke.ActivityClient(ctx).GetActivityStatus()
	if err != nil {
		log.Error("ReqActivityStatus invoke activity error:%v", err)
		return
	}

	if reply == nil {
		log.Error("ReqActivityStatus reply is nil")
		return
	}

	resp := new(proto_activity.S2CActivityStatus)
	resp.Info = reply
	ctx.Send(resp)
}

// ReqActivityData  获取活动数据
func ReqActivityData(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SActivityData) {
	if req.ActivityId == 0 {
		log.Error("ReqActivityData id error")
		return
	}

	data, err := invoke.ActivityClient(ctx).GetActivityData(pl.ToContext(), req.ActivityId)
	if err != nil {
		log.Error("ReqActivityData invoke activity error:%v", err)
		return
	}

	if data == nil {
		log.Error("ReqActivityData reply is nil")
		return
	}

	resp := new(proto_activity.S2CActivityData)
	resp.Data = data
	ctx.Send(resp)
}

// ReqActivityDataList  获取多个活动数据
func ReqActivityDataList(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SActivityDataList) {
	if len(req.Ids) == 0 {
		log.Error("ReqActivityDataList id error")
		return
	}

	data := invoke.ActivityClient(ctx).GetActivityDataList(pl.ToContext(), req.Ids)
	if data == nil {
		log.Error("ReqActivityDataList reply is nil")
		return
	}

	resp := new(proto_activity.S2CActivityDataList)
	resp.List = data
	ctx.Send(resp)
}

// ReqGetActivityAward  领取活动奖励
func ReqGetActivityAward(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SActivityAward) {
	resp := new(proto_activity.S2CActivityAward)
	if req.ActivityId == 0 {
		log.Error("ReqGetActivityAward id error")
		resp.Success = false
		ctx.Send(resp)
		return
	}

	reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActivityId, req)
	if err != nil {
		log.Error("ReqGetActivityAward invoke activity error:%v", err)
		resp.Success = false
		ctx.Send(resp)
		return
	}

	if reply == nil {
		log.Error("ReqGetActivityAward reply is nil:%v", reply)
		resp.Success = false
		ctx.Send(resp)
		return
	}

	//获取奖励结构
	awards := reply.([]conf.ItemE)

	if len(awards) > 0 {
		bag.AddAward(ctx, pl, awards, true)
	}

	resp.Success = true
	ctx.Send(resp)
}

// PushActivityData 活动数据推送
func PushActivityData(ctx global.IPlayer, pl *model.Player, activityId int64) {
	//同步数据变化
	data, err := invoke.ActivityClient(ctx).GetActivityData(pl.ToContext(), activityId)
	if err != nil {
		log.Error("ReqGetActivityAward invoke activity error:%v", err)
		return
	}

	if data == nil {
		log.Error("ReqGetActivityAward reply is nil")
		return
	}

	ctx.Send(&proto_activity.PushActivityDataChange{
		Data: data,
	})
}

// ReqActivityBuy  TODO:活动购买 活动内部没有接入
func ReqActivityBuy(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2SActivityBuy) {
	if req.ActivityId == 0 {
		log.Error("ReqActivityBuy id error")
		return
	}

	reply, err := invoke.ActivityClient(ctx).OnRouterMsg(pl.ToContext(), req.ActivityId, req)
	if err != nil {
		log.Error("ReqActivityBuy invoke activity error:%v", err)
		return
	}

	if reply == nil {
		log.Error("ReqActivityBuy reply is nil")
		return
	}

	resp := new(proto_activity.S2CActivityDataList)
	resp.List = reply.([]*proto_activity.ActivityData)
	ctx.Send(resp)
}
