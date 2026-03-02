package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"xfx/core/cache"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

var (
	Cache    *cache.WriteBackCache[int64, any]
	ServerId int
)

// =================================活动数据===========================================

func LoadAllActivityData() ([]*model.ActivityData, error) {
	rdb, _ := db.GetEngine(ServerId)

	keys, err := redis.Strings(rdb.RedisExec("KEYS", fmt.Sprintf("%s:*", define.ActivityRedisKey)))
	if err != nil {
		return nil, fmt.Errorf("KEYS error: %v", err)
	}

	if len(keys) == 0 {
		return nil, nil
	}

	values, err := redis.Values(rdb.RedisExec("MGET", redis.Args{}.AddFlat(keys)...))
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
	rdb, _ := db.GetEngine(ServerId)
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.ActivityRedisKey, id))
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
	rdb, _ := db.GetEngine(ServerId)

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.ActivityRedisKey, data.Id), string(b))
	if err != nil {
		return err
	}

	//log.Debug("保存活动数据:%v", data)
	return nil
}

func DelActivityData(id int64) {
	rdb, _ := db.GetEngine(ServerId)
	_, err := rdb.RedisExec("DEL", fmt.Sprintf("%s:%d", define.ActivityRedisKey, id))
	if err != nil {
		log.Error("DelActivityData error:%v", err)
	}
}

func UnmarshalActivityData(actData *model.ActivityData) any {
	if actData.Data == nil {
		return nil
	}

	var d any

	switch actData.Type {
	case define.ActivityTypeDailyAccRecharge:
		d = new(model.ActDataDailyAccumulateRecharge)
	case define.ActivityTypeNormalMonthCard:
		d = new(model.ActDataMonthCard)
	case define.ActivityTypeTheCompetition:
		d = new(model.ActDataTheCompetition)
	case define.ActivityTypeLadderRace:
		d = new(model.ActDataLadderRace)
	case define.ActivityTypeGoFish:
		d = new(model.ActDataGoFish)
	default:
		log.Error("ConvertDataType error:actData.Type:%v", actData.Type)
		return nil
	}

	b, err := json.Marshal(actData.Data)
	if err != nil {
		log.Error("UnmarshalActivityData error:%v", err)
		return nil
	}

	err = json.Unmarshal(b, d)
	if err != nil {
		log.Error("convert data type failed for activity type: %v,error:%v", actData.Type, err)
		return nil
	}

	return d
}

// =================================活动数据 END===========================================

// SavePlayerData 玩家数据落库
func SavePlayerData(key int64, data any) bool {
	rdb, _ := db.GetEngine(ServerId)

	b, err := json.Marshal(data)
	if err != nil {
		log.Error("save player activity data marshal error:%v", err)
		return false
	}

	playerId := key % define.ActivityPlayerDataBase
	baseKet := key - playerId
	actId := baseKet / define.ActivityPlayerDataBase
	_, err = rdb.RedisExec("HSET", fmt.Sprintf("%s:%d", define.ActivityPlayerRedisKey, actId), fmt.Sprintf("%d", playerId), b)
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
	key := actId*define.ActivityPlayerDataBase + playerId
	Cache.Set(key, pd)
}

// LoadPlayerData 获取活动对应玩家数据
func LoadPlayerData[T comparable](actId, playerId int64) T {
	key := actId*define.ActivityPlayerDataBase + playerId

	pd, ok := Cache.Get(key)
	if ok {
		return pd.(T)
	}

	var ret T

	rdb, _ := db.GetEngine(ServerId)

	bytes, err := redis.Bytes(rdb.RedisExec("HGET", fmt.Sprintf("%s:%d", define.ActivityPlayerRedisKey, actId), fmt.Sprintf("%d", playerId)))
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

// PurgeActivityPlayerData 删除活动所有对应玩家数据
func PurgeActivityPlayerData(actId int64) {
	rdb, _ := db.GetEngine(ServerId)

	_, err := rdb.RedisExec("DEL", fmt.Sprintf("%s:%d", define.ActivityPlayerRedisKey, actId))
	if err != nil {
		log.Error("PurgeActivityPlayerData error:%v", err)
	}
}
