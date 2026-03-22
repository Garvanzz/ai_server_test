package impl

import (
	"errors"
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/logic/activity/data"
	"xfx/pkg/utils"

	"github.com/golang/protobuf/proto"
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
	consumeConfs, ok := GetTypedConf[conf.ActDailyAccumulateRecharge](a.GetCfgId(), config.ActDailyAccRecharge.All())
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

	if utils.ContainsInt32(pd.GetList, int32(config.Id)) {
		return nil, errors.New("get activity has get")
	}
	pd.GetList = append(pd.GetList, int32(config.Id))
	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	//奖励
	return config.Award, nil
}

func (a *ActivityDailyAccRecharge) Update(now time.Time) {
	// 跨天逻辑已迁移到 OnDayReset
}

// OnDayReset 跨天重置：清空每日累充金额和已领取列表
func (a *ActivityDailyAccRecharge) OnDayReset(now time.Time) {
	data.IterateActivityPlayerData[*model.DailyAccumulateRechargePd](a.GetId(), func(_ int64, pd *model.DailyAccumulateRechargePd) bool {
		if pd == nil {
			return true
		}
		pd.Money = 0
		pd.GetList = pd.GetList[:0]
		return true
	})
	log.Debug("ActivityDailyAccRecharge OnDayReset: actId=%v", a.GetId())
}

func (a *ActivityDailyAccRecharge) OnClose() {
	//活动结束补发奖励
}

func init() {
	RegisterActivity(define.ActivityTypeDailyAccRecharge, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityDailyAccRecharge) },
		NewActivityData: func() any { return new(model.ActDataDailyAccumulateRecharge) },
		NewPlayerData: func() any {
			return &model.DailyAccumulateRechargePd{
				GetList: make([]int32, 0),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.ActivityConsume = data.(*proto_activity.DailyAccumulateRecharge)
		},
		InjectFunc: func(handler IActivity, data any) {
			h := handler.(*ActivityDailyAccRecharge)
			if data == nil {
				h.data = new(model.ActDataDailyAccumulateRecharge)
				return
			}
			h.data = data.(*model.ActDataDailyAccumulateRecharge)
		},
		ExtractFunc: func(handler IActivity) any { return handler.(*ActivityDailyAccRecharge).data },
	})
}
