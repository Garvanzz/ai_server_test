package impl

import (
	"sort"
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
)

// ActivityGoFish 钓鱼
type ActivityGoFish struct {
	BaseActivity
	data *model.ActDataGoFish
}

func (a *ActivityGoFish) OnInit() {}

func (a *ActivityGoFish) OnStart() {
	// 初始化鱼池
	a.InitPool()
	a.data.PoolRefreshTime = utils.Now().Unix()

	//这里有点没设计好，先这样吧
	activityConfs := config.ActGoFish.All()
	for _, v := range activityConfs {
		a.data.StartTime = v.StartTime
		a.data.EndTime = v.EndTime
		a.data.PoolRefreshOffseTime = v.RefreshTime
		break
	}
}

func (a *ActivityGoFish) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.GoFishPd](a, ctx.Id)

	//判断签到时间
	if pd.SignDay >= 7 {
		if !utils.CheckIsSameDayBySec(pd.LastSignTime, utils.Now().Unix(), 0) {
			pd.SignDay = 0
		}
	}

	pools := make(map[int32]*proto_activity.GoFishPoolOption)
	for k, v := range a.data.Pool {
		if _, ok := pools[k]; !ok {
			pools[k] = new(proto_activity.GoFishPoolOption)
			pools[k].Pool = make(map[int32]int32)
		}

		for id, num := range v {
			pools[k].Pool[id] = num
		}
	}

	offseTime := a.data.PoolRefreshTime + 3600 - utils.Now().Unix()
	if offseTime <= 0 {
		offseTime = 0
	}
	isTodaySign := true
	if !utils.CheckIsSameDayBySec(pd.LastSignTime, utils.Now().Unix(), 0) {
		isTodaySign = false
	}
	return &proto_activity.GoFish{
		Pools:           pools,
		Fish:            pd.Fish,
		PoolRefreshTime: offseTime,
		SignDay:         pd.SignDay,
		LastSignTime:    pd.LastSignTime,
		Exp:             pd.Exp,
		LevelIds:        pd.GetList,
		ToDaySign:       isTodaySign,
	}
}

func (a *ActivityGoFish) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	default:
	}
}

// 初始鱼池
func (a *ActivityGoFish) InitPool() {
	activityConfs := config.ActGoFish.All()
	a.data.Pool = make(map[int32]map[int32]int32)
	for _, v := range activityConfs {
		if _, ok := a.data.Pool[v.Type]; !ok {
			a.data.Pool[v.Type] = make(map[int32]int32)
		}
		for typ, num := range v.Fish {
			if _, ok := a.data.Pool[v.Type][typ]; !ok {
				a.data.Pool[v.Type][typ] = 0
			}

			a.data.Pool[v.Type][typ] += num
		}
	}
}

func (a *ActivityGoFish) Update(now time.Time) {
	if a.data.PoolRefreshOffseTime > 0 {
		// 每隔一个小时刷新鱼池
		if now.Unix()-a.data.PoolRefreshTime > int64(3600*a.data.PoolRefreshOffseTime) {
			a.InitPool()
			a.data.PoolRefreshTime = now.Unix()
			//让玩家自己去刷新吧，获取数据的时候
		}
	}

	if a.data.EndTime > 0 {
		t1 := time.Unix(utils.Now().Unix(), int64(0))
		time, _ := utils.GetTodayEndUnixInHour(&t1, int(a.data.EndTime))

		//结束的时候要补发排行榜的邮件
		if now.Unix() >= time && !a.data.FireRankAward {
			//结束补发奖励
			sendRankReward(a, define.RankTypeGoFish, nil)

			//删除排行榜
			deleteActivityRank(a, define.RankTypeGoFish)
			a.data.FireRankAward = true
		}
	}

	if a.data.StartTime > 0 {
		t1 := time.Unix(utils.Now().Unix(), int64(0))
		time, _ := utils.GetTodayEndUnixInHour(&t1, int(a.data.StartTime))

		if now.Unix() <= time && a.data.FireRankAward {
			a.data.FireRankAward = false
		}
	}
}

func (a *ActivityGoFish) Router(ctx *proto_player.Context, _req proto.Message) (any, error) {
	switch req := _req.(type) {
	case *proto_activity.C2SGoFish: // 钓鱼
		return a.GoFish(ctx, req)
	case *proto_activity.C2SFishLevelAward: // 等级
		return a.GetAward(ctx, req)
	case *proto_activity.C2SFishSign: // 签到
		return a.Sign(ctx, req)
	}
	return nil, nil
}

func (a *ActivityGoFish) OnStop() {
	//结束补发奖励
	sendRankReward(a, define.RankTypeGoFish, nil)

	//删除排行榜
	deleteActivityRank(a, define.RankTypeGoFish)
}

