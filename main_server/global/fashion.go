package global

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

// GetPlayerEquip 获取玩家时装
func GetPlayerFashion(dbId int64) *model.Fashion {
	fashion := new(model.Fashion)
	rdb, err := db.GetEngineByPlayerId(dbId)
	if err != nil {
		log.Error("fashion error, no this server:%v", err)
		return fashion
	}

	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerFashion, dbId))
	if err != nil {
		log.Error("player[%v],load fashion error:%v", dbId, err)
		return fashion
	}

	if reply == nil {
		return fashion
	}

	err = json.Unmarshal(reply.([]byte), &fashion)
	if err != nil {
		log.Error("player[%v],load fashion unmarshal error:%v", dbId, err)
		return fashion
	}
	return fashion
}
