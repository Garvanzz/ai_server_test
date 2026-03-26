package logic

import (
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"path/filepath"
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

const dateTimeLayout = "2006-01-02 15:04:05"

const (
	manageModeManual       = "manual"
	manageModeLocalCommand = "local_command"
)

func serverGroupTypeToText(groupType int) string {
	switch groupType {
	case 1:
		return "推荐"
	case 2:
		return "历史"
	default:
		return "常规"
	}
}

func formatServerTime(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format(dateTimeLayout)
}

func parseServerTime(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parsed, err := time.ParseInLocation(dateTimeLayout, value, time.Local)
	if err != nil {
		return 0, err
	}
	return parsed.Unix(), nil
}

func buildManagedServerStartShell(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", command}
	}
	return "sh", []string{"-c", command}
}

func processCandidates(processName string) []string {
	processName = strings.TrimSpace(processName)
	if processName == "" {
		return nil
	}
	if runtime.GOOS == "windows" {
		return windowsProcessCandidates(processName)
	}
	candidates := []string{processName}
	nameWithoutExt := strings.TrimSuffix(processName, filepath.Ext(processName))
	if nameWithoutExt != "" && nameWithoutExt != processName {
		candidates = append(candidates, nameWithoutExt)
	}
	return candidates
}

func buildServerGroupMap() map[int64]model.ServerGroup {
	groups := make([]model.ServerGroup, 0)
	_ = db.AccountDb.Table(define.ServerGroupTable).Asc("sort_order", "id").Find(&groups)
	groupMap := make(map[int64]model.ServerGroup, len(groups))
	for _, group := range groups {
		groupMap[group.Id] = group
	}
	return groupMap
}

func buildManagedServerItem(item model.ServerItem, groupMap map[int64]model.ServerGroup) dto.GMRespServerItem {
	groupName := ""
	if group, ok := groupMap[int64(item.GroupId)]; ok {
		groupName = group.Name
	}
	runState := "离线"
	if managedServerReachable(item) {
		runState = "运行中"
	}
	return dto.GMRespServerItem{
		Id:                item.Id,
		LogicServerId:     item.LogicServerId,
		MergeState:        item.MergeState,
		MergeStateText:    mergeStateToText(item.MergeState),
		MergeTime:         item.MergeTime,
		ServerName:        item.ServerName,
		GroupId:           item.GroupId,
		GroupName:         groupName,
		Channel:           item.Channel,
		Ip:                item.Ip,
		Port:              item.Port,
		MainServerHttpUrl: item.MainServerHttpUrl,
		ServerState:       serverStateToText(item.ServerState),
		ServerStateCode:   item.ServerState,
		OpenServerTime:    formatServerTime(item.OpenServerTime),
		StopServerTime:    formatServerTime(item.StopServerTime),
		RunState:          runState,
	}
}

func managedServerReachable(item model.ServerItem) bool {
	return mainServerHTTPReachable(item.MainServerHttpUrl)
}

func mainServerHTTPReachable(rawURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return false
	}
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || strings.TrimSpace(parsed.Host) == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", parsed.Host, 800*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func execCommandExists(exeName string) bool {
	exeName = strings.TrimSpace(exeName)
	if exeName == "" {
		return false
	}
	if runtime.GOOS == "windows" {
		return windowsProcessExists(exeName)
	}
	nameWithoutExt := strings.TrimSuffix(exeName, filepath.Ext(exeName))
	return execCommandRun("pgrep", "-x", exeName) == nil || (nameWithoutExt != "" && execCommandRun("pgrep", "-x", nameWithoutExt) == nil)
}

func windowsProcessExists(exeName string) bool {
	for _, candidate := range windowsProcessCandidates(exeName) {
		cmd := execCommand("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", candidate), "/FO", "CSV", "/NH")
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		content := strings.ToLower(string(output))
		if strings.Contains(content, strings.ToLower(candidate)) && !strings.Contains(content, "no tasks are running") {
			return true
		}
	}
	return false
}

func windowsProcessCandidates(exeName string) []string {
	base := strings.TrimSpace(exeName)
	if base == "" {
		return nil
	}
	if strings.EqualFold(filepath.Ext(base), ".exe") {
		trimmed := strings.TrimSuffix(base, filepath.Ext(base))
		if trimmed == "" {
			return []string{base}
		}
		return []string{base, trimmed}
	}
	return []string{base, base + ".exe"}
}

func execCommandRun(name string, arg ...string) error {
	cmd := execCommand(name, arg...)
	return cmd.Run()
}

