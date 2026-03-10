package logic

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

const (
	SUCCESS                              = 0
	ERR_SERVER_INTERNAL                  = 1    // 服务器内部错误
	ERR_PAY_ORDER_NOT_FOUND              = 2    // 订单不存在
	ERR_PAY_SIGN                         = 3    // 签名不正确
	ERR_INVITE_CODE_FAIL                 = 29   // 邀请码验证不通过
	ERR_DB                               = 1801 // 数据库错误
	ERR_ACCOUNT_EXISTS                   = 1802 // 账号已存在
	ERR_ACCOUNT_PASSWORD_FAILED          = 1803 // 账号密码错误
	ERR_ACCOUNT_TYPE_UNKNOWN             = 1804 // 账号类型错误
	ERR_ACCOUNT_NOT_FOUND                = 1805 // 账号不存在
	ERR_ACCOUNT_VERIFY_CODE_INCORRECT    = 1806 // 验证码不正确
	ERR_ACCOUNT_GET_VERIFY_CODE_FAILED   = 1807 // 获取验证码失败
	ERR_ACCOUNT_REGISTER_CLOSED          = 1808 // 注册服务关闭
	ERR_ACCOUNT_LOGIN_SERVER_MAINTAIN    = 1809 // 服务器维护 其实只有白名单账号可以进
	ERR_ACCOUNT_BANNED                   = 1810 // 账号被ban中
	ERR_ACCOUNT_PARAMS_ERROR             = 1811 // 参数错误
	ERR_ACCOUNT_CLIENT_VERSION_UNMATCHED = 1812 // 客户端版本不匹配
	ERR_ACCOUNT_SDK_TOKEN_AUTH_FAILED    = 1813 // 登录SDK Token效验失败
	ERR_ACCOUNT_SDK_TOKEN_EXPIRED        = 1814 // 登录SDK Token过期
	ERR_ACCOUNT_HAS_NO_NFT_HERO          = 1815 // 帐号没有nft英雄
	ERR_ACCOUNT_FORCED_OFFLINE           = 1816 // 帐号强制下线中
	ERR_GIT_ERROR                        = 1817 // git错误
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

	js, _ := json.Marshal(servers)
	HTTPRetGame(c, SUCCESS, "success", map[string]any{
		"list": string(js),
	})
}

// GmStartServer 启动大厅服（区服进程，数据来自 game_server，group_id>0）
func GmStartServer(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer unmarshal err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartServer missing serverId")
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId required")
		return
	}

	log.Debug("请求启动大厅服 serverId:%v", serverId)

	serverItem := new(model.ServerItem)
	serverItem.Id = int64(serverId)
	has, err := db.AccountDb.Table(define.GameServerTable).Where("group_id > ?", 0).Get(serverItem)
	if err != nil {
		log.Error("GmStartServer get err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "区服不存在")
		return
	}
	if serverItem.ExeName == "" || serverItem.ExePath == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "区服未配置 ExeName/ExePath")
		return
	}

	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	_, err = cmd.Output()
	if err == nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "该大厅服进程已在运行")
		return
	}
	cmd = exec.Command(serverItem.ExePath)
	cmd.Dir = "/usr/local/games/xiyou/server"
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Start()
	if err != nil {
		log.Error("Start failed: %v, Stderr: %s", err, stderr.String())
		HTTPRetGame(c, ERR_DB, fmt.Sprintf("start failed: %v", err))
		return
	}
	time.Sleep(1 * time.Second)
	if cmd.Process == nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "process failed to start")
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Error("Process crashed: %v, Stderr: %s", err, stderr.String())
		}
	}()
}

// GmStopServer 停止大厅服（区服进程，数据来自 game_server，group_id>0）
func GmStopServer(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStopServer unmarshal err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId required")
		return
	}

	log.Debug("请求停止大厅服 serverId:%v", serverId)

	serverItem := new(model.ServerItem)
	serverItem.Id = int64(serverId)
	has, err := db.AccountDb.Table(define.GameServerTable).Where("group_id > ?", 0).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "区服不存在")
		return
	}

	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	output, err := cmd.Output()
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "该大厅服进程未在运行")
		return
	}
	pidStr := strings.TrimSpace(string(output))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if err = process.Signal(syscall.SIGTERM); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	HTTPRetGame(c, SUCCESS, "success")
}

