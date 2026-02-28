package model

import (
	"xfx/proto/proto_public"
	"xfx/proto/proto_transaction"
)

type Transaction struct {
	CooldownTime int64
}

type TransactionRecord struct {
	AttachmentInfo  *proto_public.AttachmentOption
	Status          int32
	PriceType       int32
	Price           int32
	CreateTime      int64
	OtherPlayerId   int64
	OtherPlayerName string
}

func (t *TransactionRecord) ToProto() *proto_transaction.TransactionRecord {
	record := &proto_transaction.TransactionRecord{
		AttachmentInfo: t.AttachmentInfo,
		Status:         t.Status,
	}
	record.AttachmentInfo.Price = t.Price
	record.AttachmentInfo.PriceType = t.PriceType
	return record
}

type AttachmentData struct {
	Id       int64
	Type     int32
	ItemId   int32
	Level    int32
	Stage    int32
	Star     int32
	Name     string
	Pet      *PetItem         //宠物
	Mount    *MountItemOption //坐骑
	Weaponry *WeaponryItem
	Brace    *BraceItem
	HeadWear *HeadWearItem

	// 其他类型可扩展
}

type TransactionOrder struct {
	Id             int64
	SellerId       int64
	SellerName     string
	ReceiverId     int64  // 接收者ID（Type=1私人交易时使用）
	ReceiverName   string // 接收者名称（Type=1私人交易时使用）
	AttachmentData *AttachmentData
	Price          int32
	PriceType      int32
	Status         int32 // 0:待处理 1:已完成 2:退还 3:下架
	Type           int32 // 1:私人交易 2:交易所 3:世界
	CreateTime     int64
	UpdateTime     int64
}
