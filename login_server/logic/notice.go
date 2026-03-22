package logic

import (
	"encoding/json"
	"xfx/core/define"
	"xfx/core/model"
	dto2 "xfx/login_server/dto"

	"slices"
	"xfx/login_server/internal/middleware"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

// GetNotices 获取公告
func GetNotices(c *gin.Context) {
	var req dto2.NoticeListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RetGame(c, dto2.ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	log.Debug(" 获取公告： %v", req)
	logicServerId := req.ServerID
	if req.ServerID > 0 {
		serverItem := new(model.ServerItem)
		has, getErr := AccountEngine.Table(define.GameServerTable).Where("id = ?", req.ServerID).Get(serverItem)
		if getErr != nil {
			log.Error("获取区服映射错误: %s", getErr)
			middleware.RetGame(c, dto2.ERR_DB, "server map err")
			return
		}
		if has && serverItem.LogicServerId > 0 {
			logicServerId = int(serverItem.LogicServerId)
		}
	}

	var items []model.NoticeItem
	err := AccountEngine.Table(define.NoticeTable).
		Where("(channel = ? OR channel = 0) AND (server_id = ? OR server_id = 0)",
			req.Channel, logicServerId).
		Find(&items)
	if err != nil {
		log.Error("获取公告错误: %s", err)
		middleware.RetGame(c, dto2.ERR_DB, "params err1")
		return
	}

	slices.SortFunc(items, func(a, b model.NoticeItem) int {
		return int(a.EffectTime - b.EffectTime)
	})

	log.Debug("公告列表： %v", items)

	//获取最新的5条数据
	latestCount := 5
	if len(items) > latestCount {
		items = items[:latestCount]
	}

	js, _ := json.Marshal(items)
	resp := dto2.NoticeListResponse{Notices: items}
	middleware.RetGameData(c, dto2.SUCCESS, "success", resp, map[string]any{"data": js})
}
