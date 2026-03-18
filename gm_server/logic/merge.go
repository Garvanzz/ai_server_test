package logic

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
)

const (
	mergePlanPending   = 0
	mergePlanRunning   = 1
	mergePlanSucceeded = 2
	mergePlanFailed    = 3
	mergePlanRolled    = 4
)

const redisScriptTemplatePath = "tools/scripts/redis/merge_redis_migration_template.ps1"

func normalizeRedisMode(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "shared", "independent":
		return mode
	default:
		return "independent"
	}
}

func joinIntList(values []int) string {
	return strings.Trim(strings.Replace(fmt.Sprint(values), " ", ",", -1), "[]")
}

func buildRedisKeyPatterns() []map[string]interface{} {
	return []map[string]interface{}{
		{"group": "role_route", "pattern": fmt.Sprintf("%s:<uid>:<entryServerId>", define.AccountRole), "action": "must_copy", "note": "入口服角色到 playerId 映射"},
		{"group": "player_core", "pattern": fmt.Sprintf("%s:<playerId>", define.Player), "action": "must_copy", "note": "玩家主数据 Hash"},
		{"group": "player_modules", "pattern": "Base/Bag/Shop/...:<playerId>", "action": "must_copy", "note": "玩家各模块 Redis 数据"},
		{"group": "activity", "pattern": fmt.Sprintf("%s:<actId>", define.ActivityRedisKey), "action": "copy_or_reset", "note": "活动主数据"},
		{"group": "activity_player", "pattern": fmt.Sprintf("%s:<actId>", define.ActivityPlayerRedisKey), "action": "copy_or_reset", "note": "活动玩家参与明细 Hash"},
		{"group": "rank", "pattern": "rank_*", "action": "copy_or_reset", "note": "排行榜 ZSET 与竞技/天梯记录"},
		{"group": "guild_state", "pattern": fmt.Sprintf("%s:<logicServerId>", define.GuildRedisKey), "action": "should_copy", "note": "公会运行态缓存"},
		{"group": "guild_player", "pattern": fmt.Sprintf("%s:<playerId>", define.PlayerGuildKey), "action": "must_copy", "note": "玩家公会关系缓存"},
		{"group": "mail_state", "pattern": "systemMailId:<logicServerId>, dailyMail:<logicServerId>", "action": "check_target", "note": "邮件游标与日常状态，目标服为主"},
		{"group": "chat_optional", "pattern": "guild_chat_history:<guildId>", "action": "optional", "note": "公会聊天历史，可不迁移"},
	}
}

func buildRedisMergeChecklist(targetServerId int, sourceServerIds []int, redisMode string) map[string]interface{} {
	logicIds := make([]int, 0, len(sourceServerIds))
	steps := make([]string, 0)
	for _, sid := range sourceServerIds {
		logicIds = append(logicIds, resolveLogicServerID(sid))
	}
	if redisMode == "shared" {
		steps = append(steps,
			"确认来源服与目标服 main_server 使用同一套 Redis。",
			"停服后校验来源入口服角色在目标逻辑服可读到玩家主数据、活动数据与排行榜数据。",
			"无需迁移 Redis，仅执行 MySQL 合服、角色路由切换和开服验证。",
		)
	} else {
		steps = append(steps,
			"停服并冻结登录。",
			"备份来源服与目标服 Redis。",
			"先迁移玩家主数据、角色映射、公会缓存，再迁移活动与排行榜 Key。",
			"完成 Redis 迁移后，再执行 GM 合服工单。",
			"开服前抽检登录、活动、排行、邮件、公会、好友链路。",
		)
	}
	commands := map[string]string{
		"exportTemplate": fmt.Sprintf("powershell -ExecutionPolicy Bypass -File %s -Mode export -SourceRedis <host:port> -TargetRedis <host:port> -SourceEntryServers %s -TargetLogicServer %d", redisScriptTemplatePath, joinIntList(sourceServerIds), targetServerId),
		"importTemplate": fmt.Sprintf("powershell -ExecutionPolicy Bypass -File %s -Mode import -SourceRedis <host:port> -TargetRedis <host:port> -SourceEntryServers %s -TargetLogicServer %d", redisScriptTemplatePath, joinIntList(sourceServerIds), targetServerId),
	}
	return map[string]interface{}{
		"redisMode":            redisMode,
		"targetServerId":       targetServerId,
		"sourceServerIds":      sourceServerIds,
		"sourceLogicServerIds": logicIds,
		"keyPatterns":          buildRedisKeyPatterns(),
		"steps":                steps,
		"scriptPath":           redisScriptTemplatePath,
		"commands":             commands,
		"docPath":              "tools/docs/merge_activity_redis_sop.md",
	}
}

