package internal

import (
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
)

// 获取是否购买月卡，以及月卡配置
func GetNormalMonthCard(ctx global.IPlayer, pl *model.Player) (bool, conf.MonthCard) {
	//获取月卡活动
	info, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeNormalMonthCard)
	if err != nil || info == nil {
		return false, conf.MonthCard{}
	}

	if !info.IsOpen {
		return false, conf.MonthCard{}
	}

	info1, err := invoke.ActivityClient(ctx).GetActivityData(pl.ToContext(), info.GetActivityId())
	if err != nil || info1 == nil {
		return false, conf.MonthCard{}
	}

	if info1.GetMonthCard().Day <= 0 {
		return false, conf.MonthCard{}
	}

	//获取月卡配置
	confs := config.MonthCard.All()
	for _, v := range confs {
		if v.Type == define.MonthCard_Month {
			return true, v
		}
	}

	return false, conf.MonthCard{}
}

// 通行证
func AddPassportScore(ctx global.IPlayer, pl *model.Player, score int32) {
	// 通过全局事件系统触发通行证积分事件
	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":         "passport_task_score",
		"player":      pl.ToContext(),
		"score":       score,
		"playermodel": pl,
		"IPlayer":     ctx,
	})
}
