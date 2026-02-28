package global

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

// GetPlayerHandbook 获取玩家图鉴
func GetPlayerHandbook(dbId int64) *model.Handbook {
	handbook := new(model.Handbook)
	rdb, err := db.GetEngineByPlayerId(dbId)
	if err != nil {
		log.Error("handbook error, no this server:%v", err)
		return handbook
	}

	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerHandbook, dbId))
	if err != nil {
		log.Error("player[%v],load Handbook error:%v", dbId, err)
		return handbook
	}

	if reply == nil {
		return handbook
	}

	err = json.Unmarshal(reply.([]byte), &handbook)
	if err != nil {
		log.Error("player[%v],load handbook unmarshal error:%v", dbId, err)
		return handbook
	}
	return handbook
}
