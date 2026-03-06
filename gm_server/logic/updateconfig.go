package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/pkg/log"
)

type GitResult struct {
	Success    bool
	Message    string
	Directory  string
	Branch     string
	LastCommit string
}

// 更新配置表
func GmUpdateConfig(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	rawData, _ := c.GetRawData()
	var _result map[string]interface{}
	if err := json.Unmarshal(rawData, &_result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := _result["serverId"].(float64)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	log.Debug("请求更新配置表数据:%v", serverId)

	//if len(os.Args) != 3 {
	//	fmt.Println("用法: go run main.go [目标目录] [Git仓库URL]")
	//	os.Exit(1)
	//}
	//
	//targetDir := os.Args[1]
	//repoURL := os.Args[2]
	targetDir := "/usr/local/games/xiyou/server/json"
	repoURL := "ssh://git@47.121.121.101:2222/root/server_config_json.git"

	result := GitPullOrClone(targetDir, repoURL)

	if result.Success {
		fmt.Println("✅", result.Message)
		fmt.Println("──────────────────────────")
		fmt.Println("目录:", result.Directory)
		fmt.Println("分支:", result.Branch)
		fmt.Println("最新提交:", result.LastCommit)
		fmt.Println("──────────────────────────")
		httpRetGame(c, SUCCESS, "success", map[string]any{
			"msg": result.Message,
		})
	} else {
		fmt.Println("❌ 错误:", result.Message)
		httpRetGame(c, ERR_GIT_ERROR, "faild", map[string]any{
			"msg": result.Message,
		})
	}
}

// 更新配置表
func GmGameUpdateConfig(c *gin.Context) {
	p := new(model.ServerItem)
	if has, _ := db.AccountDb.IsTableExist(define.ServerGroup); !has {
		// 同步结构体与数据库表
		err := db.AccountDb.Sync2(p)
		if err != nil {
			log.Error("Failed to sync database: %v", err)
			httpRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	rawData, _ := c.GetRawData()
	var _result map[string]interface{}
	if err := json.Unmarshal(rawData, &_result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := _result["serverId"].(float64)
	if !ok {
		log.Error("GmStartServer find serverName err")
		httpRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	log.Debug("请求更新配置表数据:%v", serverId)

	//if len(os.Args) != 3 {
	//	fmt.Println("用法: go run main.go [目标目录] [Git仓库URL]")
	//	os.Exit(1)
	//}
	//
	//targetDir := os.Args[1]
	//repoURL := os.Args[2]
	targetDir := "/usr/local/games/xiyou/server/json"
	repoURL := "ssh://git@47.121.121.101:2222/root/server_config_json.git"

	result := GitPullOrClone(targetDir, repoURL)

	if result.Success {
		fmt.Println("✅", result.Message)
		fmt.Println("──────────────────────────")
		fmt.Println("目录:", result.Directory)
		fmt.Println("分支:", result.Branch)
		fmt.Println("最新提交:", result.LastCommit)
		fmt.Println("──────────────────────────")
		httpRetGame(c, SUCCESS, "success", map[string]any{
			"msg": result.Message,
		})
	} else {
		fmt.Println("❌ 错误:", result.Message)
		httpRetGame(c, ERR_GIT_ERROR, "faild", map[string]any{
			"msg": result.Message,
		})
	}
}
