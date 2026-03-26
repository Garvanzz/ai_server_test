package logic

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

// serverStateToText 区服状态码转展示文案（与 core/define/server.go 常量一致）
func serverStateToText(state int) string {
	switch state {
	case define.ServerStateNormal:
		return "正常"
	case define.ServerStateYongji:
		return "拥挤"
	case define.ServerStateBaoMan:
		return "爆满"
	case define.ServerStateMaintenance:
		return "维护"
	case define.ServerStateNoOpen:
		return "未开服"
	case define.ServerStateStop:
		return "停服"
	default:
		return "未知"
	}
}

func mergeStateToText(state int) string {
	switch state {
	case 1:
		return "待合服"
	case 2:
		return "已合服"
	case 3:
		return "回滚中"
	default:
		return "正常"
	}
}

// GmGetServerList 获取区服列表（与 login 一致：区服来自 game_server，group_id>0；区服组来自 server_group）
func GmGetServerList(c *gin.Context) {
	metaList := make([]model.ServerGroup, 0)
	_ = db.AccountDb.Table(define.ServerGroupTable).Asc("sort_order", "id").Find(&metaList)
	metaMap := make(map[int64]model.ServerGroup)
	for _, m := range metaList {
		metaMap[m.Id] = m
	}

	items := make([]model.ServerItem, 0)
	err := db.AccountDb.Table(define.GameServerTable).Where("group_id > ?", 0).Asc("group_id", "id").Find(&items)
	if err != nil {
		log.Error("GmGetServerList find err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	log.Debug("获取区服列表, 共 %d 条", len(items))
	servers := make([]*dto.GMGetServerList, 0, len(items))
	for i := range items {
		groupName := ""
		if g, ok := metaMap[int64(items[i].GroupId)]; ok {
			groupName = g.Name
		}
		servers = append(servers, &dto.GMGetServerList{
			Name:      items[i].ServerName,
			Id:        items[i].Id,
			Time:      time.Now().Format("2006-01-02 15:04:05"),
			GroupId:   items[i].GroupId,
			GroupName: groupName,
		})
	}

	HTTPRetGameData(c, SUCCESS, "success", servers, map[string]any{
		"list": servers,
	})
}

// GmStartServer 启动大厅服（区服进程，数据来自 game_server，group_id>0）
type managedServerActionReq struct {
	ServerID int64 `json:"serverId"`
}

func parseManagedServerActionReq(c *gin.Context) (int64, error) {
	rawData, _ := c.GetRawData()
	var req managedServerActionReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		return 0, err
	}
	if req.ServerID <= 0 {
		return 0, simpleError("serverId required")
	}
	return req.ServerID, nil
}

func isManagedServerUserError(err error) bool {
	switch err {
	case errManagedServerManualStart,
		errManagedServerManualStop,
		errManagedServerProcessNameRequired,
		errManagedServerStartCommandRequired,
		errManagedServerAlreadyRunning,
		errManagedServerNotRunning:
		return true
	default:
		return false
	}
}

func writeManagedServerActionErr(c *gin.Context, err error) {
	if isManagedServerUserError(err) {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	HTTPRetGame(c, ERR_SERVER_INTERNAL, err.Error())
}

func handleManagedServerLifecycle(c *gin.Context, action string) {
	serverID, err := parseManagedServerActionReq(c)
	if err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}

	// Find process(es) linked to this entry server by server_ref_id.
	// A single entry server may have multiple process records (e.g. main_server + sidecar).
	var processes []model.ServerProcess
	if err = db.AccountDb.Table(define.ServerProcessTable).Where("server_ref_id = ?", serverID).Find(&processes); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if len(processes) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "该服未配置进程，请先在进程管理中添加")
		return
	}

	var lastErr error
	for _, p := range processes {
		var pErr error
		switch action {
		case "start":
			pErr = startProcess(p)
		case "stop":
			pErr = stopProcess(p)
		case "restart":
			if strings.TrimSpace(p.ManageMode) != manageModeLocalCommand {
				pErr = errManagedServerManualStart
				break
			}
			if strings.TrimSpace(p.ProcessBinName) != "" && execCommandExists(p.ProcessBinName) {
				if stopErr := stopProcess(p); stopErr != nil && stopErr != errManagedServerNotRunning {
					pErr = stopErr
					break
				}
				time.Sleep(500 * time.Millisecond)
			}
			pErr = startProcess(p)
		default:
			pErr = simpleError("unsupported action")
		}
		if pErr != nil && lastErr == nil {
			lastErr = pErr
		}
	}

	if lastErr != nil {
		writeManagedServerActionErr(c, lastErr)
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

func GmStartServer(c *gin.Context) {
	handleManagedServerLifecycle(c, "start")
}

// GmStopServer 停止大厅服（通过关联进程记录操作）
func GmStopServer(c *gin.Context) {
	handleManagedServerLifecycle(c, "stop")
}

// GmReStartServer 重启大厅服（通过关联进程记录操作）
func GmReStartServer(c *gin.Context) {
	handleManagedServerLifecycle(c, "restart")
}

// GmGetGameServerList 获取区服列表（含区服组、关联游戏服、进程状态）；区服来自 game_server（group_id>0）
func GmGetGameServerList(c *gin.Context) {
	log.Debug("请求区服列表（含游戏服关联）")

	metaList := make([]model.ServerGroup, 0)
	_ = db.AccountDb.Table(define.ServerGroupTable).Asc("sort_order", "id").Find(&metaList)
	metaMap := make(map[int64]model.ServerGroup)
	for _, m := range metaList {
		metaMap[m.Id] = m
	}

	var serverItem []model.ServerItem
	err := db.AccountDb.Table(define.GameServerTable).Where("group_id > ?", 0).Asc("group_id", "id").Find(&serverItem)
	if err != nil {
		log.Error("GmGetGameServerList find err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	items := make([]*dto.GMRespServerItem, 0, len(serverItem))
	for i := range serverItem {
		managed := buildManagedServerItem(serverItem[i], metaMap)
		items = append(items, &managed)
	}

	HTTPRetGameData(c, SUCCESS, "success", items, map[string]any{"totalCount": len(items)})
}

// GmGetHotUpdate 获取热更版本列表（从 hot_update 表）
func GmGetHotUpdate(c *gin.Context) {
	var list []dto.HotUpdateItem
	err := db.AccountDb.Table(define.HotUpdateTable).Find(&list)
	if err != nil {
		log.Error("GmGetHotUpdate find err: %v", err)
		HTTPRetGame(c, ERR_DB, "获取热更列表失败: "+err.Error())
		return
	}
	items := make([]*dto.GMRespHotUpdateItem, 0, len(list))
	for i := range list {
		items = append(items, &dto.GMRespHotUpdateItem{
			Id:          list[i].Id,
			Channel:     list[i].Channel,
			ChannelName: list[i].ChannelName,
			Version:     list[i].Version,
		})
	}
	HTTPRetGameData(c, SUCCESS, "success", items, map[string]any{"totalCount": len(items)})
}

// GmEditHotUpdateVersion 编辑指定热更版本的 version 字段（按 channel 查记录）
func GmEditHotUpdateVersion(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmHotUpdateVersionReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	channel := strings.TrimSpace(req.Channel)
	version := strings.TrimSpace(req.Version)
	if channel == "" || version == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "channel 和 version 必填")
		return
	}

	row := new(dto.HotUpdateItem)
	row.Channel = channel
	has, err := db.AccountDb.Table(define.HotUpdateTable).Get(row)
	if err != nil {
		log.Error("GmEditHotUpdateVersion get err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_PAY_ORDER_NOT_FOUND, "channel 对应热更记录不存在")
		return
	}
	row.Version = version
	_, err = db.AccountDb.Table(define.HotUpdateTable).Where("id = ?", row.Id).Cols("version").Update(row)
	if err != nil {
		log.Error("GmEditHotUpdateVersion update err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmCreateHotUpdateVersion 创建热更版本（插入 hot_update 表，channel 唯一）
func GmCreateHotUpdateVersion(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmHotUpdateVersionReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	channel := strings.TrimSpace(req.Channel)
	channelName := strings.TrimSpace(req.ChannelName)
	version := strings.TrimSpace(req.Version)
	if channel == "" || version == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "channel 和 version 必填")
		return
	}

	exist, err := db.AccountDb.Table(define.HotUpdateTable).Where("channel = ?", channel).Exist()
	if err != nil {
		log.Error("GmCreateHotUpdateVersion exist err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if exist {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "channel 已存在")
		return
	}

	row := &dto.HotUpdateItem{
		Channel:     channel,
		ChannelName: channelName,
		Version:     version,
	}
	_, err = db.AccountDb.Table(define.HotUpdateTable).Insert(row)
	if err != nil {
		log.Error("GmCreateHotUpdateVersion insert err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmDeleteHotUpdateVersion 按 channel 列表删除热更版本（支持单个或多个）
func GmDeleteHotUpdateVersion(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmHotUpdateDeleteReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if len(req.Channels) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "channel 不能为空")
		return
	}
	for _, channel := range req.Channels {
		channel = strings.TrimSpace(channel)
		if channel == "" {
			continue
		}
		affected, err := db.AccountDb.Table(define.HotUpdateTable).Where("channel = ?", channel).Delete(new(dto.HotUpdateItem))
		if err != nil {
			log.Error("GmDeleteHotUpdateVersion delete err: %v", err)
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
		if affected == 0 {
			log.Debug("GmDeleteHotUpdateVersion channel not found: %s", channel)
		}
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmCreateHotUpdatePath 创建热更路径（仅创建 channel/version 对应磁盘目录，不写 hot_update 表）
func GmCreateHotUpdatePath(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmHotUpdatePathReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	channel := strings.TrimSpace(req.Channel)
	version := strings.TrimSpace(req.Version)
	if channel == "" || version == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "channel 和 version 必填")
		return
	}
	if !safePathSegment.MatchString(channel) || !safePathSegment.MatchString(version) {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "channel/version 仅允许字母、数字、下划线、横线")
		return
	}

	hotUpdatePath := filepath.Join(hotUpdateBaseDir, channel, version)
	_, statErr := os.Stat(hotUpdatePath)
	if statErr == nil {
		HTTPRetGameData(c, SUCCESS, "success", map[string]any{"state": true, "msg": "目录已存在"}, map[string]any{"state": true, "msg": "目录已存在"})
		return
	}
	if statErr != nil && !os.IsNotExist(statErr) {
		log.Error("GmCreateHotUpdatePath stat err: %v", statErr)
		HTTPRetGame(c, ERR_DB, "检查热更目录失败")
		return
	}
	if err := os.MkdirAll(hotUpdatePath, 0755); err != nil {
		log.Error("GmCreateHotUpdatePath mkdir err: %v", err)
		HTTPRetGame(c, ERR_DB, "创建热更目录失败")
		return
	}
	HTTPRetGameData(c, SUCCESS, "success", map[string]any{"state": true}, map[string]any{"state": true})
}

// hotUpdateBaseDir 热更文件根目录，可按部署环境修改
const hotUpdateBaseDir = "/usr/local/games/xiyou/hotupdate"

// safePathSegment 仅允许字母数字、下划线、横线，防止路径遍历
var safePathSegment = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// GmUpload 上传文件（如热更包、配置等），要求 form: file, Channel, Version
func GmUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "缺少 file 或格式错误")
		return
	}
	channel := strings.TrimSpace(c.PostForm("Channel"))
	version := strings.TrimSpace(c.PostForm("Version"))
	if channel == "" || version == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "Channel 和 Version 必填")
		return
	}
	if !safePathSegment.MatchString(channel) || !safePathSegment.MatchString(version) {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "Channel/Version 仅允许字母、数字、下划线、横线")
		return
	}

	filename := filepath.Base(file.Filename)
	if filename == "" || filename == "." {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "非法文件名")
		return
	}
	uploadDir := filepath.Join(hotUpdateBaseDir, channel, version)
	dst := filepath.Join(uploadDir, filename)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "创建目录失败: "+err.Error())
		return
	}
	if err := c.SaveUploadedFile(file, dst); err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "保存文件失败: "+err.Error())
		return
	}
	if runtime.GOOS == "linux" {
		if err := os.Chown(dst, 1001, 1001); err != nil {
			log.Debug("Chown skipped or failed: %v", err)
		}
	}

	if err := unzip(dst, uploadDir); err != nil {
		_ = os.Remove(dst)
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "解压失败: "+err.Error())
		return
	}
	if err := os.Remove(dst); err != nil {
		log.Debug("remove zip after unzip: %v", err)
	}
	HTTPRetGameData(c, SUCCESS, "success", map[string]any{"filename": filename}, map[string]any{"filename": filename})
}

