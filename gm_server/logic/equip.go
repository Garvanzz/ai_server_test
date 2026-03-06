package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"sort"
	"xfx/core/model"
	"xfx/gm_server/conf"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
)

// 装备
func GmEquip(c *gin.Context) {
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
	dsts := make([]*gm_model.GmRespPlayerEquip, 0)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerEquip, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Equip)
	err = json.Unmarshal(values.([]byte), &dst)

	for _, v := range dst.Equips {
		indexName := ""
		if v.Index == 0 {
			indexName = "无"
		} else if v.Index == 1 {
			indexName = "主武器"
		} else if v.Index == 2 {
			indexName = "头盔"
		} else if v.Index == 3 {
			indexName = "项链"
		} else if v.Index == 4 {
			indexName = "外衣"
		} else if v.Index == 5 {
			indexName = "腰带"
		} else if v.Index == 5 {
			indexName = "鞋子"
		}
		dsts = append(dsts, &gm_model.GmRespPlayerEquip{
			EquipId:    v.Id,
			EquipCId:   v.CId,
			EquipNum:   v.Num,
			EquipLevel: v.Level,
			EquipIndex: indexName,
			EquipIsUse: v.IsUse,
		})
	}

	//排序
	sort.Slice(dsts, func(i, j int) bool {
		if dsts[i].EquipIsUse == true {
			return true
		}

		if dsts[j].EquipIsUse == true {
			return true
		}

		return false
	})

	js, _ := json.Marshal(dsts)
	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(dsts),
	})
}

// 删除装备
func GmDeleteEquip(c *gin.Context) {
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

	log.Debug("删除装备 : %d, %s", req.ServerId, req.Uid)

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
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerEquip, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(model.Equip)
	err = json.Unmarshal(values.([]byte), &dst)

	indexs := make([]int, 0)
	for _, v := range req.Ids {
		for i, b := range dst.Equips {
			if b.Id == int32(v) {
				indexs = append(indexs, i)
				continue
			}
		}
	}

	for i := 0; i < len(indexs); i++ {
		dst.Equips = append(dst.Equips[:indexs[i]], dst.Equips[indexs[i]+1:]...)
	}

	js, _ := json.Marshal(dst)
	_redis.RedisExec("set", fmt.Sprintf("%s:%d", define.PlayerEquip, pl.RedisId), js)

	httpRetGame(c, SUCCESS, "success")
}
