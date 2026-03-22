package impl

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/logic/activity/data"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_game"
	"xfx/proto/proto_lineup"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
)

// ActivityArena 竞技场
type ActivityArena struct {
	BaseActivity
}

func (a *ActivityArena) OnInit() {

}

func (a *ActivityArena) OnStart() {
	//初始人机
	log.Debug("初始竞技场人机")
	arenaConf, ok := a.getArenaConf()
	if !ok {
		return
	}

	robots, error := a.Module().Invoke(define.ModuleCommon, "matchRobots", define.RobotMode_Arena, int64(arenaConf.RobotPower[0]), int64(arenaConf.RobotPower[1]), arenaConf.PowerCount)
	if error != nil {
		log.Debug("初始竞技场人机失败:%v", error)
		return
	}

	robot := robots.([]*model.Robot)
	for i := 0; i < len(robot); i++ {
		//上榜
		ctx := new(proto_player.Context)
		ctx.Id = int64(robot[i].Id)
		updateActivityRank(a, ctx, 0, int32(robot[i].Power), define.RankTypeArena)
	}
}

func (a *ActivityArena) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.ArenaOptionPd](a, ctx.Id)
	log.Debug("加载竞技场数据:%v", pd)
	nowUnix := utils.Now().Unix()

	if len(pd.PlayerIds) <= 0 {
		//初始敌人
		ids := a.OnRankIndex(ctx.Id)
		for i := 0; i < len(ids); i++ {
			id := utils.MustParseInt64(ids[i])
			pd.PlayerIds = append(pd.PlayerIds, id)
		}
	}

	//敌人
	enemys := make([]*proto_public.CommonPlayerInfo, 0, len(pd.PlayerIds))
	for _, v := range pd.PlayerIds {
		if v <= 0 {
			continue
		}

		//需要区别人机和真人
		if v <= define.PlayerIdBase {
			_enInfo := global.ToCommonPlayerByRobot(v)
			enemys = append(enemys, _enInfo)
		} else {
			_en := global.GetPlayerInfo(v)
			_enInfo := _en.ToCommonPlayer()
			enemys = append(enemys, _enInfo)
		}
	}

	//布阵
	lineUps := getLineUp(ctx.Id, pd.LineUp)

	if pd.LastRefreshTime == 0 {
		pd.LastRefreshTime = nowUnix
	}

	offseTime := pd.RefreshCD - nowUnix
	if offseTime <= 0 {
		offseTime = 0
	}

	if !utils.CheckIsSameDayBySec(pd.LastRefreshTime, nowUnix, 0) {
		pd.RefreshTime = 0
		pd.LastRefreshTime = nowUnix
		offseTime = 0
	}

	if pd.LastChallengeTime <= 0 {
		pd.LastChallengeTime = nowUnix
	}

	if !utils.CheckIsSameDayBySec(pd.LastChallengeTime, nowUnix, 0) {
		pd.ChallengeTime = 0
		pd.LastChallengeTime = nowUnix
	}

	//获取
	return &proto_activity.Arena{
		ChallengeTime: pd.ChallengeTime,
		RefreshTime:   pd.RefreshTime,
		RefreshCD:     int32(offseTime),
		LineUp:        lineUps,
		PlayerInfo:    enemys,
	}
}

func (a *ActivityArena) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "arena_refreshbattleplayer":
		a.RefreshBattlePlayer(ctx, params)
		break
	case "arena_lineup":
		a.SetArenaLineUp(ctx, params)
		break
	default:
	}
}

// 刷新对手
func (a *ActivityArena) RefreshBattlePlayer(ctx *proto_player.Context, params EventParams) {
	res := &proto_activity.S2CArenaRefreshBattlePlayer{}

	arenaConf, ok := a.getArenaConf()
	if !ok {
		log.Error("get activity typed config error:%v", a.GetCfgId())
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	pd := LoadPd[*model.ArenaOptionPd](a, ctx.Id)
	nowUnix := utils.Now().Unix()

	if pd.RefreshTime >= arenaConf.RefreshTime {
		if nowUnix <= pd.RefreshCD {
			res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
			invoke.Dispatch(a.Module(), ctx.Id, res)
			return
		}

		pd.RefreshCD = nowUnix + int64(arenaConf.RefreshCD)
	} else {
		pd.RefreshTime += 1
	}

	//刷新
	a.RangeOtherPlayer(ctx)

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))

	res.Code = proto_public.CommonErrorCode_ERR_OK
	invoke.Dispatch(a.Module(), ctx.Id, res)
}

