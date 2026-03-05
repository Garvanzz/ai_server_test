package event

import (
	"strconv"

	"xfx/pkg/agent"
	"xfx/pkg/module"
)

var Router *EventRouter

type EventRouter struct {
	System *agent.System
}

// Event 业务层事件 payload，通过 eventbus 投递
type Event struct {
	Type int            // 类型
	M    map[string]any // 发送内容
}

// topic 将 int 事件类型映射为 eventbus 的 string topic，避免与其他 topic 冲突
func topic(evType int) string {
	return "e:" + strconv.Itoa(evType)
}

func Init(sys *agent.System) {
	Router = &EventRouter{System: sys}
}

func AddEventListener(eventType int, listener module.PID) bool {
	if Router == nil || listener == nil {
		return false
	}
	Router.System.Subscribe(topic(eventType), listener)
	return true
}

func DelEventListener(eventType int, listener module.PID) {
	if Router == nil || listener == nil {
		return
	}
	Router.System.Unsubscribe(topic(eventType), listener)
}

func DoEvent(evType int, data map[string]any) {
	if Router == nil {
		return
	}
	Router.System.Publish(topic(evType), &Event{
		Type: evType,
		M:    data,
	})
}
