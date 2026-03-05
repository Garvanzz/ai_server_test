package impl

import (
	"encoding/json"
	"strings"
	"xfx/core/model"
	"xfx/main_server/logic/activity/data"
	"xfx/pkg/log"
)

func Trim(s string) string {
	return strings.Trim(s, "\"")
}

func getSubArray(arr []int64, start, end int) []int64 {
	length := len(arr)
	if length == 0 || start > end || start < 0 || end < 0 || start > length {
		return []int64{}
	}

	end = min(end, length)

	sub := make([]int64, end-start+1)
	copy(sub, arr[start-1:end])

	return sub
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
