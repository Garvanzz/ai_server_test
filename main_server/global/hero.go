package global

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

// GetPlayerHero 获取玩家角色信息
func GetPlayerHero(dbId int64) *model.Hero {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerHero, dbId))
	if err != nil {
		log.Error("player[%v],load hero error:%v", dbId, err)
		return new(model.Hero)
	}

	if reply == nil {
		return new(model.Hero)
	}

	m := new(model.Hero)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load hero unmarshal error:%v", dbId, err)
	}

	return m
}
