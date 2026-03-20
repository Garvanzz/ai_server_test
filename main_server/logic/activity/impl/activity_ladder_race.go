package impl

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
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

// ActivityLadderRace 天梯
type ActivityLadderRace struct {
	BaseActivity
	data *model.ActDataLadderRace
}

func (a *ActivityLadderRace) OnInit() {
}

func (a *ActivityLadderRace) OnStart() {
	if a.data == nil {
		a.data = new(model.ActDataLadderRace)
		a.data.RankPlayer = make(map[int64]*model.ActDataLadderRaceRankPlayer)
	}
	a.data.Season = 1
	//初始人机
	log.Debug("初始竞技场人机")
	ladderRaceConfs, ok := GetTypedConf[conf.ActLadderRace](a.GetCfgId(), config.ActLadderRace.All())
	if !ok {
		return
	}

	var conf conf.ActLadderRace
	for _, v := range ladderRaceConfs {
		if v.Season == a.data.Season {
			conf = v
			break
		}
	}
	if conf.Id <= 0 {
		return
	}

	a.data.SeasonTime = utils.Now().Unix() + int64(conf.SeasonTime*86400)

	robots, error := a.Module().Invoke(define.ModuleCommon, "matchRobots", define.RobotMode_Tianti, int64(conf.RobotPower[0]), int64(conf.RobotPower[1]), conf.PowerCount)
	if error != nil {
		log.Debug("初始天梯人机失败:%v", error)
		return
	}

	robot := robots.([]*model.Robot)
	//根据战力排序
	sort.Slice(robot, func(i, j int) bool {
		return robot[i].Power < robot[j].Power
	})

	for i := int32(len(robot) - 1); i >= 0; i-- {
		//上榜
		ctx := new(proto_player.Context)
		ctx.Id = int64(robot[i].Id)
		updateActivityRank(a, ctx, 0, i*conf.BasicScore, define.RankTypeTianti)
	}

	if a.data.RankPlayer == nil {
		a.data.RankPlayer = make(map[int64]*model.ActDataLadderRaceRankPlayer)
	}
	// 初始化RankPlayer
	a.initRankPlayerFromRedis()
}

func (a *ActivityLadderRace) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.LadderRacePd](a, ctx.Id)
	log.Debug("加载天梯数据:%s", pd)

	if pd.LastChallengeTime <= 0 {
		pd.LastChallengeTime = utils.Now().Unix()
	}

	if !utils.CheckIsSameDayBySec(pd.LastChallengeTime, utils.Now().Unix(), 0) {
		pd.ChallengeTime = 0
		pd.LastChallengeTime = utils.Now().Unix()
	}

	//布阵
	lineUps := make([]*proto_public.CommonPlayerLineUpInfo, 0)
	if len(pd.LineUp) > 0 {
		_hero := global.GetPlayerHero(ctx.Id)
		for _, v := range pd.LineUp {
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

	//获取
	return &proto_activity.LadderRace{
		LineUp:     lineUps,
		Score:      pd.Score,
		Season:     a.data.Season,
		SeasonTime: a.data.SeasonTime,
	}
}

func (a *ActivityLadderRace) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "tianti_lineup":
		a.SetTiantiLineUp(ctx, params)
		break
	default:
	}
}

