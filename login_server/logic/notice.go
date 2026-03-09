package logic

import (
	"encoding/json"

	"slices"
	"xfx/login_server/define"
	"xfx/login_server/internal/middleware"
	"xfx/login_server/model/dto"
	"xfx/login_server/model/entity"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

// GetNotices 获取公告
func GetNotices(c *gin.Context) {
	var req dto.ReqNoticeList
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RetGame(c, define.ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	log.Debug(" 获取公告： %v", req)

	var items []entity.NoticeItem
	err := AccountEngine.Table(define.TableNotice).
		Where("(channel = ? OR channel = 0) AND (server_id = ? OR server_id = 0)",
			req.Channel, req.ServerId).
		Find(&items)
	if err != nil {
		log.Error("获取公告错误: %s", err)
		middleware.RetGame(c, define.ERR_DB, "params err1")
		return
	}

	slices.SortFunc(items, func(a, b entity.NoticeItem) int {
		return int(a.EffectTime - b.EffectTime)
	})

	log.Debug("公告列表： %v", items)

	//获取最新的5条数据
	latestCount := 5
	if len(items) > latestCount {
		items = items[:latestCount]
	}

	js, _ := json.Marshal(items)
	middleware.RetGame(c, define.SUCCESS, "success",
		map[string]any{
			"data": js,
		})
}
