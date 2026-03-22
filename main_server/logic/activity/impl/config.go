package impl

import (
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/pkg/log"
)

type IActivityConfig interface {
	GetActivityId() int64
}

func GetAllCommonConf() map[int64]conf.Activity {
	return config.Activity.All()
}

func GetCommonConf(cfgId int64) (conf.Activity, bool) {
	activityConf, ok := config.Activity.Find(cfgId)
	if !ok {
		log.Error("register new activity get config id error:%v", cfgId)
		return activityConf, false
	}

	return activityConf, true
}

func GetTypedConf[T IActivityConfig](cfgId int64, allConfs map[int64]T) (map[int64]T, bool) {
	result := make(map[int64]T)
	for id, t := range allConfs {
		if t.GetActivityId() == cfgId {
			result[id] = t
		}
	}
	return result, len(result) > 0
}

func FindTypedConf[T IActivityConfig](cfgId int64, allConfs map[int64]T, predicate func(T) bool) (T, bool) {
	for _, t := range allConfs {
		if t.GetActivityId() != cfgId {
			continue
		}
		if predicate != nil && !predicate(t) {
			continue
		}
		return t, true
	}
	var zero T
	return zero, false
}
