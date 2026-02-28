package event

import (
	"xfx/pkg/agent"
	"xfx/pkg/module"
)

var Router *EventRouter

type EventRouter struct {
	MapEvents map[int][]module.PID
	System    *agent.System
}

type Event struct {
	Type int            //类型
	M    map[string]any //发送内容
}

func Init(sys *agent.System) {
	Router = new(EventRouter)
	Router.MapEvents = make(map[int][]module.PID)
	Router.System = sys
}

func AddEventListener(eventType int, listener module.PID) bool {
	if Router == nil {
		return false
	}

	listeners, ok := Router.MapEvents[eventType]
	if !ok {
		listeners = []module.PID{listener}
	} else {
		for _, v := range listeners {
			if v == listener {
				return false
			}
		}

		listeners = append(listeners, listener)
	}

	Router.MapEvents[eventType] = listeners
	return true
}

func DelEventListener(eventType int, listener module.PID) {
	if Router == nil {
		return
	}

	listeners, ok := Router.MapEvents[eventType]
	if ok {
		for index, v := range listeners {
			if v == listener {
				listeners = append(listeners[:index], listeners[index+1:]...)
				Router.MapEvents[eventType] = listeners
				return
			}
		}
	}
}

func DoEvent(evType int, data map[string]any) {
	if Router == nil {
		return
	}

	if listeners, ok := Router.MapEvents[evType]; !ok {
		return
	} else {
		event := &Event{
			Type: evType,
			M:    data,
		}

		for _, listener := range listeners {
			Router.System.Cast(listener, event)
		}
	}
}