func mergePlanStatusText(status int) string {
	switch status {
	case mergePlanRunning:
		return "执行中"
	case mergePlanSucceeded:
		return "成功"
	case mergePlanFailed:
		return "失败"
	case mergePlanRolled:
		return "已回滚"
	default:
		return "待执行"
	}
}

func normalizeServerIDs(ids []int) []int {
	set := make(map[int]struct{})
	for _, id := range ids {
		if id > 0 {
			set[id] = struct{}{}
		}
	}
	ret := make([]int, 0, len(set))
	for id := range set {
		ret = append(ret, id)
	}
	sort.Ints(ret)
	return ret
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}

func parseServerIds(v interface{}) []int {
	if arr, ok := v.([]interface{}); ok {
		ret := make([]int, 0, len(arr))
		for _, a := range arr {
			if id := toInt(a); id > 0 {
				ret = append(ret, id)
			}
		}
		return normalizeServerIDs(ret)
	}
	if arr, ok := v.([]int); ok {
		return normalizeServerIDs(arr)
	}
	if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
		parts := strings.Split(s, ",")
		ret := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			var id int
			_, _ = fmt.Sscanf(p, "%d", &id)
			if id > 0 {
				ret = append(ret, id)
			}
		}
		return normalizeServerIDs(ret)
	}
	return []int{}
}

func validateServers(target int, source []int) (bool, string, []model.ServerItem) {
	if target <= 0 {
		return false, "targetServerId required", nil
	}
	if len(source) == 0 {
		return false, "sourceServerIds required", nil
	}
	for _, sid := range source {
		if sid == target {
			return false, "source server cannot include target server", nil
		}
	}

	need := append([]int{target}, source...)
	rows := make([]model.ServerItem, 0)
	if err := db.AccountDb.Table(define.GameServerTable).In("id", need).Find(&rows); err != nil {
		return false, err.Error(), nil
	}
	if len(rows) != len(need) {
		return false, "some server ids not found", nil
	}
	return true, "", rows
}

func listGuildNameConflicts(targetServerId, sourceServerId int) ([]model.GuildDB, error) {
	rows := make([]model.GuildDB, 0)
	err := db.AccountDb.Table(define.GuildTable).Where("server_id = ?", sourceServerId).Find(&rows)
	if err != nil {
		return nil, err
	}
	ret := make([]model.GuildDB, 0)
	for _, g := range rows {
		exist, e := db.AccountDb.Table(define.GuildTable).Where("server_id = ? AND guild_name = ?", targetServerId, g.GuildName).Exist()
		if e != nil {
			return nil, e
		}
		if exist {
			ret = append(ret, g)
		}
	}
	return ret, nil
}

func countTableRows(table string, serverId int) int64 {
	if serverId <= 0 {
		return 0
	}
	n, err := db.AccountDb.Table(table).Where("server_id = ?", serverId).Count()
	if err != nil {
		return 0
	}
	return n
}

func countAccountRolesByEntryServer(entryServerId int) int64 {
	n, err := db.AccountDb.Table(define.AccountRoleTable).Where("entry_server_id = ?", entryServerId).Count()
	if err != nil {
		return 0
	}
	return n
}

