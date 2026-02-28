package logic

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"slices"
	"xfx/login_server/define"
	"xfx/login_server/model"
	"xfx/pkg/log"
)

// GetServerList 获取服务器列表
func GetServerList(c *gin.Context) {
	var req model.ReqServerList
	if err := c.ShouldBindJSON(&req); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err")
		return
	}
	log.Debug("请求获取服务器列表 :%v", req)
	if req.Channel < 1 || req.Channel > 100 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "platform type error")
		return
	}

	//p := new(model.ServerItem)
	//if has, _ := ServerGroupEngine.IsTableExist(define.ServerGroup); !has {
	//	// 同步结构体与数据库表
	//	err := ServerGroupEngine.Sync2(p)
	//	if err != nil {
	//		log.Error("Failed to sync database: %v", err)
	//	}
	//}

	items := make([]model.ServerItem, 0)
	err := AccountEngine.Table(define.ServerGroup).Where("channel = ?", req.Channel).Find(&items)
	if err != nil {
		log.Error("get server list find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	log.Debug("获取服务器列表 :%v", items)

	groups := make(map[int][]model.LoginServerItem)
	ids := make([]int, 0)
	groupNames := make(map[int]string) // 使用 map 存储组名

	for _, v := range items {
		// 假设 ServerItem 中有 ServerGroupName 字段
		groupNames[v.ServerGroup] = v.ServerName // 或者使用其他字段

		// 获取游戏服
		var gameItem model.GameServerItem
		// TODO: 从数据库获取 gameItem

		if _, ok := groups[v.ServerGroup]; !ok {
			groups[v.ServerGroup] = make([]model.LoginServerItem, 0)
			ids = append(ids, v.ServerGroup) // 只在首次遇到该组时添加到 ids
		}

		groups[v.ServerGroup] = append(groups[v.ServerGroup], model.LoginServerItem{
			Id:             v.Id,
			Ip:             v.Ip,
			Port:           v.Port,
			Channel:        v.Channel,
			RedisPort:      v.RedisPort,
			MysqlAddr:      v.MysqlAddr,
			ServerState:    v.ServerState,
			OpenServerTime: v.OpenServerTime,
			StopServerTime: v.StopServerTime,
			ServerName:     v.ServerName,
			LoginServerUrl: v.LoginServerUrl,
			ServerGroup:    v.ServerGroup,
			UdpIp:          gameItem.Ip,
			UdpPort:        gameItem.Port,
		})
	}

	// 排序
	slices.Sort(ids)

	// 重组
	serverMap := make([]model.ServerGroup, 0)
	for _, id := range ids {
		name := groupNames[id]
		if name == "" {
			name = fmt.Sprintf("服务器组 %d", id) // 默认名称
		}

		serverMap = append(serverMap, model.ServerGroup{
			Group:   int32(id),
			Name:    name,
			Servers: groups[id],
		})
	}
	js, _ := json.Marshal(serverMap)

	//获取上一次登录的服务器
	lastServerId := 0
	if len(req.Account) > 0 {
		player := new(model.Account)
		player.Account = req.Account
		player.Password = req.Password
		has, err := AccountEngine.Table(define.AccountTable).Get(player)
		if err != nil {
			log.Error("account failed, err : %v", err)
			httpRetGame(c, ERR_DB, "db err")
			return
		}

		if has {
			//获得redisPort
			var serverItem model.ServerItem
			has, err = AccountEngine.Table(define.ServerGroup).Where("id = ?", player.ServerId).Get(&serverItem)
			if err != nil {
				log.Error("getlastloginserver find err :%v", err.Error())
				httpRetGame(c, ERR_DB, err.Error())
				return
			}
			//Id大于0才可以取
			if player.RedisId > 0 {
				//获取上一次登陆服务器
				reply, err := RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerLastLoginServer, player.RedisId))
				if err != nil {
					log.Error("player[%v],load bag error:%v", player.RedisId, err)
					return
				}

				if reply != nil {
					lastServerId, err = redis.Int(reply, nil)
					if err != nil {
						log.Error("player[%v],load lastServerId error:%v", lastServerId, err)
						return
					}
					log.Debug("上一次登录服务器:%d, %d", player.RedisId, lastServerId)
				}
			}
		}
	}

	//历史登录服
	historyIds := make([]int, 0)
	historyIds = append(historyIds, lastServerId)

	httpRetGame(c, SUCCESS, "success", map[string]interface{}{
		"ServerList":         string(js),
		"LastLoginServer":    lastServerId,
		"HistoryLoginServer": historyIds,
	})
}
