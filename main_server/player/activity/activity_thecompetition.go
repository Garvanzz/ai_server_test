package activity

import (
	"time"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_public"
)

// ReqActivityTheCompetitionChooseGroupId  巅峰对决选择阵营
func ReqActivityTheCompetitionChooseGroupId(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2STheCompetitionChooseGroup) {
	res := new(proto_activity.S2CTheCompetitionChooseGroup)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeTheCompetition)
	if err != nil {
		log.Error("ReqActivityTheCompetitionChooseGroupId invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityTheCompetitionChooseGroupId reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	resp := reply
	if time.Now().Unix() >= resp.CloseTime {
		res.Code = proto_public.CommonErrorCode_ERR_ACTIVITYCLOSE
		ctx.Send(res)
		return
	}

	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":    "thecompetition_choosegroup",
		"req":    req,
		"player": pl.ToContext(),
	})
}

// ReqActivityTheCompetitionStake  巅峰对决押注
func ReqActivityTheCompetitionStake(ctx global.IPlayer, pl *model.Player, req *proto_activity.C2STheCompetitionStake) {
	res := new(proto_activity.S2CTheCompetitionStake)
	//先判断活动是否结束
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeTheCompetition)
	if err != nil {
		log.Error("ReqActivityTheCompetitionChooseGroupId invoke activity error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	if reply == nil {
		log.Error("ReqActivityTheCompetitionChooseGroupId reply is nil")
		res.Code = proto_public.CommonErrorCode_ERR_NOACTIVITY
		ctx.Send(res)
		return
	}

	resp := reply
	if time.Now().Unix() >= resp.CloseTime {
		res.Code = proto_public.CommonErrorCode_ERR_ACTIVITYCLOSE
		ctx.Send(res)
		return
	}

	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":    "thecompetition_stake",
		"req":    req,
		"player": pl.ToContext(),
	})
}
