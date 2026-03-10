package logic

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/name5566/leaf/log"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
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
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerInfo
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

// 获取玩家游戏数据（经 main_server 读 Redis Player）
func GmGetPlayerGameInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerInfo
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家游戏数据 : %d, %s", req.ServerId, req.Uid)

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
	if len(pl) == 0 {
		httpRetGame(c, SUCCESS, "success", map[string]any{"data": "[]", "totalCount": 0})
		return
	}

	playerIds := make([]int64, 0, len(pl))
	for i := range pl {
		playerIds = append(playerIds, pl[i].RedisId)
	}
	body, _ := json.Marshal(model.GMPlayerIdsReq{PlayerIds: playerIds})
	err, respBody := HttpRequest(body, "/gm/player/game-info")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// 角色（经 main_server 读 Redis Hero+LineUp）
func GmHero(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerHero
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家角色数据 : %d, %s", req.ServerId, req.Uid)

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	body, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequest(body, "/gm/hero")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// 编辑角色（经 main_server 写 Redis Hero）
func GmEditHero(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerHero
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家编辑角色数据 : %d, %s", req.ServerId, req.Uid)

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	// 先拉取当前 hero，在内存中改单条后写回
	bodyGet, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequest(bodyGet, "/gm/hero")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		httpRetGame(c, ERR_SERVER_INTERNAL, "parse hero response err")
		return
	}
	var heroLineup struct {
		Hero   *model.Hero `json:"Hero"`
		LineUp interface{} `json:"LineUp"`
	}
	if err := json.Unmarshal([]byte(wrap.Data), &heroLineup); err != nil || heroLineup.Hero == nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, "parse hero data err")
		return
	}
	if heroLineup.Hero.Hero != nil {
		if v := heroLineup.Hero.Hero[req.Data.HeroId]; v != nil {
			v.Exp = req.Data.HeroExp
			v.Star = req.Data.HeroStar
			v.Stage = req.Data.HeroStage
			v.Level = req.Data.HeroLevel
		}
	}
	dataJs, _ := json.Marshal(heroLineup.Hero)
	setBody, _ := json.Marshal(struct {
		PlayerId int64           `json:"player_id"`
		Data     json.RawMessage `json:"data"`
	}{PlayerId: playerId, Data: dataJs})
	err, setResp := HttpRequest(setBody, "/gm/hero/set")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, setResp)
}
