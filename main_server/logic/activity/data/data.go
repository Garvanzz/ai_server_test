package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"xfx/core/cache"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"

	"github.com/gomodule/redigo/redis"
)

var (
	Cache    *cache.WriteBackCache[int64, any]
	ServerId int
)

// =================================活动数据===========================================

func LoadAllActivityData() ([]*model.ActivityData, error) {
	keys, err := redis.Strings(db.RedisExec("KEYS", fmt.Sprintf("%s:*", define.ActivityRedisKey)))
	if err != nil {
		return nil, fmt.Errorf("KEYS error: %v", err)
	}

	if len(keys) == 0 {
		return nil, nil
	}

	values, err := redis.Values(db.RedisExec("MGET", redis.Args{}.AddFlat(keys)...))
	if err != nil {
		return nil, fmt.Errorf("MGET error: %v", err)
	}

	results := make([]*model.ActivityData, 0)
	for i := range keys {
		if values[i] == nil {
			continue
		}

		dataBytes, ok := values[i].([]byte)
		if !ok {
			continue
		}

		activityData := new(model.ActivityData)
		if err := json.Unmarshal(dataBytes, activityData); err != nil {
			continue
		}
		//log.Debug("活动数据:%v", activityData)
		results = append(results, activityData)
	}

	return results, nil
}

func LoadActivityData(id int32) (*model.ActivityData, error) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.ActivityRedisKey, id))
	if err != nil {
		log.Error("load activity data from redis error:%v", err)
		return nil, err
	}

	if reply == nil {
		return nil, nil
	}

	result := new(model.ActivityData)
	err = json.Unmarshal(reply.([]byte), result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func SaveActivityData(data *model.ActivityData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = db.RedisExec("SET", fmt.Sprintf("%s:%d", define.ActivityRedisKey, data.Id), string(b))
	if err != nil {
		return err
	}

	//log.Debug("保存活动数据:%v", data)
	return nil
}

func DelActivityData(id int64) {
	_, err := db.RedisExec("DEL", fmt.Sprintf("%s:%d", define.ActivityRedisKey, id))
	if err != nil {
		log.Error("DelActivityData error:%v", err)
	}
}

// =================================活动数据 END===========================================

// SavePlayerData 玩家数据落库
func SavePlayerData(key int64, data any) bool {
	b, err := json.Marshal(data)
	if err != nil {
		log.Error("save player activity data marshal error:%v", err)
		return false
	}

	actId, playerId := decodePlayerDataKey(key)
	_, err = db.RedisExec("HSET", activityPlayerRedisKey(actId), fmt.Sprintf("%d", playerId), b)
	if err != nil {
		log.Error("save player activity data db error:%v", err)
		return false
	}

	//log.Debug("保存活动玩家数据:%v,%v,%v", actId, playerId, data)
	return true
}

// SetPlayerData 保存角色数据
func SetPlayerData(actId, playerId int64, pd any) {
	//log.Debug("玩家数据活动变化: %v, %v, %s", actId, playerId, pd)
	key := encodePlayerDataKey(actId, playerId)
	Cache.Set(key, pd)
}

// LoadPlayerData 获取活动对应玩家数据
func LoadPlayerData[T comparable](actId, playerId int64) T {
	key := encodePlayerDataKey(actId, playerId)

	pd, ok := Cache.Get(key)
	if ok {
		return pd.(T)
	}

	var ret T

	bytes, err := redis.Bytes(db.RedisExec("HGET", activityPlayerRedisKey(actId), fmt.Sprintf("%d", playerId)))
	if err != nil && !errors.Is(err, redis.ErrNil) {
		log.Error("load activity player data error:%v", err)
		return ret
	}

	if bytes == nil {
		return ret
	}

	err = json.Unmarshal(bytes, &ret)
	if err != nil {
		log.Error("load activity player data unmarshal error:%v,%v,%v", err, actId, playerId)
	}

	Cache.SetClean(key, ret)
	return ret
}

func IterateActivityPlayerData[T any](actId int64, fn func(playerId int64, pd T) bool) {
	if fn == nil {
		return
	}

	visited := make(map[int64]struct{})
	if Cache != nil {
		Cache.Iterate(func(key int64, value any) bool {
			cacheActId, playerId := decodePlayerDataKey(key)
			if cacheActId != actId {
				return true
			}
			pd, ok := value.(T)
			if !ok {
				return true
			}
			visited[playerId] = struct{}{}
			return fn(playerId, pd)
		})
	}

	entries, err := redis.ByteSlices(db.RedisExec("HGETALL", activityPlayerRedisKey(actId)))
	if err != nil && !errors.Is(err, redis.ErrNil) {
		log.Error("IterateActivityPlayerData HGETALL error:%v", err)
		return
	}

	for i := 0; i+1 < len(entries); i += 2 {
		playerID, err := strconv.ParseInt(string(entries[i]), 10, 64)
		if err != nil {
			log.Error("IterateActivityPlayerData parse player id error:%v", err)
			continue
		}
		if _, ok := visited[playerID]; ok {
			continue
		}

		var pd T
		if err = json.Unmarshal(entries[i+1], &pd); err != nil {
			log.Error("IterateActivityPlayerData unmarshal error:%v, actId=%v, playerId=%v", err, actId, playerID)
			continue
		}

		if Cache != nil {
			Cache.SetClean(encodePlayerDataKey(actId, playerID), any(pd))
		}
		if !fn(playerID, pd) {
			return
		}
	}
}

// PurgeActivityPlayerData 删除活动所有对应玩家数据
func PurgeActivityPlayerData(actId int64) {
	var keysToDel []int64
	if Cache != nil {
		Cache.Iterate(func(key int64, _ any) bool {
			if cacheActId, _ := decodePlayerDataKey(key); cacheActId == actId {
				keysToDel = append(keysToDel, key)
			}
			return true
		})
		for _, k := range keysToDel {
			Cache.Del(k)
		}
	}

	_, err := db.RedisExec("DEL", activityPlayerRedisKey(actId))
	if err != nil {
		log.Error("PurgeActivityPlayerData redis DEL error:%v", err)
	}
}

func encodePlayerDataKey(actId, playerId int64) int64 {
	return actId*define.ActivityPlayerDataBase + playerId
}

func decodePlayerDataKey(key int64) (actId, playerId int64) {
	playerId = key % define.ActivityPlayerDataBase
	actId = (key - playerId) / define.ActivityPlayerDataBase
	return actId, playerId
}

func activityPlayerRedisKey(actId int64) string {
	return fmt.Sprintf("%s:%d", define.ActivityPlayerRedisKey, actId)
}
