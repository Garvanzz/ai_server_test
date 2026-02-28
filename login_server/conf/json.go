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
}

func init() {
	data, err := os.ReadFile("./conf/loginserver.json")
	if err != nil {
		log.Fatal("read conf file error:", err)
	}
	err = json.Unmarshal(data, &Server)
	if err != nil {
		log.Fatal("read conf file unmarshal error:", err)
	}
}
