package logic

import (
	"encoding/json"
	"xfx/core/define"
	"xfx/core/model"
	dto2 "xfx/login_server/dto"
	"xfx/pkg/utils"

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
		has, getErr := AccountEngine.Table(define.GameServerTable).Cols("logic_server_id").Where("id = ?", req.ServerID).Get(serverItem)
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
		return int(b.EffectTime - a.EffectTime)
	})

	log.Debug("公告列表： %v", items)

	//去除过期的
	temp := make([]model.NoticeItem, 0)
	for k := 0; k < len(items); k++ {
		if items[k].ExpireTime > 0 && items[k].ExpireTime <= utils.Now().Unix() {
			continue
		}

		temp = append(temp, items[k])
		if len(temp) >= 5 {
			break
		}
	}

	js, _ := json.Marshal(items)
	resp := dto2.NoticeListResponse{Notices: items}
	middleware.RetGameData(c, dto2.SUCCESS, "success", resp, map[string]any{"data": js})
}
