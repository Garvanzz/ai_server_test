package impl

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"time"
	"xfx/core/common"
	"xfx/core/config/conf"
	"xfx/core/model"
	//"xfx/main_server/logic/activity/data"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"
)

// ActivityDailyAccRecharge 每日累充
type ActivityDailyAccRecharge struct {
	BaseActivity
	data *model.ActDataDailyAccumulateRecharge
}

func (a *ActivityDailyAccRecharge) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.DailyAccumulateRechargePd](a, ctx.Id)

	return &proto_activity.DailyAccumulateRecharge{
		Money:   pd.Money,
		GetList: pd.GetList,
	}
}

func (a *ActivityDailyAccRecharge) OnInit() {
}

func (a *ActivityDailyAccRecharge) OnStart() {
	//commonConf, ok := GetCommonConf(a.GetCfgId())
}

func (a *ActivityDailyAccRecharge) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "recharge":
		rechargeConf, ok := Key[conf.Recharge](params, "rechargeconf")
		if !ok {
			return
		}

		pd := LoadPd[*model.DailyAccumulateRechargePd](a, ctx.Id)
		pd.Money += rechargeConf.Price

		// 推送活动数据
		a.PushActivityData(ctx.Id, a.Format(ctx))
	default:
	}
}

func (a *ActivityDailyAccRecharge) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	switch msg := req.(type) {
	case *proto_activity.C2SActivityAward:
		return a.GetAward(ctx, msg)
	default:
		return nil, nil
	}
}

func (a *ActivityDailyAccRecharge) GetAward(ctx *proto_player.Context, req *proto_activity.C2SActivityAward) ([]conf.ItemE, error) {
	consumeConfs, ok := GetTypedConf[conf.ActDailyAccumulateRecharge](a.GetCfgId())
	if !ok {
		log.Error("get activity typed config error:%v", a.GetCfgId())
		return nil, errors.New("get activity typed config error")
	}

	// 没有该奖励配置
	var config conf.ActDailyAccumulateRecharge
	for _, consumeConf := range consumeConfs {
		if consumeConf.Id == int64(req.Index) {
			config = consumeConf
		}
	}

	if config.Id == 0 {
		log.Error("get activity award index error:%v", req.Index)
		return nil, errors.New("get activity typed config error")
	}

	pd := LoadPd[*model.DailyAccumulateRechargePd](a, ctx.Id)
	if pd.Money < config.Progress {
		return nil, errors.New("get activity money is not enghth")
	}

	if common.IsHaveValueIntArray(pd.GetList, int32(config.Id)) {
		return nil, errors.New("get activity has get")
	}
	pd.GetList = append(pd.GetList, int32(config.Id))
	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	//奖励
	return config.Award, nil
}

func (a *ActivityDailyAccRecharge) Update(now time.Time) {
	// 检查跨天 按照排行榜发奖励
}

func (a *ActivityDailyAccRecharge) OnClose() {
	//活动结束补发奖励
}

func (a *ActivityDailyAccRecharge) Inject(data any) {
	if data == nil {
		a.data = new(model.ActDataDailyAccumulateRecharge)
		return
	}
	a.data = data.(*model.ActDataDailyAccumulateRecharge)
}

func (a *ActivityDailyAccRecharge) Extract() any { return a.data }

func init() {
	RegisterActivity(define.ActivityTypeDailyAccRecharge, &ActivityDesc{
		NewHandler: func() IActivity { return new(ActivityDailyAccRecharge) },
		NewActivityData: func() any { return new(model.ActDataDailyAccumulateRecharge) },
		NewPlayerData: func() any { return new(model.DailyAccumulateRechargePd) },
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.DailyAccumulateRecharge = data.(*proto_activity.DailyAccumulateRecharge)
		},
	})
}

