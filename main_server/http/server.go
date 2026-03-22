package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// http返回 游戏专用 码只返回200 不然不返
func (m *HttpModule) httpRetGame(c *gin.Context, code int, message string, data ...map[string]any) {

	ret := gin.H{
		"errcode": code,
		"errmsg":  message,
	}
	if data != nil && len(data) > 0 {
		//只认第一个data 其他直接抛 用...主要是为了省参
		for k, v := range data[0] {
			ret[k] = v
		}
	}

	c.JSON(http.StatusOK, ret)
}

func (m *HttpModule) httpRetGameData(c *gin.Context, code int, message string, payload any, legacy ...map[string]any) {
	ret := gin.H{
		"errcode": code,
		"errmsg":  message,
		"data":    payload,
	}
	if len(legacy) > 0 {
		for k, v := range legacy[0] {
			ret[k] = v
		}
	}
	c.JSON(http.StatusOK, ret)
}

// http返回
func (m *HttpModule) httpRet(c *gin.Context, code int, message string, data ...map[string]any) {
	if data != nil && len(data) > 0 {
		//这里只走data[0] ...只是为了省参用 后续多参默认无效
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
			"data":    data[0],
		})
	} else {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	}

}
