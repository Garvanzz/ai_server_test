package invoke

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

type CommonModClient struct {
	invoke Invoker
	Type   string
}

func CommonClient(invoker Invoker) CommonModClient {
	return CommonModClient{
		invoke: invoker,
		Type:   define.ModuleCommon,
	}
}

// ReleaseRobot 释放机器人
func (m CommonModClient) ReleaseRobot(robotId int32) {
	_, err := m.invoke.Invoke(m.Type, "releaseRobot", robotId)
	if err != nil {
		log.Error("ReleaseRobot err:%v", err)
	}
}

// MatchRobot 匹配机器人
func (m CommonModClient) MatchRobot(startPower int64, endPower int64) (*model.Robot, error) {
	result, err := m.invoke.Invoke(m.Type, "matchRobot", startPower, endPower)
	if err != nil {
		log.Error("GetSystemMailById err:%v", err)
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	return result.(*model.Robot), nil
}
