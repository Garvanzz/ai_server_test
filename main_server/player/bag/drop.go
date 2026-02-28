package bag

import (
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/pkg/log"
	"xfx/pkg/utils"
)

type dropFn func(drop conf.Drop, appointId int32) []conf.ItemE

var dropDeal = map[int16]dropFn{
	// 常规掉落
	define.DropTypeNormal: func(dropConf conf.Drop, appointId int32) []conf.ItemE {
		if appointId <= 0 {
			return dropConf.Rewards
		}

		l := make([]conf.ItemE, 0)
		for i, v := range dropConf.Rewards {
			if v.ItemId == appointId {
				l = append(l, dropConf.Rewards[i])
			}
		}
		return l
	},

	// 权重随机
	define.DropTypeWeight: func(dropConf conf.Drop, appointId int32) []conf.ItemE {
		return utils.WeightedRandom(dropConf.Weight, dropConf.Rewards, int(dropConf.Num))
	},

	// 概率随机
	define.DropTypeIndependentRandom: func(dropConf conf.Drop, appointId int32) []conf.ItemE {
		l := make([]conf.ItemE, 0)
		for i, v := range dropConf.Probability {
			randRst := utils.RandInt(0, 10000)
			if randRst < int(v) {
				l = append(l, dropConf.Rewards[i])
			}
		}
		return l
	},
}

// GetDrop 获取掉落
func GetDrop(dropId, apppointId int32) []conf.ItemE {
	dropConf, ok := config.CfgMgr.AllJson["Drop"].(map[int64]conf.Drop)[int64(dropId)]
	if !ok {
		log.Error("getDrop conf not found %v", dropId)
		return nil
	}

	return dropDeal[int16(dropConf.Type)](dropConf, apppointId)
}
