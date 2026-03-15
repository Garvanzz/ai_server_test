package shop

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_shop"
)

func Init(pl *model.Player) {
	pl.Shop = new(model.PlayerShop)
	pl.Shop.Shops = make(map[int32]*model.ShopType)
}

func Save(pl *model.Player, isSync bool) {
	//商城
	j, err := json.Marshal(pl.Shop)
	if err != nil {
		log.Error("player[%v],save shop marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerShop, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerShop, pl.Id))
	if err != nil {
		log.Error("player[%v],load shop error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	shop := new(model.PlayerShop)
	err = json.Unmarshal(reply.([]byte), &shop)
	if err != nil {
		log.Error("player[%v],load shop unmarshal error:%v", pl.Id, err)
	}
	pl.Shop = shop
}

// ReqShopData 请求商店数据
func ReqShopData(ctx global.IPlayer, pl *model.Player, req *proto_shop.C2SShopData) {
	shopData := new(proto_shop.S2CShopData)
	//刷新
	refreshShopData(ctx, pl)
	shopData.TypeShopItem = model.ToShopTypeProto(pl.Shop.Shops)
	ctx.Send(shopData)
}

// 刷新商城数据
func refreshShopData(ctx global.IPlayer, pl *model.Player) {
	for shopType, v := range pl.Shop.Shops {
		rt, ok := define.ShopRefreshType[shopType]

		if !ok {
			continue
		}

		switch rt {
		case define.ShopLimitTypeNull:
		case define.ShopLimitTypeYongJiu:
		case define.ShopLimitTypeDay:
			if !utils.CheckIsSameDayBySec(utils.Now().Unix(), v.LastTime, 0) {
				pl.Shop.Shops[shopType].ShopItems = make(map[int]*model.ShopItem)
				pl.Shop.Shops[shopType].LastTime = utils.TimestampToday()
			}
		case define.ShopLimitTypeWeek:
			if !utils.IsSameWeekBySec(utils.Now().Unix(), v.LastTime) {
				pl.Shop.Shops[shopType].ShopItems = make(map[int]*model.ShopItem)
				pl.Shop.Shops[shopType].LastTime = utils.GetTargetDayStartUnix(utils.GetWeekday(time.Monday)) //周一的
			}
		case define.ShopLimitTypeMonth:
			if utils.DaysBetweenTwoTimeUnix(utils.Now().Unix(), v.LastTime) > 30 {
				//直接删除
				delete(pl.Shop.Shops, shopType)
			}
		case define.ShopLimitTypeSeason:
			//获取赛季数据
			status, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeSeason)
			if err != nil {
				log.Debug("load shop season: %v", err)
				continue
			}
			if status == nil {
				continue
			}

			if v.LastTime < status.StartTime || v.LastTime > status.EndTime {
				pl.Shop.Shops[shopType].ShopItems = make(map[int]*model.ShopItem)
				pl.Shop.Shops[shopType].LastTime = utils.Now().Unix()
			}
		}
	}
}

// ReqShopBuyData 请求商店购买
func ReqShopBuyData(ctx global.IPlayer, pl *model.Player, req *proto_shop.C2SBuyShop) {
	res := new(proto_shop.S2CBuyShop)
	//刷新
	refreshShopData(ctx, pl)

	//判断配置
	conf := config.Shop.All()[int64(req.Id)]
	if conf.Id <= 0 {
		res.Code = proto_shop.ERRORCODE_ERR_ConfigErr
		ctx.Send(res)
		return
	}

	if conf.IsBuy {
		res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
		ctx.Send(res)
		return
	}

	//次数加成
	addNum := 0
	//快捷兑换要考虑加成
	if req.Type == define.SHOPTYPE_QUICKBUY {
		addNum = getExchangeAddNum(pl, conf)
	}

	//判断个数
	if data, ok := pl.Shop.Shops[req.Type]; ok {
		if _, ok = data.ShopItems[int(req.Id)]; ok {
			if conf.LimitType != define.ShopLimitTypeNull && data.ShopItems[int(req.Id)].Num >= conf.LimitNum+int32(addNum) {
				res.Code = proto_shop.ERRORCODE_ERR_OutLimit
				ctx.Send(res)
				return
			}
		} else {
			pl.Shop.Shops[req.Type].ShopItems[int(req.Id)] = new(model.ShopItem)
			pl.Shop.Shops[req.Type].ShopItems[int(req.Id)].Id = req.Id
		}
	} else {
		pl.Shop.Shops[req.Type] = new(model.ShopType)
		pl.Shop.Shops[req.Type].ShopItems = make(map[int]*model.ShopItem)
		if conf.LimitType == define.ShopLimitTypeDay {
			pl.Shop.Shops[req.Type].LastTime = utils.TimestampToday()
		} else if conf.LimitType == define.ShopLimitTypeWeek { //每周限购
			pl.Shop.Shops[req.Type].LastTime = utils.GetTargetDayStartUnix(utils.GetWeekday(time.Monday)) //周一的
		} else if conf.LimitType == define.ShopLimitTypeMonth {
			pl.Shop.Shops[req.Type].LastTime = utils.Now().Unix()
		} else if conf.LimitType == define.ShopLimitTypeSeason {
			pl.Shop.Shops[req.Type].LastTime = utils.Now().Unix()
		}

		pl.Shop.Shops[req.Type].ShopItems[int(req.Id)] = new(model.ShopItem)
		pl.Shop.Shops[req.Type].ShopItems[int(req.Id)].Id = req.Id
	}

	//判断道具够不够
	cost := make(map[int32]int32)
	for _, v := range conf.CostItem {
		cost[v.ItemId] += cost[v.ItemId] + v.ItemNum*req.Num
	}

	if !internal.CheckItemsEnough(pl, cost) {
		res.Code = proto_shop.ERRORCODE_ERR_NoEnough
		ctx.Send(res)
		return
	}

	internal.SubItems(ctx, pl, cost)

	awards := conf.GetItem

	//判断首充
	if pl.Shop.Shops[req.Type].ShopItems[int(req.Id)].Num <= 0 && len(conf.FirstAward) > 0 {
		awards = append(awards, conf.FirstAward...)
	}

	bag.AddAward(ctx, pl, awards, true)

	pl.Shop.Shops[req.Type].ShopItems[int(req.Id)].Num += req.Num
	pushChangeShopData(ctx, pl, req.Type, req.Id)
	res.Code = proto_shop.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// ReqGMShopBuyData 请求商店GM购买
func ReqGMShopBuyData(ctx global.IPlayer, pl *model.Player, req *proto_shop.C2SGMBuyShop) {
	res := new(proto_shop.S2CGMBuyShop)
	//刷新
	refreshShopData(ctx, pl)

	//判断配置
	shopConf := config.Shop.All()[int64(req.Id)]
	if shopConf.Id <= 0 {
		res.Code = proto_shop.ERRORCODE_ERR_ConfigErr
		ctx.Send(res)
		return
	}

	if !shopConf.IsBuy {
		res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
		ctx.Send(res)
		return
	}

	rechargeConf := config.Recharge.All()[int64(shopConf.RechargeId)]
	if rechargeConf.Id <= 0 {
		res.Code = proto_shop.ERRORCODE_ERR_ConfigErr
		ctx.Send(res)
		return
	}

	//判断个数
	if data, ok := pl.Shop.Shops[shopConf.Type]; ok {
		if _, ok = data.ShopItems[int(req.Id)]; ok {
			if shopConf.LimitType != define.ShopLimitTypeNull && data.ShopItems[int(req.Id)].Num >= shopConf.LimitNum {
				res.Code = proto_shop.ERRORCODE_ERR_OutLimit
				ctx.Send(res)
				return
			}
		} else {
			pl.Shop.Shops[shopConf.Type].ShopItems[int(req.Id)] = new(model.ShopItem)
			pl.Shop.Shops[shopConf.Type].ShopItems[int(req.Id)].Id = req.Id
		}
	} else {
		pl.Shop.Shops[shopConf.Type] = new(model.ShopType)
		pl.Shop.Shops[shopConf.Type].ShopItems = make(map[int]*model.ShopItem)
		if shopConf.LimitType == define.ShopLimitTypeDay {
			pl.Shop.Shops[shopConf.Type].LastTime = utils.TimestampToday()
		} else if shopConf.LimitType == define.ShopLimitTypeWeek { //每周限购
			pl.Shop.Shops[shopConf.Type].LastTime = utils.GetTargetDayStartUnix(utils.GetWeekday(time.Monday)) //周一的
		} else if shopConf.LimitType == define.ShopLimitTypeMonth {
			pl.Shop.Shops[shopConf.Type].LastTime = utils.Now().Unix()
		} else if shopConf.LimitType == define.ShopLimitTypeSeason {
			pl.Shop.Shops[shopConf.Type].LastTime = utils.Now().Unix()
		}

		pl.Shop.Shops[shopConf.Type].ShopItems[int(req.Id)] = new(model.ShopItem)
		pl.Shop.Shops[shopConf.Type].ShopItems[int(req.Id)].Id = req.Id
	}

	//生成订单信息
	oid := randOrderId()
	ProductName := fmt.Sprintf("%d", rechargeConf.ItemId)
	ProductId := fmt.Sprintf("%d", rechargeConf.Id)
	order := &model.RechargeOrder{
		Amount:        float32(rechargeConf.Price),
		ProductId:     ProductId,
		ProductName:   ProductName,
		UserId:        strconv.FormatInt(pl.Id, 10),
		OrderId:       oid,
		GameUserId:    pl.Uid,
		ServerId:      int(pl.GetProp(define.PlayerPropServerId)),
		PaymentTime:   utils.GetTimeNowFormat(),
		ChannelNumber: "hgm",
	}
	log.Debug("订单信息: %v", order)
	awards := shopConf.GetItem

	//判断首充
	if pl.Shop.Shops[shopConf.Type].ShopItems[int(req.Id)].Num <= 0 && len(shopConf.FirstAward) > 0 {
		awards = append(awards, shopConf.FirstAward...)
	}

	//根据类型链路到各个模块
	shopBuyToMode(ctx, pl, shopConf)

	//订单信息进入订单缓存数据库
	_, err := db.Engine.Mysql.Table(define.PayCacheOrderTable).Insert(order)
	if err != nil {
		log.Error("插入失败: %v", err)
		res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
		ctx.Send(res)
		return
	}

	//奖励进入redis
	js, _ := json.Marshal(awards)
	db.RedisExec("hset", fmt.Sprintf("%s:%d", define.PayChache, pl.Id), oid, js)
	pl.Shop.Shops[shopConf.Type].ShopItems[int(req.Id)].Num += 1
	pushChangeShopData(ctx, pl, shopConf.Type, req.Id)
	res.Code = proto_shop.ERRORCODE_ERR_Ok
	//订单ID
	res.OrderId = oid
	ctx.Send(res)
}

// 获取兑换加成次数
func getExchangeAddNum(pl *model.Player, shopConf conf.Shop) int {
	//这里要区别下鉴宝的快捷兑换，ID先写死
	if shopConf.Type == define.SHOPTYPE_QUICKBUY && shopConf.Id == define.ShopIdGemAppraisal {
		//判断有没有购买鉴宝月卡
		reply, err := db.RedisExec("HGET", define.GemAppraisal_MonthCard, pl.Id)
		if err != nil {
			log.Error("[%v],load getExchangeAddNum error:%v", pl.Id, err)
			return 0
		}

		if reply != nil {
			//获取配置
			confs := config.MonthCard.All()
			conf := conf.MonthCard{}
			for _, v := range confs {
				if v.Type == define.MonthCard_GemAppraisal {
					conf = v
					break
				}
			}

			if conf.Id > 0 {
				return int(conf.ExchangeCount)
			}
		}
	}

	return 0
}

// 根据类型链路到各个模块,需要处理对应逻辑
func shopBuyToMode(ctx global.IPlayer, pl *model.Player, shopConf conf.Shop) {
	switch shopConf.Type {
	case define.SHOPTYPE_GEMAPPRAISALLMONTHCATD:
		ShopToGemAppraisalMonthCard(ctx, pl)
		break
	}

	//获取充值金额
	rechargeConf := config.Recharge.All()[int64(shopConf.RechargeId)]
	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":          "recharge",
		"player":       pl.ToContext(),
		"playermodel":  pl,
		"shopconf":     shopConf,
		"rechargeconf": rechargeConf,
		"IPlayer":      ctx,
	})

	//通告相关
	internal.SyncNotice_ShopBuy(ctx, pl, shopConf.Id)
}

// ReqBackShopBuyAward 支付后 请求奖励
func ReqBackShopBuyAward(ctx global.IPlayer, pl *model.Player, req *proto_shop.C2SReqRechargeBackAward) {
	res := new(proto_shop.S2CReqRechargeBackAward)

	//判断订单存不存在
	order := model.RechargeOrder{}
	has, err := db.Engine.Mysql.Table(define.PayCacheOrderTable).Where("order_id = ?", req.OrderId).Get(&order)
	if err != nil {
		res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
		ctx.Send(res)
		return
	}

	if !has {
		res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
		ctx.Send(res)
		return
	}

	log.Debug("进入支付库:%v", order)
	//订单信息进入完成订单数据库
	_, err = db.Engine.Mysql.Table(define.PayOrderTable).Insert(order)
	if err != nil {
		log.Error("进入支付库错误:%v", err)
		res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
		ctx.Send(res)
		return
	}

	//删除缓存订单信息
	db.Engine.Mysql.Table(define.PayCacheOrderTable).Delete(order)

	reply, err := db.RedisExec("hget", fmt.Sprintf("%s:%d", define.PayChache, pl.Id), req.OrderId)
	if err != nil {
		res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
		ctx.Send(res)
		return
	}

	awards := []conf.ItemE{}
	if reply == nil {
		res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
		ctx.Send(res)
		return
	} else {
		err = json.Unmarshal(reply.([]byte), &awards)
		if err != nil {
			res.Code = proto_shop.ERRORCODE_ERR_RechargeErr
			ctx.Send(res)
			return
		}
	}

	bag.AddAward(ctx, pl, awards, true)

	//删除缓存
	db.RedisExec("hdel", fmt.Sprintf("%s:%d", define.PayChache, pl.Id), req.OrderId)

	res.Code = proto_shop.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// 生成订单号信息
func randOrderId() string {
	time := fmt.Sprintf("%d", utils.Now().Unix())
	rangNum := fmt.Sprintf("%d", utils.RandInt(0, 1000000))
	return "hgmgameGm-" + time + "-" + rangNum
}

// 变化
func pushChangeShopData(ctx global.IPlayer, pl *model.Player, typ int32, id int32) {
	types := make(map[int32]*proto_shop.TypeShopItem)
	types[typ] = new(proto_shop.TypeShopItem)
	types[typ].LastTime = pl.Shop.Shops[typ].LastTime
	items := make(map[int32]*proto_shop.ShopItem)
	items[id] = new(proto_shop.ShopItem)
	items[id].Id = id
	items[id].Num = pl.Shop.Shops[typ].ShopItems[int(id)].Num
	types[typ].ShopItem = items
	ctx.Send(&proto_shop.PushShopChange{
		TypeShopItem: types,
	})
}
