package impl

import (
	"sort"
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
)

// ActivityPassport 通行证活动
type ActivityPassport struct {
	BaseActivity
	data         *model.ActDataPassport
	configByID   map[int32]conf.ActPassport
	levelConfigs []conf.ActPassport
}

func (a *ActivityPassport) OnInit() {
	a.ensureState()
	a.reloadConfigs()
}

func (a *ActivityPassport) OnStart() {
	a.ensureState()
	a.reloadConfigs()
}

// Format 格式化玩家数据返回给客户端
func (a *ActivityPassport) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.PassportPd](a, ctx.Id)
	return &proto_activity.Passport{
		Score:      pd.Score,
		Level:      pd.Level,
		NormalIds:  pd.NormalIds,
		AdvanceIds: pd.AdvanceIds,
	}
}

// OnEvent 处理事件(积分获取、购买高级通行证等)
func (a *ActivityPassport) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "passport_task_score":
		// 添加积分事件
		score, ok := Key[int32](params, "score")
		if !ok {
			return
		}
		a.addScore(ctx, score)
	case "recharge":
		shopconf, ok := Key[conf.Shop](params, "shopconf")
		if !ok {
			return
		}
		if shopconf.Type == define.SHOPTYPE_PASSPORT || shopconf.Type == define.SHOPTYPE_PASSPORT_ADVANCE {
			// 购买高级通行证事件
			a.buyAdvancePassport(ctx, shopconf.GetItem)
		} else if shopconf.Type == define.SHOPTYPE_PASSPORT_SCOREGIFT {
			if len(shopconf.GetItem) > 0 {
				a.addScore(ctx, shopconf.GetItem[0].ItemNum)
			}
		}
	default:
	}
}

// Router 处理协议消息
func (a *ActivityPassport) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	switch req.(type) {
	case *proto_activity.C2SPassportGetAward:
		return a.getAward(ctx, req.(*proto_activity.C2SPassportGetAward))
	}
	return nil, nil
}

// addScore 增加积分并自动升级
func (a *ActivityPassport) addScore(ctx *proto_player.Context, score int32) {
	pd := LoadPd[*model.PassportPd](a, ctx.Id)
	pd.Score += score

	// 根据积分计算等级
	newLevel := a.calculateLevel(pd.Score)
	if newLevel > pd.Level {
		pd.Level = newLevel
		log.Debug("通行证升级: 玩家=%d, 新等级=%d, 积分=%d", ctx.Id, pd.Level, pd.Score)
	}

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
}

// buyAdvancePassport 购买高级通行证
func (a *ActivityPassport) buyAdvancePassport(ctx *proto_player.Context, item []conf.ItemE) {
	pd := LoadPd[*model.PassportPd](a, ctx.Id)
	if pd.IsBuy {
		log.Warn("玩家已购买高级通行证: 玩家=%d", ctx.Id)
		return
	}

	pd.IsBuy = true
	log.Debug("购买高级通行证成功: 玩家=%d", ctx.Id)

	firstAward := item[0]
	if firstAward.ItemType == define.ItemTypePassportScore {
		a.addScore(ctx, firstAward.ItemNum)
	}
	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
}

