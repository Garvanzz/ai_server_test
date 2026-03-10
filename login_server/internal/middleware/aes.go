package middleware

import (
	"bytes"
	"encoding/hex"
	"io"
	"xfx/login_server/dto"
	"xfx/pkg/log"
	"xfx/pkg/utils/crypto"

	"github.com/gin-gonic/gin"
)

// AesDecryptGame 游戏请求体 AES 解密中间件（原始 body 为密文，解密后替换 body）
func AesDecryptGame(key []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, err := c.GetRawData()
		if err != nil {
			RetGame(c, dto.ERR_SERVER_INTERNAL, "ERR_SERVER_INTERNAL")
			c.Abort()
			return
		}
		if len(payload) == 0 {
			RetGame(c, dto.ERR_ACCOUNT_PARAMS_ERROR, "params error")
			c.Abort()
			return
		}
		data, err := crypto.AesPkcs7Decrypt(payload, key)
		if err != nil {
			log.Error("login server aes decrypt error: %v", err)
			RetGame(c, dto.ERR_ACCOUNT_PARAMS_ERROR, "params error")
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(data))
	}
}

// AesDecryptHomeWeb 官网请求体解密中间件（body 先 hex 解码再 AES 解密）
func AesDecryptHomeWeb(key []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, err := c.GetRawData()
		if err != nil {
			Ret(c, 500, "ERR_SERVER_INTERNAL")
			c.Abort()
			return
		}
		if len(payload) == 0 {
			Ret(c, 400, "params error")
			c.Abort()
			return
		}
		raw, err := hex.DecodeString(string(payload))
		if err != nil {
			log.Debug("AesDecryptHomeWeb hex error: %v", err)
			Ret(c, 400, "params error")
			c.Abort()
			return
		}
		data, err := crypto.AesPkcs7Decrypt(raw, key)
		if err != nil {
			log.Debug("AesDecryptHomeWeb aes error: %v", err)
			Ret(c, 400, "params error")
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(data))
	}
}
