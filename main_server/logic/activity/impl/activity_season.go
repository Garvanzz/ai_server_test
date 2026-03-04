package impl

import (
	"time"
	"xfx/core/define"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"

	"github.com/golang/protobuf/proto"
)

// ActivitySeason 赛季活动（纯展示型，无业务逻辑）
type ActivitySeason struct {
	BaseActivity
}

// OnInit 初始化（空实现）
func (a *ActivitySeason) OnInit() {}

// OnStart 活动开始（空实现）
func (a *ActivitySeason) OnStart() {}

// OnClose 活动结束（空实现）
func (a *ActivitySeason) OnClose() {}

// OnStop 活动停止（空实现）
func (a *ActivitySeason) OnStop() {}

// OnEvent 处理事件（空实现，无事件处理）
func (a *ActivitySeason) OnEvent(key string, ctx *proto_player.Context, params EventParams) {}

// Update 定时更新（空实现）
func (a *ActivitySeason) Update(now time.Time) {}

// Format 格式化玩家数据（空实现，无玩家维度数据）
func (a *ActivitySeason) Format(ctx *proto_player.Context) proto.Message {
	return nil
}

// Router 处理协议消息（空实现，无协议处理）
func (a *ActivitySeason) Router(ctx *proto_player.Context, req proto.Message) (any, error) {
	return nil, nil
}

func init() {
	RegisterActivity(define.ActivityTypeSeason, &ActivityDesc{
		NewHandler:      func() IActivity { return new(ActivitySeason) },
		NewActivityData: func() any { return nil },
		NewPlayerData:   func() any { return nil },
		SetProto:        func(msg *proto_activity.ActivityData, data proto.Message) {},
		InjectFunc:      func(handler IActivity, data any) {},
		ExtractFunc:     func(handler IActivity) any { return nil },
	})
}
