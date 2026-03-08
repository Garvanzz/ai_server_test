package mail

import (
	"encoding/json"
	"fmt"
	"strconv"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_mail"
)

// ReqMailList 请求邮件列表
func ReqMailList(ctx global.IPlayer, pl *model.Player, req *proto_mail.C2SMailList) {
	// 检查新邮件
	checkNewSysMail(ctx, pl)

	pushMails := make([]*proto_mail.Mail, 0)
	mails := getMailsFromDB(pl.Cache.App.GetEnv().ID, pl.Id)
	for _, mail := range mails {
		pushMails = append(pushMails, mail.ToProto())
	}

	ctx.Send(&proto_mail.S2CMailList{Mails: pushMails})
}

// ReqOpenMail 打开邮件
func ReqOpenMail(ctx global.IPlayer, pl *model.Player, req *proto_mail.C2SOpenMail) {
	resp := new(proto_mail.S2COpenMail)
	resp.Mail = make([]*proto_mail.Mail, 0)
	for _, id := range req.Id {
		if id == 0 {
			log.Error("reqOpenMail id is 0")
			continue
		}

		mail := getMailById(pl.Cache.App.GetEnv().ID, id)
		if mail == nil {
			continue
		}

		if mail.OpenTime != 0 {
			continue
		}

		mail.OpenTime = utils.Now().Unix()
		if ok := updateDBMail(pl.Cache.App.GetEnv().ID, mail); ok {
			resp.Mail = append(resp.Mail, mail.ToProto())
		}
	}

	ctx.Send(resp)
}

// ReqDelMail 删除邮件
func ReqDelMail(ctx global.IPlayer, pl *model.Player, req *proto_mail.C2SDelMail) {
	if req.Id == 0 {
		log.Error("reqDelMail id error:%v", req.Id)
		return
	}

	mail := getMailById(pl.Cache.App.GetEnv().ID, req.Id)
	if mail == nil {
		log.Error("reqDelMail get mail error:%v", req.Id)
		return
	}

	if mail.OpenTime == 0 {
		return
	}

	//删除
	if req.Action == define.MailAction_Delete {
		//邮件有奖励不能删除
		if !mail.GotItem {
			log.Error("reqDelMail id error:%v", req.Id)
			return
		}
	} else if req.Action == define.MailAction_ReturnAttachment {
		if mail.GotItem {
			log.Error("MailAction_ReturnAttachment is GotItem")
			return
		}
		// 检查是否是私人交易邮件，如果是则需要退还附件
		handlePrivateTransactionReturn(ctx, pl, mail)
	}

	if !deleteDBMail(pl.Cache.App.GetEnv().ID, req.Id) {
		log.Error("reqDelMail db error:%v", req.Id)
		return
	}

	ctx.Send(&proto_mail.S2CDelMail{
		Success: true,
		Id:      req.Id,
	})
}

// ReqDelAllMails TODO:删除所有已读邮件
func ReqDelAllMails(ctx global.IPlayer, pl *model.Player, req *proto_mail.C2SDelAllMail) {
	rdb, err := db.GetEngine(pl.Cache.App.GetEnv().ID)
	if err != nil {
		log.Error("ReqDelAllMails error, no this server:%v", err)
		return
	}

	conn := rdb.Mysql

	mails := make([]*model.PlayerMailInfo, 0)
	sql := fmt.Sprintf("select * FROM %s where account_id = %s ORDER BY id limit %d",
		define.PlayerMailInfoTable, pl.Uid, define.MailStorageLimit)
	err = conn.SQL(sql).Find(&mails)
	if err != nil {
		log.Error("ReqDelAllMails get ids error:%v", err)
		return
	}

	ids := make([]int64, 0)
	for _, mail := range mails {
		if mail.OpenTime == 0 || !mail.GotItem {
			continue
		}

		if !deleteDBMail(pl.Cache.App.GetEnv().ID, mail.Id) {
			log.Error("reqDelMail db error:%v", mail.Id)
			continue
		}

		ids = append(ids, mail.Id)
	}

	ctx.Send(&proto_mail.S2CDelAllMail{
		Ids: ids,
	})
}

