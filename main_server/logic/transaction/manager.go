package transaction

import (
	"encoding/json"
	"fmt"
	"time"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
)

var Module = func() module.Module {
	return &Manager{
		orders:             make(map[int64]*model.TransactionOrder),
		playerOrders:       make(map[int64][]int64),
		receiverOrders:     make(map[int64][]int64),
		TransactionRecords: make(map[int64][]*model.TransactionRecord),
	}
}

type Manager struct {
	modules.BaseModule
	orderId            int64
	orders             map[int64]*model.TransactionOrder
	playerOrders       map[int64][]int64 // 发送者的订单列表
	receiverOrders     map[int64][]int64 // 接收者的订单列表（私人交易）
	lastSaveTime       int64
	TransactionRecords map[int64][]*model.TransactionRecord
}

func (m *Manager) OnInit(app module.App) {
	m.BaseModule.OnInit(app)
	m.loadData()

	m.Register("CreateOrder", m.CreateOrder)
	m.Register("UpdateOrder", m.UpdateOrder)
	m.Register("GetOrder", m.GetOrder)
	m.Register("GetOrderList", m.GetOrderList)
	m.Register("ProcessOrder", m.ProcessOrder)
	m.Register("CancelOrder", m.CancelOrder)
	m.Register("addTransactionRecord", m.addTransactionRecord)
	m.Register("getTransactionRecords", m.getTransactionRecords)
	m.Register("updateTransactionRecords", m.updateTransactionRecord)
}

func (m *Manager) loadData() {
	rdb, err := db.GetEngine(m.App.GetEnv().ID)
	if err != nil {
		log.Error("transaction loadData error: %v", err)
		return
	}

	reply, err := rdb.RedisExec("GET", define.TransactionOrder)
	if err != nil {
		log.Error("load transaction order error: %v", err)
		return
	}

	if reply != nil {
		data := make(map[string]interface{})
		err = json.Unmarshal(reply.([]byte), &data)
		if err != nil {
			log.Error("unmarshal transaction data error: %v", err)
			return
		}

		if id, ok := data["orderId"]; ok {
			m.orderId = int64(id.(float64))
		}
		if orders, ok := data["orders"]; ok {
			ordersData, _ := json.Marshal(orders)
			json.Unmarshal(ordersData, &m.orders)
		}
		if playerOrders, ok := data["playerOrders"]; ok {
			playerOrdersData, _ := json.Marshal(playerOrders)
			json.Unmarshal(playerOrdersData, &m.playerOrders)
		}
		if receiverOrders, ok := data["receiverOrders"]; ok {
			receiverOrdersData, _ := json.Marshal(receiverOrders)
			json.Unmarshal(receiverOrdersData, &m.receiverOrders)
		}
	}

	log.Debug("transaction loadData success, orderId: %d, orders: %d", m.orderId, len(m.orders))

	reply_records, err := rdb.RedisExec("GET", define.TransactionRecords)
	if err != nil {
		log.Error("load transaction  reply_records order error: %v", err)
		return
	}

	if reply_records != nil {
		if m.TransactionRecords == nil {
			m.TransactionRecords = make(map[int64][]*model.TransactionRecord)
		}
		data := make(map[int64][]*model.TransactionRecord)
		err = json.Unmarshal(reply_records.([]byte), &data)
		if err != nil {
			log.Error("unmarshal reply_records data error: %v", err)
			return
		}
		m.TransactionRecords = data
	}
}

func (m *Manager) GetType() string { return define.ModuleTransaction }

func (m *Manager) OnTick(delta time.Duration) {
	now := time.Now().Unix()
	if now-m.lastSaveTime >= 60 {
		m.saveToRedis()
		m.lastSaveTime = now
	}
}

func (m *Manager) OnDestroy() {
	m.saveToRedis()
}

func (m *Manager) saveToRedis() {
	rdb, err := db.GetEngine(m.App.GetEnv().ID)
	if err != nil {
		log.Error("save transaction error: %v", err)
		return
	}

	data := map[string]interface{}{
		"orderId":        m.orderId,
		"orders":         m.orders,
		"playerOrders":   m.playerOrders,
		"receiverOrders": m.receiverOrders,
	}

	b, err := json.Marshal(data)
	if err != nil {
		log.Error("marshal transaction data error: %v", err)
		return
	}

	rdb.RedisExec("SET", define.TransactionOrder, string(b))

	c, err := json.Marshal(m.TransactionRecords)
	if err != nil {
		log.Error("save TransactionRecords Init error:%v", err)
		return
	}

	rdb.RedisExec("SET", define.TransactionRecords, string(c))

}

func (m *Manager) OnMessage(msg interface{}) interface{} {
	return nil
}

