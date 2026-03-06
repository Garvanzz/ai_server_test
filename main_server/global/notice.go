package global

import (
	"time"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"
)

// SyncHorse 同步跑马灯
func SyncHorse(ctx IPlayer, pl *model.Player, conf conf2.BroadCast, param []int32) {
	res := &proto_public.S2CHorseOption{
		Id:         conf.Id,
		Channel:    int32(0),
		ServerId:   int32(pl.Cache.App.GetEnv().ID),
		EffectTime: time.Now().Unix(),
		ExpireTime: time.Now().Unix() + 20,
		Value:      param,
		PlayerInfo: GetPlayerInfo(pl.Id).ToCommonPlayer(),
	}
	log.Debug("通知跑马灯:%v", res)
	invoke.DispatchAllPlayer(ctx, res)
}

// SyncHorse 同步跑马灯通过id
func SyncHorseById(m invoke.Invoker, pid int64, conf conf2.BroadCast, param []int32) {
	serverId := pid / define.PlayerIdBase
	res := &proto_public.S2CHorseOption{
		Id:         conf.Id,
		Channel:    int32(0),
		ServerId:   int32(serverId),
		EffectTime: time.Now().Unix(),
		ExpireTime: time.Now().Unix() + 20,
		Value:      param,
		PlayerInfo: GetPlayerInfo(pid).ToCommonPlayer(),
	}
	log.Debug("通知跑马灯:%v", res)
	invoke.DispatchAllPlayer(m, res)
}

// 通告相关-排名更新
func SyncNotice_RankUpdate(m invoke.Invoker, pl *proto_player.Context, rankType int, rankIndex int64) {
	//跑马灯
	confs := config.BroadCast.All()
	for _, v := range confs {
		if v.Type == define.HorseType_RankUpdate {
			condition := v.Condition
			params := v.Param
			pass := true
			param := make([]int32, 0)
			for index, v := range condition {
				switch v {
				case define.HorseType_Condition_PlayerName:
					param = append(param, 0)
					break
				case define.HorseType_Condition_RankIndex:
					_index := params[index]
					if int64(_index) >= rankIndex {
						param = append(param, int32(rankIndex))
					} else {
						pass = false
					}
					break
				case define.HorseType_Condition_RankType:
					_type := params[index]
					if _type == int32(rankType) {
						param = append(param, int32(rankType))
					} else {
						pass = false
					}
					break
				default:
					pass = false
					break
				}

				if !pass {
					break
				}
			}

			if !pass {
				continue
			}

			//发送
			SyncHorseById(m, pl.Id, v, param)
		}
	}
}
