package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
)

// 获取充值列表
func GmGetOrderList(c *gin.Context) {
	p := new(model.RechargeOrder)
	if has, _ := db.AccountDb.IsTableExist(define.PayOrder); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	log.Debug("请求游戏订单列表")

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
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
		err := db.AccountDb.Table(define.PayOrder).Where("order_id = ? AND game_user_id =?", OrderId, Uid).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else if len(OrderId) > 0 {
		err := db.AccountDb.Table(define.PayOrder).Where("order_id = ? ", OrderId).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else if len(Uid) > 0 {
		err := db.AccountDb.Table(define.PayOrder).Where("game_user_id = ? ", Uid).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		err := db.AccountDb.Table(define.PayOrder).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	items := make([]*gm_model.GMRespRechargeOrder, 0)
	for i := 0; i < len(orderItem); i++ {
		amount := fmt.Sprintf("%.0f", orderItem[i].Amount)
		items = append(items, &gm_model.GMRespRechargeOrder{
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

	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
}

// 获取缓存充值列表
func GmGetCacheOrderList(c *gin.Context) {
	p := new(model.RechargeOrder)
	if has, _ := db.AccountDb.IsTableExist(define.PayCacheOrder); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	log.Debug("请求游戏订单列表")

	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
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
		err := db.AccountDb.Table(define.PayCacheOrder).Where("order_id = ? AND game_user_id =?", OrderId, Uid).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else if len(OrderId) > 0 {
		err := db.AccountDb.Table(define.PayCacheOrder).Where("order_id = ? ", OrderId).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else if len(Uid) > 0 {
		err := db.AccountDb.Table(define.PayCacheOrder).Where("game_user_id = ? ", Uid).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		err := db.AccountDb.Table(define.PayCacheOrder).Find(&orderItem)
		if err != nil {
			log.Error("getserverlist2 find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	err := db.AccountDb.Table(define.PayCacheOrder).Find(&orderItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	items := make([]*gm_model.GMRespRechargeOrder, 0)
	for i := 0; i < len(orderItem); i++ {
		amount := fmt.Sprintf("%.0f", orderItem[i].Amount)
		items = append(items, &gm_model.GMRespRechargeOrder{
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

	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
}