func buildMergePrecheck(targetServerId int, sourceServerIds []int) (map[string]interface{}, []map[string]interface{}, error) {
	conflicts := make([]map[string]interface{}, 0)
	serverStats := make([]map[string]interface{}, 0, len(sourceServerIds))
	warnings := make([]string, 0)

	for _, sid := range sourceServerIds {
		logicSid := resolveLogicServerID(sid)
		guildConflicts, err := listGuildNameConflicts(targetServerId, logicSid)
		if err != nil {
			return nil, nil, err
		}
		for _, g := range guildConflicts {
			conflicts = append(conflicts, map[string]interface{}{
				"serverId":      sid,
				"logicServerId": logicSid,
				"type":          "guild_name",
				"key":           g.GuildName,
				"suggestName":   fmt.Sprintf("%s_S%d", g.GuildName, sid),
			})
		}

		stats := map[string]interface{}{
			"entryServerId":      sid,
			"logicServerId":      logicSid,
			"accountRoleCount":   countAccountRolesByEntryServer(sid),
			"guildCount":         countTableRows(define.GuildTable, logicSid),
			"guildApplyCount":    countTableRows(define.GuildApplyTable, logicSid),
			"guildLogCount":      countTableRows(define.GuildLogTable, logicSid),
			"sysMailCount":       countTableRows(define.SysMailInfoTable, logicSid),
			"adminMailCount":     countTableRows(define.AdminMailTable, logicSid),
			"playerMailCount":    countTableRows(define.PlayerMailInfoTable, logicSid),
			"friendApplyCount":   countTableRows(define.FriendApplyTable, logicSid),
			"friendBlockCount":   countTableRows(define.FriendBlockTable, logicSid),
			"payOrderCount":      countTableRows(define.PayOrderTable, logicSid),
			"payCacheOrderCount": countTableRows(define.PayCacheOrderTable, logicSid),
			"guildConflictCount": len(guildConflicts),
		}
		serverStats = append(serverStats, stats)

		if logicSid == targetServerId {
			warnings = append(warnings, fmt.Sprintf("入口服 %d 当前已路由到目标逻辑服 %d，本次只会补齐入口映射与状态。", sid, targetServerId))
		}
		if stats["guildApplyCount"].(int64) > 0 {
			warnings = append(warnings, fmt.Sprintf("入口服 %d 存在 %d 条公会申请缓存/记录，建议停服窗口执行并重点抽检公会申请列表。", sid, stats["guildApplyCount"].(int64)))
		}
		if stats["adminMailCount"].(int64) > 0 || stats["sysMailCount"].(int64) > 0 {
			warnings = append(warnings, fmt.Sprintf("入口服 %d 存在系统/延迟邮件数据，合服后将统一迁到逻辑服 %d。", sid, targetServerId))
		}
	}
	warnings = append(warnings,
		"请确认来源服与目标服的 main_server 是否共享同一套 Redis；若 Redis 独立，需额外迁移玩家主数据、活动数据、排行榜和公会缓存。",
		"活动/排行榜 Redis Key 的处理建议见 tools/docs/merge_activity_redis_sop.md。",
	)

	summary := map[string]interface{}{
		"targetServerId":  targetServerId,
		"sourceServerIds": sourceServerIds,
		"conflictCount":   len(conflicts),
		"conflicts":       conflicts,
		"serverStats":     serverStats,
		"warnings":        warnings,
	}
	return summary, conflicts, nil
}

func listMergePlans(c *gin.Context) {
	plans := make([]model.MergePlan, 0)
	err := db.AccountDb.Table(define.MergePlanTable).Desc("id").Limit(200).Find(&plans)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	typed := make([]map[string]interface{}, 0, len(plans))
	for _, p := range plans {
		typed = append(typed, map[string]interface{}{
			"id":              p.Id,
			"name":            p.Name,
			"targetServerId":  p.TargetServerId,
			"sourceServerIds": p.SourceServerIds,
			"status":          p.Status,
			"statusText":      mergePlanStatusText(p.Status),
			"operator":        p.Operator,
			"startTime":       p.StartTime,
			"endTime":         p.EndTime,
			"rollbackTime":    p.RollbackTime,
			"remark":          p.Remark,
		})
	}

	HTTPRetGameData(c, SUCCESS, "success", typed, map[string]interface{}{"totalCount": len(typed)})
}

