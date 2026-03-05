package impl

import (
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
)

// ActivityNormalMonthCard 月卡
type ActivityNormalMonthCard struct {
	BaseActivity
	data *model.ActDataMonthCard
}

func (a *ActivityNormalMonthCard) OnInit() {}

func (a *ActivityNormalMonthCard) OnStart() {}

func (a *ActivityNormalMonthCard) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.MonthCardPd](a, ctx.Id)
	log.Debug("加载月卡玩家数据:%s", pd)
	return &proto_activity.MonthCard{
		Day: pd.Day,
	}
}

func (a *ActivityNormalMonthCard) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "recharge":
		shopConf, ok := Key[conf.Shop](params, "shopconf")
		if !ok {
			return
		}

		if shopConf.Type != define.SHOPTYPE_NORMALMONTHCARD {
			return
		}

		confs := config.CfgMgr.AllJson()["MonthCard"].(map[int64]conf.MonthCard)
		var conf conf.MonthCard
		for _, v := range confs {
			if v.Type == define.MonthCard_Month {
				conf = v
				break
			}
		}
		if conf.Id <= 0 {
			return
		}

		pd := LoadPd[*model.MonthCardPd](a, ctx.Id)
		pd.Day += conf.Day
		pd.Count += 1
		pd.LastTime = time.Now().Unix()

		// 推送活动数据
		a.PushActivityData(ctx.Id, a.Format(ctx))
		log.Debug("推送活动数据:%v", pd)
	default:
	}
}

func (a *ActivityNormalMonthCard) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	return nil, nil
}

func (a *ActivityNormalMonthCard) OnClose() {
	//活动结束补发奖励
}

func init() {
	RegisterActivity(define.ActivityTypeNormalMonthCard, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityNormalMonthCard) },
		NewActivityData: func() any { return new(model.ActDataMonthCard) },
		NewPlayerData:   func() any { return new(model.MonthCardPd) },
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.MonthCard = data.(*proto_activity.MonthCard)
		},
		InjectFunc: func(handler IActivity, data any) {
			h := handler.(*ActivityNormalMonthCard)
			if data == nil {
				h.data = new(model.ActDataMonthCard)
				return
			}
			h.data = data.(*model.ActDataMonthCard)
		},
		ExtractFunc: func(handler IActivity) any { return handler.(*ActivityNormalMonthCard).data },
	})
}
