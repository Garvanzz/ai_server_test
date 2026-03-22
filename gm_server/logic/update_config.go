package logic

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"xfx/pkg/log"
)

// 配置仓库拉取目录与地址（与大厅服/游戏服共用同一配置仓库时可相同）
const (
	configTargetDir = "/usr/local/games/xiyou/server/json"
	configRepoURL   = "ssh://git@47.121.121.101:2222/root/server_config_json.git"
)

// doConfigUpdate 拉取配置仓库到目标目录（注意 GitPullOrClone 会 Chdir，并发调用需串行）
func doConfigUpdate(targetDir, repoURL string) GitResult {
	return GitPullOrClone(targetDir, repoURL)
}

func respondConfigUpdate(c *gin.Context, result GitResult) {
	if result.Success {
		log.Debug("config update ok: %s, dir=%s branch=%s", result.Message, result.Directory, result.Branch)
		HTTPRetGameData(c, SUCCESS, "success", map[string]any{"msg": result.Message}, map[string]any{"msg": result.Message})
	} else {
		log.Debug("config update failed: %s", result.Message)
		HTTPRetGameData(c, ERR_GIT_ERROR, "failed", map[string]any{"msg": result.Message}, map[string]any{"msg": result.Message})
	}
}

// GmUpdateConfig 拉取配置仓库并更新大厅服配置
func GmUpdateConfig(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req map[string]interface{}
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if sid, ok := req["serverId"].(float64); ok {
		log.Debug("GmUpdateConfig serverId: %v", sid)
	}
	result := doConfigUpdate(configTargetDir, configRepoURL)
	respondConfigUpdate(c, result)
}

// GmGameUpdateConfig 拉取配置仓库并更新游戏服配置（当前与大厅服共用同一仓库路径，可后续拆分配置）
func GmGameUpdateConfig(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req map[string]interface{}
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if sid, ok := req["serverId"].(float64); ok {
		log.Debug("GmGameUpdateConfig serverId: %v", sid)
	}
	result := doConfigUpdate(configTargetDir, configRepoURL)
	respondConfigUpdate(c, result)
}
