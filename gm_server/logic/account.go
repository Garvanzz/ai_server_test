package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/name5566/leaf/log"
	"net/http"
	"sort"
	"xfx/core/model"
	"xfx/gm_server/conf"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
)

// http返回 游戏专用 码只返回200 不然不返
func httpRetGame(c *gin.Context, code int, message string, data ...map[string]interface{}) {

	ret := gin.H{
		"errcode": code,
		"errmsg":  message,
	}
	if data != nil && len(data) > 0 {
		//只认第一个data 其他直接抛 用...主要是为了省参
		for k, v := range data[0] {
			ret[k] = v
		}
	}
	log.Debug(" ret %v", ret)
	c.JSON(http.StatusOK, ret)

}

// http返回
func httpRet(c *gin.Context, code int, message string, data ...map[string]interface{}) {
	if data != nil && len(data) > 0 {
		//这里只走data[0] ...只是为了省参用 后续多参默认无效
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"data":    data[0],
		})
	} else {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	}

}

// 获取玩家信息
func GmGetPlayerInfo(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerInfo
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家数据 : %d, %s", req.ServerId, req.Uid)
	pl := make([]model.Account, 0)
	if len(req.Uid) <= 0 {
		err := db.AccountDb.Table(define.AccountTable).Where("server_id = ?", req.ServerId).Find(&pl)
		if err != nil {
			log.Error("getserverlist find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		err := db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", req.ServerId, req.Uid).Find(&pl)
		if err != nil {
			log.Error("getserverlist find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	js, _ := json.Marshal(pl)
	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(pl),
	})
}

// 获取玩家游戏数据
func GmGetPlayerGameInfo(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerInfo
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家游戏数据 : %d, %s", req.ServerId, req.Uid)

	var serverItem model.ServerItem
	has, err := db.AccountDb.Table(define.ServerGroup).Where("id = ?", req.ServerId).Get(&serverItem)
	if err != nil {
		log.Error("getserverlist find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		log.Error("getserverlist find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	pl := make([]model.Account, 0)
	if len(req.Uid) <= 0 {
		err := db.AccountDb.Table(define.AccountTable).Where("server_id = ?", req.ServerId).Find(&pl)
		if err != nil {
			log.Error("getserverlist find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		err := db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", req.ServerId, req.Uid).Find(&pl)
		if err != nil {
			log.Error("getserverlist find err :%v", err.Error())
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	_redis := db.InitRedis(fmt.Sprintf("%s:%d", conf.Server.RedisAddr, serverItem.RedisPort), conf.Server.RedisPassword)

	dsts := make([]*gm_model.GMPlayerInfo, 0)
	for i := 0; i < len(pl); i++ {
		values, err := redis.Values(_redis.RedisExec("hgetall", fmt.Sprintf("%s:%d", define.Player, pl[i].RedisId)))
		if err != nil {
			httpRetGame(c, ERR_DB, err.Error())
			return
		}

		dst := new(gm_model.GMPlayerInfo)
		err = redis.ScanStruct(values, dst)
		dsts = append(dsts, dst)
	}

	js, _ := json.Marshal(dsts)
	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(dsts),
	})
}

// 角色
func GmHero(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerHero
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家角色数据 : %d, %s", req.ServerId, req.Uid)

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

	dsts := make([]*gm_model.GMRespHero, 0)
	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerHero, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	valueslineup, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerLineUp, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(gm_model.GMModHero)
	err = json.Unmarshal(values.([]byte), &dst)

	dstlineup := new(gm_model.GMLineUp)
	err = json.Unmarshal(valueslineup.([]byte), &dstlineup)

	for _, v := range dst.Hero {
		isMain, isUse := "否", "否"
		if v.Id <= 3004 && v.Id >= 3001 {
			isMain = "是"
		} else {
			isMain = "否"
		}

		//主布阵
		for _, lid := range dstlineup.LineUps[1].HeroId {
			if lid == v.Id {
				isUse = "是"
				break
			}
		}

		dsts = append(dsts, &gm_model.GMRespHero{
			HeroId:         v.Id,
			HeroName:       "",
			HeroExp:        v.Exp,
			HeroLevel:      v.Level,
			HeroStage:      v.Stage,
			HeroStar:       v.Star,
			HeroIsMainHero: isMain,
			HeroIsUse:      isUse,
		})
	}

	//排序
	sort.Slice(dsts, func(i, j int) bool {
		if dsts[i].HeroIsMainHero == "是" {
			return true
		}

		if dsts[j].HeroIsMainHero == "是" {
			return true
		}

		if dsts[i].HeroIsUse == "是" {
			return true
		}

		if dsts[j].HeroIsUse == "是" {
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

// 编辑角色
func GmEditHero(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
		}
	}

	rawData, _ := c.GetRawData()

	var req gm_model.GmReqPlayerHero
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家编辑角色数据 : %d, %s", req.ServerId, req.Uid)

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

	values, err := _redis.RedisExec("get", fmt.Sprintf("%s:%d", define.PlayerHero, pl.RedisId))
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	dst := new(gm_model.GMModHero)
	err = json.Unmarshal(values.([]byte), &dst)

	for _, v := range dst.Hero {
		if v.Id == req.Data.HeroId {
			v.Exp = req.Data.HeroExp
			v.Star = req.Data.HeroStar
			v.Stage = req.Data.HeroStage
			v.Level = req.Data.HeroLevel
			break
		}
	}

	js, _ := json.Marshal(dst)
	_redis.RedisExec("set", fmt.Sprintf("%s:%d", define.PlayerHero, pl.RedisId), js)

	httpRetGame(c, SUCCESS, "success")
}
