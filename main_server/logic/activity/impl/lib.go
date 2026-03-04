package impl

import (
	"encoding/json"
	"strings"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/model"
	"xfx/main_server/logic/activity/data"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"

	"github.com/golang/protobuf/proto"
)

type IActivityConfig interface {
	GetActivityId() int64
}

func GetAllCommonConf() map[int64]conf.Activity {
	return config.CfgMgr.AllJson["Activity"].(map[int64]conf.Activity)
}

func GetCommonConf(cfgId int64) (conf.Activity, bool) {
	activityConf, ok := config.CfgMgr.AllJson["Activity"].(map[int64]conf.Activity)[cfgId]
	if !ok {
		log.Error("register new activity get config id error:%v", cfgId)
		return activityConf, false
	}

	return activityConf, true
}

func GetTypedConf[T IActivityConfig](cfgId int64) (map[int64]T, bool) {
	commonConf, ok := GetCommonConf(cfgId)
	if !ok {
		return nil, false
	}

	activityConfs, ok := config.CfgMgr.AllJson[commonConf.Type].(map[int64]T)
	if !ok {
		log.Error("register new activity get config id error:%v", cfgId)
		return activityConfs, false
	}

	result := make(map[int64]T)
	for id, t := range activityConfs {
		if t.GetActivityId() != cfgId {
			continue
		}
		result[id] = t
	}

	return result, true
}

func GetOneTypedConf[T IActivityConfig](cfgId int64) (T, bool) {
	var t T
	confs, ok := GetTypedConf[T](cfgId)
	if !ok {
		return t, false
	}

	for _, v := range confs {
		return v, true
	}

	return t, false
}

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

func UnmarshalActivityData(actData *model.ActivityData) any {
	if actData.Data == nil {
		return nil
	}

	desc := GetActivityDesc(actData.Type)
	if desc == nil || desc.NewPlayerData == nil {
		log.Error("UnmarshalActivityData: no activity data factory for type: %v", actData.Type)
		return nil
	}

	d := desc.NewActivityData()
	if d == nil {
		return nil
	}

	b, err := json.Marshal(actData.Data)
	if err != nil {
		log.Error("UnmarshalActivityData error:%v", err)
		return d
	}

	err = json.Unmarshal(b, d)
	if err != nil {
		log.Error("convert data type failed for activity type: %v,error:%v", actData.Type, err)
		return d
	}

	return d
}

func LoadPd[T comparable](a BaseInfo, playerId int64) T {
	var zero T
	d := data.LoadPlayerData[T](a.GetId(), playerId)
	if d != zero {
		return d
	}

	desc := GetActivityDesc(a.GetType())
	if desc == nil || desc.NewPlayerData == nil {
		log.Error("LoadPd: no player data factory for type: %v", a.GetType())
		return zero
	}

	ret := desc.NewPlayerData()
	data.SetPlayerData(a.GetId(), playerId, ret)
	return ret.(T)
}
