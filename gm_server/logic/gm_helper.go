package logic

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/conf"
	"xfx/gm_server/db"
)

// getMainServerURL 根据区服 id 从 game_server 表取 main_server HTTP 地址；未配置或 serverId<=0 时用 conf 默认
func getMainServerURL(serverId int) string {
	if serverId <= 0 {
		return defaultMainServerBaseURL()
	}
	var item model.ServerItem
	ok, _ := db.AccountDb.Table(define.GameServerTable).Where("id = ?", serverId).Get(&item)
	if !ok || strings.TrimSpace(item.MainServerHttpUrl) == "" {
		return defaultMainServerBaseURL()
	}
	return strings.TrimRight(strings.TrimSpace(item.MainServerHttpUrl), "/")
}

func defaultMainServerBaseURL() string {
	if u := strings.TrimSpace(conf.Server.MainServerHttpUrl); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "http://127.0.0.1:9505"
}

// getPlayerIdByServerAndUid 根据 server_id、uid 查 Account 得到 RedisId（player_id），供转发 main_server 用
// serverId 须 >0，uid 须非空，否则返回 0,nil
func getPlayerIdByServerAndUid(serverId int, uid string) (int64, error) {
	if serverId <= 0 || strings.TrimSpace(uid) == "" {
		return 0, nil
	}
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