// ReqCollectMailItem 请求收取邮件物品
func ReqCollectMailItem(ctx global.IPlayer, pl *model.Player, req *proto_mail.C2SCollectMailItem) {
	if req.Id == 0 {
		return
	}

	mail := getMailById(pl.Cache.App.GetEnv().ID, req.Id)
	if mail == nil {
		log.Error("ReqCollectMailItem id error:%v", req.Id)
		return
	}

	if mail.OpenTime == 0 {
		log.Error("ReqCollectMailItem has no item:%v", req.Id)
		return
	}

	if req.Action == define.MailAction_GetAward {
		if mail.GotItem {
			log.Error("ReqCollectMailItem has no item:%v", req.Id)
			return
		}

		if utils.Now().Unix() >= mail.ExpireTime {
			return
		}

		//邮件没有奖励
		if mail.Items != nil && len(mail.Items) > 0 {
			if len(mail.Items) > 0 {
				bag.AddAward(ctx, pl, mail.Items, true)
			}
		}
	} else if req.Action == define.MailAction_BuyAttachment {
		if mail.IsHasAttachment && len(mail.Params) >= 2 {
			//解析附件
			// 尝试解析订单ID
			orderIdStr := mail.Params[0]
			orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
			if err != nil || orderId == 0 {
				return
			}

			// 查询订单
			order := invoke.TransactionClient(ctx).GetOrder(orderId)
			if order == nil {
				return
			}

			// 检查是否是私人交易订单且当前玩家是接收者
			if order.Type != 1 || order.ReceiverId != pl.Id || (order.Status != 0 && order.Status != 2) {
				return
			}

			err, _ = internal.ProcessBuyOrder(ctx, pl, order)
			if err != nil {
				return
			}
		}
	}

	mail.GotItem = true
	if ok := updateDBMail(pl.Cache.App.GetEnv().ID, mail); !ok {
		log.Error("ReqCollectMailItem db error:%v", req.Id)
		return
	}

	ctx.Send(&proto_mail.S2CCollectMailItem{Mail: mail.ToProto()})
}

// ReqCollectAllMailItems 请求收取所有邮件物品
func ReqCollectAllMailItems(ctx global.IPlayer, pl *model.Player, req *proto_mail.C2SCollectAllMailItems) {
	mails := getMailsFromDB(pl.Cache.App.GetEnv().ID, pl.Id)

	items := make([]conf.ItemE, 0)
	ret := make([]*proto_mail.Mail, 0)

	now := utils.Now().Unix()
	for _, mail := range mails {
		if mail.OpenTime != 0 && mail.GotItem {
			continue
		}

		if now >= mail.ExpireTime {
			continue
		}

		if mail.IsHasAttachment && len(mail.Params) >= 2 {
			continue
		}

		if mail.OpenTime == 0 {
			mail.OpenTime = utils.Now().Unix()
		}

		mail.GotItem = true

		if ok := updateDBMail(pl.Cache.App.GetEnv().ID, mail); !ok {
			continue
		}

		items = append(items, mail.Items...)

		ret = append(ret, mail.ToProto())
	}

	if len(items) > 0 {
		bag.AddAward(ctx, pl, items, true)
	}

	ctx.Send(&proto_mail.S2CCollectAllMailItems{Mails: ret})
}

