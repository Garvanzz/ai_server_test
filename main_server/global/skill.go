package global

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

// GetPlayerSkill 获取玩家技能
func GetPlayerSkill(dbId int64) *model.Skill {
	skill := new(model.Skill)
	rdb, err := db.GetEngineByPlayerId(dbId)
	if err != nil {
		log.Error("skill error, no this server:%v", err)
		return skill
	}

	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerSkill, dbId))
	if err != nil {
		log.Error("player[%v],load skill error:%v", dbId, err)
		return skill
	}

	if reply == nil {
		return skill
	}

	err = json.Unmarshal(reply.([]byte), &skill)
	if err != nil {
		log.Error("player[%v],load skill unmarshal error:%v", dbId, err)
		return skill
	}
	return skill
}
