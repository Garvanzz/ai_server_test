package http

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"

	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
)

// GMGetBag 查询玩家背包（读 Redis）
func (m *HttpModule) GMGetBag(c *gin.Context) {
	var req model.GMPlayerIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id")
		return
	}
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerBag, req.PlayerId))
	if err != nil {
		log.Error("GMGetBag redis err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	var bag *model.Bag
	if reply != nil {
		bag = new(model.Bag)
		if err := json.Unmarshal(reply.([]byte), bag); err != nil {
			log.Error("GMGetBag unmarshal err: %v", err)
			m.httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}
	if bag == nil || bag.Items == nil {
		bag = &model.Bag{Items: make(map[int32]int32)}
	}
	// 返回与 gm_server 一致的列表格式（ItemId, ItemNum 与 dto.GmRespPlayerBag 一致）
	type itemRow struct {
		ItemId  int32 `json:"ItemId"`
		ItemNum int32 `json:"ItemNum"`
	}
	list := make([]itemRow, 0, len(bag.Items))
	for k, v := range bag.Items {
		list = append(list, itemRow{ItemId: k, ItemNum: v})
	}
	js, _ := json.Marshal(list)
	m.httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(list),
	})
}

// GMDeleteItem 删除玩家背包道具（读-改-写 Redis）
func (m *HttpModule) GMDeleteItem(c *gin.Context) {
	var req model.GMItemDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id and item_ids")
		return
	}
	if len(req.ItemIds) == 0 {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "item_ids empty")
		return
	}
	key := fmt.Sprintf("%s:%d", define.PlayerBag, req.PlayerId)
	reply, err := db.RedisExec("GET", key)
	if err != nil {
		log.Error("GMDeleteItem redis get err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	bag := &model.Bag{Items: make(map[int32]int32)}
	if reply != nil {
		_ = json.Unmarshal(reply.([]byte), bag)
		if bag.Items == nil {
			bag.Items = make(map[int32]int32)
		}
	}
	for _, id := range req.ItemIds {
		delete(bag.Items, id)
	}
	js, err := json.Marshal(bag)
	if err != nil {
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	_, err = db.RedisExec("SET", key, js)
	if err != nil {
		log.Error("GMDeleteItem redis set err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMGetEquip 查询玩家装备（读 Redis）
func (m *HttpModule) GMGetEquip(c *gin.Context) {
	var req model.GMPlayerIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id")
		return
	}
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerEquip, req.PlayerId))
	if err != nil {
		log.Error("GMGetEquip redis err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	var equip *model.Equip
	if reply != nil {
		equip = new(model.Equip)
		if err := json.Unmarshal(reply.([]byte), equip); err != nil {
			log.Error("GMGetEquip unmarshal err: %v", err)
			m.httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}
	if equip == nil {
		equip = new(model.Equip)
	}
	js, _ := json.Marshal(equip)
	m.httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": 0,
	})
}

// GMSetEquip 设置玩家装备（写 Redis）
func (m *HttpModule) GMSetEquip(c *gin.Context) {
	var req model.GMEquipSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id and data")
		return
	}
	key := fmt.Sprintf("%s:%d", define.PlayerEquip, req.PlayerId)
	_, err := db.RedisExec("SET", key, []byte(req.Data))
	if err != nil {
		log.Error("GMSetEquip redis set err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMDeleteEquip 删除玩家装备（读-删-写 Redis Equips 切片）
func (m *HttpModule) GMDeleteEquip(c *gin.Context) {
	var req model.GMEquipDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id and ids")
		return
	}
	if len(req.Ids) == 0 {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "ids empty")
		return
	}
	key := fmt.Sprintf("%s:%d", define.PlayerEquip, req.PlayerId)
	reply, err := db.RedisExec("GET", key)
	if err != nil {
		log.Error("GMDeleteEquip redis get err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	equip := new(model.Equip)
	if reply != nil {
		_ = json.Unmarshal(reply.([]byte), equip)
	}
	if equip.Equips == nil {
		equip.Equips = make([]*model.EquipOption, 0)
	}
	idSet := make(map[int32]struct{})
	for _, id := range req.Ids {
		idSet[id] = struct{}{}
	}
	newEquips := make([]*model.EquipOption, 0, len(equip.Equips))
	for _, e := range equip.Equips {
		if e == nil {
			continue
		}
		if _, ok := idSet[e.Id]; !ok {
			newEquips = append(newEquips, e)
		}
	}
	equip.Equips = newEquips
	js, err := json.Marshal(equip)
	if err != nil {
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	_, err = db.RedisExec("SET", key, js)
	if err != nil {
		log.Error("GMDeleteEquip redis set err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMGetStage 查询玩家关卡（读 Redis）
func (m *HttpModule) GMGetStage(c *gin.Context) {
	var req model.GMPlayerIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id")
		return
	}
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerStage, req.PlayerId))
	if err != nil {
		log.Error("GMGetStage redis err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	var stage *model.Stage
	if reply != nil {
		stage = new(model.Stage)
		if err := json.Unmarshal(reply.([]byte), stage); err != nil {
			log.Error("GMGetStage unmarshal err: %v", err)
			m.httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}
	if stage == nil {
		stage = new(model.Stage)
	}
	js, _ := json.Marshal(stage)
	m.httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": 0,
	})
}

// GMSetStage 设置玩家关卡（写 Redis）
func (m *HttpModule) GMSetStage(c *gin.Context) {
	var req model.GMStageSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id and data")
		return
	}
	key := fmt.Sprintf("%s:%d", define.PlayerStage, req.PlayerId)
	_, err := db.RedisExec("SET", key, []byte(req.Data))
	if err != nil {
		log.Error("GMSetStage redis set err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMGetHero 查询玩家英雄与布阵（读 Redis）
func (m *HttpModule) GMGetHero(c *gin.Context) {
	var req model.GMPlayerIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id")
		return
	}
	hero := global.GetPlayerHero(req.PlayerId)
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerLineUp, req.PlayerId))
	if err != nil {
		log.Error("GMGetHero lineup redis err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	var lineup *model.LineUp
	if reply != nil {
		lineup = new(model.LineUp)
		_ = json.Unmarshal(reply.([]byte), lineup)
	}
	type out struct {
		Hero   *model.Hero   `json:"hero"`
		LineUp *model.LineUp `json:"lineup"`
	}
	js, _ := json.Marshal(&out{Hero: hero, LineUp: lineup})
	m.httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": 0,
	})
}

// GMSetHero 设置玩家英雄（写 Redis）
func (m *HttpModule) GMSetHero(c *gin.Context) {
	var req model.GMHeroSetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id and data")
		return
	}
	key := fmt.Sprintf("%s:%d", define.PlayerHero, req.PlayerId)
	_, err := db.RedisExec("SET", key, []byte(req.Data))
	if err != nil {
		log.Error("GMSetHero redis set err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMGetPlayerGameInfo 查询玩家游戏内信息（Redis Player 表，支持单人或多人）
func (m *HttpModule) GMGetPlayerGameInfo(c *gin.Context) {
	body, _ := c.GetRawData()
	var single model.GMPlayerIdReq
	if err := json.Unmarshal(body, &single); err == nil && single.PlayerId != 0 {
		info := global.GetPlayerInfo(single.PlayerId)
		if info == nil {
			m.httpRetGame(c, SUCCESS, "success", map[string]any{"data": "[]", "totalCount": 0})
			return
		}
		js, _ := json.Marshal([]*model.PlayerInfo{info})
		m.httpRetGame(c, SUCCESS, "success", map[string]any{"data": string(js), "totalCount": 1})
		return
	}
	var multi model.GMPlayerIdsReq
	if err := json.Unmarshal(body, &multi); err != nil || len(multi.PlayerIds) == 0 {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id or player_ids")
		return
	}
	list := make([]*model.PlayerInfo, 0, len(multi.PlayerIds))
	for _, pid := range multi.PlayerIds {
		info := global.GetPlayerInfo(pid)
		if info != nil {
			list = append(list, info)
		}
	}
	js, _ := json.Marshal(list)
	m.httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(list),
	})
}

// GMGetPlayerInfo 单玩家 Redis Player 表（hgetall），返回与 gm 后台一致结构
func (m *HttpModule) GMGetPlayerInfo(c *gin.Context) {
	var req model.GMPlayerIdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need player_id")
		return
	}
	values, err := redis.Values(db.RedisExec("HGETALL", fmt.Sprintf("%s:%d", define.Player, req.PlayerId)))
	if err != nil {
		log.Error("GMGetPlayerInfo redis err: %v", err)
		m.httpRetGame(c, ERR_DB, err.Error())
		return
	}
	info := new(model.PlayerInfo)
	if len(values) > 0 {
		if err := redis.ScanStruct(values, info); err != nil {
			log.Error("GMGetPlayerInfo scan err: %v", err)
			m.httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}
	js, _ := json.Marshal(info)
	m.httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": 1,
	})
}
