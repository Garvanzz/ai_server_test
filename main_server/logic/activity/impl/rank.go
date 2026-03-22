package impl

import (
	"errors"
	"fmt"
	"strconv"
	"xfx/core/config"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_player"
	"xfx/proto/proto_rank"

	"github.com/gomodule/redigo/redis"
)

// updateActivityRank 更新排名
func updateActivityRank(a BaseInfo, ctx *proto_player.Context, Id int32, score int32, rankType int) {
	//最终值 = 当前杯数 + (默认1个1和9个9去减 - 当前时间)(时间越早 数值越大) 杯数一般最大就是99999
	uTime, _ := strconv.ParseFloat(fmt.Sprintf("0.%d", 1999999999-utils.Now().Unix()), 64)
	_finalAmount := float64(score) + uTime

	var err error
	switch rankType {
	case define.RankTypeDrawHero:
		_, err = db.RedisExec("ZINCRBY", fmt.Sprintf("%s:%d", define.RankDrawHeroKey, a.GetId()), _finalAmount, ctx.Id)
	case define.RankTypeRecharge:
		_, err = db.RedisExec("ZINCRBY", fmt.Sprintf("%s:%d", define.RankRechargeKey, a.GetId()), _finalAmount, ctx.Id)
	case define.RankTypeTheCompetition:
		_, err = db.RedisExec("ZINCRBY", fmt.Sprintf("%s:%d_%d", define.RankTypeTheCompetitionKey, a.GetId(), Id), _finalAmount, ctx.Id)
	case define.RankTypeGoFish:
		_, err = db.RedisExec("ZINCRBY", fmt.Sprintf("%s:%d", define.RankTypeGoFishKey, a.GetId()), _finalAmount, ctx.Id)
	case define.RankTypeArena:
		_, err = db.RedisExec("ZINCRBY", fmt.Sprintf("%s:%d", define.RankTypeArenaKey, a.GetId()), _finalAmount, ctx.Id)
	case define.RankTypeTianti:
		_, err = db.RedisExec("ZINCRBY", fmt.Sprintf("%s:%d", define.RankTypeTiantiKey, a.GetId()), _finalAmount, ctx.Id)
	default:
	}

	if err != nil {
		log.Error("updateRank error : %v", err)
	}

	//人机
	if ctx.Id < define.PlayerIdBase {
		return
	}

	serverId := ctx.Id / define.PlayerIdBase
	//通知
	rankItem := new(proto_rank.RankItem)
	switch rankType {
	case define.RankTypeDrawHero:
		rankItem = getSelfRank(int(serverId), fmt.Sprintf("%s:%d", define.RankDrawHeroKey, a.GetId()), ctx.Id)
	case define.RankTypeRecharge:
		rankItem = getSelfRank(int(serverId), fmt.Sprintf("%s:%d", define.RankRechargeKey, a.GetId()), ctx.Id)
	case define.RankTypeTheCompetition:
		rankItem = getSelfRank(int(serverId), fmt.Sprintf("%s:%d_%d", define.RankTypeTheCompetitionKey, a.GetId(), Id), ctx.Id)
	case define.RankTypeGoFish:
		rankItem = getSelfRank(int(serverId), fmt.Sprintf("%s:%d", define.RankTypeGoFishKey, a.GetId()), ctx.Id)
	case define.RankTypeArena:
		rankItem = getSelfRank(int(serverId), fmt.Sprintf("%s:%d", define.RankTypeArenaKey, a.GetId()), ctx.Id)
	case define.RankTypeTianti:
		rankItem = getSelfRank(int(serverId), fmt.Sprintf("%s:%d", define.RankTypeTiantiKey, a.GetId()), ctx.Id)
	default:
	}

	if rankItem.Rank > 0 {
		//通告相关
		global.SyncNotice_RankUpdate(a.Module(), ctx, rankType, rankItem.Rank)
	}
}

func getSelfRank(serverId int, rankKey string, id int64) *proto_rank.RankItem {
	ret := new(proto_rank.RankItem)

	// 先检查用户是否在排行榜中
	score, err := redis.Float64(db.RedisExec("ZSCORE", rankKey, id))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			// 用户不在排行榜中
			return ret
		}
		log.Error("getSelfRank score error:%v", err)
		return ret
	}

	// 获取排名
	rank, err := redis.Int64(db.RedisExec("ZREVRANK", rankKey, id))
	if err != nil {
		// 理论上如果ZSCORE成功，ZREVRANK也应该成功
		// 但为了安全起见，这里还是处理错误
		log.Error("getSelfRank rank error:%v", err)
		return ret
	}

	ret.Score = int64(score)
	ret.Rank = rank + 1
	return ret
}

// 清空排行榜
func deleteActivityRank(a BaseInfo, rankType int) {
	key, ok := define.RankTypeToKey[rankType]
	if !ok {
		log.Error("deleteActivityRank error : rankType not exist, rankType=%v", rankType)
		return
	}

	_, err := db.RedisExec("DEL", fmt.Sprintf("%s:%d", key, a.GetId()))
	if err != nil {
		log.Error("deleteActivityRank error : %v", err)
	}
}

// 发送排行榜奖励
func sendRankReward(a BaseInfo, rankType int, ignoreList []int64) {

	key, ok := define.RankTypeToKey[rankType]
	if !ok {
		log.Error("sendRankReward error : rankType not exist, rankType=%v", rankType)
		return
	}

	rankKey := fmt.Sprintf("%s:%d", key, a.GetId())

	reply, err := db.RedisExec("zrevrange", rankKey, 0, define.RankTop-1)
	if err != nil {
		log.Error("sendRankReward db error : %v", err)
		return
	}

	res, _ := reply.([]interface{})
	ids := make([]int64, 0)
	for i := 0; i < len(res); i++ {
		id, _ := strconv.ParseInt(string(res[i].([]byte)), 10, 64)
		ids = append(ids, id)
	}

	ignore := make(map[int64]struct{})
	for _, id := range ignoreList {
		ignore[id] = struct{}{}
	}

	rankAwardConfs := config.RankAward.All()
	for _, rankAwardConf := range rankAwardConfs {
		if int(rankAwardConf.Type) == rankType {

			start := rankAwardConf.Rank[0]
			end := rankAwardConf.Rank[1]

			l := getSubArray(ids, int(start), int(end))

			sendList := make([]int64, 0)
			for _, id := range l {
				if _, ok = ignore[id]; ok {
					continue
				}

				sendList = append(sendList, id)
			}

			if len(sendList) == 0 {
				continue
			}

			// 发邮件
			cfgId := int32(0)
			switch rankType {
			case define.RankTypeDrawHero:
				cfgId = 1
			case define.RankTypeRecharge:
				cfgId = 2
			case define.RankTypeTheCompetition:
				cfgId = 3
			default:
				log.Error("sendRankReward mail cfg id error:%v", rankType)
				return
			}

			ok = invoke.MailClient(a.Module()).SendMail(define.PlayerMail, "", "", "", "", "", rankAwardConf.Award, sendList, int64(0), cfgId, false, []string{})
			if !ok {
				log.Error("sendRankReward error : %v", err)
				return
			}
		}
	}
}

// TODO:
// dailyaccrecharge 每日累充 每天23 59 补发邮件，没有领取的
//  normalmonthcard 月卡时间内每日要发邮件
