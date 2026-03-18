package http

import (
	"time"
	"xfx/pkg/utils"

	"github.com/gin-gonic/gin"
)

// GM 时间调试：查询/设置游戏逻辑时间偏移，便于在后台调时间测活动等。

// GMTimeGet 查询当前真实时间、游戏逻辑时间与偏移（天）。
// GET /gm/time
func (m *HttpModule) GMTimeGet(c *gin.Context) {
	realNow := time.Now()
	gameNow := utils.Now()
	offset := utils.GetTimeOffset()
	offsetDays := int64(offset / (24 * time.Hour))
	m.httpRetGameData(c, SUCCESS, "success", map[string]any{
		"realTime":      realNow.Unix(),
		"gameTime":      gameNow.Unix(),
		"offsetDays":    offsetDays,
		"offsetEnabled": utils.TimeOffsetEnabled(),
		"realIso":       realNow.Format(time.RFC3339),
		"gameIso":       gameNow.Format(time.RFC3339),
	}, map[string]any{
		"real_time":      realNow.Unix(),
		"game_time":      gameNow.Unix(),
		"offset_days":    offsetDays,
		"offset_enabled": utils.TimeOffsetEnabled(),
		"real_iso":       realNow.Format(time.RFC3339),
		"game_iso":       gameNow.Format(time.RFC3339),
	})
}

// GMTimeSetOffset 设置游戏逻辑时间偏移（单位：天）。仅 Debug 模式生效，正式服不允许修改。
// POST /gm/time/set_offset  body: { "offset_days": 7 }
func (m *HttpModule) GMTimeSetOffset(c *gin.Context) {
	if !utils.TimeOffsetEnabled() {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "正式服不允许修改时间偏移，仅 Debug 模式可用")
		return
	}
	var req struct {
		OffsetDays int64 `json:"offset_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err, need offset_days (int)")
		return
	}
	offset := time.Duration(req.OffsetDays) * 24 * time.Hour
	utils.SetTimeOffset(offset)
	m.httpRetGameData(c, SUCCESS, "success", map[string]any{
		"offsetDays": req.OffsetDays,
		"gameTime":   utils.Now().Unix(),
		"gameIso":    utils.Now().Format(time.RFC3339),
	}, map[string]any{
		"offset_days": req.OffsetDays,
		"game_time":   utils.Now().Unix(),
		"game_iso":    utils.Now().Format(time.RFC3339),
	})
}
