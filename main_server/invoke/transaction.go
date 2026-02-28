package invoke

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

type TransactionModClient struct {
	invoke Invoker
	Type   string
}

func TransactionClient(invoker Invoker) TransactionModClient {
	return TransactionModClient{
		invoke: invoker,
		Type:   define.ModuleTransaction,
	}
}

func (m TransactionModClient) CreateOrder(sellerId int64, sellerName string, attachmentData *model.AttachmentData, priceType, price int32, orderType int32, receiverId int64, receiverName string) *model.TransactionOrder {
	result, err := m.invoke.Invoke(m.Type, "CreateOrder", sellerId, sellerName, attachmentData, priceType, price, orderType, receiverId, receiverName)
	if err != nil || result == nil {
		return nil
	}
	return result.(*model.TransactionOrder)
}

func (m TransactionModClient) GetOrder(orderId int64) *model.TransactionOrder {
	result, err := m.invoke.Invoke(m.Type, "GetOrder", orderId)
	if err != nil || result == nil {
		return nil
	}
	return result.(*model.TransactionOrder)
}

func (m TransactionModClient) GetOrderList() []*model.TransactionOrder {
	result, err := m.invoke.Invoke(m.Type, "GetOrderList")
	if err != nil || result == nil {
		return []*model.TransactionOrder{}
	}
	return result.([]*model.TransactionOrder)
}

func (m TransactionModClient) ProcessOrder(orderId int64, buyerId int64) (*model.TransactionOrder, error) {
	result, err := m.invoke.Invoke(m.Type, "ProcessOrder", orderId, buyerId)
	if err != nil {
		return nil, err
	}
	return result.(*model.TransactionOrder), nil
}

func (m TransactionModClient) CancelOrder(orderId int64, playerId int64) (*model.TransactionOrder, error) {
	result, err := m.invoke.Invoke(m.Type, "CancelOrder", orderId, playerId)
	if err != nil {
		return nil, err
	}
	return result.(*model.TransactionOrder), nil
}

func (m TransactionModClient) UpdateOrder(orderId int64, order *model.TransactionOrder) (*model.TransactionOrder, error) {
	result, err := m.invoke.Invoke(m.Type, "UpdateOrder", orderId, order)
	if err != nil {
		return nil, err
	}
	return result.(*model.TransactionOrder), nil
}

// AddTransactionRecord
func (m TransactionModClient) AddTransactionRecord(playerId int64, transaction *model.TransactionRecord) {
	_, err := m.invoke.Invoke(m.Type, "addTransactionRecord", playerId, transaction)
	if err != nil {
		log.Error("AddTransactionRecord err:%v", err)
	}
}

// GetTransactionRecords
func (m TransactionModClient) GetTransactionRecords(playerId int64) []*model.TransactionRecord {
	result, err := m.invoke.Invoke(m.Type, "getTransactionRecords", playerId)
	if err != nil {
		log.Error("GetTransactionRecords err:%v", err)
		return nil
	}
	if result == nil {
		return nil
	}

	return result.([]*model.TransactionRecord)
}

// UpdateTransactionRecords
func (m TransactionModClient) UpdateTransactionRecords(playerId int64, transaction *model.TransactionRecord) {
	result, err := m.invoke.Invoke(m.Type, "updateTransactionRecords", playerId)
	if err != nil {
		log.Error("updateTransactionRecords err:%v", err)
		return
	}
	if result == nil {
		return
	}
}
