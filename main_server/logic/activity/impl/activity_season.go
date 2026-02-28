package impl

import (
	"github.com/golang/protobuf/proto"
	"time"
	"xfx/proto/proto_player"
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

// Inject 注入活动数据（空实现，无活动专属数据）
func (a *ActivitySeason) Inject(data any) {}

// Extract 提取活动数据（空实现，无活动专属数据）
func (a *ActivitySeason) Extract() any {
	return nil
}
