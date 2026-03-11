package logic

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/name5566/leaf/log"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
)

const maxPlayerListLimit = 500 // 无 uid 时单次最多返回账号数，防止全表扫

// 获取玩家信息（仅 MySQL account 表）
func GmGetPlayerInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqPlayerInfo
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId required")
		return
	}

	log.Debug("请求玩家数据 : %d, %s", req.ServerId, req.Uid)
	pl := make([]model.Account, 0)
	if strings.TrimSpace(req.Uid) == "" {
		err := db.AccountDb.Table(define.AccountTable).Where("server_id = ?", req.ServerId).Limit(maxPlayerListLimit).Find(&pl)
		if err != nil {
			log.Error("GmGetPlayerInfo find err :%v", err)
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		err := db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", req.ServerId, req.Uid).Find(&pl)
		if err != nil {
			log.Error("GmGetPlayerInfo find err :%v", err)
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	}
	js, _ := json.Marshal(pl)
	HTTPRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(pl),
	})
}

// 获取玩家游戏数据（经 main_server 读 Redis Player）
func GmGetPlayerGameInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqPlayerInfo
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId required")
		return
	}

	log.Debug("请求玩家游戏数据 : %d, %s", req.ServerId, req.Uid)
	pl := make([]model.Account, 0)
	if strings.TrimSpace(req.Uid) == "" {
		err := db.AccountDb.Table(define.AccountTable).Where("server_id = ?", req.ServerId).Limit(maxPlayerListLimit).Find(&pl)
		if err != nil {
			log.Error("GmGetPlayerGameInfo find err :%v", err)
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	} else {
		err := db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", req.ServerId, req.Uid).Find(&pl)
		if err != nil {
			log.Error("GmGetPlayerGameInfo find err :%v", err)
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	}
	if len(pl) == 0 {
		HTTPRetGame(c, SUCCESS, "success", map[string]any{"data": "[]", "totalCount": 0})
		return
	}

	playerIds := make([]int64, 0, len(pl))
	for i := range pl {
		playerIds = append(playerIds, pl[i].RedisId)
	}
	body, _ := json.Marshal(model.GMPlayerIdsReq{PlayerIds: playerIds})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/player/game-info")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// 角色（经 main_server 读 Redis Hero+LineUp）
func GmHero(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerHero
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("请求玩家角色数据 : %d, %s", req.ServerId, req.Uid)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	body, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/hero")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// 编辑角色（经 main_server 写 Redis Hero）
func GmEditHero(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerHero
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" || req.Data == nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId, uid and data required")
		return
	}

	log.Debug("请求玩家编辑角色数据 : %d, %s", req.ServerId, req.Uid)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	// 先拉取当前 hero，在内存中改单条后写回
	bodyGet, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequestToServer(req.ServerId, bodyGet, "/gm/hero")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse hero response err")
		return
	}
	var heroLineup struct {
		Hero   *model.Hero `json:"Hero"`
		LineUp interface{} `json:"LineUp"`
	}
	if err := json.Unmarshal([]byte(wrap.Data), &heroLineup); err != nil || heroLineup.Hero == nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse hero data err")
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
	err, setResp := HttpRequestToServer(req.ServerId, setBody, "/gm/hero/set")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, setResp)
}
