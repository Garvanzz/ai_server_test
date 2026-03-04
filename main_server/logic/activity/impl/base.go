package impl

import (
	"time"
	"xfx/main_server/invoke"
	"xfx/pkg/module"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
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
	SetBaseInfo(baseInfo BaseInfo)
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

func (base *BaseActivity) SetBaseInfo(baseInfo BaseInfo) {
	base.BaseInfo = baseInfo
}
