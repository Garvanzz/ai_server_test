package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RetGame 游戏接口统一返回：HTTP 200，body 含 errcode、errmsg 及可选 data 字段
func RetGame(c *gin.Context, code int, message string, data ...map[string]interface{}) {
	ret := gin.H{"errcode": code, "errmsg": message}
	if len(data) > 0 {
		for k, v := range data[0] {
			ret[k] = v
		}
	}
	c.JSON(http.StatusOK, ret)
}

// RetGameData 游戏接口统一返回 data 载荷；legacy 字段仅用于兼容旧调用方。
func RetGameData(c *gin.Context, code int, message string, payload any, legacy ...map[string]interface{}) {
	ret := gin.H{"errcode": code, "errmsg": message, "data": payload}
	if len(legacy) > 0 {
		for k, v := range legacy[0] {
			ret[k] = v
		}
	}
	c.JSON(http.StatusOK, ret)
}

// Ret 通用 HTTP 返回：使用传入的 HTTP 状态码
func Ret(c *gin.Context, statusCode int, message string, data ...map[string]interface{}) {
	if len(data) > 0 {
		c.JSON(statusCode, gin.H{"code": statusCode, "message": message, "data": data[0]})
	} else {
		c.JSON(statusCode, gin.H{"code": statusCode, "message": message})
	}
}
