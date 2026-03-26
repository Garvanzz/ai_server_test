package logic

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/pkg/log"
)

// 发送公告
func GmSendNotice(c *gin.Context) {
	var Info model.NoticeItem
	if err := c.ShouldBindJSON(&Info); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err: "+err.Error())
		return
	}

	log.Debug("Info %v", Info)

	if len(Info.Content) <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "content required")
		return
	}

	// 使用正确的结构体，包含所有字段
	item := model.NoticeItem{
		Channel:    Info.Channel,
		ServerId:   Info.ServerId,
		Title:      Info.Title,
		Content:    Info.Content,
		ExpireTime: Info.ExpireTime,
		EffectTime: time.Now().Unix(),
	}
	_, err := db.AccountDb.Table(define.NoticeTable).Insert(&item)
	if err != nil {
		log.Error("GmSendNotice insert err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	// 若指定了区服，则转发到该区服 main_server 实时推送给在线玩家
	if Info.ServerId > 0 {
		js, _ := json.Marshal(item)
		if err, _ := HttpRequestToServer(int(Info.ServerId), js, "/gm/notice"); err != nil {
			log.Debug("GmSendNotice forward main_server err: %v", err)
			// 已入库，仅记录；仍返回成功
		}
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// 发送跑马灯
func GmSendHorse(c *gin.Context) {
	var Info model.HorseItem
	if err := c.ShouldBindJSON(&Info); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err: "+err.Error())
		return
	}

	log.Debug("horse Info %v", Info)

	if len(Info.Content) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "content required")
		return
	}

	if Info.ServerId <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId required")
		return
	}

	js, _ := json.Marshal(Info)
	// 按区服转发到对应 main_server
	err, _ := HttpRequestToServer(int(Info.ServerId), js, "/gm/horse")
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmGetNoticeList 查询公告列表（从 DB 读取，支持按区服过滤）
func GmGetNoticeList(c *gin.Context) {
	var req struct {
		ServerId int `json:"serverId"`
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.PageSize <= 0 {
		req.PageSize = 50
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.PageSize

	sess := db.AccountDb.Table(define.NoticeTable)
	if req.ServerId > 0 {
		sess = sess.Where("server_id = ?", req.ServerId)
	}
	total, err := sess.Count(new(model.NoticeItem))
	if err != nil {
		log.Error("GmGetNoticeList count err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	list := make([]model.NoticeItem, 0)
	sess2 := db.AccountDb.Table(define.NoticeTable)
	if req.ServerId > 0 {
		sess2 = sess2.Where("server_id = ?", req.ServerId)
	}
	if err := sess2.OrderBy("id DESC").Limit(req.PageSize, offset).Find(&list); err != nil {
		log.Error("GmGetNoticeList find err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGameData(c, SUCCESS, "success", list, map[string]any{"totalCount": total})
}
