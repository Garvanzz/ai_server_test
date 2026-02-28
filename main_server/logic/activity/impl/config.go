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
		if t.GetActivityId() == cfgId {
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
