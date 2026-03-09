package rank

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_rank"
)

// ReqRankingData 请求排行榜数据
func ReqRankingData(ctx global.IPlayer, pl *model.Player, req *proto_rank.C2SRankData) {
	result := new(proto_rank.S2CRankData)
	result.Type = int32(req.Type)
	log.Debug("请求排行榜:%v", req.Type)

	switch req.Type {
	case define.RankTypePerfect:
		guildId := invoke.GuildClient(ctx).GetPlayerGuildId(pl.Id)
		if guildId == 0 {
			ctx.Send(result)
			return
		}
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypePerfect, guildId}, "zrevrange", fmt.Sprintf("%s:%d", define.RankPerfectKey, guildId), 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeGrow:
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeGrow}, "zrevrange", define.RankGrowKey, 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeGuildBattle:
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeGuildBattle}, "zrevrange", define.RankGuildBattleKey, 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeDrawHero:
		if len(req.Id) <= 0 {
			ctx.Send(result)
			return
		}
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeDrawHero, int64(req.Id[0])}, "zrevrange", fmt.Sprintf("%s:%d", define.RankDrawHeroKey, req.Id[0]), 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeRecharge:
		if len(req.Id) <= 0 {
			ctx.Send(result)
			return
		}
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeRecharge, int64(req.Id[0])}, "zrevrange", fmt.Sprintf("%s:%d", define.RankRechargeKey, req.Id[0]), 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeClimbTower:
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeClimbTower}, "zrevrange", define.RankClimbTowerKey, 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeTheCompetition:
		if len(req.Id) <= 2 {
			ctx.Send(result)
			return
		}
		id := req.Id[0]
		group := req.Id[1]
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeTheCompetition, int64(id), int64(group)}, "zrevrange", fmt.Sprintf("%s:%d_%d", define.RankTypeTheCompetitionKey, id, group), 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeArena:
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeArena, int64(req.Id[0])}, "zrevrange", fmt.Sprintf("%s:%d", define.RankTypeArenaKey, req.Id[0]), 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeTianti:
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeTianti, int64(req.Id[0])}, "zrevrange", fmt.Sprintf("%s:%d", define.RankTypeTiantiKey, req.Id[0]), 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypeGoFish:
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypeGoFish, int64(req.Id[0])}, "zrevrange", fmt.Sprintf("%s:%d", define.RankTypeGoFishKey, req.Id[0]), 0, define.RankTop-1, "WITHSCORES")
	case define.RankTypePower:
		db.RedisAsyncExec(ctx.Self(), define.RedisRetRank, []int64{define.RankTypePower}, "zrevrange", fmt.Sprintf("%s", define.RankTypePowerKey), 0, define.RankTop-1, "WITHSCORES")
	default:
		log.Error("ReqRankingData type error:%v", req.Type)
	}
}

func GetSelfRank(serverId int, rankKey string, id int64) *proto_rank.RankItem {
	ret := new(proto_rank.RankItem)

	log.Debug("获取自己排名:%v,%v", rankKey, id)

	// 先检查用户是否在排行榜中
	score, err := redis.Float64(db.RedisExec("ZSCORE", rankKey, id))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			// 用户不在排行榜中
			return ret
		}
		log.Error("GetSelfRank score error:%v", err)
		return ret
	}

	// 获取排名
	rank, err := redis.Int64(db.RedisExec("ZREVRANK", rankKey, id))
	if err != nil {
		// 理论上如果ZSCORE成功，ZREVRANK也应该成功
		// 但为了安全起见，这里还是处理错误
		log.Error("GetSelfRank rank error:%v", err)
		return ret
	}

	ret.Score = int64(score)
	ret.Rank = rank + 1
	return ret
}

