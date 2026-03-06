package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/model"
	conf2 "xfx/gm_server/conf"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
)

// 获取道具
func GmItem(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家游戏道具数据 : %d, %s", req.ServerId, req.Uid)

	var serverItem model.ServerItem
	has, err := db.AccountDb.Table(define.ServerGroup).Where("id = ?", req.ServerId).Get(&serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist1 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	pl := new(model.Account)
	has, err = db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", req.ServerId, req.Uid).Get(pl)
	if err != nil {
		log.Error("getserverlist123 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf2.Server.RedisAddr, serverItem.RedisPort), conf2.Server.RedisPassword)

	dsts := make([]*gm_model.GmRespPlayerBag, 0)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerBag, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Bag)
	err = json.Unmarshal(values.([]byte), &dst)

	for k, v := range dst.Items {
		dsts = append(dsts, &gm_model.GmRespPlayerBag{
			ItemId:  k,
			ItemNum: v,
		})
	}

	js, _ := json.Marshal(dsts)
	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(dsts),
	})
}

// 添加道具
func GmAddItem(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)

	var serverItem model.ServerItem
	has, err := db.AccountDb.Table(define.ServerGroup).Where("id = ?", req.ServerId).Get(&serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist1 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	pl := new(model.Account)
	has, err = db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", req.ServerId, req.Uid).Get(pl)
	if err != nil {
		log.Error("getserverlist123 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf2.Server.RedisAddr, serverItem.RedisPort), conf2.Server.RedisPassword)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerBag, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Bag)
	err = json.Unmarshal(values.([]byte), &dst)

	has = false
	for k, _ := range dst.Items {
		if k == int32(req.ItemId) {
			dst.Items[k] = int32(req.ItemNum)
			has = true
			break
		}
	}

	if !has {
		dst.Items[int32(req.ItemId)] = int32(req.ItemNum)
	}

	js, _ := json.Marshal(dst)
	_redis.RedisExec("set", fmt.Sprintf("%s:%d", define.PlayerBag, pl.RedisId), js)

	httpRetGame(c, SUCCESS, "success")
}

// 一键添加道具
func GmOneKeyAddItem(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)

	var serverItem model.ServerItem
	has, err := db.AccountDb.Table(define.ServerGroup).Where("id = ?", req.ServerId).Get(&serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist1 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	pl := new(model.Account)
	has, err = db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", req.ServerId, req.Uid).Get(pl)
	if err != nil {
		log.Error("getserverlist123 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf2.Server.RedisAddr, serverItem.RedisPort), conf2.Server.RedisPassword)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerBag, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Bag)
	err = json.Unmarshal(values.([]byte), &dst)

	//获取道具配置表
	confs := config.Item.All()
	for _, v := range confs {
		if _, ok := dst.Items[v.Id]; ok {
			dst.Items[v.Id] += 50000
		} else {
			dst.Items[v.Id] = 50000
		}
	}

	js, _ := json.Marshal(dst)
	_redis.RedisExec("set", fmt.Sprintf("%s:%d", define.PlayerBag, pl.RedisId), js)

	httpRetGame(c, SUCCESS, "success")
}

// 删除道具
func GmDeleteItem(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)

	var serverItem model.ServerItem
	has, err := db.AccountDb.Table(define.ServerGroup).Where("id = ?", req.ServerId).Get(&serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist1 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	pl := new(model.Account)
	has, err = db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", req.ServerId, req.Uid).Get(pl)
	if err != nil {
		log.Error("getserverlist123 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf2.Server.RedisAddr, serverItem.RedisPort), conf2.Server.RedisPassword)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerBag, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Bag)
	err = json.Unmarshal(values.([]byte), &dst)

	for _, v := range req.Ids {
		if _, ok := dst.Items[int32(v)]; ok {
			delete(dst.Items, int32(v))
		}
	}

	js, _ := json.Marshal(dst)
	_redis.RedisExec("set", fmt.Sprintf("%s:%d", define.PlayerBag, pl.RedisId), js)

	httpRetGame(c, SUCCESS, "success")
}