// GmReStartServer 重启大厅服（区服进程，数据来自 game_server，group_id>0）
func GmReStartServer(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmReStartServer unmarshal err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId required")
		return
	}

	log.Debug("请求重启大厅服 serverId:%v", serverId)

	serverItem := new(model.ServerItem)
	serverItem.Id = int64(serverId)
	has, err := db.AccountDb.Table(define.GameServerTable).Where("group_id > ?", 0).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "区服不存在")
		return
	}

	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	output, err := cmd.Output()
	if err == nil {
		pidStr := strings.TrimSpace(string(output))
		pid, pErr := strconv.Atoi(pidStr)
		if pErr != nil {
			HTTPRetGame(c, ERR_DB, pErr.Error())
			return
		}
		process, pErr := os.FindProcess(pid)
		if pErr != nil {
			HTTPRetGame(c, ERR_DB, pErr.Error())
			return
		}
		if pErr = process.Signal(syscall.SIGTERM); pErr != nil {
			HTTPRetGame(c, ERR_DB, pErr.Error())
			return
		}
	}

	cmd = exec.Command(serverItem.ExePath)
	cmd.Dir = "/usr/local/games/xiyou/server"
	if err = cmd.Start(); err != nil {
		log.Error("GmReStartServer start err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Error("GmReStartServer wait err: %v", err)
		}
	}()
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
		groupName := ""
		if g, ok := metaMap[int64(serverItem[i].GroupId)]; ok {
			groupName = g.Name
		}
		openTime := time.Unix(serverItem[i].OpenServerTime, 0).Format("2006-01-02 15:04:05")
		closeTime := time.Unix(serverItem[i].StopServerTime, 0).Format("2006-01-02 15:04:05")

		runState := "离线"
		if cmd := exec.Command("pgrep", "-x", serverItem[i].ExeName); cmd.Run() == nil {
			runState = "运行中"
		}

		gameServerName := ""
		gameServerId := 0
		if serverItem[i].GameServer > 0 {
			var gameRow dto.GameServerItem
			if ok, _ := db.AccountDb.Table(define.GameServerTable).Where("id = ?", serverItem[i].GameServer).Get(&gameRow); ok {
				gameServerName = gameRow.ServerName
				gameServerId = int(gameRow.Id)
			}
		}

		items = append(items, &dto.GMRespServerItem{
			Id:             serverItem[i].Id,
			ServerName:     serverItem[i].ServerName,
			GroupId:        serverItem[i].GroupId,
			GroupName:      groupName,
			Channel:        serverItem[i].Channel,
			Ip:             serverItem[i].Ip,
			Port:           serverItem[i].Port,
			RedisPort:         serverItem[i].RedisPort,
			MysqlAddr:         serverItem[i].MysqlAddr,
			LoginServerUrl:    serverItem[i].LoginServerUrl,
			MainServerHttpUrl: serverItem[i].MainServerHttpUrl,
			ServerState:       serverStateToText(serverItem[i].ServerState),
			OpenServerTime: openTime,
			StopServerTime: closeTime,
			RunState:       runState,
			GameServer:     gameServerName,
			GameServerId:   gameServerId,
		})
	}

	js, _ := json.Marshal(items)
	HTTPRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
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
	js, _ := json.Marshal(items)
	HTTPRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
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
		HTTPRetGame(c, SUCCESS, "success", map[string]any{"state": true, "msg": "目录已存在"})
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
	HTTPRetGame(c, SUCCESS, "success", map[string]any{"state": true})
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
	HTTPRetGame(c, SUCCESS, "success", map[string]any{"filename": filename})
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

// GmSetServerTime 设置 GM 所在机器系统时间（执行 date -s，需 root，慎用）
func GmSetServerTime(c *gin.Context) {
	var Info dto.GmSetServerTime
	if err := c.ShouldBindJSON(&Info); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(Info.SetTime), time.Local)
	if err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "时间格式错误，需 2006-01-02 15:04:05")
		return
	}
	if t.Unix() <= time.Now().Unix() {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "不允许将时间设置为过去或当前")
		return
	}
	timeStr := t.Format("2006-01-02 15:04:05")
	cmd := exec.Command("date", "-s", timeStr)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Error("GmSetServerTime failed: %v, output: %s", err, string(out))
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "设置时间失败: "+err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmStartGameServer 启动游戏服进程（仅 game_server 表中 group_id=0 的记录）
func GmStartGameServer(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartGameServer unmarshal err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	serverId, ok := result["serverId"].(float64)
	if !ok {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId required")
		return
	}
	log.Debug("请求启动游戏服 serverId:%v", serverId)

	serverItem := new(dto.GameServerItem)
	serverItem.Id = int64(serverId)
	has, err := db.AccountDb.Table(define.GameServerTable).Where("group_id = ?", 0).Get(serverItem)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "游戏服不存在")
		return
	}
	if serverItem.ExeName == "" || serverItem.ExePath == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "游戏服未配置 ExeName/ExePath")
		return
	}
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	if _, err = cmd.Output(); err == nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "该游戏服进程已在运行")
		return
	}
	cmd = exec.Command(serverItem.ExePath)
	cmd.Dir = "/usr/local/games/xiyou/server"
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		log.Error("GmStartGameServer start failed: %v, Stderr: %s", err, stderr.String())
		HTTPRetGame(c, ERR_DB, fmt.Sprintf("start failed: %v", err))
		return
	}
	time.Sleep(1 * time.Second)
	if cmd.Process == nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "process failed to start")
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Error("GmStartGameServer process crashed: %v, Stderr: %s", err, stderr.String())
		}
	}()
}

