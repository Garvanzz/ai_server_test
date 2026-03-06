package http

import (
	"github.com/gin-gonic/gin"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/main_server/messages"
	"xfx/pkg/log"
)

// GMGrantItem 给指定玩家发放道具（通过系统指令下发，在玩家进程内执行 bag.AddAward）
func (m *HttpModule) GMGrantItem(c *gin.Context) {
	var req model.GMGrantItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if len(req.Items) == 0 {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "items empty")
		return
	}
	items := make([]messages.SysGrantItemEntry, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, messages.SysGrantItemEntry{
			ItemId:   it.Id,
			ItemType: it.Type,
			ItemNum:  it.Num,
		})
	}
	log.Debug("gm grant item player_id=%d items=%d", req.PlayerId, len(items))
	invoke.DispatchSystemMessage(m, req.PlayerId, &messages.SysGrantItems{Items: items})
	m.httpRetGame(c, SUCCESS, "success")
}
