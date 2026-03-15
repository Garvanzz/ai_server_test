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

	js, _ := json.Marshal(typed)
	HTTPRetGame(c, SUCCESS, "success", map[string]interface{}{"data": string(js), "totalCount": len(typed)})
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

	HTTPRetGame(c, SUCCESS, "success", map[string]interface{}{"planId": plan.Id})
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

	conflicts := make([]map[string]interface{}, 0)
	for _, sid := range req.SourceServerIds {
		rows, err := listGuildNameConflicts(req.TargetServerId, sid)
		if err != nil {
			HTTPRetGame(c, ERR_DB, err.Error())
			return
		}
		for _, g := range rows {
			conflicts = append(conflicts, map[string]interface{}{
				"serverId":    sid,
				"type":        "guild_name",
				"key":         g.GuildName,
				"suggestName": fmt.Sprintf("%s_S%d", g.GuildName, sid),
			})
		}
	}

	js, _ := json.Marshal(conflicts)
	HTTPRetGame(c, SUCCESS, "success", map[string]interface{}{
		"targetServerId":  req.TargetServerId,
		"sourceServerIds": req.SourceServerIds,
		"conflictCount":   len(conflicts),
		"conflicts":       string(js),
	})
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

		tables := []string{define.GuildTable, define.GuildApplyTable, define.GuildLogTable, define.PlayerMailInfoTable, define.FriendApplyTable, define.FriendBlockTable, define.AccountTable, define.PayOrderTable, define.PayCacheOrderTable}
		for _, tb := range tables {
			q := fmt.Sprintf("UPDATE %s SET server_id = ? WHERE server_id = ?", tb)
			if _, err = db.AccountDb.Exec(q, plan.TargetServerId, logicSid); err != nil {
				success = false
				serverOK = false
			}
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
		HTTPRetGame(c, SUCCESS, "success")
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

	HTTPRetGame(c, SUCCESS, "success")
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
	js, _ := json.Marshal(rows)
	HTTPRetGame(c, SUCCESS, "success", map[string]interface{}{"data": string(js), "totalCount": len(rows)})
}
