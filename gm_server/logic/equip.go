package logic

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"xfx/core/model"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

// 装备（经 main_server 读 Redis）
func GmEquip(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerEquip
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("请求玩家装备数据 : %d, %s", req.ServerId, req.Uid)
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
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/equip")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		HTTPRetGame(c, SUCCESS, "success", map[string]any{"data": "[]", "totalCount": 0})
		return
	}
	dst := new(model.Equip)
	if err := json.Unmarshal([]byte(wrap.Data), dst); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse equip data err")
		return
	}
	dsts := make([]*dto.GmRespPlayerEquip, 0)
	if dst.Equips != nil {
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
			} else if v.Index == 6 {
				indexName = "鞋子"
			}
			dsts = append(dsts, &dto.GmRespPlayerEquip{
				EquipId:    v.Id,
				EquipCId:   v.CId,
				EquipNum:   v.Num,
				EquipLevel: v.Level,
				EquipIndex: indexName,
				EquipIsUse: v.IsUse,
			})
		}
	}
	sort.Slice(dsts, func(i, j int) bool {
		if dsts[i].EquipIsUse {
			return true
		}
		if dsts[j].EquipIsUse {
			return false
		}
		return false
	})
	js, _ := json.Marshal(dsts)
	HTTPRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(dsts),
	})
}

// 删除装备（经 main_server 写 Redis）
func GmDeleteEquip(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerEquip
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}
	if len(req.Ids) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "ids required")
		return
	}

	log.Debug("删除装备 : %d, %s", req.ServerId, req.Uid)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	ids := make([]int32, 0, len(req.Ids))
	for _, v := range req.Ids {
		ids = append(ids, int32(v))
	}
	body, _ := json.Marshal(model.GMEquipDeleteReq{PlayerId: playerId, Ids: ids})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/equip/delete")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}