// OnDayReset 跨天重置：重置签到天数（如果已连续签到7天）
func (a *ActivityGoFish) OnDayReset(now time.Time) {
	// 跨天重置所有玩家的签到天数（如果已连续签到7天且昨天未签到）
	// 逻辑在 Format 中已经处理，这里可以遍历缓存进行清理
	log.Debug("ActivityGoFish OnDayReset: actId=%v", a.GetId())
}

func (a *ActivityGoFish) OnClose() {
}

func (a *ActivityGoFish) GoFish(ctx *proto_player.Context, req *proto_activity.C2SGoFish) (any, error) {
	// TODO:需要消耗物品
	//基本判断在player里面去做

	resp := new(model.GoFishBack)
	//没有鱼了
	if a.GetPoolCount(req.PoolType) <= 0 {
		resp.Code = 1
		log.Debug("没有鱼了:%v", req.PoolType)
		return resp, nil
	}

	pd := LoadPd[*model.GoFishPd](a, ctx.Id)

	//钓鱼
	//判断成功率
	activityConfs := config.ActGoFish.All()
	var gofishConf conf.ActGoFish
	for _, v := range activityConfs {
		if v.Type == req.PoolType {
			gofishConf = v
			break
		}
	}

	addRate := config.Global.Get().GofishBasicRate
	addRate += gofishConf.PointAddRare * req.PointNum
	isSuc := false
	//必成功
	if addRate >= 100 {
		isSuc = true
	} else {
		rate := utils.RandInt(0, 100)
		if int32(rate) <= addRate {
			isSuc = true
		}
	}

	// 不管是否钓到都+经验
	pd.Exp += gofishConf.Exp

	if !isSuc {
		resp.Code = 2
		log.Debug("没有成功:%v", req.PoolType)
		return resp, nil
	}

	var weight []int32
	if req.CostId == define.ItemIdFishNormal {
		weight = config.Global.Get().GofishWeightNormal
	} else if req.CostId == define.ItemIdFishAdvance {
		weight = config.Global.Get().GofishWeightAdvance
	}
	log.Info("weight:%v", weight)

	rateIndex := 1
	if len(weight) > 0 {
		rateIndex = utils.WeightIndex(weight)
		rateIndex++
	}

	//获取该品质有没有鱼
	pool := a.data.Pool[req.PoolType]
	rareMap := make(map[int32][]int32)
	for k, v := range pool {
		if v <= 0 {
			continue
		}
		fish := getFishConfig(k)
		if fish.Id <= 0 {
			continue
		}
		if _, ok := rareMap[fish.Rate]; !ok {
			rareMap[fish.Rate] = make([]int32, 0)
		}
		rareMap[fish.Rate] = append(rareMap[fish.Rate], int32(fish.Id))
	}

	var foundRate int32 = -1
	if _, ok := rareMap[int32(rateIndex)]; !ok {
		// 找低品质的
		// 判断品质是不是最低的,不是最低的，找低1品质的鱼，要判断数量，如果没有数量，继续找，找到最低都没有 返回空

		// 获取所有品质等级并排序（从高到低）
		var rates []int32
		for rate := range rareMap {
			rates = append(rates, rate)
		}

		// 排序：从高到低（假设rate值越大品质越高）
		sort.Slice(rates, func(i, j int) bool {
			return rates[i] > rates[j]
		})

		// 查找当前品质在排序中的位置
		currentRate := int32(rateIndex)
		// 在排序后的rates中，只找比当前rate小的
		for i := 0; i < len(rates); i++ {
			if rates[i] < currentRate && len(rareMap[rates[i]]) > 0 {
				foundRate = rates[i]
				break
			}
		}

		if foundRate == -1 {
			log.Debug("找低品质没有找到鱼: %d", currentRate)
			resp.Code = 2
			return resp, nil
		}
	} else {
		foundRate = int32(rateIndex)
	}

	log.Debug("匹配到品质: %d, %v", foundRate, rareMap)
	arr := rareMap[foundRate]
	if len(arr) <= 0 {
		resp.Code = 2
		return resp, nil
	}

	count := 1
	//判断双钩
	rate := utils.RandInt(0, 100)
	if int32(rate) <= gofishConf.DoubleRate {
		count = 2
	}

	var ids []int32
	for i := 0; i < count; i++ {
		_index := utils.RandInt(0, len(arr)-1)
		id := arr[_index]
		log.Debug("掉到了鱼:id : %d", id)
		ids = append(ids, id)
	}
	resp.Code = 3
	resp.Ids = ids

	if pd.Fish == nil {
		pd.Fish = make(map[int32]int32)
	}

	for _, v := range ids {
		//钓到鱼也要增加经验值
		fishConf := config.Fish.All()[int64(v)]
		pd.Exp += fishConf.Exp

		if _, ok := pd.Fish[v]; !ok {
			pd.Fish[v] = 0
		}
		pd.Fish[v] += 1

		//去掉鱼池里面的
		pool[fishConf.Type] -= 1
		if pool[fishConf.Type] <= 0 {
			pool[fishConf.Type] = 0
		}
	}
	a.data.Pool[req.PoolType] = pool

	//判断是不是在赛事内
	t1 := time.Unix(utils.Now().Unix(), int64(0))
	time1, _ := utils.GetTodayUnixInHour(&t1, int(gofishConf.StartTime))
	time2, _ := utils.GetTodayUnixInHour(&t1, int(gofishConf.EndTime))
	if utils.Now().Unix() >= time1 && utils.Now().Unix() < time2 {
		//进入排行榜
		updateActivityRank(a, ctx, 0, foundRate*int32(count), define.RankTypeGoFish)
	}

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	return resp, nil
}