// unzip 解压 ZIP 到目标目录，防止 Zip Slip，循环内及时关闭句柄
func unzip(src, dest string) error {
	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		abs, _ := filepath.Abs(fpath)
		if !strings.HasPrefix(abs, destAbs+string(os.PathSeparator)) && abs != destAbs {
			return fmt.Errorf("非法文件路径: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, 0755)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// GmSetServerTime 设置游戏服务器时间偏移（转发到 main_server 的 /gm/time/set_offset 接口）
func GmSetServerTime(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var Info dto.GmSetServerTime
	if len(bytes.TrimSpace(rawData)) > 0 {
		if err := json.Unmarshal(rawData, &Info); err != nil {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
			return
		}
	}
	if Info.Server == 0 && Info.ServerID > 0 {
		Info.Server = Info.ServerID
	}
	if strings.TrimSpace(Info.SetTime) == "" {
		Info.SetTime = strings.TrimSpace(Info.SetTimeCamel)
	}
	if strings.TrimSpace(Info.SetTime) == "" {
		c.Request.Body = io.NopCloser(bytes.NewReader(rawData))
		GmGetServerTime(c)
		return
	}
	if Info.Server <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}

	// 解析目标时间
	targetTime, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(Info.SetTime), time.Local)
	if err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "时间格式错误，需 2006-01-02 15:04:05")
		return
	}

	// 计算目标时间与当前时间的差值（天数）
	now := time.Now()
	diff := targetTime.Sub(now)
	offsetDays := int64(diff.Hours() / 24)

	// 构造请求 main_server 的参数
	reqBody := map[string]int64{
		"offset_days": offsetDays,
	}
	js, _ := json.Marshal(reqBody)

	// 转发到 main_server 的时间偏移接口
	err, respStr := HttpRequestToServer(int(Info.Server), js, "/gm/time/set_offset")
	if err != nil {
		log.Error("GmSetServerTime forward to main_server err: %v, resp: %s", err, respStr)
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "设置时间失败: "+err.Error())
		return
	}

	code, message, payload, normalizeErr := normalizeMainServerTimeResponse(int64(Info.Server), []byte(respStr))
	if normalizeErr != nil {
		log.Error("GmSetServerTime normalize main_server response err: %v, resp: %s", normalizeErr, respStr)
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "main_server response invalid")
		return
	}
	if code != SUCCESS {
		HTTPRetGame(c, code, message)
		return
	}
	HTTPRetGameData(c, SUCCESS, message, payload, buildServerTimeLegacy(payload))
}