// updateGuildRank 更新排名
func updateGuildRank(ctx global.IPlayer, Id int32, pl *model.Player, score float64, rankType int) {
	log.Debug("刷新排名:类型: %v, 值:%v", rankType, score)
	var err error
	switch rankType {
	case define.RankTypePerfect:
		_, err = db.RedisExec("ZINCRBY", fmt.Sprintf("%s:%d", define.RankPerfectKey, Id), score, pl.Id)
	case define.RankTypeGrow:
		_, err = db.RedisExec("ZINCRBY", define.RankGrowKey, score, Id)
	case define.RankTypeGuildBattle:
		_, err = db.RedisExec("ZINCRBY", define.RankGuildBattleKey, score, Id)
	case define.RankTypeClimbTower:
		_, err = db.RedisExec("ZINCRBY", define.RankClimbTowerKey, score, pl.Id)
	case define.RankTypePower:
		_, err = db.RedisExec("ZINCRBY", define.RankTypePowerKey, score, pl.Id)
	default:
	}

	if err != nil {
		log.Error("updateRank error : %v", err)
	}

	//获取自己的排名
	rankItem := new(proto_rank.RankItem)
	switch rankType {
	case define.RankTypeClimbTower:
		rankItem = GetSelfRank(pl.Cache.App.GetEnv().ID, define.RankClimbTowerKey, pl.Id)
	default:
	}

	if rankItem.Rank > 0 {
		//通告相关
		internal.SyncNotice_RankUpdate(ctx, pl, rankType, rankItem.Rank)
	}
}

// updateZAddRank 更新排名
func updateZAddRank(ctx global.IPlayer, Id int64, pl *model.Player, score float64, rankType int) {
	log.Debug("刷新排名:类型: %v, 值:%v", rankType, score)
	var err error
	switch rankType {
	case define.RankTypeArena:
		_, err = db.RedisExec("ZADD", fmt.Sprintf("%s:%d", define.RankTypeArenaKey, Id), score, pl.Id)
	default:
	}

	if err != nil {
		log.Error("updateZAddRank error : %v", err)
	}

	//获取自己的排名
	rankItem := new(proto_rank.RankItem)
	switch rankType {
	case define.RankTypeArena:
		rankItem = GetSelfRank(pl.Cache.App.GetEnv().ID, define.RankTypeArenaKey, pl.Id)
	default:
	}

	if rankItem.Rank > 0 {
		//通告相关
		internal.SyncNotice_RankUpdate(ctx, pl, rankType, rankItem.Rank)
	}
}