// 检查新系统邮件 返回拉取到的新系统邮件
func checkNewSysMail(ctx global.IPlayer, pl *model.Player) {
	account := new(model.Account)
	_, err := db.CommonEngine.Mysql.Table("account").Where("uid = ?", pl.Uid).ForUpdate().Get(account)
	if err != nil {
		log.Error("check new mail error:%v", err)
		return
	}

	id := invoke.MailClient(ctx).GetMaxSystemMailId()

	maxSystemMailId, _ := utils.Int64(id)
	curSysId := account.SystemMailId

	if curSysId >= maxSystemMailId {
		return
	}

	account.SystemMailId = maxSystemMailId
	_, err = db.CommonEngine.Mysql.Table("account").Where("id = ?", account.Id).MustCols("sys_mail_id").Update(account)
	if err != nil {
		log.Error("check new system mail error:%v", err)
		return
	}

	now := utils.Now().Unix()
	for i := curSysId + 1; i <= maxSystemMailId; i++ {
		sysMail := invoke.MailClient(ctx).GetSystemMailById(i)
		if sysMail == nil {
			continue
		}
		//比账号创建时间还早 就不发了
		if sysMail.CreateTime < pl.Base.CreateTime {
			continue
		}

		gotItem := true
		if len(sysMail.Items) > 0 {
			gotItem = false
		}

		mailType := define.MailTypeNormal

		// 生成新的个人邮件
		newMail := &model.PlayerMailInfo{
			SysId:      sysMail.Id,
			MailInfos:  sysMail.MailInfos,
			CreateTime: now,
			Items:      sysMail.Items,
			GotItem:    gotItem,
			CfgId:      sysMail.CfgId,                       //系统邮件默认0
			Params:     sysMail.Params,                      //默认无参数
			ExpireTime: define.MailStorageLimit*86400 + now, // 根据配置设置过期时间
			AccountId:  pl.Uid,
			Type:       mailType,
		}

		if ok := insertDBMail(pl.Cache.App.GetEnv().ID, newMail); !ok {
			log.Error("System mail send to player failed, player id %v, system mail id %v", pl.Id, sysMail.Id)
			continue
		}
	}
}

// =====================DB========================
// 更新邮件信息
func updateDBMail(serverId int, mail *model.PlayerMailInfo) (ok bool) {
	rdb, err := db.GetEngine(serverId)
	if err != nil {
		log.Error("checkNewSysMail error, no this server:%v", err)
		return false
	}
	conn := rdb.Mysql

	if mail == nil {
		log.Error("update DB Mail is nil")
		return false
	}

	num, err := conn.Where("id = ?", mail.Id).MustCols("got_item").Update(mail)
	if err != nil {
		log.Error("update DB mail error:%v", err)
		return false
	}

	if num == 0 {
		log.Error("update DB mail num is 0")
		return false
	}

	return true
}

// 从db中获取邮件
func getMailsFromDB(serverId int, dbId int64) []*model.PlayerMailInfo {
	rdb, err := db.GetEngine(serverId)
	if err != nil {
		log.Error("checkNewSysMail error, no this server:%v", err)
		return nil
	}
	conn := rdb.Mysql

	now := utils.Now().Unix()

	mails := make([]*model.PlayerMailInfo, 0)
	sql := fmt.Sprintf("select * from %s where (db_id = %d AND sys_id = %d) OR sys_id = %d ORDER BY id DESC limit %d",
		define.PlayerMailInfoTable, dbId, 0, serverId, define.MailStorageLimit)

	err = conn.SQL(sql).Find(&mails)
	if err != nil {
		log.Error("get DB mails error:%v", err)
		return nil
	}

	mailList := make([]*model.PlayerMailInfo, 0)
	for _, mail := range mails {
		if now >= mail.ExpireTime {
			continue
		}
		mailList = append(mailList, mail)
	}

	return mailList
}

// 根据邮件id获取邮件
func getMailById(serverId int, id int64) *model.PlayerMailInfo {
	rdb, err := db.GetEngine(serverId)
	if err != nil {
		log.Error("getMailById error, no this server:%v", err)
		return nil
	}

	conn := rdb.Mysql

	mail := new(model.PlayerMailInfo)
	_, err = conn.Where("id = ?", id).Get(mail)
	if err != nil {
		log.Error("get mail from db by id error %v", err)
		return nil
	}

	if mail.Id == 0 {
		log.Error("get mail from db id is 0")
		return nil
	}

	return mail
}

