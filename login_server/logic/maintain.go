package logic

import (
	"bytes"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"xfx/pkg/log"
)

const (
	SUCCESS                              = 0
	ERR_SERVER_INTERNAL                  = 1    // 服务器内部错误
	ERR_PAY_ORDER_NOT_FOUND              = 2    // 订单不存在
	ERR_PAY_SIGN                         = 3    // 签名不正确
	ERR_INVITE_CODE_FAIL                 = 29   // 邀请码验证不通过
	ERR_DB                               = 1801 // 数据库错误
	ERR_ACCOUNT_EXISTS                   = 1802 // 账号已存在
	ERR_ACCOUNT_PASSWORD_FAILED          = 1803 // 账号密码错误
	ERR_ACCOUNT_TYPE_UNKNOWN             = 1804 // 账号类型错误
	ERR_ACCOUNT_NOT_FOUND                = 1805 // 账号不存在
	ERR_ACCOUNT_VERIFY_CODE_INCORRECT    = 1806 // 验证码不正确
	ERR_ACCOUNT_GET_VERIFY_CODE_FAILED   = 1807 // 获取验证码失败
	ERR_ACCOUNT_REGISTER_CLOSED          = 1808 // 注册服务关闭
	ERR_ACCOUNT_LOGIN_SERVER_MAINTAIN    = 1809 // 服务器维护 其实只有白名单账号可以进
	ERR_ACCOUNT_BANNED                   = 1810 // 账号被ban中
	ERR_ACCOUNT_PARAMS_ERROR             = 1811 // 参数错误
	ERR_ACCOUNT_CLIENT_VERSION_UNMATCHED = 1812 // 客户端版本不匹配
	ERR_ACCOUNT_SDK_TOKEN_AUTH_FAILED    = 1813 // 登录SDK Token效验失败
	ERR_ACCOUNT_SDK_TOKEN_EXPIRED        = 1814 // 登录SDK Token过期
	ERR_ACCOUNT_HAS_NO_NFT_HERO          = 1815 // 帐号没有nft英雄
	ERR_ACCOUNT_FORCED_OFFLINE           = 1816 // 帐号强制下线中
)

// aes中间件解密 游戏专用
func AesDecryptMiddleFuncForGame(c *gin.Context) {

	//解析原始数据
	payload, err := c.GetRawData()
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, "ERR_SERVER_INTERNAL")
		c.Abort()
		return
	}

	//判长
	if payload == nil || len(payload) <= 0 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params error")
		c.Abort()
		return
	}

	data, err := AesPkcs7Decrypt(payload, []byte(Key))
	if err != nil {
		log.Error("login server EcbDecrypt error : %v", err)
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params error")
		c.Abort()
		return
	}

	//解码完成 其他交由具体逻辑去处理
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

}

// aes中间件解密 官网主页专用
func AesDecryptMiddleFuncForHomeWeb(c *gin.Context) {

	//解析原始数据
	payload, err := c.GetRawData()
	if err != nil {
		httpRet(c, http.StatusInternalServerError, "ERR_SERVER_INTERNAL")
		c.Abort()
		return
	}

	//判长
	if payload == nil || len(payload) <= 0 {
		log.Debug("AesDecryptMiddleFuncForHomeWeb error : payload nil or payload <= 0")
		httpRet(c, http.StatusBadRequest, "params error")
		c.Abort()
		return
	}

	//先解hex
	raw, err := hex.DecodeString(string(payload))
	if err != nil {
		log.Debug("AesDecryptMiddleFuncForHomeWeb hex error : %s", err.Error())
		httpRet(c, http.StatusBadRequest, "params error")
		c.Abort()
		return
	}

	//再解aes
	data, err := AesPkcs7Decrypt(raw, []byte(Key))
	if err != nil {
		log.Debug("AesDecryptMiddleFuncForHomeWeb AesPkcs7Decrypt error : %s", err.Error())
		httpRet(c, http.StatusBadRequest, "params error")
		c.Abort()
		return
	}

	//解码完成 其他交由具体逻辑去处理
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

}

// http返回 游戏专用 码只返回200 不然不返
func httpRetGame(c *gin.Context, code int, message string, data ...map[string]interface{}) {

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

// http返回
func httpRet(c *gin.Context, code int, message string, data ...map[string]interface{}) {
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
