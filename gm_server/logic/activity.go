package logic

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"xfx/gm_server/dto"
)

// GmActivityList 列出指定区服所有活动（含状态）
func GmActivityList(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityList
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId")
		return
	}

	err, respBody := HttpRequestToServer(req.ServerId, nil, "/gm/activity/list")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityGetByActId 按活动实例 ID 查询
func GmActivityGetByActId(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByActId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{"act_id": req.ActId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/get_by_act_id")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityGetByCfgId 按配置 ID 查询
func GmActivityGetByCfgId(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByCfgId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.CfgId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and cfgId")
		return
	}

	body, _ := json.Marshal(map[string]any{"cfg_id": req.CfgId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/get_by_cfg_id")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityStop 暂停活动（Running -> Stopped）
func GmActivityStop(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByActId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{"act_id": req.ActId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/stop")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityRecover 恢复活动（Stopped -> Running）
func GmActivityRecover(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByActId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{"act_id": req.ActId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/recover")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityClose 强制结束活动
func GmActivityClose(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByActId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{"act_id": req.ActId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/close")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityRestart 重启活动（Stopped/Closed -> Waiting，新 actId）
func GmActivityRestart(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByActId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{"act_id": req.ActId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/restart")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityRemove 彻底移除活动实例
func GmActivityRemove(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByActId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{"act_id": req.ActId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/remove")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityCloseByCfgId 按配置 ID 强制结束
func GmActivityCloseByCfgId(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByCfgId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.CfgId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and cfgId")
		return
	}

	body, _ := json.Marshal(map[string]any{"cfg_id": req.CfgId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/close_by_cfg_id")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityStopByType 按类型暂停活动
func GmActivityStopByType(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityStopByType
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.Type == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and type")
		return
	}

	body, _ := json.Marshal(map[string]any{"type": req.Type})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/stop_by_type")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityAdjustTime 调整活动时间（startTime/endTime/closeTime，Unix 秒，传 0 不修改）
func GmActivityAdjustTime(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityAdjustTime
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{
		"act_id":     req.ActId,
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
		"close_time": req.CloseTime,
	})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/adjust_time")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityForceStart 强制开启等待中的活动
func GmActivityForceStart(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByActId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{"act_id": req.ActId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/force_start")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}

// GmActivityPlayerCount 查询活动参与玩家数
func GmActivityPlayerCount(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmReqActivityByActId
	if err := json.Unmarshal(rawData, &req); err != nil || req.ServerId <= 0 || req.ActId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need serverId and actId")
		return
	}

	body, _ := json.Marshal(map[string]any{"act_id": req.ActId})
	err, respBody := HttpRequestToServer(req.ServerId, body, "/gm/activity/player_count")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	forwardMainServerResponse(c, respBody)
}
