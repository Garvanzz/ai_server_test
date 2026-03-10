package logic

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"

	"xfx/core/config"
	"xfx/core/model"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

// 获取道具（经 main_server 读 Redis）
func GmItem(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家游戏道具数据 : %d, %s", req.ServerId, req.Uid)

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
	err, respBody := HttpRequest(body, "/gm/bag")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// 添加道具（经 main_server 下发，玩家进程内执行）
func GmAddItem(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	itemCfgs := config.Item.All()
	itemCfg, ok := itemCfgs[int64(req.ItemId)]
	if !ok {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, fmt.Sprintf("item %d not found in config", req.ItemId))
		return
	}

	gmReq := model.GMGrantItemReq{
		PlayerId: playerId,
		Items: []model.MailItem{
			{
				Id:   int32(req.ItemId),
				Num:  int32(req.ItemNum),
				Type: itemCfg.Type,
			},
		},
	}
	js, _ := json.Marshal(gmReq)
	if err, _ := HttpRequest(js, "/gm/item"); err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}

	httpRetGame(c, SUCCESS, "success")
}

// 一键添加道具（经 main_server 下发）
func GmOneKeyAddItem(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	confs := config.Item.All()
	items := make([]model.MailItem, 0, len(confs))
	for _, v := range confs {
		items = append(items, model.MailItem{
			Id:   v.Id,
			Num:  50000,
			Type: v.Type,
		})
	}

	gmReq := model.GMGrantItemReq{
		PlayerId: playerId,
		Items:    items,
	}
	js, _ := json.Marshal(gmReq)
	if err, _ := HttpRequest(js, "/gm/item"); err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}

	httpRetGame(c, SUCCESS, "success")
}

// 删除道具（经 main_server 写 Redis）
func GmDeleteItem(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家背包数据 : %d, %s", req.ServerId, req.Uid)

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	ids := make([]int32, 0, len(req.Ids))
	for _, v := range req.Ids {
		ids = append(ids, int32(v))
	}
	body, _ := json.Marshal(model.GMItemDeleteReq{PlayerId: playerId, ItemIds: ids})
	err, respBody := HttpRequest(body, "/gm/item/delete")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}
