package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

func serverProcessTypeName(t int) string {
	switch t {
	case dto.ServerProcessTypeLogin:
		return "登录服"
	case dto.ServerProcessTypeMain:
		return "大厅服"
	case dto.ServerProcessTypeGame:
		return "游戏服"
	default:
		return "未知"
	}
}

func processRunState(p model.ServerProcess) string {
	if processHealthCheck(p) {
		return "运行中"
	}
	return "离线"
}

func processHealthCheck(p model.ServerProcess) bool {
	// 优先 HTTP 健康检查
	if strings.TrimSpace(p.HttpHealthUrl) != "" {
		rawURL := strings.TrimSpace(p.HttpHealthUrl)
		if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
			rawURL = "http://" + rawURL
		}
		parsed, err := url.Parse(rawURL)
		if err == nil && strings.TrimSpace(parsed.Host) != "" {
			conn, err := net.DialTimeout("tcp", parsed.Host, 800*time.Millisecond)
			if err == nil {
				_ = conn.Close()
				return true
			}
		}
	}
	// 降级：进程名检测
	bin := strings.TrimSpace(p.ProcessBinName)
	if bin == "" {
		return false
	}
	return execCommandExists(bin)
}

func buildProcessRespItem(p model.ServerProcess) dto.GMRespProcessItem {
	return dto.GMRespProcessItem{
		Id:              p.Id,
		ServerType:      p.ServerType,
		ServerTypeName:  serverProcessTypeName(p.ServerType),
		ServerRefId:     p.ServerRefId,
		ServerName:      p.ServerName,
		ManageMode:      p.ManageMode,
		ProcessBinName:  p.ProcessBinName,
		StartCommand:    p.StartCommand,
		WorkDir:         p.WorkDir,
		HttpHealthUrl:   p.HttpHealthUrl,
		BuildRepoUrl:    p.BuildRepoUrl,
		BuildSourceDir:  p.BuildSourceDir,
		BuildOutputDir:  p.BuildOutputDir,
		BuildOutputName: p.BuildOutputName,
		SortOrder:       p.SortOrder,
		Remark:          p.Remark,
		RunState:        processRunState(p),
	}
}