// CreateOrder 创建订单
func (m *Manager) CreateOrder(sellerId int64, sellerName string, attachmentData *model.AttachmentData, priceType, price int32, orderType int32, receiverId int64, receiverName string) *model.TransactionOrder {
	m.orderId++
	order := &model.TransactionOrder{
		Id:             m.orderId,
		SellerId:       sellerId,
		SellerName:     sellerName,
		ReceiverId:     receiverId,
		ReceiverName:   receiverName,
		AttachmentData: attachmentData,
		PriceType:      priceType,
		Price:          price,
		Status:         0,
		Type:           orderType,
		CreateTime:     time.Now().Unix(),
		UpdateTime:     time.Now().Unix(),
	}
	m.orders[order.Id] = order
	m.playerOrders[sellerId] = append(m.playerOrders[sellerId], order.Id)

	// Type=1私人交易时，添加到接收者索引
	if orderType == 1 && receiverId > 0 {
		m.receiverOrders[receiverId] = append(m.receiverOrders[receiverId], order.Id)
	}

	return order
}

// GetOrder 获取订单
func (m *Manager) GetOrder(orderId int64) *model.TransactionOrder {
	return m.orders[orderId]
}

// GetOrderList 获取订单列表
func (m *Manager) GetOrderList() []*model.TransactionOrder {
	list := make([]*model.TransactionOrder, 0)
	for _, order := range m.orders {
		if order.Status == 0 && order.Type == 2 {
			list = append(list, order)
		}
	}
	return list
}

// ProcessOrder 处理订单(购买)
func (m *Manager) ProcessOrder(orderId int64, buyerId int64) (*model.TransactionOrder, error) {
	order := m.orders[orderId]
	if order == nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if order.Status != 0 {
		return nil, fmt.Errorf("订单已被处理")
	}
	if order.SellerId == buyerId {
		return nil, fmt.Errorf("不能购买自己的商品")
	}

	order.Status = 1
	order.UpdateTime = time.Now().Unix()
	m.removeOrderFromPlayer(order.SellerId, orderId)
	delete(m.orders, orderId)
	return order, nil
}

// UpdateOrder 更新订单
func (m *Manager) UpdateOrder(orderId int64, _order *model.TransactionOrder) (*model.TransactionOrder, error) {
	order := m.orders[orderId]
	if order == nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if order.Status != 0 {
		return nil, fmt.Errorf("订单已被处理")
	}

	order = _order
	return order, nil
}

// CancelOrder 取消订单(下架)
func (m *Manager) CancelOrder(orderId int64, playerId int64) (*model.TransactionOrder, error) {
	order := m.orders[orderId]
	if order == nil {
		return nil, fmt.Errorf("订单不存在")
	}
	if order.SellerId != playerId {
		return nil, fmt.Errorf("无权操作")
	}
	if order.Status != 0 {
		return nil, fmt.Errorf("订单已被处理")
	}

	order.Status = 3
	order.UpdateTime = time.Now().Unix()
	m.removeOrderFromPlayer(playerId, orderId)
	delete(m.orders, orderId)
	return order, nil
}

func (m *Manager) removeOrderFromPlayer(playerId int64, orderId int64) {
	orders := m.playerOrders[playerId]
	for i, id := range orders {
		if id == orderId {
			m.playerOrders[playerId] = append(orders[:i], orders[i+1:]...)
			return
		}
	}
}

func (m *Manager) removeOrderFromReceiver(receiverId int64, orderId int64) {
	orders := m.receiverOrders[receiverId]
	for i, id := range orders {
		if id == orderId {
			m.receiverOrders[receiverId] = append(orders[:i], orders[i+1:]...)
			return
		}
	}
}

func (mgr *Manager) addTransactionRecord(playerId int64, record *model.TransactionRecord) {
	if mgr.TransactionRecords == nil {
		mgr.TransactionRecords = make(map[int64][]*model.TransactionRecord)
	}

	records, ok := mgr.TransactionRecords[playerId]
	if !ok {
		mgr.TransactionRecords[playerId] = []*model.TransactionRecord{record}
	} else {
		mgr.TransactionRecords[playerId] = append(records, record)
	}
}

func (mgr *Manager) updateTransactionRecord(playerId int64, record *model.TransactionRecord) {
	if mgr.TransactionRecords == nil {
		return
	}

	records, ok := mgr.TransactionRecords[playerId]
	if !ok {
		return
	} else {
		for k := 0; k < len(records); k++ {
			if records[k].AttachmentInfo.Id == record.AttachmentInfo.Id {
				records[k] = record
				break
			}
		}
		mgr.TransactionRecords[playerId] = records
	}
}

func (mgr *Manager) getTransactionRecords(playerId int64) []*model.TransactionRecord {
	if mgr.TransactionRecords == nil {
		mgr.TransactionRecords = make(map[int64][]*model.TransactionRecord)
	}

	records, ok := mgr.TransactionRecords[playerId]
	if !ok {
		return nil
	}

	return records
}
