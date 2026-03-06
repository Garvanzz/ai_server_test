package impl

import (
	"errors"
	"xfx/core/common"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
)

// ActivityBoxFund 成长基金
type ActivityBoxFund struct {
	BaseActivity
}

func (a *ActivityBoxFund) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.FundOptionPd](a, ctx.Id)

	return &proto_activity.BoxFund{
		Opt: &proto_activity.FundOption{
			Type:       pd.Type,
			NormalIds:  pd.NormalIds,
			AdvanceIds: pd.AdvanceIds,
			IsBuy:      pd.IsBuy,
		},
	}
}

func (a *ActivityBoxFund) OnInit() {

}

func (a *ActivityBoxFund) OnStart() {

}

func (a *ActivityBoxFund) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "recharge":
		shopConf, ok := Key[conf.Shop](params, "shopconf")
		if !ok {
			return
		}

		//判断是不是成长基金
		if shopConf.Type != define.SHOPTYPE_BOXFUND {
			return
		}

		pd := LoadPd[*model.FundOptionPd](a, ctx.Id)
		if pd.IsBuy {
			return
		}

		pd.IsBuy = true
		pd.Type = define.ActivityFund_Box

		// 推送活动数据
		a.PushActivityData(ctx.Id, a.Format(ctx))
	default:
	}
}

func (a *ActivityBoxFund) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	switch msg := req.(type) {
	case *proto_activity.C2SGetFundAward:
		return a.GetAward(ctx, msg)
	default:
		return nil, nil
	}
}

func (a *ActivityBoxFund) GetAward(ctx *proto_player.Context, req *proto_activity.C2SGetFundAward) ([]conf.ItemE, error) {
	if req.Type != define.ActivityFund_Box {
		return nil, errors.New("get activity typed error")
	}

	mainLineConfs, ok := GetTypedConf[conf.ActBoxFund](a.GetCfgId(), config.ActBoxFund.All())
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
			if common.IsHaveValueIntArray(pd.AdvanceIds, v) {
				return nil, errors.New("activity is al buy")
			}
		}

	} else {
		for _, v := range req.Ids {
			if common.IsHaveValueIntArray(pd.NormalIds, v) {
				return nil, errors.New("activity is al buy")
			}
		}
	}

	//判断条件
	for _, v := range req.Ids {
		conf := mainLineConfs[int64(v)]
		if conf.Level > int32(ctx.Level) {
			return nil, errors.New("activity level is outlimit")
		}
	}

	//写入
	awards := make([]conf.ItemE, 0)
	for _, v := range req.Ids {
		conf := mainLineConfs[int64(v)]
		if req.IsPay {
			pd.AdvanceIds = append(pd.AdvanceIds, v)
			awards = append(awards, conf.AdvanceAward...)
			if !common.IsHaveValueIntArray(pd.NormalIds, v) {
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

func (a *ActivityBoxFund) OnClose() {
	//活动结束补发奖励
}

func init() {
	RegisterActivity(define.ActivityTypeBoxFund, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityBoxFund) },
		NewActivityData: func() any { return nil },
		NewPlayerData: func() any {
			return &model.FundOptionPd{
				NormalIds:  make([]int32, 0),
				AdvanceIds: make([]int32, 0),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.BoxFund = data.(*proto_activity.BoxFund)
		},
		InjectFunc:  func(handler IActivity, data any) {},
		ExtractFunc: func(handler IActivity) any { return nil },
	})
}
