package cdkey

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_item"
	"xfx/proto/proto_public"

	"github.com/gomodule/redigo/redis"
)

func Init(pl *model.Player) {
	pl.Cdkey = &model.PlayerCdkey{
		UsedKeys: make(map[string]int32),
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load cdkey error, no this server:%v", err)
		return
	}

	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerCdkey, pl.Id))
	if err != nil {
		log.Error("player[%v],load cdkey error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.PlayerCdkey)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load cdkey unmarshal error:%v", pl.Id, err)
	}

	pl.Cdkey = m
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Cdkey)
	if err != nil {
		log.Error("player[%v],save cdkey marshal error:%v", pl.Id, err)
		return
	}

	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save cdkey error, no this server:%v", err)
		return
	}

	if isSync {
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerCdkey, pl.Id), j)
	} else {
		rdb.RedisAsyncExec(pl.Cache.Self, define.RedisRetNone, nil, "SET", fmt.Sprintf("%s:%d", define.PlayerCdkey, pl.Id), j)
	}
}

// ReqExchangeCDKey 兑换码兑换
func ReqExchangeCDKey(ctx global.IPlayer, pl *model.Player, req *proto_item.C2SExchangeCDKey) {
	res := &proto_item.S2CExchangeCDKey{}

	// 参数校验
	if req.Code == "" {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(res)
		return
	}

	code := strings.TrimSpace(req.Code)

	// 查找兑换码配置
	cdkeyConfs := config.Cdkey.All()
	var targetConf *conf2.Cdkey
	for _, conf := range cdkeyConfs {
		for _, key := range conf.Keys {
			if strings.TrimSpace(key) == code {
				targetConf = &conf
				break
			}
		}
		if targetConf != nil {
			break
		}
	}

	if targetConf == nil {
		res.Code = proto_public.CommonErrorCode_ERR_CDKeyInvaild
		ctx.Send(res)
		return
	}

	// 检查时间有效性
	now := utils.Now()
	if targetConf.StartTime != "" {
		startTime, err := time.ParseInLocation("2006-01-02 15:04:05", targetConf.StartTime, time.Local)
		if err == nil && now.Before(startTime) {
			res.Code = proto_public.CommonErrorCode_ERR_CDKeyExpireCode
			ctx.Send(res)
			return
		}
	}
	if targetConf.EndTime != "" {
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", targetConf.EndTime, time.Local)
		if err == nil && now.After(endTime) {
			res.Code = proto_public.CommonErrorCode_ERR_CDKeyExpireCode
			ctx.Send(res)
			return
		}
	}

	// 获取Redis连接
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("ReqExchangeCDKey get db engine error:%v", err)
		res.Code = proto_public.CommonErrorCode_ERR_REDISERROR
		ctx.Send(res)
		return
	}

	// 非通用兑换码检查玩家个人数据
	if !targetConf.Iscommon {
		if count, ok := pl.Cdkey.UsedKeys[code]; ok && count >= targetConf.Count {
			res.Code = proto_public.CommonErrorCode_ERR_CDKeyTodayGet
			ctx.Send(res)
			return
		}
	} else {
		// 通用兑换码检查全局次数限制
		if targetConf.Count > 0 {
			// 获取当前全局使用次数
			globalCountKey := fmt.Sprintf("%s:%s", define.CommonCdkey, code)
			globalCount, err := redis.Int(rdb.RedisExec("GET", globalCountKey))
			if err != nil && err != redis.ErrNil {
				log.Error("ReqExchangeCDKey get global count error:%v", err)
				res.Code = proto_public.CommonErrorCode_ERR_REDISERROR
				ctx.Send(res)
				return
			}
			if int32(globalCount) >= targetConf.Count {
				res.Code = proto_public.CommonErrorCode_ERR_CDKeyExpireCode
				ctx.Send(res)
				return
			}
		}
	}

	// 发放奖励
	awards := make([]conf2.ItemE, 0, len(targetConf.Awards))
	items := make([]*proto_public.Item, 0, len(targetConf.Awards))
	for _, award := range targetConf.Awards {
		awards = append(awards, award)
		items = append(items, &proto_public.Item{
			ItemId:   award.ItemId,
			ItemNum:  award.ItemNum,
			ItemType: award.ItemType,
		})
	}
	log.Debug("兑换码奖励:%v", awards)
	if len(awards) > 0 {
		bag.AddAward(ctx, pl, awards, true)
	}

	// 标记已使用
	if !targetConf.Iscommon {
		// 非通用：记录到玩家个人数据
		pl.Cdkey.UsedKeys[code]++
		Save(pl, false)
	} else if targetConf.Count > 0 {
		// 通用且有次数限制：增加全局计数
		globalCountKey := fmt.Sprintf("%s:%s", define.CommonCdkey, code)
		_, err := rdb.RedisExec("INCR", globalCountKey)
		if err != nil {
			log.Error("ReqExchangeCDKey incr global count error:%v", err)
		}
	}

	res.Code = proto_public.CommonErrorCode_ERR_OK
	res.Items = items
	res.CdKeys = &proto_item.CDKeyOption{
		Code:       code,
		Count:      0,
		ExpireTime: 0,
	}
	ctx.Send(res)

	log.Debug("player[%v] exchange cdkey success, code:%v", pl.Id, code)
}

// ReqInitCDKey 初始化兑换码数据（获取已使用的兑换码列表）
func ReqInitCDKey(ctx global.IPlayer, pl *model.Player, req *proto_item.C2SInitCDKey) {
	res := &proto_item.S2CInitCDKey{}

	cdKeys := make([]*proto_item.CDKeyOption, 0)
	cdkeyConfs := config.Cdkey.All()

	// 获取Redis连接
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("ReqInitCDKey get db engine error:%v", err)
	}

	// 非通用兑换码：从玩家数据中获取
	for code, count := range pl.Cdkey.UsedKeys {
		if count <= 0 {
			continue
		}
		// 查找该兑换码对应的配置
		for _, conf := range cdkeyConfs {
			if !conf.Iscommon {
				for _, key := range conf.Keys {
					if strings.TrimSpace(key) == code {
						// 获取过期时间
						var expireTime int64
						if conf.EndTime != "" {
							endTime, err := time.ParseInLocation("2006-01-02 15:04:05", conf.EndTime, time.Local)
							if err == nil {
								expireTime = endTime.Unix()
							}
						}
						cdKeys = append(cdKeys, &proto_item.CDKeyOption{
							Code:       code,
							Count:      count,
							ExpireTime: expireTime,
						})
						break
					}
				}
			}
		}
	}

	// 通用兑换码且有次数限制：从Redis获取全局使用次数
	for _, conf := range cdkeyConfs {
		if conf.Iscommon && conf.Count > 0 {
			for _, key := range conf.Keys {
				code := strings.TrimSpace(key)
				globalCountKey := fmt.Sprintf("%s:%s", define.CommonCdkey, code)
				globalCount, _ := redis.Int(rdb.RedisExec("GET", globalCountKey))

				// 获取过期时间
				var expireTime int64
				if conf.EndTime != "" {
					endTime, err := time.ParseInLocation("2006-01-02 15:04:05", conf.EndTime, time.Local)
					if err == nil {
						expireTime = endTime.Unix()
					}
				}

				cdKeys = append(cdKeys, &proto_item.CDKeyOption{
					Code:       code,
					Count:      int32(globalCount),
					ExpireTime: expireTime,
				})
			}
		}
	}

	res.CdKeys = cdKeys
	ctx.Send(res)
}
