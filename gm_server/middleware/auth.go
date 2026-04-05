package middleware

import (
	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/gm_server/logic"
	"xfx/pkg/log"
)

// GmAuth GM 鉴权中间件：从请求头 xiaoxiaoxiyou 读取 token，校验通过后写入当前用户到 context
func GmAuth(c *gin.Context) {
	token := c.GetHeader(logic.TokenHeaderName)
	if token == "" {
		log.Debug("gm auth: token empty")
		logic.HTTPRetGame(c, logic.ERR_ACCOUNT_NOT_FOUND, "accessToken required")
		c.Abort()
		return
	}

	player := new(dto.GmAccount)
	player.Token = token
	has, err := db.AccountDb.Table(define.AdminTable).Where("token = ?", token).Get(player)
	if err != nil {
		log.Error("gm auth db err: %v", err)
		logic.HTTPRetGame(c, logic.ERR_DB, "db err")
		c.Abort()
		return
	}
	if !has || player.Token != token {
		log.Debug("gm auth: token invalid or expired")
		logic.HTTPRetGame(c, logic.ERR_ACCOUNT_NOT_FOUND, "accessToken invalid or expired")
		c.Abort()
		return
	}

	c.Set(logic.ContextKeyGmUser, player)
	c.Next()
}

// RequirePermission 返回一个 Gin 中间件，校验当前用户的权限等级是否在 allowed 列表中。
// 必须在 GmAuth 之后使用（依赖 ContextKeyGmUser）。
func RequirePermission(allowed ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(logic.ContextKeyGmUser)
		if !exists {
			logic.HTTPRetGame(c, logic.ERR_ACCOUNT_NOT_FOUND, "user not found in context")
			c.Abort()
			return
		}
		user, ok := val.(*dto.GmAccount)
		if !ok {
			logic.HTTPRetGame(c, logic.ERR_SERVER_INTERNAL, "invalid user in context")
			c.Abort()
			return
		}
		for _, perm := range allowed {
			if user.Permission == perm {
				c.Next()
				return
			}
		}
		logic.HTTPRetGame(c, logic.ERR_PERMISSION_DENIED, "permission denied")
		c.Abort()
	}
}
