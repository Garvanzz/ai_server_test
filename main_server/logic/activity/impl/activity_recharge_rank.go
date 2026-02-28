package impl

import (
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/proto/proto_player"
)

// ActivityRechargeRank 充值排行榜
type ActivityRechargeRank struct {
	BaseActivity
}

func (a *ActivityRechargeRank) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "recharge":
		rechargeConf, ok := Key[conf.Recharge](params, "rechargeconf")
		if !ok {
			return
		}

		// 更新排行榜
		updateActivityRank(a, ctx, 0, rechargeConf.Price, define.RankTypeRecharge)
	default:
	}
}

func (a *ActivityRechargeRank) OnClose() {
	//活动结束补发奖励
	sendRankReward(a, define.RankTypeRecharge, nil)

	//删除排行榜
	deleteActivityRank(a, define.RankTypeRecharge)
	//活动结束，发放奖励
}