func (a *ActivityArena) RangeOtherPlayer(ctx *proto_player.Context) {
	pd := LoadPd[*model.ArenaOptionPd](a, ctx.Id)
	//获取自己的排名
	serverId := ctx.Id / define.PlayerIdBase
	rankKey := fmt.Sprintf("%s:%d", define.RankTypeArenaKey, a.GetId())
	rankItem := getSelfRank(int(serverId), rankKey, ctx.Id)
	myIdStr := strconv.FormatInt(ctx.Id, 10)

	var result []string

	// 如果自己排名 <= 6，直接返回前6名（排除自己）
	if rankItem.Rank > 0 && rankItem.Rank <= 6 {
		// 获取前6名（排名1-6，索引0-5）
		var err error
		candidates, err := redis.Strings(db.RedisExec("ZREVRANGE", rankKey, 0, 5))
		if err != nil && !errors.Is(err, redis.ErrNil) {
			log.Error("获取前6名失败 error:%v", err)
			candidates = []string{}
		}

		// 排除自己
		result = filterOutPlayerID(candidates, myIdStr)
	} else if rankItem.Rank > 6 {
		// 自己排名 > 6，从比自己排名靠前的20名中随机选6个
		// 计算候选范围：从排名 max(1, 自己排名-20) 到 自己排名-1
		startRank := rankItem.Rank - 20
		if startRank < 1 {
			startRank = 1
		}
		endRank := rankItem.Rank - 1

		// 转换为Redis索引（从0开始）
		startIndex := startRank - 1
		endIndex := endRank - 1

		// 从Redis获取候选范围内的所有玩家ID
		candidates, err := redis.Strings(db.RedisExec("ZREVRANGE", rankKey, startIndex, endIndex))
		if err != nil && !errors.Is(err, redis.ErrNil) {
			log.Error("获取候选玩家失败 error:%v", err)
			candidates = []string{}
		}

		// 排除自己
		filteredCandidates := filterOutPlayerID(candidates, myIdStr)

		// 随机选择6个
		if len(filteredCandidates) <= 6 {
			// 候选数量不足6个，全部返回
			result = filteredCandidates
		} else {
			// 使用 MicsSlice 随机选6个
			result = utils.MicsSlice(filteredCandidates, 6)
		}
	} else {
		// 自己不在排行榜中，从最后一名开始向前取20名，随机选取6名
		var err error
		// 先获取排行榜总长度
		count, err := redis.Int64(db.RedisExec("ZCARD", rankKey))
		if err != nil && !errors.Is(err, redis.ErrNil) {
			log.Error("获取排行榜长度失败 error:%v", err)
			result = []string{}
		} else {
			// 从最后一名（分数最低）开始向前取20名，即获取分数最低的最多20名
			end := int64(19) // 索引19是第20个元素（从0开始）
			if count <= 20 {
				// 如果总数不足20名，就取全部
				end = count - 1
			}
			// 获取分数最低的最多20名玩家（ZRANGE按分数升序返回）
			candidates, err := redis.Strings(db.RedisExec("ZRANGE", rankKey, 0, end))
			if err != nil && !errors.Is(err, redis.ErrNil) {
				log.Error("获取后20名失败 error:%v", err)
				result = []string{}
			} else {
				// 排除自己
				filteredCandidates := filterOutPlayerID(candidates, myIdStr)

				// 随机选择6个
				if len(filteredCandidates) <= 6 {
					// 候选数量不足6个，全部返回
					result = filteredCandidates
				} else {
					// 使用 MicsSlice 随机选6个
					result = utils.MicsSlice(filteredCandidates, 6)
				}
			}
		}
	}

	// 将结果转换为 int64 并存储
	pd.PlayerIds = make([]int64, 0)
	for _, idStr := range result {
		id := utils.MustParseInt64(idStr)
		if id > 0 {
			pd.PlayerIds = append(pd.PlayerIds, id)
		}
	}
}

