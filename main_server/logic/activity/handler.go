package activity

import (
	"fmt"
	"xfx/core/define"
	"xfx/main_server/logic/activity/impl"
)

// // 分配活动处理方法
// func setActivityHandler(ent *entity) impl.IActivity {
// 	var handler impl.IActivity
// 	switch ent.Type {
// 	case define.ActivityTypeDailyAccRecharge:
// 		h := new(impl.ActivityDailyAccRecharge)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeDrawHeroRank:
// 		h := new(impl.ActivityDrawHeroRank)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeRechargeRank:
// 		h := new(impl.ActivityRechargeRank)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeNormalMonthCard:
// 		h := new(impl.ActivityNormalMonthCard)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeTheCompetition:
// 		h := new(impl.ActivityTheCompetition)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeMainLineFund:
// 		h := new(impl.ActivityMainLineFund)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeLevelFund:
// 		h := new(impl.ActivityLevelFund)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeBoxFund:
// 		h := new(impl.ActivityBoxFund)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeArena:
// 		h := new(impl.ActivityArena)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeLadderRace:
// 		h := new(impl.ActivityLadderRace)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeGoFish:
// 		h := new(impl.ActivityGoFish)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypePassport:
// 		h := new(impl.ActivityPassport)
// 		h.BaseInfo = ent
// 		handler = h
// 	case define.ActivityTypeSeason:
// 		h := new(impl.ActivitySeason)
// 		h.BaseInfo = ent
// 		handler = h
// 	default:
// 		panic(fmt.Sprintf("missing activity handler:%v", ent.Type))
// 	}
// 	return handler
// }

func setActivityHandler(ent *entity) impl.IActivity {
    desc := impl.GetActivityDesc(ent.Type)
    if desc == nil {
        panic(fmt.Sprintf("missing activity handler: %v", ent.Type))
    }
    h := desc.NewHandler()
    // 通过反射或接口注入 BaseInfo
    h.(impl.BaseInfoSetter).SetBaseInfo(ent)
    return h
}