var execCommand = exec.Command

func validateServerGroup(groupId int) error {
	if groupId <= 0 {
		return nil
	}
	exist, err := db.AccountDb.Table(define.ServerGroupTable).Where("id = ?", groupId).Exist()
	if err != nil {
		return err
	}
	if !exist {
		return errServerGroupNotFound
	}
	return nil
}

func validateLogicServer(logicServerId int64, selfId int64) error {
	if logicServerId <= 0 || logicServerId == selfId {
		return nil
	}
	exist, err := db.AccountDb.Table(define.GameServerTable).Where("id = ?", logicServerId).Exist()
	if err != nil {
		return err
	}
	if !exist {
		return errLogicServerNotFound
	}
	return nil
}

func validateServerState(state int) error {
	if state < define.ServerStateNormal || state > define.ServerStateStop {
		return errInvalidServerState
	}
	return nil
}

func hydrateServerItem(req dto.GmServerManageUpsertReq, current *model.ServerItem) (*model.ServerItem, error) {
	if req.Id < 0 {
		return nil, errInvalidServerID
	}
	if req.Channel < 0 || req.GroupId < 0 || req.Port < 0 {
		return nil, errInvalidServerConfig
	}
	if req.ServerState < define.ServerStateNormal || req.ServerState > define.ServerStateStop {
		return nil, errInvalidServerState
	}
	name := strings.TrimSpace(req.ServerName)
	if name == "" {
		return nil, errServerNameRequired
	}
	if err := validateServerGroup(req.GroupId); err != nil {
		return nil, err
	}
	openTime, err := parseServerTime(req.OpenServerTime)
	if err != nil {
		return nil, errOpenServerTimeInvalid
	}
	stopTime, err := parseServerTime(req.StopServerTime)
	if err != nil {
		return nil, errStopServerTimeInvalid
	}
	item := &model.ServerItem{}
	if current != nil {
		*item = *current
	}
	item.Id = req.Id
	item.Channel = req.Channel
	item.GroupId = req.GroupId
	item.LogicServerId = req.LogicServerId
	item.Ip = strings.TrimSpace(req.Ip)
	item.Port = req.Port
	item.MainServerHttpUrl = strings.TrimSpace(req.MainServerHttpUrl)
	item.ServerState = req.ServerState
	item.OpenServerTime = openTime
	item.StopServerTime = stopTime
	item.ServerName = name
	if item.LogicServerId == 0 && item.Id > 0 {
		item.LogicServerId = item.Id
	}
	if err := validateLogicServer(item.LogicServerId, item.Id); err != nil {
		return nil, err
	}
	return item, nil
}

func serverDeleteBlocked(serverId int64) (string, error) {
	if serverId <= 0 {
		return "", nil
	}
	linkCount, err := db.AccountDb.Table(define.GameServerTable).Where("logic_server_id = ? AND id <> ?", serverId, serverId).Count()
	if err != nil {
		return "", err
	}
	if linkCount > 0 {
		return "仍有其他服务器指向该逻辑服", nil
	}
	roleCount, err := db.AccountDb.Table(define.AccountRoleTable).Where("entry_server_id = ? OR logic_server_id = ? OR origin_server_id = ?", serverId, serverId, serverId).Count()
	if err != nil {
		return "", err
	}
	if roleCount > 0 {
		return "存在玩家角色数据，禁止删除", nil
	}
	mergeTargetCount, err := db.AccountDb.Table(define.MergePlanTable).Where("target_server_id = ?", serverId).Count()
	if err != nil {
		return "", err
	}
	if mergeTargetCount > 0 {
		return "存在合服计划引用，禁止删除", nil
	}
	mergeMapCount, err := db.AccountDb.Table(define.MergeServerMapTable).Where("source_server_id = ? OR target_server_id = ?", serverId, serverId).Count()
	if err != nil {
		return "", err
	}
	if mergeMapCount > 0 {
		return "存在合服映射引用，禁止删除", nil
	}
	return "", nil
}

