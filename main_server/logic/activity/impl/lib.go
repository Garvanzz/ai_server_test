package impl

import (
	"github.com/golang/protobuf/proto"
	"strings"
	"xfx/core/define"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
)

type EventParams map[string]any

func Key[T any](params EventParams, key string) (T, bool) {
	v, ok := params[key]
	if !ok {
		var zero T
		return zero, false
	}
	result, ok := v.(T)
	if !ok {
		log.Error("event params key error:%v", key)
		return result, false
	}
	return result, true
}

func Trim(s string) string {
	return strings.Trim(s, "\"")
}

func SetProtoByType(actType string, msg *proto_activity.ActivityData, data proto.Message) {
	switch actType {
	case define.ActivityTypeDailyAccRecharge:
		nd := data.(*proto_activity.DailyAccumulateRecharge)
		msg.ActivityConsume = nd
	case define.ActivityTypeNormalMonthCard:
		nd := data.(*proto_activity.MonthCard)
		msg.MonthCard = nd
	case define.ActivityTypeTheCompetition:
		nd := data.(*proto_activity.TheCompetition)
		msg.TheCompetition = nd
	case define.ActivityTypeMainLineFund:
		nd := data.(*proto_activity.MainLineFund)
		msg.MainLineFund = nd
	case define.ActivityTypeLevelFund:
		nd := data.(*proto_activity.LevelFund)
		msg.LevelFund = nd
	case define.ActivityTypeBoxFund:
		nd := data.(*proto_activity.BoxFund)
		msg.BoxFund = nd
	case define.ActivityTypeArena:
		nd := data.(*proto_activity.Arena)
		msg.Arena = nd
	case define.ActivityTypeLadderRace:
		nd := data.(*proto_activity.LadderRace)
		msg.LadderRace = nd
	case define.ActivityTypeGoFish:
		nd := data.(*proto_activity.GoFish)
		msg.GoFish = nd
	case define.ActivityTypePassport:
		nd := data.(*proto_activity.Passport)
		msg.Passport = nd
	default:
		log.Error("set proto by type error:%v", actType)
	}
}

//func MergeItem(m map[int32]uint32, rewards []global.Reward) {
//	for _, reward := range rewards {
//		m[reward.ItemID] += reward.Num
//	}
//}
//
//// 取权重
//func weightedRandomIndex(weights []int64) int {
//	if len(weights) == 1 {
//		return 0
//	}
//
//	sum := int64(0)
//	for _, w := range weights {
//		sum += w
//	}
//	r := int64(global.ServerG.GetRandSrc().Float64() * float64(sum))
//	t := int64(0)
//
//	for i, w := range weights {
//		t += w
//		if t > r {
//			return i
//		}
//	}
//	return len(weights) - 1
//
//}

func SetProtoByType(actType string, msg *proto_activity.ActivityData, d proto.Message) {
	desc := GetActivityDesc(actType)
	if desc == nil || desc.SetProto == nil {
		log.Error("SetProtoByType: unknown type: %v", actType)
		return
	}
	desc.SetProto(msg, d)
}
