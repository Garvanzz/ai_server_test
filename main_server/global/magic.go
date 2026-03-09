package global

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

// GetPlayerMagic 获取玩家功法
func GetPlayerMagic(dbId int64) *model.Magic {
	magic := new(model.Magic)

	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerMagic, dbId))
	if err != nil {
		log.Error("player[%v],load magic error:%v", dbId, err)
		return magic
	}

	if reply == nil {
		return magic
	}

	err = json.Unmarshal(reply.([]byte), &magic)
	if err != nil {
		log.Error("player[%v],load magic unmarshal error:%v", dbId, err)
		return magic
	}
	return magic
}
