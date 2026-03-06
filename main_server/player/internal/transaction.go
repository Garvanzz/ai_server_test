package internal

import (
	"errors"
	"fmt"
	"time"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_public"
)

// ProcessBuyOrder 处理购买
func ProcessBuyOrder(ctx global.IPlayer, pl *model.Player, order *model.TransactionOrder) (error, *model.TransactionOrder) {
	id := define.ItemIdBoxLingyu
	if order.PriceType == 1 {
		id = define.ItemIdXianyu
	}
	costs := map[int32]int32{int32(id): order.Price}
	if !CheckItemsEnough(pl, costs) {
		return errors.New("item is not Enough"), nil
	}

	processedOrder, err := invoke.TransactionClient(ctx).ProcessOrder(order.Id, pl.Id)
	if err != nil {
		return err, nil
	}

	SubItems(ctx, pl, costs)
	GrantAttachmentFromData(ctx, pl, processedOrder.AttachmentData, false)

	mailItems := []conf.ItemE{{
		ItemType: define.ItemTypeItem,
		ItemId:   int32(id),
		ItemNum:  processedOrder.Price,
	}}

	invoke.MailClient(ctx).SendMail(
		define.PlayerMail,
		"交易成功",
		"您的商品已售出",
		"", "", "交易所",
		mailItems,
		[]int64{processedOrder.SellerId},
		int64(0), int32(0), false, []string{},
	)

	return nil, processedOrder
}

// GrantAttachmentFromData 使用AttachmentData发放物品（使用服务端保存的真实数据）
func GrantAttachmentFromData(ctx global.IPlayer, pl *model.Player, data *model.AttachmentData, cancel bool) {
	popAward := []conf.ItemE{}
	switch data.Type {
	case define.ItemTypePet:
		pl.Pet.Pets[data.ItemId] = data.Pet
		SyncPetChange(ctx, pl, data.ItemId)

		//推送恭喜获得
		popAward = append(popAward, conf.ItemE{
			ItemType: define.ItemTypePet,
			ItemId:   data.ItemId,
			ItemNum:  1,
		})
	case define.ItemTypeMount:
		pl.Equip.Mount.Mount[data.ItemId] = data.Mount
		SyncMountChange(ctx, pl)

		//推送恭喜获得
		popAward = append(popAward, conf.ItemE{
			ItemType: define.ItemTypeMount,
			ItemId:   data.ItemId,
			ItemNum:  1,
		})
	case define.ItemTypeWeaponry:
		pl.Equip.Weaponry.WeaponryItems[data.ItemId] = data.Weaponry
		SyncWeaponryChange(ctx, pl)

		//推送恭喜获得
		popAward = append(popAward, conf.ItemE{
			ItemType: define.ItemTypeWeaponry,
			ItemId:   data.ItemId,
			ItemNum:  1,
		})
	case define.ItemTypeBraces:
		pl.Equip.Brace.BraceItems[data.ItemId] = data.Brace

		//推送恭喜获得
		popAward = append(popAward, conf.ItemE{
			ItemType: define.ItemTypeBraces,
			ItemId:   data.ItemId,
			ItemNum:  1,
		})
	case define.ItemTypeHeadWear:
		pl.Fashion.HeadWear[data.ItemId] = data.HeadWear
		SyncHeadWearChange(ctx, pl)

		//推送恭喜获得
		popAward = append(popAward, conf.ItemE{
			ItemType: define.ItemTypeHeadWear,
			ItemId:   data.ItemId,
			ItemNum:  1,
		})
	}

	if !cancel && len(popAward) > 0 {
		PushPopReward(ctx, global.ItemFormat(popAward))
	}
}