func GmCreateMergePlan(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmCreateMergePlanReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}

	req.TargetServerId = resolveLogicServerID(req.TargetServerId)
	req.SourceServerIds = normalizeServerIDs(req.SourceServerIds)
	ok, msg, _ := validateServers(req.TargetServerId, req.SourceServerIds)
	if !ok {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, msg)
		return
	}

	plan := model.MergePlan{
		Name:            strings.TrimSpace(req.Name),
		TargetServerId:  req.TargetServerId,
		SourceServerIds: req.SourceServerIds,
		Status:          mergePlanPending,
		Operator:        c.GetString(ContextKeyGmUser),
		Remark:          strings.TrimSpace(req.Remark),
	}
	if plan.Name == "" {
		plan.Name = fmt.Sprintf("merge_%d_%d", plan.TargetServerId, time.Now().Unix())
	}

	if _, err := db.AccountDb.Table(define.MergePlanTable).Insert(&plan); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	maps := make([]model.MergeServerMap, 0, len(plan.SourceServerIds))
	for _, sid := range plan.SourceServerIds {
		maps = append(maps, model.MergeServerMap{
			PlanId:         plan.Id,
			SourceServerId: sid,
			TargetServerId: plan.TargetServerId,
			State:          0,
		})
	}
	if len(maps) > 0 {
		batch := make([]interface{}, 0, len(maps))
		for i := range maps {
			batch = append(batch, &maps[i])
		}
		if _, err := db.AccountDb.Table(define.MergeServerMapTable).Insert(batch...); err != nil {
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
	}

	HTTPRetGameData(c, SUCCESS, "success", map[string]interface{}{"planId": plan.Id}, map[string]interface{}{"planId": plan.Id})
}

func GmPrecheckMerge(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmPrecheckMergeReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	req.TargetServerId = resolveLogicServerID(req.TargetServerId)
	req.SourceServerIds = normalizeServerIDs(req.SourceServerIds)

	ok, msg, _ := validateServers(req.TargetServerId, req.SourceServerIds)
	if !ok {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, msg)
		return
	}

	summary, _, err := buildMergePrecheck(req.TargetServerId, req.SourceServerIds)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGameData(c, SUCCESS, "success", summary, summary)
}

func GmRedisMergeCheck(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmRedisMergeCheckReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	req.TargetServerId = resolveLogicServerID(req.TargetServerId)
	req.SourceServerIds = normalizeServerIDs(req.SourceServerIds)
	ok, msg, _ := validateServers(req.TargetServerId, req.SourceServerIds)
	if !ok {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, msg)
		return
	}
	result := buildRedisMergeChecklist(req.TargetServerId, req.SourceServerIds, normalizeRedisMode(req.RedisMode))
	HTTPRetGameData(c, SUCCESS, "success", result, result)
}

func GmExportRedisMergeScript(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmRedisMergeCheckReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	req.TargetServerId = resolveLogicServerID(req.TargetServerId)
	req.SourceServerIds = normalizeServerIDs(req.SourceServerIds)
	ok, msg, _ := validateServers(req.TargetServerId, req.SourceServerIds)
	if !ok {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, msg)
		return
	}
	redisMode := normalizeRedisMode(req.RedisMode)
	checklist := buildRedisMergeChecklist(req.TargetServerId, req.SourceServerIds, redisMode)
	script := fmt.Sprintf(`# Redis merge script template
# File: %s
# Redis mode: %s

powershell -ExecutionPolicy Bypass -File %s -Mode export -SourceRedis <host:port> -TargetRedis <host:port> -SourceEntryServers %s -TargetLogicServer %d
powershell -ExecutionPolicy Bypass -File %s -Mode import -SourceRedis <host:port> -TargetRedis <host:port> -SourceEntryServers %s -TargetLogicServer %d
`, redisScriptTemplatePath, redisMode, redisScriptTemplatePath, joinIntList(req.SourceServerIds), req.TargetServerId, redisScriptTemplatePath, joinIntList(req.SourceServerIds), req.TargetServerId)
	result := map[string]interface{}{
		"scriptPath": redisScriptTemplatePath,
		"script":     script,
		"checklist":  checklist,
	}
	HTTPRetGameData(c, SUCCESS, "success", result, result)
}

