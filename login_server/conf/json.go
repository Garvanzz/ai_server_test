package conf

import (
	"encoding/json"
	"os"
	"xfx/pkg/env"
	"xfx/pkg/log"
)

var Server struct {
	Log           *env.Log
	RedisAddr     string
	RedisPassword string
	RedisDbNum    int
	HttpPort      string
	AccountAddr   string
	AesKey        string `json:"aesKey"` // AES 解密密钥，32 字节；未配置或为空则不启用解密中间件
}

func init() {
	data, err := os.ReadFile("./conf/loginserver.json")
	if err != nil {
		log.Fatal("read conf file error: %v", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("read conf file unmarshal error: %v", err)
	}
}