// ChatSendTransaction 聊天附件
func ChatSendTransaction(ctx global.IPlayer, pl *model.Player, opt *proto_public.AttachmentOption) (*model.AttachmentData, error) {
	// 验证物品合法性
	if opt == nil {
		return nil, errors.New("opt is null")
	}

	// 验证物品是否拥有和是否在使用中
	valid, err := ValidateAttachment(pl, opt)
	if !valid || err != nil {
		log.Error("validateAttachment error: %v", err)
		return nil, err
	}

	// 提取真实物品数据（防止客户端篡改）
	realData := ExtractAttachmentData(pl, opt)

	if opt.Src != define.TransactionTypeChat {
		return nil, errors.New("this is type is error")
	}

	data, err := processChatTransaction(ctx, pl, realData, opt)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ExtractAttachmentData 从玩家数据中提取真实的物品信息
func ExtractAttachmentData(pl *model.Player, att *proto_public.AttachmentOption) *model.AttachmentData {
	data := &model.AttachmentData{
		Type:   att.Type,
		ItemId: att.Value,
	}

	switch att.Type {
	case define.ItemTypePet:
		if pet, ok := pl.Pet.Pets[att.Value]; ok {
			data.Level = pet.Level
			data.Stage = pet.Stage
			data.Star = pet.Star
			data.Name = pet.Name
			data.Pet = pet
		}
	case define.ItemTypeMount:
		if mount, ok := pl.Equip.Mount.Mount[att.Value]; ok {
			data.Level = mount.Level
			data.Name = mount.Name
			data.Mount = mount
		}
	case define.ItemTypeWeaponry:
		if weaponry, ok := pl.Equip.Weaponry.WeaponryItems[att.Value]; ok {
			data.Level = weaponry.Level
			data.Weaponry = weaponry
		}
	case define.ItemTypeBraces:
		if brace, ok := pl.Equip.Brace.BraceItems[att.Value]; ok {
			data.Level = brace.Level
			data.Brace = brace
		}
	case define.ItemTypeFashion:
		// 时装一般没有等级数据
	case define.ItemTypeHeadWear:
		if headWear, ok := pl.Fashion.HeadWear[att.Value]; ok {
			data.HeadWear = headWear
		}
	}

	return data
}

// processChatTransaction
func processChatTransaction(ctx global.IPlayer, pl *model.Player, realData *model.AttachmentData, opt *proto_public.AttachmentOption) (*model.AttachmentData, error) {
	// 验证目标玩家
	if opt.TargetInfo == nil || opt.TargetInfo.PlayerId != 0 {
		log.Error("processPrivateTransaction: TargetInfo is nil or id is not 0")
		return nil, errors.New("processPrivateTransaction: TargetInfo is nil or id is not 0")
	}

	// 创建私人交易订单
	order := invoke.TransactionClient(ctx).CreateOrder(
		pl.Id,
		pl.Base.Name,
		realData,
		opt.PriceType,
		opt.Price,
		define.TransactionTypeChat,
		opt.TargetInfo.PlayerId,
		opt.TargetInfo.Name,
	)

	if order == nil {
		return nil, errors.New("processPrivateTransaction: order is null")
	}

	// 扣除物品
	if !DeductAttachment(ctx, pl, opt) {
		return nil, errors.New("processPrivateTransaction: DeductAttachment is null")
	}

	//组装下附件信息
	attachmentObj := &model.AttachmentData{
		Id:     order.Id,
		Type:   opt.Type,
		ItemId: opt.Value,
		Level:  realData.Level,
		Star:   realData.Star,
		Stage:  realData.Stage,
	}
	// 发送邮件通知，带上订单ID
	invoke.MailClient(ctx).SendMail(
		define.PlayerMail,
		"交易",
		fmt.Sprintf("您的商品在世界频道中已发出，请查看"),
		"",
		fmt.Sprintf(""),
		pl.Base.Name,
		[]conf.ItemE{}, // 无奖励，只有附件
		[]int64{pl.Id},
		int64(0),
		int32(0),
		false,
		[]string{},
	)

	record := &model.TransactionRecord{
		AttachmentInfo:  ToAttachmentOption(realData),
		Status:          2,
		PriceType:       opt.PriceType,
		Price:           opt.Price,
		CreateTime:      time.Now().Unix(),
		OtherPlayerId:   opt.TargetInfo.PlayerId,
		OtherPlayerName: opt.TargetInfo.Name,
	}
	record.AttachmentInfo.Id = order.Id
	record.AttachmentInfo.Src = define.TransactionTypeChat
	// 添加交易记录
	AddTransactionRecord(ctx, pl, record)

	return attachmentObj, nil
}

// ValidateAttachment 验证物品合法性
func ValidateAttachment(pl *model.Player, att *proto_public.AttachmentOption) (bool, error) {
	switch att.Type {
	case define.ItemTypePet:
		pet, ok := pl.Pet.Pets[att.Value]
		if !ok {
			return false, fmt.Errorf("pet not found")
		}
		for petId := range pl.Pet.DispatchPets {
			if petId == att.Value {
				return false, fmt.Errorf("pet is in dispatch")
			}
		}
		if pet.IsUse {
			return false, fmt.Errorf("pet is in use")
		}

	case define.ItemTypeMount:
		if _, ok := pl.Equip.Mount.Mount[att.Value]; !ok {
			return false, fmt.Errorf("mount not found")
		}
		if pl.Equip.Mount.UseId == att.Value {
			return false, fmt.Errorf("mount is in use")
		}
	case define.ItemTypeWeaponry:
		if _, ok := pl.Equip.Weaponry.WeaponryItems[att.Value]; !ok {
			return false, fmt.Errorf("weaponry not found")
		}
		if pl.Equip.Weaponry.UseId == att.Value {
			return false, fmt.Errorf("weaponry is in use")
		}
	case define.ItemTypeBraces:
		item, ok := pl.Equip.Brace.BraceItems[att.Value]
		if !ok {
			return false, fmt.Errorf("brace not found")
		}
		if item.IsUse {
			return false, fmt.Errorf("brace is in use")
		}
	case define.ItemTypeFashion:
		item, ok := pl.Fashion.FashionItems[att.Value]
		if !ok {
			return false, fmt.Errorf("fashion not found")
		}
		if item.Use {
			return false, fmt.Errorf("fashion is in use")
		}
	case define.ItemTypeHeadWear:
		item, ok := pl.Fashion.HeadWear[att.Value]
		if !ok {
			return false, fmt.Errorf("headwear not found")
		}
		if item.Use {
			return false, fmt.Errorf("headwear is in use")
		}
	default:
		return false, fmt.Errorf("unsupported item type")
	}
	return true, nil
}

// DeductAttachment 扣除物品
func DeductAttachment(ctx global.IPlayer, pl *model.Player, att *proto_public.AttachmentOption) bool {
	switch att.Type {
	case define.ItemTypePet:
		if _, ok := pl.Pet.Pets[att.Value]; !ok {
			return false
		}
		delete(pl.Pet.Pets, att.Value)
		SyneAllPet(ctx, pl)
	case define.ItemTypeMount:
		if _, ok := pl.Equip.Mount.Mount[att.Value]; !ok {
			return false
		}
		delete(pl.Equip.Mount.Mount, att.Value)
		SyncMountChange(ctx, pl)
	case define.ItemTypeWeaponry:
		if _, ok := pl.Equip.Weaponry.WeaponryItems[att.Value]; !ok {
			return false
		}
		delete(pl.Equip.Weaponry.WeaponryItems, att.Value)
		SyncWeaponryChange(ctx, pl)
	case define.ItemTypeBraces:
		if _, ok := pl.Equip.Brace.BraceItems[att.Value]; !ok {
			return false
		}
		delete(pl.Equip.Brace.BraceItems, att.Value)
	case define.ItemTypeFashion:
		if _, ok := pl.Fashion.FashionItems[att.Value]; !ok {
			return false
		}
		delete(pl.Fashion.FashionItems, att.Value)
		SyncFashionChange(ctx, pl)
	case define.ItemTypeHeadWear:
		if _, ok := pl.Fashion.HeadWear[att.Value]; !ok {
			return false
		}
		delete(pl.Fashion.HeadWear, att.Value)
		SyncHeadWearChange(ctx, pl)
	default:
		return false
	}
	return true
}

// AddTransactionRecord 添加交易记录
func AddTransactionRecord(ctx global.IPlayer, pl *model.Player, record *model.TransactionRecord) {
	invoke.TransactionClient(ctx).AddTransactionRecord(pl.Id, record)
}

// ToAttachmentOption 将AttachmentData转换为AttachmentOption
func ToAttachmentOption(data *model.AttachmentData) *proto_public.AttachmentOption {
	return &proto_public.AttachmentOption{
		Id:    data.Id,
		Type:  data.Type,
		Value: data.ItemId,
		Level: data.Level,
		Stage: data.Stage,
		Star:  data.Star,
	}
}