func rankDataToProto(ctx global.IPlayer, pl *model.Player, rankType int, reply any, err error) []*proto_rank.RankItem {
	if err != nil {
		log.Error("rankDataToProto error : %v", err)
		return nil
	}

	if reply == nil {
		log.Debug("排行榜数据为空:%v", rankType)
		return nil
	}

	ret := make([]*proto_rank.RankItem, 0)
	res, _ := reply.([]interface{})
	switch rankType {
	case define.RankTypePerfect: //玩家排行榜
		ids := make([]int64, 0)
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			playerInfo := global.GetPlayerInfo(key)
			item := &proto_rank.RankItem{
				Rank:   int64(i) + 1,
				Score:  int64(score),
				Player: playerInfo.ToCommonPlayer(),
			}
			ids = append(ids, key)
			ret = append(ret, item)
		}
		playerInfos := invoke.GuildClient(ctx).GetPlayerInfoByIds(ids)
		for _, v := range ret {
			if info, ok := playerInfos[v.Player.PlayerId]; ok {
				v.Player.Position = info.Position
			}
		}
	case define.RankTypeGrow: // 成长值
		ids := make([]int64, 0)
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			item := &proto_rank.RankItem{
				Rank:  int64(i) + 1,
				Score: int64(score),
				Guild: &proto_rank.GuildRankItem{
					Id: key,
				},
			}

			ids = append(ids, key)
			ret = append(ret, item)
		}

		guildInfos := invoke.GuildClient(ctx).GetGuildRankInfoByIds(ids)
		for _, v := range ret {
			if info, ok := guildInfos[v.Guild.Id]; ok {
				v.Guild = info
			}
		}
	case define.RankTypeGuildBattle: // 工会战斗排行榜
		ids := make([]int64, 0)
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			item := &proto_rank.RankItem{
				Rank:  int64(i) + 1,
				Score: int64(score),
				Guild: &proto_rank.GuildRankItem{
					Id: key,
				},
			}
			ids = append(ids, key)
			ret = append(ret, item)
		}
		guildInfos := invoke.GuildClient(ctx).GetGuildRankInfoByIds(ids)
		for _, v := range ret {
			if info, ok := guildInfos[v.Guild.Id]; ok {
				v.Guild = info
			}
		}
	case define.RankTypeDrawHero: // 招募英雄排行榜
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			playerInfo := global.GetPlayerInfo(key)
			item := &proto_rank.RankItem{
				Rank:   int64(i) + 1,
				Score:  int64(score),
				Player: playerInfo.ToCommonPlayer(),
			}
			item.Player.Power = playerInfo.Power
			ret = append(ret, item)
		}
	case define.RankTypeRecharge: // 充值排行榜
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			playerInfo := global.GetPlayerInfo(key)
			item := &proto_rank.RankItem{
				Rank:   int64(i) + 1,
				Score:  int64(score),
				Player: playerInfo.ToCommonPlayer(),
			}
			ret = append(ret, item)
		}
	case define.RankTypeClimbTower: // 爬塔排行榜
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			playerInfo := global.GetPlayerInfo(key)
			item := &proto_rank.RankItem{
				Rank:   int64(i) + 1,
				Score:  int64(score),
				Player: playerInfo.ToCommonPlayer(),
			}
			ret = append(ret, item)
		}
	case define.RankTypeTheCompetition: // 巅峰决斗排行榜
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			playerInfo := global.GetPlayerInfo(key)
			item := &proto_rank.RankItem{
				Rank:   int64(i) + 1,
				Score:  int64(score),
				Player: playerInfo.ToCommonPlayer(),
			}
			ret = append(ret, item)
		}
	case define.RankTypeArena: //竞技场排行榜
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			item := &proto_rank.RankItem{
				Rank:  int64(i) + 1,
				Score: int64(score),
			}

			//人机
			if key < define.PlayerIdBase {
				item.Player = global.ToCommonPlayerByRobot(key)
				item.Player.ServerId = int32(pl.Cache.App.GetEnv().ID)
			} else {
				playerInfo := global.GetPlayerInfo(key)
				item.Player = playerInfo.ToCommonPlayer()
			}
			ret = append(ret, item)
		}
	case define.RankTypeTianti: //天梯排行榜
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			item := &proto_rank.RankItem{
				Rank:  int64(i) + 1,
				Score: int64(score),
			}

			//人机
			if key < define.PlayerIdBase {
				item.Player = global.ToCommonPlayerByRobot(key)
				item.Player.ServerId = int32(pl.Cache.App.GetEnv().ID)
			} else {
				playerInfo := global.GetPlayerInfo(key)
				item.Player = playerInfo.ToCommonPlayer()
			}
			ret = append(ret, item)
		}
	case define.RankTypeGoFish: //钓鱼排行榜
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			item := &proto_rank.RankItem{
				Rank:  int64(i) + 1,
				Score: int64(score),
			}

			//人机
			if key < define.PlayerIdBase {
				item.Player = global.ToCommonPlayerByRobot(key)
				item.Player.ServerId = int32(pl.Cache.App.GetEnv().ID)
			} else {
				playerInfo := global.GetPlayerInfo(key)
				item.Player = playerInfo.ToCommonPlayer()
			}
			ret = append(ret, item)
		}
	case define.RankTypePower:
		for i := 0; i < len(res)/2; i++ {
			str := string(res[i*2].([]byte))
			key, _ := strconv.ParseInt(str, 10, 64)
			score, _ := strconv.ParseFloat(string(res[i*2+1].([]byte)), 64)

			item := &proto_rank.RankItem{
				Rank:  int64(i) + 1,
				Score: int64(score),
			}

			//人机
			if key < define.PlayerIdBase {
				item.Player = global.ToCommonPlayerByRobot(key)
				item.Player.ServerId = int32(pl.Cache.App.GetEnv().ID)
			} else {
				playerInfo := global.GetPlayerInfo(key)
				item.Player = playerInfo.ToCommonPlayer()
			}
			ret = append(ret, item)
		}
	default:
		log.Error("unknown rank type %v", rankType)
	}

	return ret
}