var (
	errManagedServerManualStart          = simpleError("褰撳墠鏈嶅姟鍣ㄤ负鎵嬪姩绠＄悊妯″紡锛岃鎵嬪姩鍚姩")
	errManagedServerManualStop           = simpleError("褰撳墠鏈嶅姟鍣ㄤ负鎵嬪姩绠＄悊妯″紡锛岃鎵嬪姩鍋滄")
	errManagedServerProcessNameRequired  = simpleError("鏈厤缃繘绋嬪悕")
	errManagedServerStartCommandRequired = simpleError("鏈厤缃惎鍔ㄥ懡浠?")
	errManagedServerAlreadyRunning       = simpleError("鏈嶅姟鍣ㄨ繘绋嬪凡鍦ㄨ繍琛?")
	errManagedServerNotRunning           = simpleError("鏈嶅姟鍣ㄨ繘绋嬫湭鍦ㄨ繍琛?")
	errManagedServerStopFailed           = simpleError("鍋滄杩涚▼澶辫触")
	errInvalidServerID                   = simpleError("server id 非法")
	errInvalidServerConfig               = simpleError("服务器配置非法")
	errInvalidServerState                = simpleError("服务器状态非法")
	errServerNameRequired                = simpleError("服务器名称必填")
	errOpenServerTimeInvalid             = simpleError("开服时间格式错误")
	errStopServerTimeInvalid             = simpleError("停服时间格式错误")
	errServerGroupNotFound               = simpleError("区服组不存在")
	errLogicServerNotFound               = simpleError("逻辑服不存在")
)

type simpleError string

func (e simpleError) Error() string { return string(e) }

func GmGetServerGroupManageList(c *gin.Context) {
	groups := make([]model.ServerGroup, 0)
	if err := db.AccountDb.Table(define.ServerGroupTable).Asc("sort_order", "id").Find(&groups); err != nil {
		log.Error("GmGetServerGroupManageList find err: %v", err)
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	items := make([]dto.GmServerGroupManageItem, 0, len(groups))
	for _, group := range groups {
		serverCount, err := db.AccountDb.Table(define.GameServerTable).Where("group_id = ?", group.Id).Count()
		if err != nil {
			log.Error("GmGetServerGroupManageList count err: %v", err)
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
		items = append(items, dto.GmServerGroupManageItem{
			Id:            group.Id,
			Name:          group.Name,
			SortOrder:     group.SortOrder,
			GroupType:     group.GroupType,
			GroupTypeText: serverGroupTypeToText(group.GroupType),
			IsVisible:     group.IsVisible,
			ServerCount:   serverCount,
		})
	}
	HTTPRetGameData(c, SUCCESS, "success", items, map[string]interface{}{"totalCount": len(items)})
}

func GmCreateServerGroup(c *gin.Context) {
	var req dto.GmServerGroupUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "区服组名称必填")
		return
	}
	exist, err := db.AccountDb.Table(define.ServerGroupTable).Where("name = ?", req.Name).Exist()
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if exist {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "区服组名称已存在")
		return
	}
	group := &model.ServerGroup{
		Id:        req.Id,
		Name:      req.Name,
		SortOrder: req.SortOrder,
		GroupType: req.GroupType,
		IsVisible: req.IsVisible,
	}
	if _, err = db.AccountDb.Table(define.ServerGroupTable).Insert(group); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGameData(c, SUCCESS, "success", group)
}

func GmUpdateServerGroup(c *gin.Context) {
	var req dto.GmServerGroupUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.Id <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "区服组 id 必填")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "区服组名称必填")
		return
	}
	group := new(model.ServerGroup)
	has, err := db.AccountDb.Table(define.ServerGroupTable).Where("id = ?", req.Id).Get(group)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "区服组不存在")
		return
	}
	exist, err := db.AccountDb.Table(define.ServerGroupTable).Where("name = ? AND id <> ?", req.Name, req.Id).Exist()
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if exist {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "区服组名称已存在")
		return
	}
	group.Name = req.Name
	group.SortOrder = req.SortOrder
	group.GroupType = req.GroupType
	group.IsVisible = req.IsVisible
	if _, err = db.AccountDb.Table(define.ServerGroupTable).Where("id = ?", req.Id).Cols("name", "sort_order", "group_type", "is_visible").Update(group); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGameData(c, SUCCESS, "success", group)
}

func GmDeleteServerGroup(c *gin.Context) {
	var req dto.GmServerGroupDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if len(req.Ids) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "ids 必填")
		return
	}
	for _, id := range req.Ids {
		if id <= 0 {
			continue
		}
		serverCount, err := db.AccountDb.Table(define.GameServerTable).Where("group_id = ?", id).Count()
		if err != nil {
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
		if serverCount > 0 {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "区服组下仍有关联服务器，不能删除")
			return
		}
		if _, err = db.AccountDb.Table(define.ServerGroupTable).Where("id = ?", id).Delete(new(model.ServerGroup)); err != nil {
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	}
	HTTPRetGame(c, SUCCESS, "success")
}

