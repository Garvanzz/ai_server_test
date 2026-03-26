package http

import (
	"github.com/gin-gonic/gin"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
)

// GM 活动接口：列表、查询、暂停/恢复/关闭/重启/移除等，均走 /gm 鉴权，通过 invoke.ActivityClient 调用活动模块

// GMActivityList 列出所有活动（含状态）
func (m *HttpModule) GMActivityList(c *gin.Context) {
	client := invoke.ActivityClient(m)
	list, err := client.ListAllActivities()
	if err != nil {
		log.Error("gm activity list err: %v", err)
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGameData(c, SUCCESS, "success", list, map[string]any{"list": list})
}

// GMActivityGetByActId 按活动实例 ID 查询
func (m *HttpModule) GMActivityGetByActId(c *gin.Context) {
	var req struct {
		ActId int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	info, err := client.GetActivityByActId(req.ActId)
	if err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGameData(c, SUCCESS, "success", info, map[string]any{"activity": info})
}

// GMActivityGetByCfgId 按配置 ID 查询
func (m *HttpModule) GMActivityGetByCfgId(c *gin.Context) {
	var req struct {
		CfgId int64 `json:"cfg_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need cfg_id")
		return
	}
	client := invoke.ActivityClient(m)
	info, err := client.GetActivityByCfgId(req.CfgId)
	if err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGameData(c, SUCCESS, "success", info, map[string]any{"activity": info})
}

// GMActivityStop 暂停活动（Running -> Stopped）
func (m *HttpModule) GMActivityStop(c *gin.Context) {
	var req struct {
		ActId int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.StopActivity(req.ActId); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityRecover 恢复活动（Stopped -> Running）
func (m *HttpModule) GMActivityRecover(c *gin.Context) {
	var req struct {
		ActId int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.RecoverActivity(req.ActId); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityClose 强制结束活动
func (m *HttpModule) GMActivityClose(c *gin.Context) {
	var req struct {
		ActId int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.CloseActivity(req.ActId); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityRestart 重启活动（Stopped/Closed -> Waiting，新 actId）
func (m *HttpModule) GMActivityRestart(c *gin.Context) {
	var req struct {
		ActId int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.RestartActivity(req.ActId); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityRemove 彻底移除活动
func (m *HttpModule) GMActivityRemove(c *gin.Context) {
	var req struct {
		ActId int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.RemoveActivity(req.ActId); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityCloseByCfgId 按配置 ID 强制结束
func (m *HttpModule) GMActivityCloseByCfgId(c *gin.Context) {
	var req struct {
		CfgId int64 `json:"cfg_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need cfg_id")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.CloseActivityByCfgId(req.CfgId); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityStopByType 按类型暂停（当前 Running 的该类型实例）
func (m *HttpModule) GMActivityStopByType(c *gin.Context) {
	var req struct {
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Type == "" {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need type")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.StopActivityByType(req.Type); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityAdjustTime 调整活动时间（startTime/endTime/closeTime，Unix 秒，传 0 表示不修改）
func (m *HttpModule) GMActivityAdjustTime(c *gin.Context) {
	var req struct {
		ActId     int64 `json:"act_id"`
		StartTime int64 `json:"start_time"`
		EndTime   int64 `json:"end_time"`
		CloseTime int64 `json:"close_time"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ActId == 0 {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.AdjustActivityTime(req.ActId, req.StartTime, req.EndTime, req.CloseTime); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityForceStart 强制开启等待中的活动（Waiting -> Running）
func (m *HttpModule) GMActivityForceStart(c *gin.Context) {
	var req struct {
		ActId int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ActId == 0 {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	if err := client.ForceStartActivity(req.ActId); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGame(c, SUCCESS, "success")
}

// GMActivityPlayerCount 查询活动参与玩家数
func (m *HttpModule) GMActivityPlayerCount(c *gin.Context) {
	var req struct {
		ActId int64 `json:"act_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ActId == 0 {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need act_id")
		return
	}
	client := invoke.ActivityClient(m)
	count, err := client.GetActivityPlayerCount(req.ActId)
	if err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	m.httpRetGameData(c, SUCCESS, "success", map[string]any{"count": count})
}
