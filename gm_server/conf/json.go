package conf

import (
	"encoding/json"
	"os"
	"xfx/pkg/env"

	"github.com/name5566/leaf/log"
)

var Server struct {
	Log               *env.Log
	HttpPort          string
	AccountAddr       string
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