// 阵容调整
func (a *ActivityArena) SetArenaLineUp(ctx *proto_player.Context, params EventParams) {
	res := &proto_activity.S2CArenaSetLineUp{}
	req, ok := Key[*proto_activity.C2SArenaSetLineUp](params, "req")
	if !ok {
		log.Debug("act:%v", req.Index)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.BaseInfo.Module(), ctx.Id, res)
		return
	}

	if req.ActId != a.GetId() {
		log.Debug("act:%v, %v", req.ActId, a.GetId())
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.BaseInfo.Module(), ctx.Id, res)
		return
	}

	if req.Index <= 0 || req.Index > 5 {
		log.Debug("act:%v", req.Index)
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.BaseInfo.Module(), ctx.Id, res)
		return
	}

	pd := LoadPd[*model.ArenaOptionPd](a, ctx.Id)
	find := false
	for _, v := range pd.LineUp {
		if v.Index == req.Index {
			v.Id = req.HeroIds
			find = true
			break
		}
	}

	if !find {
		pd.LineUp = append(pd.LineUp, model.ArenaLineUpIds{
			Index: req.Index,
			Id:    req.HeroIds,
		})
	}

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))

	res.Code = proto_public.CommonErrorCode_ERR_OK
	invoke.Dispatch(a.Module(), ctx.Id, res)
}

// 获取布阵通过玩家id
func (a *ActivityArena) GetAreaLineUpByPlayerId(ctx *proto_player.Context, req *proto_activity.C2SArenaGetPlayerLineUp) (any, error) {
	if req.ActId != a.GetId() {
		return nil, errors.New("act no equip")
	}

	res := new(proto_activity.S2CArenaGetPlayerLineUp)

	//获取排名
	rankItem := getSelfRank(a.Module().GetApp().GetEnv().ID, fmt.Sprintf("%s:%d", define.RankTypeArenaKey, a.GetId()), req.PlayerId)
	//人机
	if req.PlayerId < define.PlayerIdBase {
		res.PlayerInfo = global.ToCommonPlayerByRobot(req.PlayerId)
		res.Rank = int32(rankItem.Rank)
		res.Code = proto_public.CommonErrorCode_ERR_OK
		return res, nil
	}

	_en := global.GetPlayerInfo(req.PlayerId)
	_enInfo := _en.ToCommonPlayer()
	res.Code = proto_public.CommonErrorCode_ERR_OK
	res.PlayerInfo = _enInfo
	res.Rank = int32(rankItem.Rank)

	//获取数据
	pd := LoadPd[*model.ArenaOptionPd](a, req.PlayerId)
	res.LineUp = getLineUp(req.PlayerId, pd.LineUp)
	return res, nil
}

// 组合布阵
func getLineUp(id int64, lineup []model.ArenaLineUpIds) []*proto_public.CommonPlayerLineUpInfo {
	lineUps := make([]*proto_public.CommonPlayerLineUpInfo, 0)
	if len(lineup) > 0 {
		_hero := global.GetPlayerHero(id)
		for _, v := range lineup {
			_lineTemp := new(proto_public.CommonPlayerLineUpInfo)
			_lineTemp.HeroId = make([]*proto_public.CommonPlayerLineUpItemInfo, 0)
			for _, k := range v.Id {
				if data, ok := _hero.Hero[k]; ok {
					_lineTemp.HeroId = append(_lineTemp.HeroId, &proto_public.CommonPlayerLineUpItemInfo{
						Id:    k,
						Level: data.Level,
						Star:  data.Star,
					})
				} else {
					_lineTemp.HeroId = append(_lineTemp.HeroId, &proto_public.CommonPlayerLineUpItemInfo{
						Id: k,
					})
				}
			}
			lineUps = append(lineUps, _lineTemp)
		}
	}
	return lineUps
}

// 战斗
func (a *ActivityArena) Battle(ctx *proto_player.Context, req *proto_activity.C2SArenaBattle) (*proto_activity.S2CArenaBattle, error) {
	res := new(proto_activity.S2CArenaBattle)
	//判断挑战玩家
	pd := LoadPd[*model.ArenaOptionPd](a, ctx.Id)
	if len(pd.LineUp) <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		return res, nil
	}

	isHave := false
	for _, v := range pd.PlayerIds {
		if v == req.Id {
			isHave = true
			break
		}
	}
	if !isHave {
		res.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		return res, nil
	}

	arenaConf, ok := a.getArenaConf()
	if !ok {
		log.Error("get activity typed config error:%v", a.GetCfgId())
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		return res, nil
	}

	if arenaConf.Id <= 0 {
		log.Error("get activity typed config error:%v", a.GetCfgId())
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		return res, nil
	}

	nowUnix := utils.Now().Unix()
	if pd.LastChallengeTime <= 0 {
		pd.LastChallengeTime = nowUnix
	}

	if !utils.CheckIsSameDayBySec(pd.LastChallengeTime, nowUnix, 0) {
		pd.ChallengeTime = 0
		pd.LastChallengeTime = nowUnix
	}

	//判断次数
	if pd.ChallengeTime >= arenaConf.ChallengeTime {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		return res, nil
	}

	pd.ChallengeTime++

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	res.Code = proto_public.CommonErrorCode_ERR_OK

	//获取
	return res, nil
}

