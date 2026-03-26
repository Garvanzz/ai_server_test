package logic

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"xfx/core/model"
	"xfx/pkg/log"
)

// GmKickPlayer 踢玩家下线（经 main_server 下发 SysKick 指令）
func GmKickPlayer(c *gin.Context) {
	var req struct {
		ServerId int    `json:"ServerId"`
		Uid      string `json:"Uid"`
		Reason   string `json:"Reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.ServerId <= 0 || strings.TrimSpace(req.Uid) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId and uid required")
		return
	}

	log.Debug("踢玩家下线 serverId=%d uid=%s reason=%s", req.ServerId, req.Uid, req.Reason)
	playerId, err := getPlayerIdByServerAndUid(req.ServerId, req.Uid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if playerId == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "account not found")
		return
	}

	body, _ := json.Marshal(model.GMKickReq{PlayerId: playerId, Reason: req.Reason})
	if err, _ := HttpRequestToServer(req.ServerId, body, "/gm/kick"); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}
