package impl

import (
    "github.com/golang/protobuf/proto"
    "xfx/proto/proto_activity"
)

// ActivityDesc 一个活动类型的完整描述
type ActivityDesc struct {
    // 创建 handler 实例（由 handler.go 调用）
    NewHandler func() IActivity

    // 创建活动级数据的空实例，用于 JSON 反序列化（由 data.go 调用）
    // 返回 nil 表示该活动没有活动级数据
    NewActivityData func() any

    // 创建玩家数据的初始化实例（由 LoadPd 调用）
    // 返回 nil 表示该活动没有玩家数据
    NewPlayerData func() any

    // 将 Format 返回的 proto 填入 ActivityData 对应字段（由 SetProtoByType 调用）
    SetProto func(msg *proto_activity.ActivityData, data proto.Message)
}

// activityRegistry 全局注册表，key 是活动类型字符串
var activityRegistry = map[string]*ActivityDesc{}

// RegisterActivity 注册一个活动类型
func RegisterActivity(actType string, desc *ActivityDesc) {
    if _, exists := activityRegistry[actType]; exists {
        panic("activity type already registered: " + actType)
    }
    activityRegistry[actType] = desc
}

// GetActivityDesc 获取描述，找不到返回 nil
func GetActivityDesc(actType string) *ActivityDesc {
    return activityRegistry[actType]
}