package impl

import (
	"time"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
)

type EventParams map[string]any

func Key[T any](params EventParams, key string) (T, bool) {
	v, ok := params[key]
	if !ok {
		var zero T
		return zero, false
	}
	result, ok := v.(T)
	if !ok {
		var zero T
		log.Error("event params key type error: key=%v, expected=%T, actual=%T", key, zero, v)
		return result, false
	}
	return result, true
}

type IActivity interface {
	OnInit()         // 初始化：每次加载完成都会调用一次
	OnStart()        // 开始：活动从 Waiting 转为 Running 时调用一次
	OnClose()        // 关闭：活动结束时调用
	OnStop()         // 暂停：活动暂停时调用
	OnRecover()      // 恢复：活动从 Stopped 转为 Running 时调用
	OnRestart(int64) // 重开：活动重置并分配新实例 ID 后调用，参数为旧 actId
	OnReloadConfig() // 配置重载后调用；初始化首轮配置同步也会触发一次
	OnEvent(key string, obj *proto_player.Context, params EventParams)
	Router(ctx *proto_player.Context, req proto.Message) (interface{}, error)
	Update(now time.Time)     // 更新：每帧调用，用于定时检查
	OnDayReset(now time.Time) // 跨天重置：每天首次调用时触发
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

func (base *BaseActivity) OnClose()                      {}
func (base *BaseActivity) OnStop()                       {}
func (base *BaseActivity) OnRecover()                    {}
func (base *BaseActivity) OnRestart(previousActID int64) {}

func (base *BaseActivity) OnReloadConfig() {}

func (base *BaseActivity) OnEvent(key string, ctx *proto_player.Context, params EventParams) {}

func (base *BaseActivity) Update(now time.Time) {}

func (base *BaseActivity) Format(ctx *proto_player.Context) proto.Message { return nil }

func (base *BaseActivity) OnDayReset(now time.Time) {}

func (base *BaseActivity) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	return nil, nil
}

func (base *BaseActivity) PushActivityData(playerId int64, data proto.Message) {
	if data == nil {
		log.Error("PushActivityData nil payload: actId=%v, type=%v, playerId=%v", base.GetId(), base.GetType(), playerId)
		return
	}
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
