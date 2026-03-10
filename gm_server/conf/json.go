package conf

import (
	"encoding/json"
	"os"
	"xfx/pkg/env"

	"github.com/name5566/leaf/log"
)

var Server struct {
	Log           *env.Log
	RedisAddr     string
	RedisPassword string
	RedisDbNum    int
	HttpPort      string
	AccountAddr   string
	// MainServerHttpUrl main_server 的 HTTP 基础地址，例如 http://127.0.0.1:9505
	MainServerHttpUrl string
}

func init() {
	data, err := os.ReadFile("./conf/gm_server.json")
	if err != nil {
		log.Fatal("read conf file error:", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("read conf file unmarshal error:", err)
	}
}
