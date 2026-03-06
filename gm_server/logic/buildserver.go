package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"os/exec"
	"path/filepath"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
)

type GitBuildResult struct {
	Success bool
	Message string
}

// 编译服务器
func GmBuildServer(c *gin.Context) {
	p := new(gm_model.ServerItem)
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

	log.Debug("请求编译服务器代码:%v", serverId)

	//if len(os.Args) != 3 {
	//	fmt.Println("用法: go run main.go [目标目录] [Git仓库URL]")
	//	os.Exit(1)
	//}
	//
	//targetDir := os.Args[1]
	//repoURL := os.Args[2]

	//先拉取server
	targetDir := "/usr/local/games/xiyou/server/server"
	repoURL := "ssh://git@47.121.121.101:2222/root/server.git"

	result := GitPullOrClone(targetDir, repoURL)

	if result.Success {
		fmt.Println("✅", result.Message)
		fmt.Println("──────────────────────────")
		fmt.Println("目录:", result.Directory)
		fmt.Println("分支:", result.Branch)
		fmt.Println("最新提交:", result.LastCommit)
		fmt.Println("──────────────────────────")
		//httpRetGame(c, SUCCESS, "success", map[string]any{
		//	"msg": result.Message,
		//})

		//拉取proto
		targetDir = "/usr/local/games/xiyou/server/server/proto"
		repoURL = "ssh://git@47.121.121.101:2222/root/server_proto.git"

		sresult := GitPullOrClone(targetDir, repoURL)
		if sresult.Success {
			fmt.Println("✅", sresult.Message)
			fmt.Println("──────────────────────────")
			fmt.Println("目录:", sresult.Directory)
			fmt.Println("分支:", sresult.Branch)
			fmt.Println("最新提交:", sresult.LastCommit)
			fmt.Println("──────────────────────────")

			//进入main_server 编译代码
			sourceDir := "/usr/local/games/xiyou/server/server/main_server/run"
			outputName := "run"
			targetDir := "/usr/local/games/xiyou/server"
			res := buildserver(sourceDir, targetDir, outputName)
			if res.Success {
				fmt.Printf("✅ 构建成功: %s\n", targetDir)
				fmt.Printf(" - 源目录: %s\n", sourceDir)
				fmt.Printf(" - 目标位置: %s\n", targetDir)
				httpRetGame(c, SUCCESS, "success", map[string]any{
					"msg": res.Message,
				})
			} else {
				fmt.Println("❌ 编译错误:", res.Message)
				httpRetGame(c, ERR_GIT_ERROR, "faild", map[string]any{
					"msg": res.Message,
				})
			}
		} else {
			fmt.Println("❌ 错误:", sresult.Message)
			httpRetGame(c, ERR_GIT_ERROR, "faild", map[string]any{
				"msg": sresult.Message,
			})
		}

	} else {
		fmt.Println("❌ 错误:", result.Message)
		httpRetGame(c, ERR_GIT_ERROR, "faild", map[string]any{
			"msg": result.Message,
		})
	}
}

// 编译服务器
func GmGameBuildServer(c *gin.Context) {
	p := new(gm_model.ServerItem)
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

	log.Debug("请求编译服务器代码:%v", serverId)

	//if len(os.Args) != 3 {
	//	fmt.Println("用法: go run main.go [目标目录] [Git仓库URL]")
	//	os.Exit(1)
	//}
	//
	//targetDir := os.Args[1]
	//repoURL := os.Args[2]

	//先拉取server
	targetDir := "/usr/local/games/xiyou/server/server"
	repoURL := "ssh://git@47.121.121.101:2222/root/server.git"

	result := GitPullOrClone(targetDir, repoURL)

	if result.Success {
		fmt.Println("✅", result.Message)
		fmt.Println("──────────────────────────")
		fmt.Println("目录:", result.Directory)
		fmt.Println("分支:", result.Branch)
		fmt.Println("最新提交:", result.LastCommit)
		fmt.Println("──────────────────────────")
		//httpRetGame(c, SUCCESS, "success", map[string]any{
		//	"msg": result.Message,
		//})

		//拉取proto
		targetDir = "/usr/local/games/xiyou/server/server/proto"
		repoURL = "ssh://git@47.121.121.101:2222/root/server_proto.git"

		sresult := GitPullOrClone(targetDir, repoURL)
		if sresult.Success {
			fmt.Println("✅", sresult.Message)
			fmt.Println("──────────────────────────")
			fmt.Println("目录:", sresult.Directory)
			fmt.Println("分支:", sresult.Branch)
			fmt.Println("最新提交:", sresult.LastCommit)
			fmt.Println("──────────────────────────")

			//进入main_server 编译代码
			sourceDir := "/usr/local/games/xiyou/server/server/game_server/run"
			outputName := "game"
			targetDir := "/usr/local/games/xiyou/server"
			res := buildserver(sourceDir, targetDir, outputName)
			if res.Success {
				fmt.Printf("✅ 构建成功: %s\n", targetDir)
				fmt.Printf(" - 源目录: %s\n", sourceDir)
				fmt.Printf(" - 目标位置: %s\n", targetDir)
				httpRetGame(c, SUCCESS, "success", map[string]any{
					"msg": res.Message,
				})
			} else {
				fmt.Println("❌ 编译错误:", res.Message)
				httpRetGame(c, ERR_GIT_ERROR, "faild", map[string]any{
					"msg": res.Message,
				})
			}
		} else {
			fmt.Println("❌ 错误:", sresult.Message)
			httpRetGame(c, ERR_GIT_ERROR, "faild", map[string]any{
				"msg": sresult.Message,
			})
		}

	} else {
		fmt.Println("❌ 错误:", result.Message)
		httpRetGame(c, ERR_GIT_ERROR, "faild", map[string]any{
			"msg": result.Message,
		})
	}
}

func buildserver(sourceDir, targetDir, outputName string) GitBuildResult {
	result := GitBuildResult{}
	// 1. 进入源码目录
	if err := os.Chdir(sourceDir); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("无法进入源码目录 %s", sourceDir)
		return result
	}

	// 2. 执行 go build
	buildPath := filepath.Join(".", outputName)
	if err := buildGoBinary(outputName); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("编译失败 %s", err)
		return result
	}

	// 3. 检查编译结果
	if _, err := os.Stat(buildPath); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("编译产物不存在 %s", err)
		return result
	}

	// 4. 复制到目标目录
	targetPath := filepath.Join(targetDir, outputName)
	if err := copyFile(buildPath, targetPath); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("复制文件失败 %s", err)
		return result
	}

	// 5. 设置可执行权限
	if err := os.Chmod(targetPath, 0755); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("设置可执行权限失败 %s", err)
		return result
	}
	result.Success = true
	result.Message = fmt.Sprintf("编译成功")
	return result
}

func copyFile(src, dst string) error {
	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

func buildGoBinary(outputName string) error {
	cmd := exec.Command("go", "build", "-o", outputName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