func GmCreateManagedServer(c *gin.Context) {
	var req dto.GmServerManageUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	item, err := hydrateServerItem(req, nil)
	if err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	if req.Id > 0 {
		exist, existErr := db.AccountDb.Table(define.GameServerTable).Where("id = ?", req.Id).Exist()
		if existErr != nil {
			HTTPRetGame(c, ERR_DB, existErr.Error())
			return
		}
		if exist {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "服务器 id 已存在")
			return
		}
	}
	if _, err = db.AccountDb.Table(define.GameServerTable).Insert(item); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if item.LogicServerId == 0 {
		item.LogicServerId = item.Id
		_, _ = db.AccountDb.Table(define.GameServerTable).Where("id = ?", item.Id).Cols("logic_server_id").Update(&model.ServerItem{LogicServerId: item.Id})
	}
	groupMap := buildServerGroupMap()
	HTTPRetGameData(c, SUCCESS, "success", buildManagedServerItem(*item, groupMap))
}

func GmUpdateManagedServer(c *gin.Context) {
	var req dto.GmServerManageUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.Id <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "服务器 id 必填")
		return
	}
	current := new(model.ServerItem)
	has, err := db.AccountDb.Table(define.GameServerTable).Where("id = ?", req.Id).Get(current)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "服务器不存在")
		return
	}
	item, err := hydrateServerItem(req, current)
	if err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}
	if _, err = db.AccountDb.Table(define.GameServerTable).Where("id = ?", req.Id).
		Cols("channel", "group_id", "logic_server_id", "ip", "port", "main_server_http_url", "server_state", "open_server_time", "stop_server_time", "server_name").
		Update(item); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	groupMap := buildServerGroupMap()
	HTTPRetGameData(c, SUCCESS, "success", buildManagedServerItem(*item, groupMap))
}

func GmDeleteManagedServer(c *gin.Context) {
	var req dto.GmServerManageDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if len(req.Ids) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "ids 必填")
		return
	}
	for _, id := range req.Ids {
		if id <= 0 {
			continue
		}
		blockReason, err := serverDeleteBlocked(id)
		if err != nil {
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
		if blockReason != "" {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, blockReason)
			return
		}
		if _, err = db.AccountDb.Table(define.GameServerTable).Where("id = ?", id).Delete(new(model.ServerItem)); err != nil {
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	}
	HTTPRetGame(c, SUCCESS, "success")
}

func GmBatchUpdateManagedServer(c *gin.Context) {
	var req dto.GmServerBatchUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if len(req.Ids) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "ids 必填")
		return
	}
	if req.GroupId == nil && req.LogicServerId == nil && req.ServerState == nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "至少提供一个更新字段")
		return
	}
	if req.GroupId != nil {
		if *req.GroupId < 0 {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, errInvalidServerConfig.Error())
			return
		}
		if err := validateServerGroup(*req.GroupId); err != nil {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
			return
		}
	}
	if req.ServerState != nil {
		if err := validateServerState(*req.ServerState); err != nil {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
			return
		}
	}
	if req.LogicServerId != nil {
		if err := validateLogicServer(*req.LogicServerId, 0); err != nil {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
			return
		}
	}

	groupMap := buildServerGroupMap()
	updated := make([]dto.GMRespServerItem, 0, len(req.Ids))
	for _, id := range req.Ids {
		if id <= 0 {
			continue
		}
		item := new(model.ServerItem)
		has, err := db.AccountDb.Table(define.GameServerTable).Where("id = ?", id).Get(item)
		if err != nil {
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
		if !has {
			HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "服务器不存在")
			return
		}
		cols := make([]string, 0, 3)
		if req.GroupId != nil {
			item.GroupId = *req.GroupId
			cols = append(cols, "group_id")
		}
		if req.LogicServerId != nil {
			item.LogicServerId = *req.LogicServerId
			if item.LogicServerId == 0 {
				item.LogicServerId = item.Id
			}
			if err := validateLogicServer(item.LogicServerId, item.Id); err != nil {
				HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
				return
			}
			cols = append(cols, "logic_server_id")
		}
		if req.ServerState != nil {
			item.ServerState = *req.ServerState
			cols = append(cols, "server_state")
		}
		if len(cols) == 0 {
			continue
		}
		if _, err = db.AccountDb.Table(define.GameServerTable).Where("id = ?", id).Cols(cols...).Update(item); err != nil {
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
		updated = append(updated, buildManagedServerItem(*item, groupMap))
	}
	HTTPRetGameData(c, SUCCESS, "success", updated, map[string]interface{}{"totalCount": len(updated)})
}
