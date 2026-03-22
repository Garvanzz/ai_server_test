package impl

import (
	"fmt"
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
	if desc == nil {
		panic("activity desc is nil: " + actType)
	}
	if _, exists := activityRegistry[actType]; exists {
		panic("activity type already registered: " + actType)
	}
	if desc.NewHandler == nil {
		panic("activity NewHandler is nil: " + actType)
	}
	if desc.SetProto == nil {
		panic("activity SetProto is nil: " + actType)
	}
	handler := desc.NewHandler()
	if handler == nil {
		panic("activity NewHandler returned nil: " + actType)
	}
	if desc.InjectFunc != nil && desc.NewActivityData == nil {
		panic("activity InjectFunc requires NewActivityData: " + actType)
	}
	if desc.ExtractFunc != nil && desc.NewActivityData == nil {
		panic("activity ExtractFunc requires NewActivityData: " + actType)
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
	if d == nil {
		log.Error("SetProtoByType: nil payload for type: %v", actType)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Error("SetProtoByType panic: type=%v, payload=%T, err=%v", actType, d, fmt.Sprint(r))
		}
	}()
	desc.SetProto(msg, d)
}
