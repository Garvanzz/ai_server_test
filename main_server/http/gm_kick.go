package http

import (
	"github.com/gin-gonic/gin"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/main_server/messages"
	"xfx/pkg/log"
)

// GMKick 踢指定玩家下线（示例：常用 GM 通过 invoke.DispatchSystemMessage 下发系统指令）
func (m *HttpModule) GMKick(c *gin.Context) {
	var req model.GMKickReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	log.Debug("gm kick player_id=%d reason=%s", req.PlayerId, req.Reason)
	invoke.DispatchSystemMessage(m, req.PlayerId, &messages.SysKick{Reason: req.Reason})
	m.httpRetGame(c, SUCCESS, "success")
}
