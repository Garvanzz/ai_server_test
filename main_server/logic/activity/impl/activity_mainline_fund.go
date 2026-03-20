package impl

import (
	"errors"
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
)

// ActivityMainLineFund 主线基金
type ActivityMainLineFund struct {
	BaseActivity
}

func (a *ActivityMainLineFund) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.FundOptionPd](a, ctx.Id)

	log.Debug("详情:%v:%v", pd.NormalIds, pd.AdvanceIds)
	return &proto_activity.MainLineFund{
		Opt: &proto_activity.FundOption{
			Type:       pd.Type,
			NormalIds:  pd.NormalIds,
			AdvanceIds: pd.AdvanceIds,
			IsBuy:      pd.IsBuy,
		},
	}
}

func (a *ActivityMainLineFund) OnInit() {}

func (a *ActivityMainLineFund) OnStart() {

}

func (a *ActivityMainLineFund) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "recharge":
		shopConf, ok := Key[conf.Shop](params, "shopconf")
		if !ok {
			return
		}

		//判断是不是主线基金
		if shopConf.Type != define.SHOPTYPE_MAINLINEFUND {
			return
		}

		pd := LoadPd[*model.FundOptionPd](a, ctx.Id)
		if pd.IsBuy {
			return
		}

		pd.IsBuy = true
		pd.Type = define.ActivityFund_MainLine

		// 推送活动数据
		a.PushActivityData(ctx.Id, a.Format(ctx))
	default:
	}
}

func (a *ActivityMainLineFund) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	switch msg := req.(type) {
	case *proto_activity.C2SGetFundAward:
		return a.GetAward(ctx, msg)
	default:
		return nil, nil
	}
}

func (a *ActivityMainLineFund) GetAward(ctx *proto_player.Context, req *proto_activity.C2SGetFundAward) ([]conf.ItemE, error) {
	if req.Type != define.ActivityFund_MainLine {
		return nil, errors.New("get activity typed error")
	}

	mainLineConfs, ok := GetTypedConf[conf.ActMainLineFund](a.GetCfgId(), config.ActMainLineFund.All())
	if !ok {
		log.Error("get activity typed config error:%v", a.GetCfgId())
		return nil, errors.New("get activity typed config error")
	}

	pd := LoadPd[*model.FundOptionPd](a, ctx.Id)
	if req.IsPay {
		if !pd.IsBuy {
			log.Error("activity is not buy error:%v", a.GetCfgId())
			return nil, errors.New("activity is not buy")
		}
	}

	if req.IsPay {
		for _, v := range req.Ids {
			if utils.ContainsInt32(pd.AdvanceIds, v) {
				return nil, errors.New("activity is al buy")
			}
		}

	} else {
		for _, v := range req.Ids {
			if utils.ContainsInt32(pd.NormalIds, v) {
				return nil, errors.New("activity is al buy")
			}
		}
	}

	//判断条件
	for _, v := range req.Ids {
		conf := mainLineConfs[int64(v)]
		if conf.Stage > ctx.Stage {
			return nil, errors.New("activity stage is outlimit")
		}
	}

	//写入
	awards := make([]conf.ItemE, 0)
	for _, v := range req.Ids {
		conf := mainLineConfs[int64(v)]
		if req.IsPay {
			pd.AdvanceIds = append(pd.AdvanceIds, v)
			awards = append(awards, conf.AdvanceAward...)

			if !utils.ContainsInt32(pd.NormalIds, v) {
				pd.NormalIds = append(pd.NormalIds, v)
				awards = append(awards, conf.NormalAward...)
			}
		} else {
			pd.NormalIds = append(pd.NormalIds, v)
			awards = append(awards, conf.NormalAward...)
		}
	}

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	return awards, nil
}

func (a *ActivityMainLineFund) OnClose() {
	//活动结束补发奖励
}

func (a *ActivityMainLineFund) Update(now time.Time) {
	// 跨天逻辑已迁移到 OnDayReset
}

// OnDayReset 跨天重置
func (a *ActivityMainLineFund) OnDayReset(now time.Time) {
	log.Debug("ActivityMainLineFund OnDayReset: actId=%v", a.GetId())
}

func init() {
	RegisterActivity(define.ActivityTypeMainLineFund, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityMainLineFund) },
		NewActivityData: func() any { return nil },
		NewPlayerData: func() any {
			return &model.FundOptionPd{
				NormalIds:  make([]int32, 0),
				AdvanceIds: make([]int32, 0),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.MainLineFund = data.(*proto_activity.MainLineFund)
		},
		InjectFunc:  func(handler IActivity, data any) {},
		ExtractFunc: func(handler IActivity) any { return nil },
	})
}
