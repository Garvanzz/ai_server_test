package logic

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

// 获取充值列表
func GmGetOrderList(c *gin.Context) {
	log.Debug("请求游戏订单列表")

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	OrderId, ok := result["OrderId"].(string)
	if !ok {
		OrderId = ""
	}

	Uid, ok := result["Uid"].(string)
	if !ok {
		Uid = ""
	}

	var orderItem []model.RechargeOrder
	if len(OrderId) > 0 && len(Uid) > 0 {
		err := db.AccountDb.Table(define.PayOrderTable).Where("order_id = ? AND game_user_id =?", OrderId, Uid).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	} else if len(OrderId) > 0 {
		err := db.AccountDb.Table(define.PayOrderTable).Where("order_id = ? ", OrderId).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	} else if len(Uid) > 0 {
		err := db.AccountDb.Table(define.PayOrderTable).Where("game_user_id = ? ", Uid).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		err := db.AccountDb.Table(define.PayOrderTable).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	items := make([]*dto.GMRespRechargeOrder, 0)
	for i := 0; i < len(orderItem); i++ {
		amount := fmt.Sprintf("%.0f", orderItem[i].Amount)
		items = append(items, &dto.GMRespRechargeOrder{
			Amount:        amount,
			ProductId:     orderItem[i].ProductId,
			ProductName:   orderItem[i].ProductName,
			UserId:        orderItem[i].UserId,
			GameUserId:    orderItem[i].GameUserId,
			OrderId:       orderItem[i].OrderId,
			ServerId:      orderItem[i].ServerId,
			PaymentTime:   orderItem[i].PaymentTime,
			ChannelNumber: orderItem[i].ChannelNumber,
		})
	}

	js, _ := json.Marshal(items)

	HTTPRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
}

// 获取缓存充值列表
func GmGetCacheOrderList(c *gin.Context) {
	log.Debug("请求游戏订单列表")

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	OrderId, ok := result["OrderId"].(string)
	if !ok {
		OrderId = ""
	}

	Uid, ok := result["Uid"].(string)
	if !ok {
		Uid = ""
	}

	var orderItem []model.RechargeOrder
	if len(OrderId) > 0 && len(Uid) > 0 {
		err := db.AccountDb.Table(define.PayCacheOrderTable).Where("order_id = ? AND game_user_id =?", OrderId, Uid).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	} else if len(OrderId) > 0 {
		err := db.AccountDb.Table(define.PayCacheOrderTable).Where("order_id = ? ", OrderId).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	} else if len(Uid) > 0 {
		err := db.AccountDb.Table(define.PayCacheOrderTable).Where("game_user_id = ? ", Uid).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		err := db.AccountDb.Table(define.PayCacheOrderTable).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	items := make([]*dto.GMRespRechargeOrder, 0)
	for i := 0; i < len(orderItem); i++ {
		amount := fmt.Sprintf("%.0f", orderItem[i].Amount)
		items = append(items, &dto.GMRespRechargeOrder{
			Amount:        amount,
			ProductId:     orderItem[i].ProductId,
			ProductName:   orderItem[i].ProductName,
			UserId:        orderItem[i].UserId,
			GameUserId:    orderItem[i].GameUserId,
			OrderId:       orderItem[i].OrderId,
			ServerId:      orderItem[i].ServerId,
			PaymentTime:   orderItem[i].PaymentTime,
			ChannelNumber: orderItem[i].ChannelNumber,
		})
	}

	js, _ := json.Marshal(items)

	HTTPRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
}