func GmExecuteMergePlan(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmExecuteMergePlanReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.PlanId <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "planId required")
		return
	}

	plan := new(model.MergePlan)
	has, err := db.AccountDb.Table(define.MergePlanTable).Where("id = ?", req.PlanId).Get(plan)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "merge plan not found")
		return
	}
	if plan.Status != mergePlanPending && plan.Status != mergePlanFailed {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "plan status not executable")
		return
	}

	now := time.Now().Unix()
	plan.Status = mergePlanRunning
	plan.StartTime = now
	if _, err = db.AccountDb.Table(define.MergePlanTable).Where("id = ?", plan.Id).Cols("status", "start_time").Update(plan); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	success := true
	for _, sid := range plan.SourceServerIds {
		serverOK := true
		logicSid := resolveLogicServerID(sid)
		if logicSid == plan.TargetServerId {
			if _, err = db.AccountDb.Table(define.GameServerTable).Where("id = ?", sid).
				Cols("logic_server_id", "merge_state", "merge_time").
				Update(&model.ServerItem{LogicServerId: int64(plan.TargetServerId), MergeState: 2, MergeTime: now}); err != nil {
				success = false
				serverOK = false
			}
			mapUpdate := model.MergeServerMap{State: 1, ErrMsg: ""}
			if !serverOK {
				mapUpdate.State = 2
				mapUpdate.ErrMsg = "update routing failed"
			}
			_, _ = db.AccountDb.Table(define.MergeServerMapTable).Where("plan_id = ? AND source_server_id = ?", plan.Id, sid).Cols("state", "err_msg").Update(&mapUpdate)
			continue
		}

		if _, err = db.AccountDb.Exec("UPDATE guild SET guild_name = CONCAT(guild_name, '_S', ?) WHERE server_id = ? AND guild_name IN (SELECT t.guild_name FROM (SELECT guild_name FROM guild WHERE server_id = ?) t)", sid, logicSid, plan.TargetServerId); err != nil {
			success = false
			serverOK = false
		}

		if conflictRows, cErr := listGuildNameConflicts(plan.TargetServerId, logicSid); cErr == nil {
			for _, g := range conflictRows {
				_, _ = db.AccountDb.Table(define.MergeConflictTable).Insert(&model.MergeConflictLog{
					PlanId:       plan.Id,
					ServerId:     logicSid,
					ConflictType: "guild_name",
					BizKey:       g.GuildName,
					OldValue:     g.GuildName,
					NewValue:     fmt.Sprintf("%s_S%d", g.GuildName, sid),
					Resolved:     1,
					CreatedAt:    now,
				})
			}
		}

		tables := []string{define.GuildTable, define.GuildApplyTable, define.GuildLogTable, define.SysMailInfoTable, define.AdminMailTable, define.PlayerMailInfoTable, define.FriendApplyTable, define.FriendBlockTable, define.PayOrderTable, define.PayCacheOrderTable}
		for _, tb := range tables {
			q := fmt.Sprintf("UPDATE %s SET server_id = ? WHERE server_id = ?", tb)
			if _, err = db.AccountDb.Exec(q, plan.TargetServerId, logicSid); err != nil {
				success = false
				serverOK = false
			}
		}
		if _, err = db.AccountDb.Table(define.AccountRoleTable).Where("entry_server_id = ?", sid).Cols("logic_server_id").Update(&model.AccountRole{LogicServerId: plan.TargetServerId}); err != nil {
			success = false
			serverOK = false
		}

		if _, err = db.AccountDb.Table(define.GameServerTable).Where("id = ?", sid).
			Cols("logic_server_id", "merge_state", "merge_time").
			Update(&model.ServerItem{LogicServerId: int64(plan.TargetServerId), MergeState: 2, MergeTime: now}); err != nil {
			success = false
			serverOK = false
		}

		mapUpdate := model.MergeServerMap{State: 1, ErrMsg: ""}
		if !serverOK {
			mapUpdate.State = 2
			mapUpdate.ErrMsg = "partial failed"
		}
		_, _ = db.AccountDb.Table(define.MergeServerMapTable).Where("plan_id = ? AND source_server_id = ?", plan.Id, sid).Cols("state", "err_msg").Update(&mapUpdate)
	}

	plan.EndTime = time.Now().Unix()
	if success {
		plan.Status = mergePlanSucceeded
	} else {
		plan.Status = mergePlanFailed
	}
	_, _ = db.AccountDb.Table(define.MergePlanTable).Where("id = ?", plan.Id).Cols("status", "end_time").Update(plan)

	if success {
		HTTPRetGameData(c, SUCCESS, "success", map[string]interface{}{
			"planId":          plan.Id,
			"status":          plan.Status,
			"statusText":      mergePlanStatusText(plan.Status),
			"rollbackWarning": "当前回滚仅恢复入口路由，不自动回滚已迁移业务数据；真实回滚请使用备份恢复。",
		}, map[string]interface{}{
			"planId":          plan.Id,
			"status":          plan.Status,
			"statusText":      mergePlanStatusText(plan.Status),
			"rollbackWarning": "当前回滚仅恢复入口路由，不自动回滚已迁移业务数据；真实回滚请使用备份恢复。",
		})
		return
	}
	HTTPRetGame(c, ERR_SERVER_INTERNAL, "merge executed with errors")
}

