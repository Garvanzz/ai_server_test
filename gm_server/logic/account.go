package logic

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/name5566/leaf/log"

	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
)

const maxPlayerListLimit = 500 // 无 uid 时单次最多返回账号数，防止全表扫

func queryAccountRoleProfiles(entryServerId int, uid string) ([]model.AccountRoleProfile, error) {
	rows := make([]model.AccountRoleProfile, 0)
	sql := `
SELECT
	a.id AS account_id,
	r.id AS role_id,
	a.uid AS uid,
	a.account AS account,
	r.nick_name AS nick_name,
	a.type AS type,
	a.device_id AS device_id,
	a.platform AS platform,
	a.is_white_acc AS is_white_acc,
	a.login_ban AS login_ban,
	a.login_ban_reason AS login_ban_reason,
	a.chat_ban AS chat_ban,
	a.chat_ban_reason AS chat_ban_reason,
	r.entry_server_id AS entry_server_id,
	r.logic_server_id AS logic_server_id,
	r.origin_server_id AS origin_server_id,
	r.redis_id AS redis_id,
	r.system_mail_id AS system_mail_id,
	r.create_time AS role_create_time,
	r.online_time AS role_online_time,
	r.offline_time AS role_offline_time,
	r.last_login_time AS role_last_login_time,
	a.create_time AS account_create_time,
	a.last_login_entry_server_id AS last_login_entry_server_id,
	a.last_login_logic_server_id AS last_login_logic_server_id
FROM account_role r
INNER JOIN account a ON a.id = r.account_id
WHERE r.entry_server_id = ?`
	args := []interface{}{entryServerId}
	if strings.TrimSpace(uid) != "" {
		sql += ` AND r.uid = ?`
		args = append(args, uid)
	}
	sql += ` ORDER BY r.id DESC LIMIT ?`
	args = append(args, maxPlayerListLimit)
	return rows, db.AccountDb.SQL(sql, args...).Find(&rows)
}

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
	logicServerId := resolveLogicServerID(req.ServerId)
	log.Debug("请求玩家数据 : %d(logic:%d), %s", req.ServerId, logicServerId, req.Uid)
	pl, err := queryAccountRoleProfiles(req.ServerId, req.Uid)
	if err != nil {
		log.Error("GmGetPlayerInfo find err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGameData(c, SUCCESS, "success", pl, map[string]any{"totalCount": len(pl)})
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
	logicServerId := resolveLogicServerID(req.ServerId)
	log.Debug("请求玩家游戏数据 : %d(logic:%d), %s", req.ServerId, logicServerId, req.Uid)
	pl, err := queryAccountRoleProfiles(req.ServerId, req.Uid)
	if err != nil {
		log.Error("GmGetPlayerGameInfo find err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if len(pl) == 0 {
		HTTPRetGameData(c, SUCCESS, "success", []model.AccountRoleProfile{}, map[string]any{"totalCount": 0})
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

// 角色（经 main_server 读 Redis Hero+LineUp），返回平铺英雄列表供前端展示
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
	var heroLineup struct {
		Hero   *model.Hero   `json:"hero"`
		LineUp *model.LineUp `json:"lineup"`
	}
	if err := decodeForwardedData(respBody, &heroLineup); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse hero data err")
		return
	}
	// 计算哪些英雄在阵容中
	usedHeroes := make(map[int32]bool)
	if heroLineup.LineUp != nil && heroLineup.LineUp.LineUps != nil {
		for _, lu := range heroLineup.LineUp.LineUps {
			if lu == nil {
				continue
			}
			for _, hid := range lu.HeroId {
				usedHeroes[hid] = true
			}
		}
	}
	list := make([]*dto.GMRespHero, 0)
	if heroLineup.Hero != nil && heroLineup.Hero.Hero != nil {
		for _, v := range heroLineup.Hero.Hero {
			if v == nil {
				continue
			}
			isUse := "否"
			if usedHeroes[v.Id] {
				isUse = "是"
			}
			list = append(list, &dto.GMRespHero{
				HeroId:    v.Id,
				HeroLevel: v.Level,
				HeroStage: v.Stage,
				HeroStar:  v.Star,
				HeroExp:   v.Exp,
				HeroIsUse: isUse,
			})
		}
	}
	HTTPRetGameData(c, SUCCESS, "success", list, map[string]any{"totalCount": len(list)})
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
	var heroLineup struct {
		Hero   *model.Hero `json:"hero"`
		LineUp interface{} `json:"lineup"`
	}
	if err := decodeForwardedData(respBody, &heroLineup); err != nil || heroLineup.Hero == nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse hero response err")
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
