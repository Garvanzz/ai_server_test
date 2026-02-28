package invoke

import (
	"github.com/golang/protobuf/proto"
	"xfx/core/define"
	"xfx/proto/proto_activity"
	"xfx/proto/proto_player"
)

type ActivityModClient struct {
	invoke Invoker
	Type   string
}

func ActivityClient(invoker Invoker) ActivityModClient {
	return ActivityModClient{
		invoke: invoker,
		Type:   define.ModuleActivity,
	}
}

func (m ActivityModClient) GetActivityStatus() ([]*proto_activity.ActivityStatusInfo, error) {
	result, err := m.invoke.Invoke(m.Type, "GetActivityStatus")
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.([]*proto_activity.ActivityStatusInfo), nil
}

func (m ActivityModClient) GetActivityStatusByType(typ string) (*proto_activity.ActivityStatusInfo, error) {
	result, err := m.invoke.Invoke(m.Type, "GetActivityStatusByType", typ)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*proto_activity.ActivityStatusInfo), nil
}

func (m ActivityModClient) GetActivityData(ctx *proto_player.Context, id int64) (*proto_activity.ActivityData, error) {
	result, err := m.invoke.Invoke(m.Type, "GetActivityData", ctx, id)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*proto_activity.ActivityData), nil
}

func (m ActivityModClient) GetActivityDataList(ctx *proto_player.Context, ids []int64) []*proto_activity.ActivityData {
	result, err := m.invoke.Invoke(m.Type, "GetActivityDataList", ctx, ids)
	if err != nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return result.([]*proto_activity.ActivityData)
}

func (m ActivityModClient) OnRouterMsg(ctx *proto_player.Context, actId int64, req proto.Message) (any, error) {
	return m.invoke.Invoke(m.Type, "OnRouterMsg", ctx, actId, req)
}