// 获取鱼池里鱼的数量
func (a *ActivityGoFish) GetPoolCount(typ int32) int32 {
	if data, ok := a.data.Pool[typ]; ok {
		sum := int32(0)
		for _, v := range data {
			sum += v
		}
		return sum
	}
	return 0
}

// Sign 签到
func (a *ActivityGoFish) Sign(ctx *proto_player.Context, req *proto_activity.C2SFishSign) (any, error) {
	resp := new(model.CommonActivityAwardBack)

	pd := LoadPd[*model.GoFishPd](a, ctx.Id)
	if utils.CheckIsSameDayBySec(pd.LastSignTime, utils.Now().Unix(), 0) {
		resp.Code = 1
		return resp, nil
	}
	pd.LastSignTime = utils.Now().Unix()
	pd.SignDay += 1

	fishSignConf := config.FishSign.All()
	for _, v := range fishSignConf {
		if v.Day == pd.SignDay {
			resp.Award = v.Reward
			break
		}
	}
	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	return resp, nil
}

// GetAward TODO:领奖
func (a *ActivityGoFish) GetAward(ctx *proto_player.Context, req *proto_activity.C2SFishLevelAward) (any, error) {
	pd := LoadPd[*model.GoFishPd](a, ctx.Id)

	resp := new(model.CommonActivityAwardBack)
	//判断等级
	id := int32(0)
	fishLevelAwardConf := config.FishLevelAward.All()
	var lists []conf.FishLevelAward
	for _, v := range fishLevelAwardConf {
		lists = append(lists, v)
	}

	//排序
	sort.Slice(lists, func(i, j int) bool {
		return lists[i].Exp > lists[j].Exp
	})

	for _, v := range lists {
		if v.Exp <= pd.Exp {
			id = int32(v.Id)
			break
		}
	}

	//过滤传入的参数
	for _, j := range req.Ids {
		if j > id {
			resp.Code = 2
			return resp, nil
		}
	}

	for _, j := range req.Ids {
		if utils.ContainsInt32(pd.GetList, j) {
			resp.Code = 1
			return resp, nil
		}
	}

	pd.GetList = append(pd.GetList, req.Ids...)

	//奖励
	award := make([]conf.ItemE, 0)
	for _, j := range req.Ids {
		conf := fishLevelAwardConf[int64(j)]
		award = append(award, conf.Reward...)
	}

	resp.Award = award

	// TODO:推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))

	// 返回奖励
	return resp, nil
}

func getFishConfig(typ int32) conf.Fish {
	fishConfs := config.Fish.All()
	for _, v := range fishConfs {
		if v.Type == typ {
			return v
		}
	}
	return conf.Fish{}
}

func init() {
	RegisterActivity(define.ActivityTypeGoFish, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityGoFish) },
		NewActivityData: func() any { return new(model.ActDataGoFish) },
		NewPlayerData: func() any {
			return &model.GoFishPd{
				Fish:    make(map[int32]int32),
				GetList: make([]int32, 0),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.GoFish = data.(*proto_activity.GoFish)
		},
		InjectFunc: func(handler IActivity, data any) {
			h := handler.(*ActivityGoFish)
			if data == nil {
				h.data = new(model.ActDataGoFish)
				h.data.FireRankAward = false
				return
			}
			h.data = data.(*model.ActDataGoFish)
			h.data.FireRankAward = false
		},
		ExtractFunc: func(handler IActivity) any { return handler.(*ActivityGoFish).data },
	})
}
