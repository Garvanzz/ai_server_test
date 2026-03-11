package logic

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"xfx/core/model"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

// 获取道具（经 main_server 读 Redis）
func GmItem(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("请求玩家游戏道具数据 : %d, %s", req.ServerId, req.Uid)

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
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/bag")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// 添加道具（经 main_server 下发，玩家进程内执行）
func GmAddItem(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	// 道具 id 校验与 Type 由 main_server 根据 config 完成，gm_server 只转发
	gmReq := model.GMGrantItemReq{
		PlayerId: playerId,
		Items: []model.MailItem{
			{Id: int32(req.ItemId), Num: int32(req.ItemNum), Type: 0},
		},
	}
	js, _ := json.Marshal(gmReq)
	if err, _ := HttpRequestToServer(req.ServerId, js, "/gm/item"); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}

	HTTPRetGame(c, SUCCESS, "success")
}

// 一键添加道具（经 main_server 下发）
func GmOneKeyAddItem(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	// 一键发放全部：由 main_server 根据 config.Item 构建列表并发放
	gmReq := model.GMGrantItemReq{PlayerId: playerId, GrantAll: true}
	js, _ := json.Marshal(gmReq)
	if err, _ := HttpRequestToServer(req.ServerId, js, "/gm/item"); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}

	HTTPRetGame(c, SUCCESS, "success")
}

// 删除道具（经 main_server 写 Redis）
func GmDeleteItem(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}
	if len(req.Ids) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "ids required")
		return
	}

	ids := make([]int32, 0, len(req.Ids))
	for _, v := range req.Ids {
		ids = append(ids, int32(v))
	}
	body, _ := json.Marshal(model.GMItemDeleteReq{PlayerId: playerId, ItemIds: ids})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/item/delete")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}