func GmRollbackMergePlan(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var req dto.GmExecuteMergePlanReq
	if err := json.Unmarshal(rawData, &req); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	if req.PlanId <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "planId required")
		return
	}

	plan := new(model.MergePlan)
	has, err := db.AccountDb.Table(define.MergePlanTable).Where("id = ?", req.PlanId).Get(plan)
	if err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	if !has {
		HTTPRetGame(c, ERR_ACCOUNT_NOT_FOUND, "merge plan not found")
		return
	}

	now := time.Now().Unix()
	for _, sid := range plan.SourceServerIds {
		_, _ = db.AccountDb.Table(define.GameServerTable).Where("id = ?", sid).
			Cols("logic_server_id", "merge_state", "merge_time").
			Update(&model.ServerItem{LogicServerId: int64(sid), MergeState: 0, MergeTime: now})
	}

	plan.Status = mergePlanRolled
	plan.RollbackTime = now
	_, _ = db.AccountDb.Table(define.MergePlanTable).Where("id = ?", plan.Id).Cols("status", "rollback_time").Update(plan)

	HTTPRetGameData(c, SUCCESS, "success", map[string]interface{}{
		"planId":          plan.Id,
		"status":          plan.Status,
		"statusText":      mergePlanStatusText(plan.Status),
		"rollbackWarning": "仅恢复入口路由与 merge_state，已迁移业务数据不会自动回滚。",
	}, map[string]interface{}{
		"planId":          plan.Id,
		"status":          plan.Status,
		"statusText":      mergePlanStatusText(plan.Status),
		"rollbackWarning": "仅恢复入口路由与 merge_state，已迁移业务数据不会自动回滚。",
	})
}

func GmListMergePlans(c *gin.Context) {
	listMergePlans(c)
}

func GmListMergeConflicts(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	_ = json.Unmarshal(rawData, &result)
	planId := int64(toInt(result["planId"]))

	rows := make([]model.MergeConflictLog, 0)
	q := db.AccountDb.Table(define.MergeConflictTable)
	if planId > 0 {
		q = q.Where("plan_id = ?", planId)
	}
	if err := q.Desc("id").Limit(500).Find(&rows); err != nil {
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}
	HTTPRetGameData(c, SUCCESS, "success", rows, map[string]interface{}{"totalCount": len(rows)})
}
