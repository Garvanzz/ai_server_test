package logic

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"xfx/pkg/log"
)

type GitBuildResult struct {
	Success bool
	Message string
}

// 大厅服/游戏服拉取与编译路径（可按部署修改）
const (
	serverRepoDir    = "/usr/local/games/xiyou/server/server"
	serverRepoURL    = "ssh://git@47.121.121.101:2222/root/server.git"
	protoSubDir      = "/usr/local/games/xiyou/server/server/proto"
	protoRepoURL     = "ssh://git@47.121.121.101:2222/root/server_proto.git"
	mainServerRunDir = "/usr/local/games/xiyou/server/server/main_server/run"
	gameServerRunDir = "/usr/local/games/xiyou/server/server/game_server/run"
	buildOutputDir   = "/usr/local/games/xiyou/server"
)

// GmBuildServer 拉取 server + proto 仓库并编译大厅服（main_server/run -> run）
func GmBuildServer(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req map[string]interface{}
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if sid, ok := req["serverId"].(float64); ok {
		log.Debug("GmBuildServer serverId: %v", sid)
	}

	result := GitPullOrClone(serverRepoDir, serverRepoURL)
	if !result.Success {
		HTTPRetGame(c, ERR_GIT_ERROR, "failed", map[string]any{"msg": result.Message})
		return
	}
	log.Debug("server pull ok: %s", result.Message)

	protoResult := GitPullOrClone(protoSubDir, protoRepoURL)
	if !protoResult.Success {
		HTTPRetGame(c, ERR_GIT_ERROR, "failed", map[string]any{"msg": protoResult.Message})
		return
	}
	log.Debug("proto pull ok: %s", protoResult.Message)

	res := buildServer(mainServerRunDir, buildOutputDir, "run")
	if res.Success {
		HTTPRetGame(c, SUCCESS, "success", map[string]any{"msg": res.Message})
	} else {
		HTTPRetGame(c, ERR_GIT_ERROR, "failed", map[string]any{"msg": res.Message})
	}
}

// GmGameBuildServer 拉取 server + proto 仓库并编译游戏服（game_server/run -> game）
func GmGameBuildServer(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req map[string]interface{}
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if sid, ok := req["serverId"].(float64); ok {
		log.Debug("GmGameBuildServer serverId: %v", sid)
	}

	result := GitPullOrClone(serverRepoDir, serverRepoURL)
	if !result.Success {
		HTTPRetGame(c, ERR_GIT_ERROR, "failed", map[string]any{"msg": result.Message})
		return
	}
	log.Debug("server pull ok: %s", result.Message)

	protoResult := GitPullOrClone(protoSubDir, protoRepoURL)
	if !protoResult.Success {
		HTTPRetGame(c, ERR_GIT_ERROR, "failed", map[string]any{"msg": protoResult.Message})
		return
	}
	log.Debug("proto pull ok: %s", protoResult.Message)

	res := buildServer(gameServerRunDir, buildOutputDir, "game")
	if res.Success {
		HTTPRetGame(c, SUCCESS, "success", map[string]any{"msg": res.Message})
	} else {
		HTTPRetGame(c, ERR_GIT_ERROR, "failed", map[string]any{"msg": res.Message})
	}
}

// buildServer 执行拉取代码并编译，返回构建结果
func buildServer(sourceDir, targetDir, outputName string) GitBuildResult {
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