// 更新战力排行榜
func UpdatePowerRank(ctx global.IPlayer, pl *model.Player, power int64) {
	//最终值 = 当前杯数 + (默认1个1和9个9去减 - 当前时间)(时间越早 数值越大) 杯数一般最大就是99999
	uTime, _ := strconv.ParseFloat(fmt.Sprintf("0.%d", 1999999999-utils.Now().Unix()), 64)
	_finalAmount := float64(power) + uTime

	updateGuildRank(ctx, 0, pl, _finalAmount, define.RankTypePower)
}

// 更新帮会完美榜单
func UpdateGuildPerfectRank(ctx global.IPlayer, guildId int32, pl *model.Player, power int32) {
	//最终值 = 当前杯数 + (默认1个1和9个9去减 - 当前时间)(时间越早 数值越大) 杯数一般最大就是99999
	uTime, _ := strconv.ParseFloat(fmt.Sprintf("0.%d", 1999999999-utils.Now().Unix()), 64)
	_finalAmount := float64(power) + uTime

	updateGuildRank(ctx, guildId, pl, _finalAmount, define.RankTypePerfect)
}

// 更新帮会成长值榜单
func UpdateGuildGrowRank(ctx global.IPlayer, guildId int32, pl *model.Player, power int32) {
	//最终值 = 当前杯数 + (默认1个1和9个9去减 - 当前时间)(时间越早 数值越大) 杯数一般最大就是99999
	uTime, _ := strconv.ParseFloat(fmt.Sprintf("0.%d", 1999999999-utils.Now().Unix()), 64)
	_finalAmount := float64(power) + uTime

	updateGuildRank(ctx, guildId, pl, _finalAmount, define.RankTypeGrow)
}

// 更新帮会战力榜单
func UpdateGuildBattleRank(ctx global.IPlayer, guildId int32, pl *model.Player, power int32) {
	//最终值 = 当前杯数 + (默认1个1和9个9去减 - 当前时间)(时间越早 数值越大) 杯数一般最大就是99999
	uTime, _ := strconv.ParseFloat(fmt.Sprintf("0.%d", 1999999999-utils.Now().Unix()), 64)
	_finalAmount := float64(power) + uTime

	updateGuildRank(ctx, guildId, pl, _finalAmount, define.RankTypeGuildBattle)
}

// 更新爬塔
func UpdateClimbTowerRank(ctx global.IPlayer, pl *model.Player, power int32) {
	//最终值 = 当前杯数 + (默认1个1和9个9去减 - 当前时间)(时间越早 数值越大) 杯数一般最大就是99999
	uTime, _ := strconv.ParseFloat(fmt.Sprintf("0.%d", 1999999999-utils.Now().Unix()), 64)
	_finalAmount := float64(power) + uTime

	updateGuildRank(ctx, 0, pl, _finalAmount, define.RankTypeClimbTower)
}

// 更新竞技场
func UpdateArenaRank(ctx global.IPlayer, ActId int64, pl *model.Player, power int32) {
	//最终值 = 当前杯数 + (默认1个1和9个9去减 - 当前时间)(时间越早 数值越大) 杯数一般最大就是99999
	uTime, _ := strconv.ParseFloat(fmt.Sprintf("0.%d", 1999999999-utils.Now().Unix()), 64)
	_finalAmount := float64(power) + uTime

	updateZAddRank(ctx, ActId, pl, _finalAmount, define.RankTypeArena)
}

