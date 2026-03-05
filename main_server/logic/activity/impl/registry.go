package impl

import (
	"xfx/pkg/log"
	"xfx/proto/proto_activity"

	"github.com/golang/protobuf/proto"
)

type ActivityDesc struct {
	NewHandler func() IActivity

	NewActivityData func() any

	NewPlayerData func() any

	SetProto func(msg *proto_activity.ActivityData, data proto.Message)

	InjectFunc func(handler IActivity, data any)

	ExtractFunc func(handler IActivity) any
}

var activityRegistry = map[string]*ActivityDesc{}

// RegisterActivity 注册一个活动类型
func RegisterActivity(actType string, desc *ActivityDesc) {
	if _, exists := activityRegistry[actType]; exists {
		panic("activity type already registered: " + actType)
	}
	activityRegistry[actType] = desc
}

// GetActivityDesc 获取活动描述
func GetActivityDesc(actType string) *ActivityDesc {
	return activityRegistry[actType]
}

func SetProtoByType(actType string, msg *proto_activity.ActivityData, d proto.Message) {
	desc := GetActivityDesc(actType)
	if desc == nil || desc.SetProto == nil {
		log.Error("SetProtoByType: unknown type: %v", actType)
		return
	}
	desc.SetProto(msg, d)
}
