package invoke

import (
	"xfx/core/define"
	"xfx/main_server/messages"
	"xfx/pkg/agent"
	"xfx/pkg/log"
)

type LoginModClient struct {
	invoke Invoker
	Type   string
}

func LoginClient(invoker Invoker) LoginModClient {
	return LoginModClient{
		invoke: invoker,
		Type:   define.ModuleLogin,
	}
}

// Login 玩家登录
func (m LoginModClient) Login(msg *messages.Login) (*messages.LoginResult, error) {
	result, err := m.invoke.Invoke(m.Type, "login", msg)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*messages.LoginResult), nil
}

// Logout 玩家登出
func (m LoginModClient) Logout(playerId int64) error {
	_, err := m.invoke.Invoke(m.Type, "logout", playerId)
	return err
}

// Disconnect 断开连接
func (m LoginModClient) Disconnect(playerId int64) error {
	_, err := m.invoke.Invoke(m.Type, "disconnect", playerId)
	return err
}

// CastAgent 消息转发
func (m LoginModClient) CastAgent(playerId int64, msg any) {
	_, err := m.invoke.Invoke(m.Type, "castAgent", playerId, msg)
	if err != nil {
		log.Error("CastAgent error:%v", err)
	}
}

// CastAgents 消息转发
func (m LoginModClient) CastAgents(playerIds []int64, msg any) {
	_, err := m.invoke.Invoke(m.Type, "castAgents", playerIds, msg)
	if err != nil {
		log.Error("CastAgents error:%v", err)
	}
}

// BoardCast 消息广播
func (m LoginModClient) BoardCast(msg any) {
	_, err := m.invoke.Invoke(m.Type, "boardCast", msg)
	if err != nil {
		log.Error("BoardCast error:%v", err)
	}
}

// IsOnline 获取玩家是否在线
func (m LoginModClient) IsOnline(id int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "isOnline", id))
	return result
}

// 获取玩家pid
func (m LoginModClient) getPlayerPid(dbId int64) agent.PID {
	result, err := m.invoke.Invoke(m.Type, "getPlayerPid", dbId)
	if err != nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(agent.PID)
}