// 阵容调整
func (a *ActivityLadderRace) SetTiantiLineUp(ctx *proto_player.Context, params EventParams) {
	res := &proto_activity.S2CLadderRaceSetLineUp{}
	req, ok := Key[*proto_activity.C2SLadderRaceSetLineUp](params, "req")
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

	pd := LoadPd[*model.LadderRacePd](a, ctx.Id)
	find := false
	for _, v := range pd.LineUp {
		if v.Index == req.Index {
			v.Id = req.HeroIds
			find = true
			break
		}
	}

	if !find {
		pd.LineUp = append(pd.LineUp, model.LadderRaceIds{
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
func (a *ActivityLadderRace) GetLadderRaceLineUpByPlayerId(ctx *proto_player.Context, req *proto_activity.C2SLadderRaceGetPlayerLineUp) (any, error) {
	if req.ActId != a.GetId() {
		return nil, errors.New("act no equip")
	}

	res := new(proto_activity.S2CLadderRaceGetPlayerLineUp)

	//获取排名
	rankItem := getSelfRank(a.Module().GetApp().GetEnv().ID, fmt.Sprintf("%s:%d", define.RankTypeTianti, a.GetId()), req.PlayerId)
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
	pd := LoadPd[*model.LadderRacePd](a, req.PlayerId)
	res.LineUp = getLadderRaceLineUp(req.PlayerId, pd.LineUp)
	return res, nil
}

// 组合布阵
func getLadderRaceLineUp(id int64, lineup []model.LadderRaceIds) []*proto_public.CommonPlayerLineUpInfo {
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
func (a *ActivityLadderRace) Battle(ctx *proto_player.Context, req *proto_activity.C2SLadderRaceBattle) (*proto_activity.S2CLadderRaceBattle, error) {
	res := new(proto_activity.S2CLadderRaceBattle)
	//判断挑战玩家
	pd := LoadPd[*model.LadderRacePd](a, ctx.Id)

	//判断跳转次数
	if pd.LastChallengeTime <= 0 {
		pd.LastChallengeTime = utils.Now().Unix()
	}

	if !utils.CheckIsSameDayBySec(pd.LastChallengeTime, utils.Now().Unix(), 0) {
		pd.ChallengeTime = 0
		pd.LastChallengeTime = utils.Now().Unix()
	}

	//当前积分
	score := pd.Score

	// 步骤1: 计算当前玩家的段位
	scoreConfigs := config.ActLadderRaceScore.All()

	// 遍历配置找到当前积分对应的段位
	var currentRank, currentLittleRank int32
	var maxScore int32 = -1
	for _, cfg := range scoreConfigs {
		if cfg.Score <= int32(score) && cfg.Score > maxScore {
			maxScore = cfg.Score
			currentRank = cfg.Rank
			currentLittleRank = cfg.LittleRank
		}
	}

	// 步骤2: 构建段位索引数组
	// 将配置转为数组并按积分排序
	rankList := make([]conf.ActLadderRaceScore, 0, len(scoreConfigs))
	for _, cfg := range scoreConfigs {
		rankList = append(rankList, cfg)
	}
	sort.Slice(rankList, func(i, j int) bool {
		return rankList[i].Score < rankList[j].Score
	})

	// 找到当前段位在数组中的索引
	var currentIndex int = -1
	for i, cfg := range rankList {
		if cfg.Rank == currentRank && cfg.LittleRank == currentLittleRank {
			currentIndex = i
			break
		}
	}

	// 确定匹配区间的三个段位
	var lowerRankConfig, sameRankConfig, higherRankConfig *conf.ActLadderRaceScore
	if currentIndex >= 0 {
		sameRankConfig = &rankList[currentIndex]
		if currentIndex > 0 {
			lowerRankConfig = &rankList[currentIndex-1]
		}
		if currentIndex < len(rankList)-1 {
			higherRankConfig = &rankList[currentIndex+1]
		}
	}

	// 步骤3: 按概率选择目标段位
	var targetRank, targetLittleRank int32
	hasLower := lowerRankConfig != nil
	hasHigher := higherRankConfig != nil

	if hasLower && hasHigher {
		// 三个段位都存在: 20% lower, 40% same, 40% higher
		randomValue := utils.RandInt[int32](0, 99)
		if randomValue < 20 {
			targetRank = lowerRankConfig.Rank
			targetLittleRank = lowerRankConfig.LittleRank
		} else if randomValue < 60 {
			targetRank = sameRankConfig.Rank
			targetLittleRank = sameRankConfig.LittleRank
		} else {
			targetRank = higherRankConfig.Rank
			targetLittleRank = higherRankConfig.LittleRank
		}
	} else if !hasLower && hasHigher {
		// 只有 same 和 higher (最低段位): 33% same, 67% higher
		randomValue := utils.RandInt[int32](0, 99)
		if randomValue < 33 {
			targetRank = sameRankConfig.Rank
			targetLittleRank = sameRankConfig.LittleRank
		} else {
			targetRank = higherRankConfig.Rank
			targetLittleRank = higherRankConfig.LittleRank
		}
	} else if hasLower && !hasHigher {
		// 只有 lower 和 same (最高段位): 33% lower, 67% same
		randomValue := utils.RandInt[int32](0, 99)
		if randomValue < 33 {
			targetRank = lowerRankConfig.Rank
			targetLittleRank = lowerRankConfig.LittleRank
		} else {
			targetRank = sameRankConfig.Rank
			targetLittleRank = sameRankConfig.LittleRank
		}
	} else {
		// 只有 same (最强王者): 100% same
		targetRank = sameRankConfig.Rank
		targetLittleRank = sameRankConfig.LittleRank
	}

	if a.data.RankPlayer == nil {
		a.data.RankPlayer = make(map[int64]*model.ActDataLadderRaceRankPlayer)
	}

	//判断有没有玩家
	if len(a.data.RankPlayer) <= 0 {
		//从排行榜初始
		a.initRankPlayerFromRedis()
	}

	// 步骤4: 从 RankPlayer 中过滤目标段位的玩家
	candidatePlayers := make([]int64, 0)
	for playerId, playerData := range a.data.RankPlayer {
		if playerData.Rank == targetRank &&
			playerData.LittleRank == targetLittleRank &&
			playerId != ctx.Id {
			candidatePlayers = append(candidatePlayers, playerId)
		}
	}

	// 步骤5: 随机选择一个玩家
	var matchedPlayerId int64 = 0
	if len(candidatePlayers) > 0 {
		randomIndex := utils.RandInt[int](0, len(candidatePlayers)-1)
		matchedPlayerId = candidatePlayers[randomIndex]
	}

	log.Debug("天梯匹配: 玩家[%d] 积分[%d] 段位[%d-%d] -> 目标段位[%d-%d] 匹配玩家[%d]",
		ctx.Id, score, currentRank, currentLittleRank, targetRank, targetLittleRank, matchedPlayerId)

	// TODO: 后续使用 matchedPlayerId 进行战斗逻辑

	if matchedPlayerId <= 0 {
		res.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		return res, errors.New("match no player")
	}

	//判断次数
	ladderRaceConfs, ok := GetTypedConf[conf.ActLadderRace](a.GetCfgId(), config.ActLadderRace.All())
	if !ok {
		log.Error("get activity typed config error:%v", a.GetCfgId())
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		return res, nil
	}

	var arenaConf conf.ActLadderRace
	for _, v := range ladderRaceConfs {
		arenaConf = v
		break
	}

	if arenaConf.Id <= 0 {
		log.Error("get activity typed config error:%v", a.GetCfgId())
		res.Code = proto_public.CommonErrorCode_ERR_NoConfig
		return res, nil

	}
	if pd.ChallengeTime >= arenaConf.ChallengeTime {
		res.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		return res, nil
	}

	pd.ChallengeTime++

	res.Code = proto_public.CommonErrorCode_ERR_OK
	res.Id = matchedPlayerId

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))

	//获取
	return res, nil
}

// 战报
func (a *ActivityLadderRace) BattleReport(ctx *proto_player.Context, req *proto_game.C2SChallengeBattleReport) (*proto_game.S2CChallengeBattleReport, error) {
	res := new(proto_game.S2CChallengeBattleReport)
	//判断挑战玩家
	pd := LoadPd[*model.LadderRacePd](a, ctx.Id)

	//当前积分
	score := pd.Score

	lastConf := conf.ActLadderRaceScore{}
	isSuc := req.WinId == ctx.Id
	// 步骤1: 计算当前玩家的段位
	scoreConfigs := config.ActLadderRaceScore.All()
	for _, cfg := range scoreConfigs {
		if cfg.Score <= int32(score) && cfg.Score > lastConf.Score {
			lastConf = cfg
		}
	}

	if lastConf.Id <= 0 {
		return res, errors.New("config is null")
	}

	//积分加减
	if isSuc {
		pd.Score += lastConf.SettleScore[0]
	} else {
		pd.Score -= lastConf.SettleScore[1]
		if pd.Score <= 0 {
			pd.Score = 0
		}
	}

	if a.data.RankPlayer == nil {
		a.data.RankPlayer = make(map[int64]*model.ActDataLadderRaceRankPlayer)
	}

	// 遍历配置找到当前积分对应的段位
	var currentRank, currentLittleRank int32
	var maxScore int32 = -1
	for _, cfg := range scoreConfigs {
		if cfg.Score <= int32(pd.Score) && cfg.Score > maxScore {
			maxScore = cfg.Score
			currentRank = cfg.Rank
			currentLittleRank = cfg.LittleRank
		}
	}

	//判断存不存在
	if data, ok := a.data.RankPlayer[ctx.Id]; ok {
		data.Score = int64(pd.Score)
		data.Rank = currentRank
		data.LittleRank = currentLittleRank
	} else {
		a.data.RankPlayer[ctx.Id] = new(model.ActDataLadderRaceRankPlayer)
		a.data.RankPlayer[ctx.Id].Score = int64(pd.Score)
		a.data.RankPlayer[ctx.Id].Rank = currentRank
		a.data.RankPlayer[ctx.Id].LittleRank = currentLittleRank
	}

	//更新排行
	updateActivityRank(a, ctx, 0, int32(pd.Score), define.RankTypeTianti)

	//推送
	a.PushActivityData(ctx.Id, a.Format(ctx))

	res.Code = proto_public.CommonErrorCode_ERR_OK
	//获取
	return res, nil
}

func (a *ActivityLadderRace) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	switch req := req.(type) {
	case *proto_activity.C2SLadderRaceGetPlayerLineUp: //获取竞技场阵容
		return a.GetLadderRaceLineUpByPlayerId(ctx, req)
	case *proto_lineup.C2SSetLineUp:
		if req.Type != define.LINEUP_Tianti {
			return nil, nil
		}

		pd := LoadPd[*model.LadderRacePd](a, ctx.Id)
		pd.LineUp = make([]model.LadderRaceIds, 0)

		// 推送活动数据
		a.PushActivityData(ctx.Id, a.Format(ctx))
	case *proto_activity.C2SLadderRaceBattle:
		return a.Battle(ctx, req)
	case *proto_game.C2SChallengeBattleReport:
		return a.BattleReport(ctx, req)
	}
	return nil, nil
}

func (a *ActivityLadderRace) Update(now time.Time) {
	// 跨天逻辑已迁移到 OnDayReset
}

// OnDayReset 跨天重置：重置所有玩家的挑战次数
func (a *ActivityLadderRace) OnDayReset(now time.Time) {
	// TODO: 遍历所有参与过该活动的玩家，重置 pd.ChallengeTime
	log.Debug("ActivityLadderRace OnDayReset: actId=%v", a.GetId())
}

func (a *ActivityLadderRace) OnStop() {
	//活动结束补发奖励
	sendRankReward(a, define.RankTypeTianti, nil)

	//删除排行榜
	deleteActivityRank(a, define.RankTypeTianti)
}

func (a *ActivityLadderRace) OnClose() {
}

// initRankPlayerFromRedis 从Redis排行榜初始化RankPlayer数据
func (a *ActivityLadderRace) initRankPlayerFromRedis() {
	rankKey := fmt.Sprintf("%s:%d", define.RankTypeTiantiKey, a.GetId())

	// 从Redis获取排行榜数据（所有玩家的ID和积分）
	result, err := redis.Strings(db.RedisExec("ZREVRANGE", rankKey, 0, -1, "WITHSCORES"))
	if err != nil && !errors.Is(err, redis.ErrNil) {
		log.Error("initRankPlayerFromRedis error:%v", err)
		return
	}

	if len(result) == 0 {
		log.Debug("天梯排行榜为空")
		return
	}

	// 获取段位配置
	scoreConfigs := config.ActLadderRaceScore.All()

	// 清空现有数据
	a.data.RankPlayer = make(map[int64]*model.ActDataLadderRaceRankPlayer)

	// 解析排行榜数据，填充RankPlayer
	for i := 0; i < len(result); i += 2 {
		playerIdStr := result[i]
		scoreStr := result[i+1]

		playerId, err := strconv.ParseInt(playerIdStr, 10, 64)
		if err != nil {
			log.Error("解析玩家ID失败:%v", err)
			continue
		}

		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			log.Error("解析积分失败:%v", err)
			continue
		}

		// 根据积分查找对应的段位配置
		var rank, littleRank int32
		var maxScore int32 = -1
		for _, cfg := range scoreConfigs {
			if cfg.Score <= int32(score) && cfg.Score > maxScore {
				maxScore = cfg.Score
				rank = cfg.Rank
				littleRank = cfg.LittleRank
			}
		}

		// 添加到RankPlayer
		a.data.RankPlayer[playerId] = &model.ActDataLadderRaceRankPlayer{
			Score:      int64(score),
			Rank:       rank,
			LittleRank: littleRank,
		}
	}

	log.Debug("天梯排行榜已初始化，玩家数:%d", len(a.data.RankPlayer))
}

func init() {
	RegisterActivity(define.ActivityTypeLadderRace, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityLadderRace) },
		NewActivityData: func() any { return new(model.ActDataLadderRace) },
		NewPlayerData: func() any {
			return &model.LadderRacePd{
				LineUp: make([]model.LadderRaceIds, 0),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.LadderRace = data.(*proto_activity.LadderRace)
		},
		InjectFunc: func(handler IActivity, data any) {
			h := handler.(*ActivityLadderRace)
			if data == nil {
				h.data = new(model.ActDataLadderRace)
				h.data.Season = 1
				return
			}
			h.data = data.(*model.ActDataLadderRace)
		},
		ExtractFunc: func(handler IActivity) any { return handler.(*ActivityLadderRace).data },
	})
}
