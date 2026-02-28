package invoke

import (
	"xfx/main_server/messages"
)

type Invoker interface {
	Invoke(mod, fn string, args ...any) (any, error)
}

// DispatchSystemMessage 转发系统消息给玩家
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
