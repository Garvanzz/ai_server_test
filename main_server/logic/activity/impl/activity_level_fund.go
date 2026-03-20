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

// ActivityLevelFund 成长基金
type ActivityLevelFund struct {
	BaseActivity
}

func (a *ActivityLevelFund) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.FundOptionPd](a, ctx.Id)

	return &proto_activity.LevelFund{
		Opt: &proto_activity.FundOption{
			Type:       pd.Type,
			NormalIds:  pd.NormalIds,
			AdvanceIds: pd.AdvanceIds,
			IsBuy:      pd.IsBuy,
		},
	}
}

func (a *ActivityLevelFund) OnInit() {

}

func (a *ActivityLevelFund) OnStart() {

}

func (a *ActivityLevelFund) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "recharge":
		shopConf, ok := Key[conf.Shop](params, "shopconf")
		if !ok {
			return
		}

		//判断是不是成长基金
		if shopConf.Type != define.SHOPTYPE_LEVELFUND {
			return
		}

		pd := LoadPd[*model.FundOptionPd](a, ctx.Id)
		if pd.IsBuy {
			return
		}

		pd.IsBuy = true
		pd.Type = define.ActivityFund_Level

		// 推送活动数据
		a.PushActivityData(ctx.Id, a.Format(ctx))
	default:
	}
}

func (a *ActivityLevelFund) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	switch msg := req.(type) {
	case *proto_activity.C2SGetFundAward:
		return a.GetAward(ctx, msg)
	default:
		return nil, nil
	}
}

func (a *ActivityLevelFund) GetAward(ctx *proto_player.Context, req *proto_activity.C2SGetFundAward) ([]conf.ItemE, error) {
	if req.Type != define.ActivityFund_Level {
		return nil, errors.New("get activity typed error")
	}

	mainLineConfs, ok := GetTypedConf[conf.ActLevelFund](a.GetCfgId(), config.ActLevelFund.All())
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
			log.Debug("66666666:%v,%v", v, pd.AdvanceIds)
			if !utils.ContainsInt32(pd.NormalIds, v) {
				pd.NormalIds = append(pd.NormalIds, v)
				awards = append(awards, conf.NormalAward...)
			}
		} else {
			pd.NormalIds = append(pd.NormalIds, v)
			log.Debug("555555555555:%v,%v", v, pd.NormalIds)
			awards = append(awards, conf.NormalAward...)
		}
	}

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	return awards, nil
}

func (a *ActivityLevelFund) OnClose() {
	//活动结束补发奖励
}

func (a *ActivityLevelFund) Update(now time.Time) {
	// 跨天逻辑已迁移到 OnDayReset
}

// OnDayReset 跨天重置
func (a *ActivityLevelFund) OnDayReset(now time.Time) {
	log.Debug("ActivityLevelFund OnDayReset: actId=%v", a.GetId())
}

func init() {
	RegisterActivity(define.ActivityTypeLevelFund, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityLevelFund) },
		NewActivityData: func() any { return nil },
		NewPlayerData: func() any {
			return &model.FundOptionPd{
				NormalIds:  make([]int32, 0),
				AdvanceIds: make([]int32, 0),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.LevelFund = data.(*proto_activity.LevelFund)
		},
		InjectFunc:  func(handler IActivity, data any) {},
		ExtractFunc: func(handler IActivity) any { return nil },
	})
}
