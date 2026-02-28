package invoke

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

type HuaguoshanModClient struct {
	invoke Invoker
	Type   string
}

func HuaguoshanClient(invoker Invoker) HuaguoshanModClient {
	return HuaguoshanModClient{
		invoke: invoker,
		Type:   define.ModuleHuaguoshan,
	}
}

// CreateInvite 创建伴侣邀请
func (m HuaguoshanModClient) CreateInvite(senderId int64, senderName string, receiverId int64) *model.PartnerInvite {
	result, err := m.invoke.Invoke(m.Type, "CreateInvite", senderId, senderName, receiverId)
	if err != nil || result == nil {
		log.Error("CreateInvite error: %v", err)
		return nil
	}
	return result.(*model.PartnerInvite)
}

// GetInvite 获取邀请
func (m HuaguoshanModClient) GetInvite(inviteId int64) *model.PartnerInvite {
	result, err := m.invoke.Invoke(m.Type, "GetInvite", inviteId)
	if err != nil || result == nil {
		return nil
	}
	return result.(*model.PartnerInvite)
}

// GetReceiverInvites 获取接收者的邀请列表
func (m HuaguoshanModClient) GetReceiverInvites(receiverId int64) []*model.PartnerInvite {
	result, err := m.invoke.Invoke(m.Type, "GetReceiverInvites", receiverId)
	if err != nil || result == nil {
		return []*model.PartnerInvite{}
	}
	return result.([]*model.PartnerInvite)
}

// ProcessInvite 处理邀请
func (m HuaguoshanModClient) ProcessInvite(inviteId int64, accept bool) (*model.PartnerInvite, error) {
	result, err := m.invoke.Invoke(m.Type, "ProcessInvite", inviteId, accept)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return result.(*model.PartnerInvite), nil
}
