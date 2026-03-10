package logic

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
)

// getPlayerIdByServerAndUid 根据 server_id、uid 查 Account 得到 RedisId（player_id），供转发 main_server 用
func getPlayerIdByServerAndUid(serverId int, uid string) (int64, error) {
	pl := new(model.Account)
	has, err := db.AccountDb.Table(define.AccountTable).Where("server_id = ? AND uid = ?", serverId, uid).Get(pl)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, nil // 未找到视为 0，调用方用 ERR_ACCOUNT_NOT_FOUND
	}
	return pl.RedisId, nil
}

// forwardMainServerResponse 将 main_server 返回的 body 原样写回客户端（格式已为 {errcode, errmsg, data?, totalCount?}）
func forwardMainServerResponse(c *gin.Context, body string) {
	c.Data(http.StatusOK, "application/json", []byte(body))
}
