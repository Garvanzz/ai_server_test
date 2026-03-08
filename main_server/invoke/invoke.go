package invoke

import (
	"fmt"

	"xfx/main_server/messages"
)

type Invoker interface {
	Invoke(mod, fn string, args ...any) (any, error)
}

// DispatchSystemMessage 转发给玩家系统指令
func DispatchSystemMessage(m Invoker, playerId int64, msg any) {
	LoginClient(m).CastAgent(playerId, &messages.SysMessage{
		Content: msg,
	})
}

// Dispatch 转发消息给玩家
func Dispatch(m Invoker, playerId int64, msg any) {
	LoginClient(m).CastAgent(playerId, &messages.DispatchMessage{
		Content: msg,
	})
}

// DispatchAllPlayer 玩家广播
func DispatchAllPlayer(m Invoker, msg any) {
	LoginClient(m).BoardCast(&messages.DispatchMessage{
		Content: msg,
	})
}

// DispatchPlayers 给指定玩家发送消息
func DispatchPlayers(m Invoker, playerIds []int64, msg any) {
	LoginClient(m).CastAgents(playerIds, &messages.DispatchMessage{
		Content: msg,
	})
}

// As 对 invoke 返回值做安全类型断言
func As[T any](result any) (T, error) {
	var zero T
	if result == nil {
		return zero, nil
	}
	v, ok := result.(T)
	if !ok {
		return zero, fmt.Errorf("invoke: type assertion failed, expected %T, got %T", zero, result)
	}
	return v, nil
}