func OnRetRankData(ctx global.IPlayer, pl *model.Player, ret *db.RedisRet) {
	resp := new(proto_rank.S2CRankData)
	resp.Type = int32(ret.Params[0])

	switch ret.Params[0] {
	case define.RankTypePerfect:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, fmt.Sprintf("%s:%d", define.RankPerfectKey, ret.Params[1]), pl.Id)
		log.Debug("获取工会完美排行榜 self:%v", resp)
	case define.RankTypeGrow:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)

		guildRank := invoke.GuildClient(ctx).GetGuildRankInfoByPlayerId(pl.Id)
		if guildRank != nil {
			resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, define.RankGrowKey, guildRank.Id)
			resp.Self.Guild = guildRank
		}
		log.Debug("获取工会成长排行榜 self:%v", resp)
	case define.RankTypeGuildBattle:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)

		guildRank := invoke.GuildClient(ctx).GetGuildRankInfoByPlayerId(pl.Id)
		if guildRank != nil {
			resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, define.RankGuildBattleKey, guildRank.Id)
			resp.Self.Guild = guildRank
		}
		log.Debug("获取工会战斗排行榜:%v", resp)
	case define.RankTypeDrawHero:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, fmt.Sprintf("%s:%d", define.RankDrawHeroKey, ret.Params[1]), pl.Id)
		log.Debug("获取招募英雄排行榜:%v", resp)
	case define.RankTypeRecharge:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, fmt.Sprintf("%s:%d", define.RankRechargeKey, ret.Params[1]), pl.Id)
		log.Debug("获取充值排行榜:%v", resp)
	case define.RankTypeClimbTower:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, define.RankClimbTowerKey, pl.Id)
		log.Debug("获取爬塔排行榜:%v", resp)
	case define.RankTypeTheCompetition:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, fmt.Sprintf("%s:%d_%d", define.RankTypeTheCompetitionKey, ret.Params[1], ret.Params[2]), pl.Id)
		log.Debug("获取巅峰决斗排行榜:%v", resp)
	case define.RankTypeArena:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, fmt.Sprintf("%s:%d", define.RankTypeArenaKey, ret.Params[1]), pl.Id)
		log.Debug("获取竞技场排行榜:%v", resp)
	case define.RankTypeTianti:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, fmt.Sprintf("%s:%d", define.RankTypeTiantiKey, ret.Params[1]), pl.Id)
		log.Debug("获取天梯排行榜:%v", resp)
	case define.RankTypeGoFish:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, fmt.Sprintf("%s:%d", define.RankTypeGoFishKey, ret.Params[1]), pl.Id)
		log.Debug("获取钓鱼排行榜:%v", resp)
	case define.RankTypePower:
		resp.Rankings = rankDataToProto(ctx, pl, int(resp.Type), ret.Reply, ret.Err)
		resp.Self = GetSelfRank(pl.Cache.App.GetEnv().ID, fmt.Sprintf("%s", define.RankTypePowerKey), pl.Id)
		log.Debug("获取战力排行榜:%v", resp)
	default:
	}
	ctx.Send(resp)
}

// dbId|heroId|quality
func encodeHeroInfo(v1, v2, v3 int64) int64 {
	v := v1 << 34
	v |= (v2 & 0x3fffffff) << 4
	return v | (v3 & 0xF)
}

// dbId|heroId|quality
func decodeHeroInfo(v int64) (int64, int64, int64) {
	v1 := v >> 34
	v2 := (v >> 4) & 0x3FFFFFFF
	v3 := v & 0xF
	return v1, v2, v3
}

// DeleteRankData 删除排行榜
func DeleteRankData() {
	// 排行榜
	//global.ServerG.GetDBEngine().Request(nil, 0, 0, "ZREM", "rank_hero", heroId)
}
