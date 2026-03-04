package impl

import (
	"github.com/golang/protobuf/proto"
	"time"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/main_server/logic/activity/data"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"
)

type IActivity interface {
	OnInit()  // 每次加载完成都会调用一次
	OnStart() // 只会调用一次
	OnClose() // 活动结束调用
	OnStop()  // 活动结束 可以请求数据调用
	OnEvent(key string, obj *proto_player.Context, params EventParams)
	Router(ctx *proto_player.Context, req proto.Message) (interface{}, error)
	Update(time.Time)
	Format(ctx *proto_player.Context) proto.Message
	Inject(data any)
	Extract() any
}

type BaseInfo interface {
	GetId() int64
	GetCfgId() int64
	GetType() string
	GetStartTime() int64
	GetEndTime() int64
	GetCloseTime() int64
	Module() module.Module
}

type BaseActivity struct {
	BaseInfo
}

func (base *BaseActivity) OnInit() {}

func (base *BaseActivity) OnStart() {}

func (base *BaseActivity) OnClose() {}
func (base *BaseActivity) OnStop()  {}

func (base *BaseActivity) OnEvent(key string, ctx *proto_player.Context, params EventParams) {}

func (base *BaseActivity) Update(now time.Time) {}

func (base *BaseActivity) Format(ctx *proto_player.Context) proto.Message { return nil }

func (base *BaseActivity) Inject(data any) {}

func (base *BaseActivity) Extract() any { return nil }

func (base *BaseActivity) OnDayReset() {}

func (base *BaseActivity) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	return nil, nil
}

func (base *BaseActivity) PushActivityData(playerId int64, data proto.Message) {
	result := new(proto_activity.ActivityData)
	result.ActivityId = base.GetId()
	result.ConfigId = base.GetCfgId()
	SetProtoByType(base.GetType(), result, data)

	invoke.Dispatch(base.Module(), playerId, &proto_activity.PushActivityDataChange{
		Data: result,
	})
}

func LoadPd[T comparable](a BaseInfo, playerId int64) T {
	var zero T
	d := data.LoadPlayerData[T](a.GetId(), playerId)
	if d != zero {
		return d
	}

	var ret any

	switch a.GetType() {
	case define.ActivityTypeDailyAccRecharge:
		pd := new(model.DailyAccumulateRechargePd)
		pd.GetList = make([]int32, 0)
		ret = pd
	case define.ActivityTypeNormalMonthCard:
		ret = new(model.MonthCardPd)
	case define.ActivityTypeTheCompetition:
		ret = new(model.TheCompetitionPd)
	case define.ActivityTypeMainLineFund,
		define.ActivityTypeLevelFund,
		define.ActivityTypeBoxFund:
		pd := new(model.FundOptionPd)
		pd.NormalIds = make([]int32, 0)
		pd.AdvanceIds = make([]int32, 0)
		ret = pd
	case define.ActivityTypeArena:
		pd := new(model.ArenaOptionPd)
		pd.PlayerIds = make([]int64, 0)
		pd.LineUp = make([]model.ArenaLineUpIds, 0)
		ret = pd
	case define.ActivityTypeLadderRace:
		pd := new(model.LadderRacePd)
		pd.LineUp = make([]model.LadderRaceIds, 0)
		ret = pd
	case define.ActivityTypeGoFish:
		pd := new(model.GoFishPd)
		pd.Fish = make(map[int32]int32)
		pd.GetList = make([]int32, 0)
		ret = pd
	case define.ActivityTypePassport:
		pd := new(model.PassportPd)
		pd.NormalIds = make([]int32, 0)
		pd.AdvanceIds = make([]int32, 0)
		ret = pd
	default:
		log.Error("LoadPd type error:%v", a.GetType())
	}

	data.SetPlayerData(a.GetId(), playerId, ret)
	return ret.(T)
}

// TODO:
// handler的数据初始化一直都有 看需要调整不
// 活动end time的处理

func LoadPd[T comparable](a BaseInfo, playerId int64) T {
	var zero T
	d := data.LoadPlayerData[T](a.GetId(), playerId)
	if d != zero {
		return d
	}

	desc := GetActivityDesc(a.GetType())
	if desc == nil || desc.NewPlayerData == nil {
		log.Error("LoadPd: no player data factory for type: %v", a.GetType())
		return zero
	}

	ret := desc.NewPlayerData()
	data.SetPlayerData(a.GetId(), playerId, ret)
	return ret.(T)
}