// getAward 领取奖励
func (a *ActivityPassport) getAward(ctx *proto_player.Context, req *proto_activity.C2SPassportGetAward) ([]conf.ItemE, error) {
	pd := LoadPd[*model.PassportPd](a, ctx.Id)
	a.ensureState()
	if len(a.configByID) == 0 {
		a.reloadConfigs()
	}
	normalReceived := make(map[int32]struct{}, len(pd.NormalIds))
	for _, id := range pd.NormalIds {
		normalReceived[id] = struct{}{}
	}
	advanceReceived := make(map[int32]struct{}, len(pd.AdvanceIds))
	for _, id := range pd.AdvanceIds {
		advanceReceived[id] = struct{}{}
	}

	// 收集所有奖励
	awards := make([]conf.ItemE, 0)

	// 遍历要领取的奖励ID
	for _, id := range req.Ids {
		conf, ok := a.configByID[id]
		if !ok {
			log.Error("通行证配置不存在: id=%d", id)
			return nil, nil
		}

		// 检查等级是否达到
		if pd.Level < conf.Level {
			log.Warn("通行证等级不足: 玩家=%d, 当前等级=%d, 需要等级=%d", ctx.Id, pd.Level, conf.Level)
			return nil, nil
		}

		// 检查普通奖励是否已领取
		_, normalGot := normalReceived[id]
		// 检查高级奖励是否已领取（或未购买高级通行证则视为已处理）
		_, advanceGot := advanceReceived[id]
		advanceDone := !pd.IsBuy || advanceGot

		if normalGot && advanceDone {
			log.Warn("奖励已领取: 玩家=%d, id=%d", ctx.Id, id)
			return nil, nil
		}

		//普通奖励
		if !normalGot {
			// 领取普通奖励
			if len(conf.NormalAward) > 0 {
				awards = append(awards, conf.NormalAward...)
				pd.NormalIds = append(pd.NormalIds, id)
				normalReceived[id] = struct{}{}
				log.Debug("领取普通奖励: 玩家=%d, id=%d", ctx.Id, id)
			}
		}

		// 如果购买了高级通行证且有高级奖励
		if pd.IsBuy && !advanceGot && len(conf.AdvanceAward) > 0 {
			// 领取高级奖励
			awards = append(awards, conf.AdvanceAward...)
			pd.AdvanceIds = append(pd.AdvanceIds, id)
			advanceReceived[id] = struct{}{}
			log.Debug("领取高级奖励: 玩家=%d, id=%d", ctx.Id, id)
		}
	}

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))

	return awards, nil
}

// calculateLevel 根据积分计算等级
func (a *ActivityPassport) calculateLevel(score int32) int32 {
	a.ensureState()
	if len(a.levelConfigs) == 0 {
		a.reloadConfigs()
	}
	if len(a.levelConfigs) == 0 {
		return 0
	}
	idx := sort.Search(len(a.levelConfigs), func(i int) bool {
		return a.levelConfigs[i].Score > score
	}) - 1
	if idx < 0 {
		return 0
	}
	return a.levelConfigs[idx].Level
}

func (a *ActivityPassport) OnClose() {
	// 活动结束处理
}

func (a *ActivityPassport) Update(now time.Time) {
	// 跨天逻辑已迁移到 OnDayReset
}

// OnDayReset 跨天重置：重置通行证每日任务
func (a *ActivityPassport) OnDayReset(now time.Time) {
	a.reloadConfigs()
	log.Debug("ActivityPassport OnDayReset: actId=%v", a.GetId())
}

func (a *ActivityPassport) ensureState() {
	if a.data == nil {
		a.data = new(model.ActDataPassport)
	}
	if a.data.Season == 0 {
		a.data.Season = 1
	}
}

func (a *ActivityPassport) reloadConfigs() {
	a.ensureState()
	a.configByID = make(map[int32]conf.ActPassport)
	a.levelConfigs = a.levelConfigs[:0]
	for _, passportConf := range config.ActPassport.All() {
		if passportConf.Season != a.data.Season {
			continue
		}
		a.configByID[int32(passportConf.Id)] = passportConf
		a.levelConfigs = append(a.levelConfigs, passportConf)
	}
	sort.Slice(a.levelConfigs, func(i, j int) bool {
		if a.levelConfigs[i].Score == a.levelConfigs[j].Score {
			return a.levelConfigs[i].Level < a.levelConfigs[j].Level
		}
		return a.levelConfigs[i].Score < a.levelConfigs[j].Score
	})
}

func init() {
	RegisterActivity(define.ActivityTypePassport, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityPassport) },
		NewActivityData: func() any { return new(model.ActDataPassport) },
		NewPlayerData: func() any {
			return &model.PassportPd{
				NormalIds:  make([]int32, 0),
				AdvanceIds: make([]int32, 0),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.Passport = data.(*proto_activity.Passport)
		},
		InjectFunc: func(handler IActivity, data any) {
			h := handler.(*ActivityPassport)
			if data == nil {
				h.data = new(model.ActDataPassport)
				return
			}
			h.data = data.(*model.ActDataPassport)
		},
		ExtractFunc: func(handler IActivity) any { return handler.(*ActivityPassport).data },
	})
}
