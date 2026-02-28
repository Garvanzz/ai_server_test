package logic

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"slices"
	"xfx/login_server/define"
	"xfx/login_server/model"
	"xfx/pkg/log"
)

// 获取公告
func GetNotices(c *gin.Context) {
	var reqnotice model.ReqNoticeList
	if err := c.ShouldBindJSON(&reqnotice); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	log.Debug(" 获取公告： %v", reqnotice)

	//取库
	var items []model.NoticeItem
	err := AccountEngine.Table(define.Notice).
		Where("(channel = ? OR channel = 0) AND (server_id = ? OR server_id = 0)",
			reqnotice.Channel, reqnotice.ServerId).
		Find(&items)
	if err != nil {
		log.Error("获取公告错误: %s", err)
		httpRetGame(c, ERR_DB, "params err1")
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
	httpRetGame(c, SUCCESS, "success",
		map[string]any{
			"data": js,
		})
}