// 插入新邮件
func insertDBMail(serverId int, mail *model.PlayerMailInfo) bool {
	rdb, err := db.GetEngine(serverId)
	if err != nil {
		log.Error("insertDBMail error, no this server:%v", err)
		return false
	}
	conn := rdb.Mysql

	num, err := conn.Insert(mail)
	if err != nil {
		log.Error("insert DB mail error:%v", err)
		return false
	}
	if num == 0 {
		log.Error("insert DB mail num is 0")
		return false
	}

	log.Debug("写入邮件进db, 玩家account_id:%v, 邮件id:%v, 系统id:%v", mail.AccountId, mail.Id, mail.SysId)

	return true
}

// 插入新邮件
func deleteDBMail(serverId int, id int64) bool {
	rdb, err := db.GetEngine(serverId)
	if err != nil {
		log.Error("deleteDBMail error, no this server:%v", err)
		return false
	}

	conn := rdb.Mysql

	num, err := conn.Table(define.PlayerMailInfoTable).Where("id = ?", id).Delete()
	if err != nil {
		log.Error("delete DB mail error:%v", err)
		return false
	}
	if num == 0 {
		log.Error("delete DB mail num is 0")
		return false
	}

	return true
}

// handlePrivateTransactionReturn 处理私人交易附件退还
func handlePrivateTransactionReturn(ctx global.IPlayer, pl *model.Player, mail *model.PlayerMailInfo) {
	if !mail.IsHasAttachment {
		return
	}

	// 检查是否有Params，且第一个参数是订单ID
	if len(mail.Params) == 0 {
		return
	}

	// 尝试解析订单ID
	orderIdStr := mail.Params[0]
	orderId, err := strconv.ParseInt(orderIdStr, 10, 64)
	if err != nil || orderId == 0 {
		return
	}

	// 查询订单
	order := invoke.TransactionClient(ctx).GetOrder(orderId)
	if order == nil {
		return
	}

	// 检查是否是私人交易订单且当前玩家是接收者
	if order.Type != 1 || order.ReceiverId != pl.Id || order.Status != 0 {
		return
	}

	order.Status = 2
	order.ReceiverId = order.SellerId
	order.ReceiverName = order.SellerName
	order.SellerName = pl.Base.Name
	order.SellerId = pl.Id
	order.Price = 0

	// 调用transaction退还订单
	returnedOrder, err := invoke.TransactionClient(ctx).UpdateOrder(orderId, order)
	if err != nil {
		log.Error("handlePrivateTransactionReturn error: %v", err)
		return
	}

	//组装下附件信息
	attachmentObj := &model.AttachmentData{
		Id:     order.Id,
		Type:   returnedOrder.AttachmentData.Type,
		ItemId: returnedOrder.AttachmentData.ItemId,
		Level:  returnedOrder.AttachmentData.Level,
		Star:   returnedOrder.AttachmentData.Star,
		Stage:  returnedOrder.AttachmentData.Stage,
	}
	attachmentStr, _ := json.Marshal(attachmentObj)

	// 发送纯文本通知邮件给原发送者（无附件）
	invoke.MailClient(ctx).SendMail(
		define.PlayerMail,
		"私人交易退还",
		fmt.Sprintf("您发送给 %s 的私人交易附件已被退还，物品已自动返回", order.SellerName),
		"Private Transaction Returned",
		fmt.Sprintf("Your private transaction attachment to %s has been returned, items have been automatically returned", order.SellerName),
		"系统",
		[]conf.ItemE{}, // 无附件，纯通知
		[]int64{returnedOrder.ReceiverId},
		int64(0),
		int32(0),
		true,
		[]string{fmt.Sprintf("%d", order.Id), string(attachmentStr)},
	)

	log.Debug("私人交易附件退还成功，订单ID:%d, 发送者:%d, 接收者:%d, 物品已直接发放", orderId, pl.Id, returnedOrder.ReceiverId)
}
