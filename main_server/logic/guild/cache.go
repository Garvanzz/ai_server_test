package guild

import (
	"encoding/json"
	"fmt"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

func savePlayerData(playerId int64, v *model.PlayerGuild) bool {
	b, err := json.Marshal(v)
	if err != nil {
		log.Error("callback marshal error:%v", err)
		return false
	}

	_, err = db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerGuildKey, playerId), string(b))
	if err != nil {
		log.Error("callback redis error:%v", err)
		return false
	}
	return true
}

func loadPlayerData(playerId int64) (*model.PlayerGuild, error) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerGuildKey, playerId))
	if err != nil {
		return nil, err
	}

	data := new(model.PlayerGuild)
	if reply != nil {
		err = json.Unmarshal(reply.([]byte), data)
		if err != nil {
			return nil, fmt.Errorf("load guild player from cache json marshal error:%v,%v", playerId, err)
		}
		return data, nil
	}

	data = newPlayerGuild(playerId)
	return data, nil
}

func (mgr *Manager) loadPlayerGuildFromCache(playerId int64) *model.PlayerGuild {
	if playerId == 0 {
		log.Error("load guild player from cache id error:%v", playerId)
		return nil
	}

	v, ok := mgr.cache.Get(playerId)
	if ok {
		return v
	}

	v, err := loadPlayerData(playerId)
	if err != nil {
		log.Error("loadPlayerData error:%v", err)
		return nil
	}

	// 放到缓存中
	mgr.cache.Set(playerId, v)

	return v
}
