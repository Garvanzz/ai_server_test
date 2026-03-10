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
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" Info %v", Info)
	mail := new(model.GMMailInfo)
	mail.CreatorName = c.ClientIP()
	mail.CnContent = Info.Content
	mail.CnTitle = Info.Title
	mail.CreateTime = time.Now()
	mail.SenderName = Info.SenderName
	mail.Status = 1
	mail.PlayerIds = make([]int64, 0)
	mail.Items = make([]model.MailItem, 0)

	//类型
	if Info.Type == "system" {
		mail.Type = 1
	} else if Info.Type == "persion" {
		mail.Type = 2
	}

	//uid
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
			httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
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
				httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
				return

			}

			Id1, err := strconv.ParseInt(awards[1], 10, 32)
			if err != nil {
				httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
				return

			}

			Num1, err := strconv.ParseInt(awards[2], 10, 32)
			if err != nil {
				httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
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
	//通知游戏服
	err, _ := HttpRequest(js, "/gm/mail")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, "fail")
	} else {
		httpRetGame(c, SUCCESS, "success")
	}
	////全服
	//if Info.Fullserversend {
	//
	//} else {
	//	_, err := AdminEngine.Table(define.AdminMailTable).Insert(mail)
	//	if err != nil {
	//		log.Error("插入错误:%v", err)
	//		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
	//		return
	//	}
	//}

	//httpRetGame(c, SUCCESS, "success")
}
