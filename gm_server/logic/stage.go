package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/conf"
	"xfx/gm_server/db"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
)

//1 获取关卡列表信息。 字段： 周目     章节id    关卡id。 经验  通关状态 （通关。未通关，正在通关） 挑战boss是否成功
//
//2 设置关卡信息，传入 周目 章节和关卡id。可设置字段  经验  通关状态   挑战boss是否成功。
//
//3 添加关卡信息， 传入 周目 章节 关卡id 通关状态 挑战boss是否成功。 如果传入关卡 之前的关卡信息没有也需要生成 例如我传入的10008 服务器只有10002。那中间的也要生成
//
//4 删除关卡信息，传入周目 章节 关卡id

// GmGetStageInfo 获取关卡列表信息
func GmGetStageInfo(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqGetStageInfo
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家装备数据 : %d, %s", req.ServerId, req.Uid)

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

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf.Server.RedisAddr, serverItem.RedisPort), conf.Server.RedisPassword)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerStage, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Stage)
	err = json.Unmarshal(values.([]byte), &dst)

	dsts := make([]*gm_model.GmRespGetStageInfo, 0)
	//for _, v := range dst.Stage {
	//indexName := ""
	//if v.Index == 0 {
	//	indexName = "无"
	//} else if v.Index == 1 {
	//	indexName = "主武器"
	//} else if v.Index == 2 {
	//	indexName = "头盔"
	//} else if v.Index == 3 {
	//	indexName = "项链"
	//} else if v.Index == 4 {
	//	indexName = "外衣"
	//} else if v.Index == 5 {
	//	indexName = "腰带"
	//} else if v.Index == 5 {
	//	indexName = "鞋子"
	//}
	//dsts = append(dsts, &gm_model.GmRespGetStageInfo{
	//	EquipId:    v.Id,
	//	EquipCId:   v.CId,
	//	EquipNum:   v.Num,
	//	EquipLevel: v.Level,
	//	EquipIndex: indexName,
	//	EquipIsUse: v.IsUse,
	//})
	//}

	////排序
	//sort.Slice(dsts, func(i, j int) bool {
	//	if dsts[i].EquipIsUse == true {
	//		return true
	//	}
	//
	//	if dsts[j].EquipIsUse == true {
	//		return true
	//	}
	//
	//	return false
	//})

	js, _ := json.Marshal(dsts)
	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(dsts),
	})
}

//2 设置关卡信息，传入 周目 章节和关卡id。可设置字段  经验  通关状态   挑战boss是否成功。

// GmSetStageInfo 设置关卡信息
func GmSetStageInfo(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqSetStageInfo
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家关卡数据 : %d, %s", req.ServerId, req.Uid)

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

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf.Server.RedisAddr, serverItem.RedisPort), conf.Server.RedisPassword)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerStage, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Stage)
	err = json.Unmarshal(values.([]byte), &dst)

	// TODO:

	js, _ := json.Marshal(dst)
	_redis.RedisExec("set", fmt.Sprintf("%s:%d", define.PlayerStage, pl.RedisId), js)

	httpRetGame(c, SUCCESS, "success")
}

//3 添加关卡信息， 传入 周目 章节 关卡id 通关状态 挑战boss是否成功。 如果传入关卡 之前的关卡信息没有也需要生成 例如我传入的10008 服务器只有10002。那中间的也要生成

// GmAddStageInfo 添加关卡信息
func GmAddStageInfo(c *gin.Context) {
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

	//log.Debug("请求玩家数据 : %d, %s", req.ServerId, req.Uid)

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

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf.Server.RedisAddr, serverItem.RedisPort), conf.Server.RedisPassword)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerStage, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Stage)
	err = json.Unmarshal(values.([]byte), &dst)

	js, _ := json.Marshal(dst)
	_redis.RedisExec("set", fmt.Sprintf("%s:%d", define.PlayerStage, pl.RedisId), js)

	httpRetGame(c, SUCCESS, "success")
}

//4 删除关卡信息，传入周目 章节 关卡id

// GmDeleteStageInfo 删除关卡信息
func GmDeleteStageInfo(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerEquip
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("删除关卡 : %d, %s", req.ServerId, req.Uid)

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

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf.Server.RedisAddr, serverItem.RedisPort), conf.Server.RedisPassword)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerStage, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Stage)
	err = json.Unmarshal(values.([]byte), &dst)

	js, _ := json.Marshal(dst)
	_redis.RedisExec("set", fmt.Sprintf("%s:%d", define.PlayerStage, pl.RedisId), js)

	httpRetGame(c, SUCCESS, "success")
}
