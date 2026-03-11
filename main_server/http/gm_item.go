package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"xfx/core/config"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/main_server/messages"
	"xfx/pkg/log"
)

// GMGrantItem 给指定玩家发放道具（通过系统指令下发，在玩家进程内执行 bag.AddAward）
// 道具 id 校验与 Type 均从游戏服 config.Item 读取，gm_server 无需加载游戏配置
func (m *HttpModule) GMGrantItem(c *gin.Context) {
	var req model.GMGrantItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	var entries []messages.SysGrantItemEntry
	if req.GrantAll {
		// 一键发放全部：由游戏服从 config 构建列表
		allCfgs := config.Item.All()
		entries = make([]messages.SysGrantItemEntry, 0, len(allCfgs))
		for _, cfg := range allCfgs {
			entries = append(entries, messages.SysGrantItemEntry{
				ItemId:   cfg.Id,
				ItemType: cfg.Type,
				ItemNum:  50000,
			})
		}
	} else {
		if len(req.Items) == 0 {
			m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "items empty")
			return
		}
		entries = make([]messages.SysGrantItemEntry, 0, len(req.Items))
		for _, it := range req.Items {
			cfg, ok := config.Item.Find(int64(it.Id))
			if !ok {
				m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, fmt.Sprintf("item %d not found in config", it.Id))
				return
			}
			entries = append(entries, messages.SysGrantItemEntry{
				ItemId:   it.Id,
				ItemType: cfg.Type,
				ItemNum:  it.Num,
			})
		}
	}
	log.Debug("gm grant item player_id=%d items=%d", req.PlayerId, len(entries))
	invoke.DispatchSystemMessage(m, req.PlayerId, &messages.SysGrantItems{Items: entries})
	m.httpRetGame(c, SUCCESS, "success")
}
