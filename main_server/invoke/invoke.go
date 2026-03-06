package invoke

import (
	"fmt"

	"xfx/main_server/messages"
)

type Invoker interface {
	Invoke(mod, fn string, args ...any) (any, error)
}

// DispatchSystemMessage 模块 → 玩家系统指令。玩家进程按 Content 类型处理（踢线、刷新活动等）。
// 示例：invoke.DispatchSystemMessage(m, playerId, &messages.SysKick{Reason: "重复登录"})
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

func Bool(reply any, err error) (bool, error) {
	if err != nil {
		return false, err
	}

	switch v := reply.(type) {
	case bool:
		return v, nil
	default:
		return false, nil
	}
}

func Int64(reply any, err error) (int64, error) {
	if err != nil {
		return 0, err
	}

	switch v := reply.(type) {
	case int64:
		return v, nil
	default:
		return 0, nil
	}
}

func Int32(reply any, err error) (int32, error) {
	if err != nil {
		return 0, err
	}

	switch v := reply.(type) {
	case int32:
		return v, nil
	default:
		return 0, nil
	}
}

// As 对 invoke 返回值做安全类型断言，类型不匹配时返回 error 而非 panic
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
