package impl

import (
	"xfx/core/define"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
)

// ActivityDrawHeroRank 招募排行榜
type ActivityDrawHeroRank struct {
	BaseActivity
}

func (a *ActivityDrawHeroRank) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "draw_hero":
		count, ok := Key[int32](params, "value")
		if !ok {
			return
		}
		//排行榜
		updateActivityRank(a, ctx, 0, count, define.RankTypeDrawHero)
	default:
	}
}

func (a *ActivityDrawHeroRank) OnClose() {
	//活动结束补发奖励
	sendRankReward(a, define.RankTypeDrawHero, nil)

	//删除排行榜
	deleteActivityRank(a, define.RankTypeDrawHero)
}

func init() {
	RegisterActivity(define.ActivityTypeDrawHeroRank, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityDrawHeroRank) },
		NewActivityData: func() any { return nil },
		NewPlayerData:   func() any { return nil },
		SetProto:        func(msg *proto_activity.ActivityData, data proto.Message) {},
		InjectFunc:      func(handler IActivity, data any) {},
		ExtractFunc:     func(handler IActivity) any { return nil },
	})
}