// 战报
func (a *ActivityArena) BattleReport(ctx *proto_player.Context, req *proto_game.C2SChallengeBattleReport) (*proto_game.S2CChallengeBattleReport, error) {
	res := new(proto_game.S2CChallengeBattleReport)
	if req.WinId != ctx.Id {
		return res, nil
	}

	//刷新6位
	a.RangeOtherPlayer(ctx)
	//推送
	a.PushActivityData(ctx.Id, a.Format(ctx))
	return res, nil
}

func (a *ActivityArena) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	switch req := req.(type) {
	case *proto_activity.C2SArenaGetPlayerLineUp: //获取竞技场阵容
		return a.GetAreaLineUpByPlayerId(ctx, req)
	case *proto_lineup.C2SSetLineUp:
		if req.Type != define.LINEUP_ARENA {
			return nil, nil
		}

		pd := LoadPd[*model.ArenaOptionPd](a, ctx.Id)
		pd.LineUp = make([]model.ArenaLineUpIds, 0)

		// 推送活动数据
		a.PushActivityData(ctx.Id, a.Format(ctx))
	case *proto_activity.C2SArenaBattle: //挑战
		return a.Battle(ctx, req)
	case *proto_game.C2SChallengeBattleReport:
		return a.BattleReport(ctx, req)
	}
	return nil, nil
}

func (a *ActivityArena) Update(now time.Time) {
	// 跨天逻辑已迁移到 OnDayReset
}

// OnDayReset 跨天重置：重置所有玩家的挑战次数和刷新次数
func (a *ActivityArena) OnDayReset(now time.Time) {
	data.IterateActivityPlayerData[*model.ArenaOptionPd](a.GetId(), func(_ int64, pd *model.ArenaOptionPd) bool {
		if pd == nil {
			return true
		}
		pd.ChallengeTime = 0
		pd.RefreshTime = 0
		pd.RefreshCD = 0
		pd.LastChallengeTime = now.Unix()
		pd.LastRefreshTime = now.Unix()
		return true
	})
	log.Debug("ActivityArena OnDayReset: actId=%v", a.GetId())
}

func (a *ActivityArena) OnStop() {
	//活动结束补发奖励
	sendRankReward(a, define.RankTypeArena, nil)

	//删除排行榜
	deleteActivityRank(a, define.RankTypeArena)
}

func (a *ActivityArena) OnClose() {
}

func (a *ActivityArena) getArenaConf() (conf.ActArena, bool) {
	return FindTypedConf[conf.ActArena](a.GetCfgId(), config.ActArena.All(), func(arenaConf conf.ActArena) bool {
		return arenaConf.Id > 0
	})
}

func filterOutPlayerID(candidates []string, exclude string) []string {
	filtered := make([]string, 0, len(candidates))
	for _, id := range candidates {
		if id != exclude {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

// 获取
func (a *ActivityArena) OnRankIndex(Id int64) []string {
	// 获取最后6名的成员ID（分数最低的6个）
	result, err := redis.Strings(db.RedisExec("ZRANGE", fmt.Sprintf("%s:%d", define.RankTypeArenaKey, a.GetId()), 0, 5))
	if err != nil && !errors.Is(err, redis.ErrNil) {
		log.Error("获取最后4名失败 error:%v", err)
		return nil
	}

	// result 就是最后4名的ID列表
	// 如果是空列表，说明排行榜不足6人
	if len(result) == 0 {
		log.Info("排行榜为空或不足6人")
		return []string{}
	}

	return result
}

func init() {
	RegisterActivity(define.ActivityTypeArena, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityArena) },
		NewActivityData: func() any { return nil },
		NewPlayerData: func() any {
			return &model.ArenaOptionPd{
				PlayerIds: make([]int64, 0),
				LineUp:    make([]model.ArenaLineUpIds, 0),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.Arena = data.(*proto_activity.Arena)
		},
		InjectFunc:  func(handler IActivity, data any) {},
		ExtractFunc: func(handler IActivity) any { return nil },
	})
}
