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
	var baseURL string
	if serverId <= 0 {
		baseURL = defaultMainServerBaseURL()
	} else {
		var item model.ServerItem
		ok, err := db.AccountDb.Table(define.GameServerTable).Where("id = ?", serverId).Get(&item)
		if err != nil || !ok {
			baseURL = defaultMainServerBaseURL()
		} else if strings.TrimSpace(item.MainServerHttpUrl) == "" {
			baseURL = defaultMainServerBaseURL()
		} else {
			baseURL = strings.TrimRight(strings.TrimSpace(item.MainServerHttpUrl), "/")
		}
	}
	// 确保 URL 有 http:// 或 https:// 前缀
	if baseURL != "" && !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return baseURL
}

func defaultMainServerBaseURL() string {
	u := strings.TrimSpace(conf.Server.MainServerHttpUrl)
	if u == "" {
		return "http://127.0.0.1:9505"
	}
	u = strings.TrimRight(u, "/")
	// 确保 URL 有 http:// 或 https:// 前缀
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "http://" + u
	}
	return u
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
