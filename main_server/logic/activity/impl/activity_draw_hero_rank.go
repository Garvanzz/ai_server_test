package impl

import (
	"xfx/core/define"
	"xfx/proto/proto_player"
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
