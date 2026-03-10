package logic

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"xfx/core/model"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

// GmGetStageInfo 获取关卡列表信息（经 main_server 读 Redis）
func GmGetStageInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqGetStageInfo
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("请求玩家关卡数据 : %d, %s", req.ServerId, req.Uid)
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
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/stage")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmSetStageInfo 设置关卡信息（经 main_server 读-写 Redis）
func GmSetStageInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()

	var req dto.GmReqSetStageInfo
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("请求玩家关卡数据 : %d, %s", req.ServerId, req.Uid)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	bodyGet, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequestToServer(req.ServerId, bodyGet, "/gm/stage")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse stage response err")
		return
	}
	var stage model.Stage
	if err := json.Unmarshal([]byte(wrap.Data), &stage); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse stage data err")
		return
	}
	applySetStageReq(&stage, &req)
	dataJs, _ := json.Marshal(&stage)
	setBody, _ := json.Marshal(struct {
		PlayerId int64           `json:"player_id"`
		Data     json.RawMessage `json:"data"`
	}{PlayerId: playerId, Data: dataJs})
	err, setResp := HttpRequestToServer(req.ServerId, setBody, "/gm/stage/set")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, setResp)
}

// applySetStageReq 按 req 修改 stage：找到或创建 Cycle/Chapter/StageId，设置 Exp、Pass、PassState
func applySetStageReq(s *model.Stage, req *dto.GmReqSetStageInfo) {
	if s.Stage == nil {
		s.Stage = make(map[int32]map[int32]*model.ChapterOpt)
	}
	if s.Stage[req.Cycle] == nil {
		s.Stage[req.Cycle] = make(map[int32]*model.ChapterOpt)
	}
	if s.Stage[req.Cycle][req.Chapter] == nil {
		s.Stage[req.Cycle][req.Chapter] = &model.ChapterOpt{Stages: make(map[int32]*model.StageOpt)}
	}
	ch := s.Stage[req.Cycle][req.Chapter]
	if ch.Stages == nil {
		ch.Stages = make(map[int32]*model.StageOpt)
	}
	opt := ch.Stages[req.StageId]
	if opt == nil {
		opt = &model.StageOpt{Id: req.StageId}
		ch.Stages[req.StageId] = opt
	}
	opt.Exp = req.Exp
	opt.Pass = req.State >= 1
	opt.PassState = req.State
}

// GmAddStageInfo 添加关卡信息（经 main_server 读-写 Redis），确保指定周目章节关卡存在并设置
func GmAddStageInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmAddStageInfo
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	bodyGet, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequestToServer(req.ServerId, bodyGet, "/gm/stage")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse stage response err")
		return
	}
	var stage model.Stage
	if err := json.Unmarshal([]byte(wrap.Data), &stage); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse stage data err")
		return
	}
	applyAddStageReq(&stage, &req)
	dataJs, _ := json.Marshal(&stage)
	setBody, _ := json.Marshal(struct {
		PlayerId int64           `json:"player_id"`
		Data     json.RawMessage `json:"data"`
	}{PlayerId: playerId, Data: dataJs})
	err, setResp := HttpRequestToServer(req.ServerId, setBody, "/gm/stage/set")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, setResp)
}

func applyAddStageReq(s *model.Stage, req *dto.GmAddStageInfo) {
	if s.Stage == nil {
		s.Stage = make(map[int32]map[int32]*model.ChapterOpt)
	}
	if s.Stage[req.Cycle] == nil {
		s.Stage[req.Cycle] = make(map[int32]*model.ChapterOpt)
	}
	if s.Stage[req.Cycle][req.Chapter] == nil {
		s.Stage[req.Cycle][req.Chapter] = &model.ChapterOpt{Stages: make(map[int32]*model.StageOpt)}
	}
	ch := s.Stage[req.Cycle][req.Chapter]
	if ch.Stages == nil {
		ch.Stages = make(map[int32]*model.StageOpt)
	}
	opt := ch.Stages[req.StageId]
	if opt == nil {
		opt = &model.StageOpt{Id: req.StageId}
		ch.Stages[req.StageId] = opt
	}
	opt.Exp = req.Exp
	opt.Pass = req.State >= 1
	opt.PassState = req.State
}

// GmDeleteStageInfo 删除关卡信息（经 main_server 读-写 Redis），按 req 删除指定关卡或章节/周目
func GmDeleteStageInfo(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmDeleteStageInfo
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("删除关卡 : %d, %s", req.ServerId, req.Uid)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	bodyGet, _ := json.Marshal(model.GMPlayerIdReq{PlayerId: playerId})
	err, respBody := HttpRequestToServer(req.ServerId, bodyGet, "/gm/stage")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	var wrap struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(respBody), &wrap); err != nil || wrap.Data == "" {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse stage response err")
		return
	}
	var stage model.Stage
	if err := json.Unmarshal([]byte(wrap.Data), &stage); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "parse stage data err")
		return
	}
	applyDeleteStageReq(&stage, &req)
	dataJs, _ := json.Marshal(&stage)
	setBody, _ := json.Marshal(struct {
		PlayerId int64           `json:"player_id"`
		Data     json.RawMessage `json:"data"`
	}{PlayerId: playerId, Data: dataJs})
	err, setResp := HttpRequestToServer(req.ServerId, setBody, "/gm/stage/set")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, setResp)
}

func applyDeleteStageReq(s *model.Stage, req *dto.GmDeleteStageInfo) {
	if s.Stage == nil {
		return
	}
	if req.IsDelCycle {
		delete(s.Stage, req.Cycle)
		return
	}
	if s.Stage[req.Cycle] == nil {
		return
	}
	if req.IsDelChapter {
		delete(s.Stage[req.Cycle], req.Chapter)
		return
	}
	if ch := s.Stage[req.Cycle][req.Chapter]; ch != nil && ch.Stages != nil {
		delete(ch.Stages, req.Stage)
	}
}