// GmGetServerTime 获取游戏服务器当前时间（转发到 main_server 的 /gm/time 接口）
func GmGetServerTime(c *gin.Context) {
	log.Debug("GmGetServerTime=============")
	// 从查询参数或 POST body 中获取 serverId
	serverIdStr := c.Query("server")
	if serverIdStr == "" {
		serverIdStr = c.Query("serverId")
	}
	if serverIdStr == "" {
		// 尝试从 POST body 获取
		var req struct {
			Server   int32 `json:"server"`
			ServerID int32 `json:"serverId"`
		}
		if c.Request.Method == "POST" {
			if err := c.ShouldBindJSON(&req); err == nil {
				if req.Server > 0 {
					serverIdStr = fmt.Sprintf("%d", req.Server)
				} else if req.ServerID > 0 {
					serverIdStr = fmt.Sprintf("%d", req.ServerID)
				}
			}
		}
	}

	serverId := 0
	if serverIdStr != "" {
		if id, err := strconv.Atoi(serverIdStr); err == nil {
			serverId = id
		}
	}

	baseURL := getMainServerURL(serverId)
	if baseURL == "" {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "main_server URL not configured")
		return
	}

	url := baseURL + "/gm/time"

	// 创建 GET 请求
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Error("GmGetServerTime request failed: %v", err)
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "获取时间失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("GmGetServerTime read response failed: %v", err)
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "读取响应失败: "+err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("GmGetServerTime http status: %d, body: %s", resp.StatusCode, string(body))
		HTTPRetGame(c, ERR_SERVER_INTERNAL, fmt.Sprintf("http status %d", resp.StatusCode))
		return
	}

	code, message, payload, normalizeErr := normalizeMainServerTimeResponse(int64(serverId), body)
	if normalizeErr != nil {
		log.Error("GmGetServerTime normalize main_server response err: %v, body: %s", normalizeErr, string(body))
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "main_server response invalid")
		return
	}
	if code != SUCCESS {
		HTTPRetGame(c, code, message)
		return
	}
	HTTPRetGameData(c, SUCCESS, message, payload, buildServerTimeLegacy(payload))
}

// GmStartGameServer 启动游戏服进程（通过 server_process ID 操作）
func GmStartGameServer(c *gin.Context) {
	handleProcessLifecycle(c, "start")
}

// GmStopGameServer 停止游戏服进程
func GmStopGameServer(c *gin.Context) {
	handleProcessLifecycle(c, "stop")
}

// GmReStartGameServer 重启游戏服进程
func GmReStartGameServer(c *gin.Context) {
	handleProcessLifecycle(c, "restart")
}

// GmGetGameServerProcessList 获取游戏服进程列表（来自 server_process 表，server_type=3）
func GmGetGameServerProcessList(c *gin.Context) {
	log.Debug("请求游戏服进程列表")
	list := make([]model.ServerProcess, 0)
	if err := db.AccountDb.Table(define.ServerProcessTable).Where("server_type = ?", dto.ServerProcessTypeGame).Asc("sort_order", "id").Find(&list); err != nil {
		log.Error("GmGetGameServerProcessList find err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	resp := make([]dto.GMRespProcessItem, 0, len(list))
	for _, p := range list {
		resp = append(resp, buildProcessRespItem(p))
	}
	HTTPRetGameData(c, SUCCESS, "success", resp, map[string]any{"totalCount": len(resp)})
}