// GmStopGameServer 停止游戏服进程（仅 group_id=0）
func GmStopGameServer(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartGameServer find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartGameServer find serverName err")
		HTTPRetGame(c, ERR_DB, "GmStartGameServer find serverName err")
		return
	}

	log.Debug("请求服务中心数据:%v", serverId)

	serverItem := new(dto.GameServerItem)
	serverItem.Id = int64(serverId)
	has, err := db.AccountDb.Table(define.GameServerTable).Where("group_id = ?", 0).Get(serverItem)
	if err != nil {
		log.Error("GmStopGameServer get err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "游戏服不存在")
		return
	}
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	output, err := cmd.Output()
	if err != nil {
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "该游戏服进程未在运行")
		return
	}
	pidStr := strings.TrimSpace(string(output))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if err = process.Signal(syscall.SIGTERM); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
}

// GmReStartGameServer 重启游戏服进程（仅 group_id=0）
func GmReStartGameServer(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	serverId, ok := result["serverId"].(float64)
	if !ok {
		log.Error("GmStartServer find serverName err")
		HTTPRetGame(c, ERR_DB, "GmStartServer find serverName err")
		return
	}

	log.Debug("请求服务中心数据:%v", serverId)

	serverItem := new(dto.GameServerItem)
	serverItem.Id = int64(serverId)
	has, err := db.AccountDb.Table(define.GameServerTable).Where("group_id = ?", 0).Get(serverItem)
	if err != nil {
		log.Error("GmReStartGameServer get err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "游戏服不存在")
		return
	}
	cmd := exec.Command("pgrep", "-x", serverItem.ExeName)
	output, err := cmd.Output()
	if err == nil {
		pidStr := strings.TrimSpace(string(output))
		pid, pErr := strconv.Atoi(pidStr)
		if pErr != nil {
			HTTPRetGame(c, ERR_DB, pErr.Error())
			return
		}
		process, pErr := os.FindProcess(pid)
		if pErr != nil {
			HTTPRetGame(c, ERR_DB, pErr.Error())
			return
		}
		_ = process.Signal(syscall.SIGTERM)
	}
	cmd = exec.Command(serverItem.ExePath)
	cmd.Dir = "/usr/local/games/xiyou/server"
	if err = cmd.Start(); err != nil {
		log.Error("GmReStartGameServer start err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Error("GmReStartGameServer wait err: %v", err)
		}
	}()
}

// GmGetGameServerProcessList 获取所有游戏服进程列表（game_server 表中 group_id=0 的记录）
func GmGetGameServerProcessList(c *gin.Context) {
	log.Debug("请求游戏服进程列表")

	var serverItem []dto.GameServerItem
	err := db.AccountDb.Table(define.GameServerTable).Where("group_id = ?", 0).Find(&serverItem)
	if err != nil {
		log.Error("GmGetGameServerProcessList find err :%v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	items := make([]*dto.GMGameRespServerItem, 0, len(serverItem))
	for i := range serverItem {
		runState := "离线"
		if exec.Command("pgrep", "-x", serverItem[i].ExeName).Run() == nil {
			runState = "运行中"
		}
		items = append(items, &dto.GMGameRespServerItem{
			Id:         serverItem[i].Id,
			Ip:         serverItem[i].Ip,
			Port:       serverItem[i].Port,
			ServerName: serverItem[i].ServerName,
			RedisPort:  serverItem[i].RedisPort,
			MysqlAddr:  serverItem[i].MysqlAddr,
			RunState:   runState,
		})
	}
	js, _ := json.Marshal(items)
	HTTPRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(items),
	})
}
