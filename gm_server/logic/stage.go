package logic

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"xfx/core/model"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

// GmGetStageInfo 获取关卡列表信息（经 main_server 读 Redis）
func GmGetStageInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqGetStageInfo
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家关卡数据 : %d, %s", req.ServerId, req.Uid)

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
	err, respBody := HttpRequest(body, "/gm/stage")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmSetStageInfo 设置关卡信息（经 main_server 读-写 Redis）
func GmSetStageInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqSetStageInfo
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("请求玩家关卡数据 : %d, %s", req.ServerId, req.Uid)

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	bodyGet, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequest(bodyGet, "/gm/stage")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		httpRetGame(c, ERR_SERVER_INTERNAL, "parse stage response err")
		return
	}
	// TODO: 按 req 修改 stage 后再写回
	setBody, _ := json.Marshal(struct {
		PlayerId int64           `json:"player_id"`
		Data     json.RawMessage `json:"data"`
	}{PlayerId: playerId, Data: json.RawMessage(wrap.Data)})
	err, setResp := HttpRequest(setBody, "/gm/stage/set")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, setResp)
}

// GmAddStageInfo 添加关卡信息（经 main_server 读-写 Redis）
func GmAddStageInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerBag
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	bodyGet, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequest(bodyGet, "/gm/stage")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		httpRetGame(c, ERR_SERVER_INTERNAL, "parse stage response err")
		return
	}
	// TODO: 按传入关卡补全中间关卡后再写回
	setBody, _ := json.Marshal(struct {
		PlayerId int64           `json:"player_id"`
		Data     json.RawMessage `json:"data"`
	}{PlayerId: playerId, Data: json.RawMessage(wrap.Data)})
	err, setResp := HttpRequest(setBody, "/gm/stage/set")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, setResp)
}

// GmDeleteStageInfo 删除关卡信息（经 main_server 读-写 Redis）
func GmDeleteStageInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqPlayerEquip
	err := json.Unmarshal(rawData, &req)
	if err != nil {
		log.Fatal("解析失败:", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug("删除关卡 : %d, %s", req.ServerId, req.Uid)

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		httpRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		httpRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	bodyGet, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequest(bodyGet, "/gm/stage")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		httpRetGame(c, ERR_SERVER_INTERNAL, "parse stage response err")
		return
	}
	// TODO: 按 req.Ids 删除对应关卡后再写回
	setBody, _ := json.Marshal(struct {
		PlayerId int64           `json:"player_id"`
		Data     json.RawMessage `json:"data"`
	}{PlayerId: playerId, Data: json.RawMessage(wrap.Data)})
	err, setResp := HttpRequest(setBody, "/gm/stage/set")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, setResp)
}
