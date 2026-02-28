package invoke

import (
	"xfx/core/define"
	"xfx/core/model"
)

type MatchModClient struct {
	invoke Invoker
	Type   string
}

func MatchClient(invoker Invoker) MatchModClient {
	return MatchModClient{
		invoke: invoker,
		Type:   define.ModuleMatch,
	}
}

// StartMatch 开始匹配
func (m MatchModClient) StartMatch(team *model.MatchTeam) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "StartMatch", team))
	return result
}

// CancelMatch 取消匹配
func (m MatchModClient) CancelMatch(mod, roomId int32) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "CancelMatch", mod, roomId))
	return result
}
