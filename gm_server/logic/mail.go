package logic

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

func resolveMailReceiverPlayerIDs(entryServerId int, raw string) ([]int64, []string, error) {
	ids := make([]int64, 0)
	resolvedUIDs := make([]string, 0)
	if strings.TrimSpace(raw) == "" {
		return ids, resolvedUIDs, nil
	}
	parts := strings.Split(raw, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		role := new(model.AccountRole)
		has, err := db.AccountDb.Table(define.AccountRoleTable).Where("entry_server_id = ? AND uid = ?", entryServerId, part).Get(role)
		if err != nil {
			return nil, nil, err
		}
		if has {
			if role.RedisId <= 0 {
				return nil, nil, fmt.Errorf("uid %s has no player id on entry server %d", part, entryServerId)
			}
			ids = append(ids, role.RedisId)
			resolvedUIDs = append(resolvedUIDs, part)
			continue
		}

		id, parseErr := strconv.ParseInt(part, 10, 64)
		if parseErr != nil || id <= 0 {
			return nil, nil, fmt.Errorf("uid %s not found on entry server %d", part, entryServerId)
		}
		ids = append(ids, id)
	}
	return ids, resolvedUIDs, nil
}

// 后台邮件创建
func GmCreateAdminMail(c *gin.Context) {
	var Info dto.GmMailInfo
	if err := c.ShouldBindJSON(&Info); err != nil {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err: "+err.Error())
		return
	}

	log.Debug("Info %v", Info)

	// 参数校验
	if Info.Server <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "server required")
		return
	}
	if len(Info.Title) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "title required")
		return
	}
	if len(Info.Content) == 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "content required")
		return
	}

	mail := new(model.GMMailInfo)
	mail.ServerId = resolveLogicServerID(int(Info.Server))
	mail.OriginServerId = int(Info.Server)
	mail.CreatorName = c.ClientIP()
	mail.CnContent = Info.Content
	mail.CnTitle = Info.Title
	mail.CreateTime = time.Now()
	mail.SenderName = Info.SenderName
	mail.Status = 1
	mail.PlayerIds = make([]int64, 0)
	mail.Items = make([]model.MailItem, 0)

	// 类型：system=系统邮件，person=个人邮件
	if Info.Type == "system" {
		mail.Type = 1
	} else if Info.Type == "person" || Info.Type == "persion" {
		mail.Type = 2
	}

	// Uid 优先按入口服上的 uid 列表解析；若找不到则兼容按 player_id 解析
	if len(Info.Uid) > 0 {
		resolvedIDs, _, err := resolveMailReceiverPlayerIDs(int(Info.Server), Info.Uid)
		if err != nil {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
			return
		}
		mail.PlayerIds = append(mail.PlayerIds, resolvedIDs...)
	}

	//立即发送
	if Info.Immediatelysend {
		mail.EffectTime = time.Now()
	} else {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", Info.Delaytime, time.Local)
		if err != nil {
			HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
			return
		}
		mail.EffectTime = t
	}

	//奖励
	if len(Info.Itmes) > 0 {
		items := make([]model.MailItem, 0)
		arr := strings.Split(Info.Itmes, "|")
		for i := 0; i < len(arr); i++ {
			if len(arr[i]) <= 0 {
				continue
			}
			awards := strings.Split(arr[i], ",")
			Type1, err := strconv.ParseInt(awards[0], 10, 32)
			if err != nil {
				HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
				return

			}

			Id1, err := strconv.ParseInt(awards[1], 10, 32)
			if err != nil {
				HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
				return

			}

			Num1, err := strconv.ParseInt(awards[2], 10, 32)
			if err != nil {
				HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
				return

			}

			items = append(items, model.MailItem{
				Id:   int32(Id1),
				Type: int32(Type1),
				Num:  int32(Num1),
			})
		}
		mail.Items = items
	}

	js, _ := json.Marshal(mail)
	// 按区服转发到对应 main_server；Server<=0 时返回错误
	err, respStr := HttpRequestToServer(int(Info.Server), js, "/gm/mail")
	if err != nil {
		log.Error("GmCreateAdminMail forward err: %v, resp: %s", err, respStr)
		HTTPRetGame(c, ERR_SERVER_INTERNAL, "forward failed: "+err.Error())
		return
	}
	HTTPRetGame(c, SUCCESS, "success")
	////全服
	//if Info.Fullserversend {
	//
	//} else {
	//	_, err := AdminEngine.Table(define.AdminMailTable).Insert(mail)
	//	if err != nil {
	//		log.Error("插入错误:%v", err)
	//		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
	//		return
	//	}
	//}

	//HTTPRetGame(c, SUCCESS, "success")
}