// GmListProcesses GET /gm/processes/list
func GmListProcesses(c *gin.Context) {
	var req struct {
		ServerType int `json:"serverType"` // 0=全部
	}
	rawData, _ := c.GetRawData()
	_ = json.Unmarshal(rawData, &req)

	query := db.AccountDb.Table(define.ServerProcessTable).Asc("sort_order", "id")
	if req.ServerType > 0 {
		query = query.Where("server_type = ?", req.ServerType)
	}

	list := make([]model.ServerProcess, 0)
	if err := query.Find(&list); err != nil {
		log.Error("GmListProcesses find err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	resp := make([]dto.GMRespProcessItem, 0, len(list))
	for _, p := range list {
		resp = append(resp, buildProcessRespItem(p))
	}
	HTTPRetGameData(c, SUCCESS, "success", resp, map[string]any{"list": resp})
}

// GmCreateProcess POST /gm/processes/create
func GmCreateProcess(c *gin.Context) {
	var req dto.GmProcessUpsertReq
	rawData, _ := c.GetRawData()
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if strings.TrimSpace(req.ServerName) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverName required")
		return
	}
	if req.ServerType < 1 || req.ServerType > 3 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverType must be 1/2/3")
		return
	}

	p := model.ServerProcess{
		ServerType:      req.ServerType,
		ServerRefId:     req.ServerRefId,
		ServerName:      strings.TrimSpace(req.ServerName),
		ManageMode:      req.ManageMode,
		ProcessBinName:  strings.TrimSpace(req.ProcessBinName),
		StartCommand:    strings.TrimSpace(req.StartCommand),
		WorkDir:         strings.TrimSpace(req.WorkDir),
		HttpHealthUrl:   strings.TrimSpace(req.HttpHealthUrl),
		BuildRepoUrl:    strings.TrimSpace(req.BuildRepoUrl),
		BuildSourceDir:  strings.TrimSpace(req.BuildSourceDir),
		BuildOutputDir:  strings.TrimSpace(req.BuildOutputDir),
		BuildOutputName: strings.TrimSpace(req.BuildOutputName),
		SortOrder:       req.SortOrder,
		Remark:          strings.TrimSpace(req.Remark),
	}
	if _, err := db.AccountDb.Table(define.ServerProcessTable).Insert(&p); err != nil {
		log.Error("GmCreateProcess insert err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGameData(c, SUCCESS, "success", buildProcessRespItem(p), map[string]any{"id": p.Id})
}

// GmUpdateProcess POST /gm/processes/update
func GmUpdateProcess(c *gin.Context) {
	var req dto.GmProcessUpsertReq
	rawData, _ := c.GetRawData()
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.Id <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "id required")
		return
	}

	p := model.ServerProcess{
		ServerType:      req.ServerType,
		ServerRefId:     req.ServerRefId,
		ServerName:      strings.TrimSpace(req.ServerName),
		ManageMode:      req.ManageMode,
		ProcessBinName:  strings.TrimSpace(req.ProcessBinName),
		StartCommand:    strings.TrimSpace(req.StartCommand),
		WorkDir:         strings.TrimSpace(req.WorkDir),
		HttpHealthUrl:   strings.TrimSpace(req.HttpHealthUrl),
		BuildRepoUrl:    strings.TrimSpace(req.BuildRepoUrl),
		BuildSourceDir:  strings.TrimSpace(req.BuildSourceDir),
		BuildOutputDir:  strings.TrimSpace(req.BuildOutputDir),
		BuildOutputName: strings.TrimSpace(req.BuildOutputName),
		SortOrder:       req.SortOrder,
		Remark:          strings.TrimSpace(req.Remark),
	}
	if _, err := db.AccountDb.Table(define.ServerProcessTable).Where("id = ?", req.Id).
		Cols("server_type", "server_ref_id", "server_name", "manage_mode",
			"process_bin_name", "start_command", "work_dir", "http_health_url",
			"build_repo_url", "build_source_dir", "build_output_dir", "build_output_name",
			"sort_order", "remark").Update(&p); err != nil {
		log.Error("GmUpdateProcess update err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmDeleteProcess POST /gm/processes/delete
func GmDeleteProcess(c *gin.Context) {
	var req dto.GmProcessDeleteReq
	rawData, _ := c.GetRawData()
	if err := json.Unmarshal(rawData, &req); err != nil || len(req.Ids) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "ids required")
		return
	}
	if _, err := db.AccountDb.Table(define.ServerProcessTable).In("id", req.Ids).Delete(&model.ServerProcess{}); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

func loadProcessForAction(id int64) (*model.ServerProcess, bool, error) {
	p := new(model.ServerProcess)
	has, err := db.AccountDb.Table(define.ServerProcessTable).Where("id = ?", id).Get(p)
	return p, has, err
}

func startProcess(p model.ServerProcess) error {
	p.ManageMode = strings.TrimSpace(p.ManageMode)
	if p.ManageMode != manageModeLocalCommand {
		return errManagedServerManualStart
	}
	if strings.TrimSpace(p.StartCommand) == "" {
		return errManagedServerStartCommandRequired
	}
	if strings.TrimSpace(p.ProcessBinName) != "" && execCommandExists(p.ProcessBinName) {
		return errManagedServerAlreadyRunning
	}

	name, args := buildManagedServerStartShell(p.StartCommand)
	cmd := execCommand(name, args...)
	if strings.TrimSpace(p.WorkDir) != "" {
		cmd.Dir = p.WorkDir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	go func() {
		if waitErr := cmd.Wait(); waitErr != nil {
			log.Error("process [%s] exited: %v stderr: %s", p.ServerName, waitErr, strings.TrimSpace(stderr.String()))
		}
	}()
	return nil
}

func stopProcess(p model.ServerProcess) error {
	p.ManageMode = strings.TrimSpace(p.ManageMode)
	if p.ManageMode != manageModeLocalCommand {
		return errManagedServerManualStop
	}
	bin := strings.TrimSpace(p.ProcessBinName)
	if bin == "" {
		return errManagedServerProcessNameRequired
	}
	if !execCommandExists(bin) {
		return errManagedServerNotRunning
	}

	var lastErr error
	for _, candidate := range processCandidates(bin) {
		var err error
		if runtime.GOOS == "windows" {
			err = execCommandRun("taskkill", "/IM", candidate, "/F")
		} else {
			err = execCommandRun("pkill", "-x", candidate)
		}
		if err == nil {
			time.Sleep(500 * time.Millisecond)
			if !execCommandExists(bin) {
				return nil
			}
		} else {
			lastErr = err
		}
	}
	time.Sleep(500 * time.Millisecond)
	if !execCommandExists(bin) {
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return errManagedServerStopFailed
}

func handleProcessLifecycle(c *gin.Context, action string) {
	var req dto.GmProcessActionReq
	rawData, _ := c.GetRawData()
	if err := json.Unmarshal(rawData, &req); err != nil || req.Id <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "id required")
		return
	}
	p, has, err := loadProcessForAction(req.Id)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "进程记录不存在")
		return
	}

	switch action {
	case "start":
		err = startProcess(*p)
	case "stop":
		err = stopProcess(*p)
	case "restart":
		if strings.TrimSpace(p.ManageMode) != manageModeLocalCommand {
			err = errManagedServerManualStart
			break
		}
		if strings.TrimSpace(p.ProcessBinName) != "" && execCommandExists(p.ProcessBinName) {
			if stopErr := stopProcess(*p); stopErr != nil && stopErr != errManagedServerNotRunning {
				err = stopErr
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		err = startProcess(*p)
	default:
		err = simpleError("unsupported action")
	}
	if err != nil {
		writeManagedServerActionErr(c, err)
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmStartProcess POST /gm/processes/start
func GmStartProcess(c *gin.Context) { handleProcessLifecycle(c, "start") }

// GmStopProcess POST /gm/processes/stop
func GmStopProcess(c *gin.Context) { handleProcessLifecycle(c, "stop") }

// GmRestartProcess POST /gm/processes/restart
func GmRestartProcess(c *gin.Context) { handleProcessLifecycle(c, "restart") }

// GmBuildProcess POST /gm/processes/build
// 从 server_process 读取 build 配置（build_repo_url / build_source_dir / build_output_dir /
// build_output_name）执行拉取 + 编译，替代 build_server.go 中的硬编码路径。
func GmBuildProcess(c *gin.Context) {
	var req dto.GmProcessBuildReq
	rawData, _ := c.GetRawData()
	if err := json.Unmarshal(rawData, &req); err != nil || req.Id <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "id required")
		return
	}

	p, has, err := loadProcessForAction(req.Id)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "进程记录不存在")
		return
	}

	if strings.TrimSpace(p.BuildRepoUrl) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "该进程未配置 buildRepoUrl，不支持在线编译")
		return
	}
	if strings.TrimSpace(p.BuildSourceDir) == "" || strings.TrimSpace(p.BuildOutputDir) == "" || strings.TrimSpace(p.BuildOutputName) == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "buildSourceDir / buildOutputDir / buildOutputName 均不能为空")
		return
	}

	result := GitPullOrClone(p.BuildSourceDir, p.BuildRepoUrl)
	if !result.Success {
		HTTPRetGameData(c, ERR_GIT_ERROR, "git pull failed",
			map[string]any{"msg": result.Message}, map[string]any{"msg": result.Message})
		return
	}
	log.Debug("GmBuildProcess git ok: %s dir=%s", result.Message, result.Directory)

	res := buildServer(p.BuildSourceDir, p.BuildOutputDir, p.BuildOutputName)
	if res.Success {
		HTTPRetGameData(c, SUCCESS, "success",
			map[string]any{"msg": res.Message}, map[string]any{"msg": res.Message})
	} else {
		HTTPRetGameData(c, ERR_GIT_ERROR, "build failed",
			map[string]any{"msg": res.Message}, map[string]any{"msg": res.Message})
	}
}
