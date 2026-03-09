package logic

import (
	"encoding/json"
	"fmt"
	"slices"

	"xfx/login_server/define"
	"xfx/login_server/internal/middleware"
	"xfx/login_server/model/dto"
	"xfx/login_server/model/entity"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

// GetServerList 获取区服列表
func GetServerList(c *gin.Context) {
	var req dto.ReqServerList
	_ = c.ShouldBindJSON(&req)
	if req.Channel > 0 {
		log.Debug("GetServerList channel=%d (attribution only)", req.Channel)
	}

	metaList := make([]entity.ServerGroupMeta, 0)
	_ = AccountEngine.Table(define.TableServerGroup).Asc("sort_order", "id").Find(&metaList)

	metaMap := make(map[int64]entity.ServerGroupMeta)
	sortOrders := make(map[int]int)
	for i, m := range metaList {
		metaMap[m.Id] = m
		sortOrders[int(m.Id)] = m.SortOrder*1000 + i
	}

	items := make([]entity.ServerItem, 0)
	err := AccountEngine.Table(define.TableGameServer).Asc("group_id", "id").Find(&items)
	if err != nil {
		log.Error("get server list find err :%v", err.Error())
		middleware.RetGame(c, define.ERR_DB, err.Error())
		return
	}

	groups := make(map[int][]dto.LoginServerItem)
	groupIds := make([]int, 0)

	for _, v := range items {
		gid := v.GroupId
		if _, ok := groups[gid]; !ok {
			groups[gid] = make([]dto.LoginServerItem, 0)
			groupIds = append(groupIds, gid)
		}
		groups[gid] = append(groups[gid], dto.LoginServerItem{
			Id:             v.Id,
			Ip:             v.Ip,
			Port:           v.Port,
			Channel:        v.Channel,
			ServerState:    v.ServerState,
			OpenServerTime: v.OpenServerTime,
			StopServerTime: v.StopServerTime,
			ServerName:     v.ServerName,
			GroupId:        gid,
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

	serverMap := make([]dto.ServerGroupResp, 0)
	for _, gid := range groupIds {
		meta, ok := metaMap[int64(gid)]
		name := fmt.Sprintf("服务器组 %d", gid)
		if ok && meta.Name != "" {
			name = meta.Name
		}
		serverMap = append(serverMap, dto.ServerGroupResp{
			Group:   int32(gid),
			Name:    name,
			Servers: groups[gid],
		})
	}
	js, _ := json.Marshal(serverMap)

	middleware.RetGame(c, define.SUCCESS, "success", map[string]interface{}{
		"ServerList": string(js),
	})
}
