package logic

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"xfx/core/model"
	"xfx/gm_server/dto"
	"xfx/pkg/log"

	"github.com/gin-gonic/gin"
)

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

	// Uid 格式为 player_id 列表，用 | 分隔（如 "123|456"），与 main_server 发件接口一致
	if len(Info.Uid) > 0 {
		uids := strings.Split(Info.Uid, "|")
		for i := 0; i < len(uids); i++ {
			if len(uids[i]) > 0 {
				id, err := strconv.ParseInt(uids[i], 10, 64)
				if err != nil {
					continue
				}

				mail.PlayerIds = append(mail.PlayerIds, id)
			}
		}
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
