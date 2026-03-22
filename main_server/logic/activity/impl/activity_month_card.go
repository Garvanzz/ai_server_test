package impl

import (
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/main_server/logic/activity/data"
	"xfx/pkg/log"
	"xfx/pkg/utils"
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
	log.Debug("加载月卡玩家数据:%v", pd)
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

		monthCardConf, ok := getMonthCardConf(define.MonthCard_Month)
		if !ok {
			return
		}

		pd := LoadPd[*model.MonthCardPd](a, ctx.Id)
		pd.Day += monthCardConf.Day
		pd.Count += 1
		pd.LastTime = utils.Now().Unix()

		// 推送活动数据
		a.PushActivityData(ctx.Id, a.Format(ctx))
		log.Debug("推送活动数据:%v", pd)
	default:
	}
}

func (a *ActivityNormalMonthCard) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	return nil, nil
}

func (a *ActivityNormalMonthCard) Update(now time.Time) {
	// 跨天逻辑已迁移到 OnDayReset
}

// OnDayReset 跨天处理：发放月卡每日奖励
func (a *ActivityNormalMonthCard) OnDayReset(now time.Time) {
	monthCardConf, ok := getMonthCardConf(define.MonthCard_Month)
	if !ok {
		log.Error("ActivityNormalMonthCard OnDayReset no config: actId=%v", a.GetId())
		return
	}

	receiverIDs := make([]int64, 0)
	nowUnix := now.Unix()
	data.IterateActivityPlayerData[*model.MonthCardPd](a.GetId(), func(playerID int64, pd *model.MonthCardPd) bool {
		if pd == nil || pd.Day <= 0 {
			return true
		}
		pd.Day--
		pd.LastTime = nowUnix
		receiverIDs = append(receiverIDs, playerID)
		return true
	})

	if len(receiverIDs) > 0 {
		ok = invoke.MailClient(a.Module()).SendMail(
			define.PlayerMail,
			"月卡",
			"月卡每日奖励补发",
			"",
			"",
			"游戏系统",
			monthCardConf.Reward,
			receiverIDs,
			int64(0),
			int32(0),
			false,
			[]string{},
		)
		if !ok {
			log.Error("ActivityNormalMonthCard OnDayReset send mail failed: actId=%v, receivers=%d", a.GetId(), len(receiverIDs))
		}
	}
	log.Debug("ActivityNormalMonthCard OnDayReset: actId=%v", a.GetId())
}

func (a *ActivityNormalMonthCard) OnClose() {
	//活动结束补发奖励
}

func getMonthCardConf(cardType int32) (conf.MonthCard, bool) {
	for _, cardConf := range config.MonthCard.All() {
		if cardConf.Type == cardType {
			return cardConf, true
		}
	}
	return conf.MonthCard{}, false
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
