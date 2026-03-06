package transaction

import (
	"encoding/json"
	"fmt"
	"time"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_public"
	"xfx/proto/proto_transaction"
)

func Init(pl *model.Player) {
	pl.Transaction = new(model.Transaction)
	pl.Transaction.CooldownTime = 0
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Transaction)
	if err != nil {
		log.Error("player[%v],save transaction marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save transaction error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.Transaction, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save task error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.Transaction, pl.Id))
	if err != nil {
		log.Error("player[%v],load transaction error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Transaction)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load task unmarshal error:%v", pl.Id, err)
	}
	pl.Transaction = m

	// TODO:load new tasks
}

// ReqGetTransaction 获取交易所信息(冷却时间)
func ReqGetTransaction(ctx global.IPlayer, pl *model.Player, req *proto_transaction.C2SGetTransaction) {
	resp := &proto_transaction.S2CGetTransaction{}
	resp.CoolDownTime = getPlayerCooldown(pl)
	ctx.Send(resp)
}

// ReqGetTransactionList 获取交易所列表
func ReqGetTransactionList(ctx global.IPlayer, pl *model.Player, req *proto_transaction.C2SGetTransactionList) {
	resp := &proto_transaction.S2CGetTransactionList{}
	orders := invoke.TransactionClient(ctx).GetOrderList()
	resp.List = make([]*proto_transaction.TransactionOption, 0)
	for _, order := range orders {
		//发送人
		sendInfo := model.ToCommonPlayerByParam(order.SellerId, order.SellerName)
		reciviedInfo := model.ToCommonPlayerByParam(order.ReceiverId, order.ReceiverName)
		resp.List = append(resp.List, &proto_transaction.TransactionOption{
			Id: order.Id,
			AttachmentInfo: &proto_public.AttachmentOption{
				Id:         order.Id,
				SendInfo:   sendInfo,
				TargetInfo: reciviedInfo,
				Price:      order.Price,
				PriceType:  order.PriceType,
				Src:        order.Type,
				Type:       order.AttachmentData.Type,
				Value:      order.AttachmentData.ItemId,
				Level:      order.AttachmentData.Level,
				Stage:      order.AttachmentData.Stage,
				Star:       order.AttachmentData.Star,
			},
			Status: order.Status,
		})
	}
	ctx.Send(resp)
}

// ReqSendTransaction 上架商品/发送私人交易
func ReqSendTransaction(ctx global.IPlayer, pl *model.Player, req *proto_transaction.C2SSendTransaction) {
	resp := &proto_transaction.S2CSendTransaction{}

	// 验证Type类型
	if req.Type != define.TransactionTypePrivate && req.Type != define.TransactionTypeExchange {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 验证物品合法性
	if req.AttachmentInfo == nil {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 验证物品是否拥有和是否在使用中
	valid, err := internal.ValidateAttachment(pl, req.AttachmentInfo)
	if !valid || err != nil {
		log.Error("validateAttachment error: %v", err)
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 提取真实物品数据（防止客户端篡改）
	realData := internal.ExtractAttachmentData(pl, req.AttachmentInfo)

	// Type=1 私人交易：通过邮件附件发送
	if req.Type == define.TransactionTypePrivate {
		processPrivateTransaction(ctx, pl, req, realData, resp)
		ctx.Send(resp)
		return
	}

	// Type=2 交易所交易
	if req.Type == define.TransactionTypeExchange {
		processExchangeTransaction(ctx, pl, req, realData, resp)
		ctx.Send(resp)
		return
	}
}

// ReqLogicTransaction 处理交易
func ReqLogicTransaction(ctx global.IPlayer, pl *model.Player, req *proto_transaction.C2SLogicTransaction) {
	resp := &proto_transaction.S2CLogicTransaction{}

	order := invoke.TransactionClient(ctx).GetOrder(req.Id)
	if order == nil {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// Status=1 购买, Status=2 下架（交易所）
	if req.Status == 1 {
		if order.Type == define.TransactionTypeExchange || order.Type == define.TransactionTypeChat {
			log.Debug("购买交易:%v", req.Id)
			processBuyOrder(ctx, pl, order, resp)
		}
	} else if req.Status == 2 {
		if order.Type == define.TransactionTypeExchange {
			processCancelOrder(ctx, pl, order, resp)
		}
	} else {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
	}

	ctx.Send(resp)
}

// processBuyOrder 处理购买
func processBuyOrder(ctx global.IPlayer, pl *model.Player, order *model.TransactionOrder, resp *proto_transaction.S2CLogicTransaction) {
	err, processedOrder := internal.ProcessBuyOrder(ctx, pl, order)
	if err != nil {
		log.Error("processBuyOrder is error : %v", err)
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		return
	}

	Records := invoke.TransactionClient(ctx).GetTransactionRecords(processedOrder.SellerId)
	if Records != nil {
		change := false
		var model *model.TransactionRecord
		for k := 0; k < len(Records); k++ {
			if Records[k].AttachmentInfo.Id == processedOrder.Id {
				Records[k].Status = 1
				Records[k].AttachmentInfo.TransactionTime = time.Now().Unix()
				change = true
				model = Records[k]
				break
			}
		}
		if change {
			invoke.TransactionClient(ctx).UpdateTransactionRecords(processedOrder.SellerId, model)
		}
	}
	resp.Code = proto_public.CommonErrorCode_ERR_OK
}

// processCancelOrder 处理下架
func processCancelOrder(ctx global.IPlayer, pl *model.Player, order *model.TransactionOrder, resp *proto_transaction.S2CLogicTransaction) {
	canceledOrder, err := invoke.TransactionClient(ctx).CancelOrder(order.Id, pl.Id)
	if err != nil {
		log.Error("CancelOrder error: %v", err)
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		return
	}

	grantAttachmentFromData(ctx, pl, canceledOrder.AttachmentData, true)

	Records := invoke.TransactionClient(ctx).GetTransactionRecords(pl.Id)
	if Records != nil {
		change := false
		var model *model.TransactionRecord
		for k := 0; k < len(Records); k++ {
			if Records[k].AttachmentInfo.Id == canceledOrder.Id {
				Records[k].Status = 3
				change = true
				model = Records[k]
				break
			}
		}
		if change {
			invoke.TransactionClient(ctx).UpdateTransactionRecords(pl.Id, model)
		}
	}
	resp.Code = proto_public.CommonErrorCode_ERR_OK
}

// ReqTransactionRecord 获取交易记录
func ReqTransactionRecord(ctx global.IPlayer, pl *model.Player, req *proto_transaction.C2STransactionRecord) {
	resp := &proto_transaction.S2CTransactionRecord{}

	addRecord := invoke.TransactionClient(ctx).GetTransactionRecords(pl.Id)

	records := make([]*proto_transaction.TransactionRecord, 0, len(addRecord))
	for _, r := range addRecord {
		records = append(records, r.ToProto())
	}

	resp.List = records
	log.Debug("获取交易记录:%v", resp.List)
	ctx.Send(resp)
}

// grantAttachmentFromData 使用AttachmentData发放物品（使用服务端保存的真实数据）
func grantAttachmentFromData(ctx global.IPlayer, pl *model.Player, data *model.AttachmentData, cancel bool) {
	popAward := []conf.ItemE{}
	switch data.Type {
	case define.ItemTypePet:
		pl.Pet.Pets[data.ItemId] = data.Pet
		internal.SyncPetChange(ctx, pl, data.ItemId)

		//推送恭喜获得
		popAward = append(popAward, conf.ItemE{
			ItemType: define.ItemTypePet,
			ItemId:   data.ItemId,
			ItemNum:  1,
		})
	case define.ItemTypeMount:
		pl.Equip.Mount.Mount[data.ItemId] = data.Mount
		internal.SyncMountChange(ctx, pl)

		//推送恭喜获得
		popAward = append(popAward, conf.ItemE{
			ItemType: define.ItemTypeMount,
			ItemId:   data.ItemId,
			ItemNum:  1,
		})
	case define.ItemTypeWeaponry:
		pl.Equip.Weaponry.WeaponryItems[data.ItemId] = data.Weaponry
		internal.SyncWeaponryChange(ctx, pl)

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
		internal.SyncHeadWearChange(ctx, pl)

		//推送恭喜获得
		popAward = append(popAward, conf.ItemE{
			ItemType: define.ItemTypeHeadWear,
			ItemId:   data.ItemId,
			ItemNum:  1,
		})
	}

	if !cancel && len(popAward) > 0 {
		internal.PushPopReward(ctx, global.ItemFormat(popAward))
	}
}

// getPlayerCooldown 获取玩家冷却时间
func getPlayerCooldown(pl *model.Player) int64 {
	if pl.Transaction == nil || pl.Transaction.CooldownTime == 0 {
		return 0
	}
	now := time.Now().Unix()
	if pl.Transaction.CooldownTime > now {
		return pl.Transaction.CooldownTime - now
	}
	return 0
}

// setPlayerCooldown 设置玩家冷却时间
func setPlayerCooldown(pl *model.Player, seconds int64) {
	pl.Transaction.CooldownTime = time.Now().Unix() + seconds
}

// processPrivateTransaction 处理私人交易（Type=1）：通过邮件附件发送
func processPrivateTransaction(ctx global.IPlayer, pl *model.Player, req *proto_transaction.C2SSendTransaction, realData *model.AttachmentData, resp *proto_transaction.S2CSendTransaction) {
	// 验证目标玩家
	if req.AttachmentInfo.TargetInfo == nil || req.AttachmentInfo.TargetInfo.PlayerId == 0 {
		log.Error("processPrivateTransaction: TargetInfo is nil or id is 0")
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		return
	}

	// 不能给自己发送
	if req.AttachmentInfo.TargetInfo.PlayerId == pl.Id {
		log.Error("processPrivateTransaction: cannot send to self")
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		return
	}

	// 创建私人交易订单
	order := invoke.TransactionClient(ctx).CreateOrder(
		pl.Id,
		pl.Base.Name,
		realData,
		req.AttachmentInfo.PriceType,
		req.AttachmentInfo.Price,
		req.Type,
		req.AttachmentInfo.TargetInfo.PlayerId,
		req.AttachmentInfo.TargetInfo.Name,
	)

	if order == nil {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		return
	}

	// 扣除物品
	if !internal.DeductAttachment(ctx, pl, req.AttachmentInfo) {
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		return
	}

	//组装下附件信息
	attachmentObj := &model.AttachmentData{
		Id:     order.Id,
		Type:   req.AttachmentInfo.Type,
		ItemId: req.AttachmentInfo.Value,
		Level:  realData.Level,
		Star:   realData.Star,
		Stage:  realData.Stage,
	}
	attachmentStr, _ := json.Marshal(attachmentObj)
	// 发送邮件通知，带上订单ID
	invoke.MailClient(ctx).SendMail(
		define.PlayerMail,
		"交易",
		fmt.Sprintf("%s 给你发送了一份私人交易附件，请查收", pl.Base.Name),
		"Private Transaction",
		fmt.Sprintf("%s sent you a private transaction attachment", pl.Base.Name),
		pl.Base.Name,
		[]conf.ItemE{}, // 无奖励，只有附件
		[]int64{req.AttachmentInfo.TargetInfo.PlayerId},
		int64(0),
		int32(0),
		true,
		[]string{fmt.Sprintf("%d", order.Id), string(attachmentStr)}, // 将订单ID作为参数传递
	)

	invoke.MailClient(ctx).SendMail(
		define.PlayerMail,
		"交易",
		fmt.Sprintf("您给%s发送了一封交易邮件", req.AttachmentInfo.TargetInfo.Name),
		"", "", "交易所",
		[]conf.ItemE{}, []int64{pl.Id}, 0, 0, false, nil,
	)

	record := &model.TransactionRecord{
		AttachmentInfo:  internal.ToAttachmentOption(realData),
		Status:          2,
		Price:           req.AttachmentInfo.Price,
		PriceType:       req.AttachmentInfo.PriceType,
		CreateTime:      time.Now().Unix(),
		OtherPlayerId:   req.AttachmentInfo.TargetInfo.PlayerId,
		OtherPlayerName: req.AttachmentInfo.TargetInfo.Name,
	}
	record.AttachmentInfo.Id = order.Id
	record.AttachmentInfo.Src = req.Type
	// 添加交易记录
	internal.AddTransactionRecord(ctx, pl, record)

	resp.Code = proto_public.CommonErrorCode_ERR_OK
}

// processExchangeTransaction 处理交易所交易（Type=2）
func processExchangeTransaction(ctx global.IPlayer, pl *model.Player, req *proto_transaction.C2SSendTransaction, realData *model.AttachmentData, resp *proto_transaction.S2CSendTransaction) {
	// 检查冷却时间
	cooldown := getPlayerCooldown(pl)
	if cooldown > 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		return
	}

	// 创建订单（使用真实数据）
	price := int32(req.AttachmentInfo.Price)
	order := invoke.TransactionClient(ctx).CreateOrder(
		pl.Id,
		pl.Base.Name,
		realData,
		req.AttachmentInfo.PriceType,
		price,
		req.Type,
		int64(0), // 交易所无接收者
		"",       // 交易所无接收者名称
	)

	if order == nil {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		return
	}

	// 扣除物品
	if !internal.DeductAttachment(ctx, pl, req.AttachmentInfo) {
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		return
	}

	// 设置冷却时间
	setPlayerCooldown(pl, define.CooldownSeconds)

	invoke.MailClient(ctx).SendMail(
		define.PlayerMail,
		"上架成功",
		"您的商品已上架，请查看交易中心",
		"", "", "交易所",
		[]conf.ItemE{}, // 无奖励
		[]int64{pl.Id},
		int64(0), int32(0), false, nil,
	)

	record := &model.TransactionRecord{
		AttachmentInfo: internal.ToAttachmentOption(realData),
		Status:         2,
		PriceType:      req.AttachmentInfo.PriceType,
		Price:          price,
		CreateTime:     time.Now().Unix(),
	}
	record.AttachmentInfo.Id = order.Id
	record.AttachmentInfo.Src = req.Type
	// 添加交易记录
	internal.AddTransactionRecord(ctx, pl, record)

	resp.Code = proto_public.CommonErrorCode_ERR_OK
}
