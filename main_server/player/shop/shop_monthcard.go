package shop

import (
	"encoding/json"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
)

// 商城to鉴宝月卡
func ShopToGemAppraisalMonthCard(ctx global.IPlayer, pl *model.Player) {

	//查找之前是否有购买月卡记录
	reply, err := db.RedisExec("HGET", define.GemAppraisal_MonthCard, pl.Id)
	if err != nil {
		log.Error("[%v],load ShopToGemAppraisalMonthCard error:%v", pl.Id, err)
		return
	}

	//添加进去
	if reply == nil {
		m := new(model.GemAppraisalMonthCard)
		m.EffectDay = 30
		m.PID = pl.Uid
		m.DbId = pl.Id
		js, _ := json.Marshal(m)
		db.RedisExec("HSET", define.GemAppraisal_MonthCard, pl.Id, js)

		//玩家自己数据
		pl.Welfare.MonthCard[define.MonthCard_GemAppraisal] = &model.MonthCard{
			IsGet: false,
		}

		//推送
		internal.PushMonthCard(ctx, pl, define.MonthCard_GemAppraisal)
	} else {
		m := new(model.GemAppraisalMonthCard)
		err = json.Unmarshal(reply.([]byte), &m)
		if err != nil {
			log.Error("player[%v],load task unmarshal error:%v", pl.Id, err)
		}

		m.EffectDay += 30
		js, _ := json.Marshal(m)
		db.RedisExec("HSET", define.GemAppraisal_MonthCard, pl.Id, js)
	}
}
