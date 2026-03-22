package logic

import (
	"encoding/json"
	"fmt"
	"slices"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/login_server/dto"

	"xfx/login_server/internal/middleware"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

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

// GetServerList 获取区服列表
func GetServerList(c *gin.Context) {
	var req dto.ServerListRequest
	_ = c.ShouldBindJSON(&req)
	if req.Channel > 0 {
		log.Debug("GetServerList channel=%d", req.Channel)
	}

	metaList := make([]model.ServerGroup, 0)
	_ = AccountEngine.Table(define.ServerGroupTable).Asc("sort_order", "id").Find(&metaList)

	metaMap := make(map[int64]model.ServerGroup)
	sortOrders := make(map[int]int)
	for i, m := range metaList {
		metaMap[m.Id] = m
		sortOrders[int(m.Id)] = m.SortOrder*1000 + i
	}

	items := make([]model.ServerItem, 0)
	err := AccountEngine.Table(define.GameServerTable).Where("group_id > ?", 0).Asc("group_id", "id").Find(&items)
	if err != nil {
		log.Error("get server list find err :%v", err.Error())
		middleware.RetGame(c, dto.ERR_DB, err.Error())
		return
	}

	groups := make(map[int][]dto.ServerListItem)
	groupIds := make([]int, 0)

	for _, v := range items {
		gid := v.GroupId
		if _, ok := groups[gid]; !ok {
			groups[gid] = make([]dto.ServerListItem, 0)
			groupIds = append(groupIds, gid)
		}
		groups[gid] = append(groups[gid], dto.ServerListItem{
			ID:             v.Id,
			LogicServerID:  v.LogicServerId,
			MergeState:     v.MergeState,
			MergeStateText: mergeStateToText(v.MergeState),
			MergeTime:      v.MergeTime,
			IP:             v.Ip,
			Port:           v.Port,
			Channel:        v.Channel,
			ServerState:    v.ServerState,
			OpenServerTime: v.OpenServerTime,
			StopServerTime: v.StopServerTime,
			ServerName:     v.ServerName,
			GroupID:        gid,
		})
	}

	slices.SortFunc(groupIds, func(a, b int) int {
		orderA := sortOrders[a]
		orderB := sortOrders[b]
		if orderA != orderB {
			return orderA - orderB
		}
		return a - b
	})

	serverMap := make([]dto.ServerGroupResponse, 0)
	for _, gid := range groupIds {
		meta, ok := metaMap[int64(gid)]
		name := fmt.Sprintf("服务器组 %d", gid)
		if ok && meta.Name != "" {
			name = meta.Name
		}
		serverMap = append(serverMap, dto.ServerGroupResponse{
			Group:   int32(gid),
			Name:    name,
			Servers: groups[gid],
		})
	}
	js, _ := json.Marshal(serverMap)

	middleware.RetGameData(c, dto.SUCCESS, "success", dto.ServerListResponse{ServerList: serverMap}, map[string]interface{}{
		"serverList": serverMap,
		"ServerList": string(js),
	})
}
