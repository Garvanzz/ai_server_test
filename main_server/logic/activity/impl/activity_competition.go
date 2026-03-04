package impl

import (
	"time"
	"xfx/core/common"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"
	Proto_Public "xfx/proto/proto_public"

	"github.com/golang/protobuf/proto"
)

// ActivityTheCompetition 巅峰决斗
type ActivityTheCompetition struct {
	BaseActivity
	data *model.ActDataTheCompetition
}

func (a *ActivityTheCompetition) OnInit() {
}

func (a *ActivityTheCompetition) OnStart() {
	//commonConf, ok := GetCommonConf(a.GetCfgId())
	log.Debug("巅峰决斗初始 start")
}

func (a *ActivityTheCompetition) Format(ctx *proto_player.Context) proto.Message {
	pd := LoadPd[*model.TheCompetitionPd](a, ctx.Id)

	log.Debug("加载决斗数据:%s", pd)

	isCanStake := false
	confthe, ok := GetTypedConf[conf.ActTheCompetition](a.GetCfgId())
	if !ok {
		isCanStake = false
	}

	var conf conf.ActTheCompetition
	for _, v := range confthe {
		conf = v
		break
	}

	if conf.Id > 0 {
		openTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.StakeOpenTime), time.Local)
		if err != nil {
			isCanStake = false
		} else {
			if time.Now().Unix() >= openTime.Unix() {
				isCanStake = true
			}
		}
	}

	return &proto_activity.TheCompetition{
		IsChoose:     pd.IsChoose,
		ChooseId:     pd.ChooseId,
		IsStake:      pd.IsStake,
		StakeCount:   pd.StakeCount,
		StageGroupId: pd.StageGroupId,
		Score:        a.data.Score,
		StakeGroup:   a.data.StakeGroup,
		IsCanStake:   isCanStake,
	}
}

func (a *ActivityTheCompetition) OnEvent(key string, ctx *proto_player.Context, params EventParams) {
	switch key {
	case "thecompetition_choosegroup": //选择阵营
		a.chooseGroupId(ctx, params)
	case "thecompetition_stake": //押注
		a.stakeCount(ctx, params)
	default:
	}
}

// 选择阵营
func (a *ActivityTheCompetition) chooseGroupId(ctx *proto_player.Context, params EventParams) {
	res := proto_activity.S2CTheCompetitionChooseGroup{}

	log.Debug("aaa:%v,%v", time.Now().Unix(), a.GetCloseTime())
	if time.Now().Unix() >= a.GetCloseTime() {
		res.Code = Proto_Public.CommonErrorCode_ERR_NOOPENTIME
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	pd := LoadPd[*model.TheCompetitionPd](a, ctx.Id)
	if pd.IsChoose {
		res.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	req, ok := Key[*proto_activity.C2STheCompetitionChooseGroup](params, "req")
	if req.ActId != a.GetId() {
		res.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	//判断角色 是不是阵容上的
	confthe, ok := GetTypedConf[conf.ActTheCompetition](a.GetCfgId())
	if !ok {
		res.Code = Proto_Public.CommonErrorCode_ERR_NoConfig
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	var conf conf.ActTheCompetition
	for _, v := range confthe {
		if common.IsHaveValueIntArray(v.GroupIds, req.Id) {
			conf = v
			break
		}
	}

	if conf.Id <= 0 {
		res.Code = Proto_Public.CommonErrorCode_ERR_NoConfig
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	if pd.IsChoose {
		res.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	pd.IsChoose = true
	pd.ChooseId = req.Id

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	log.Debug("推送巅峰对决活动数据:%v", pd)

	res.Code = Proto_Public.CommonErrorCode_ERR_OK
	invoke.Dispatch(a.Module(), ctx.Id, res)
	log.Debug("&&&&&&&&&&&&&&:%v", res)
}

// 押注
func (a *ActivityTheCompetition) stakeCount(ctx *proto_player.Context, params EventParams) {
	res := proto_activity.S2CTheCompetitionStake{}
	log.Debug("bbb:%v,%v", time.Now().Unix(), a.GetCloseTime())
	if time.Now().Unix() >= a.GetCloseTime() {
		res.Code = Proto_Public.CommonErrorCode_ERR_NOOPENTIME
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	pd := LoadPd[*model.TheCompetitionPd](a, ctx.Id)
	if pd.IsChoose {
		res.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	req, ok := Key[*proto_activity.C2STheCompetitionStake](params, "req")
	if req.ActId != a.GetId() {
		res.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	//判断角色 是不是阵营上的
	confthe, ok := GetTypedConf[conf.ActTheCompetition](a.GetCfgId())
	if !ok {
		res.Code = Proto_Public.CommonErrorCode_ERR_NoConfig
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	var conf conf.ActTheCompetition
	for _, v := range confthe {
		conf = v
		break
	}

	if conf.Id <= 0 {
		res.Code = Proto_Public.CommonErrorCode_ERR_NoConfig
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	if pd.IsStake {
		res.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	//判断押注时间
	openTime, err := time.ParseInLocation("2006-01-02 15:04:05", Trim(conf.StakeOpenTime), time.Local)
	if err != nil {
		log.Error("checkCfg parse endTime err:%v", err)
		res.Code = Proto_Public.CommonErrorCode_ERR_ParamTypeError
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}
	log.Debug("bbbccc:%v,%v", time.Now().Unix(), openTime.Unix())
	if time.Now().Unix() < openTime.Unix() {
		res.Code = Proto_Public.CommonErrorCode_ERR_NOOPENTIME
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	if !common.IsHaveValueIntArray(conf.Stake, req.Count) {
		res.Code = Proto_Public.CommonErrorCode_ERR_NoConfig
		invoke.Dispatch(a.Module(), ctx.Id, res)
		return
	}

	pd.IsStake = true
	pd.StakeCount = req.Count

	//押注
	stageGroup := a.data.StakeGroup[req.GroupId]
	stageGroup += int64(req.Count)
	a.data.StakeGroup[req.GroupId] = stageGroup

	// 推送活动数据
	a.PushActivityData(ctx.Id, a.Format(ctx))
	log.Debug("推送巅峰对决活动数据:%v", pd)

	res.Code = Proto_Public.CommonErrorCode_ERR_OK
	invoke.Dispatch(a.Module(), ctx.Id, res)
}

func (a *ActivityTheCompetition) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	return nil, nil
}

func (a *ActivityTheCompetition) OnClose() {
	//活动结束补发奖励
	sendRankReward(a, define.RankTypeTheCompetition, nil)

	//删除排行榜
	deleteActivityRank(a, define.RankTypeTheCompetition)
}

func (a *ActivityTheCompetition) Inject(data any) {
	if data == nil {
		a.data = new(model.ActDataTheCompetition)
		a.data.StakeGroup = make(map[int32]int64)
		a.data.Score = make(map[int32]int64)
		return
	}
	a.data = data.(*model.ActDataTheCompetition)
}

func (a *ActivityTheCompetition) Extract() any { return a.data }

func init() {
	RegisterActivity(define.ActivityTypeTheCompetition, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivityTheCompetition) },
		NewActivityData: func() any { return nil },
		NewPlayerData: func() any {
			return &model.ActDataTheCompetition{
				StakeGroup: make(map[int32]int64),
				Score:      make(map[int32]int64),
			}
		},
		SetProto: func(msg *proto_activity.ActivityData, data proto.Message) {
			msg.TheCompetition = data.(*proto_activity.TheCompetition)
		},
		InjectFunc:  func(handler IActivity, data any) {},
		ExtractFunc: func(handler IActivity) any { return nil },
	})
}
