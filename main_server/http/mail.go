package http

import (
	"github.com/gin-gonic/gin"
	"time"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
)

// 邮件
func (m *HttpModule) GMSendMail(c *gin.Context) {
	var Info model.GMMailInfo
	if err := c.ShouldBindJSON(&Info); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" Info %v", Info)

	awards := make([]conf.ItemE, 0)
	for i := 0; i < len(Info.Items); i++ {
		awards = append(awards, conf.ItemE{
			ItemId:   Info.Items[i].Id,
			ItemType: Info.Items[i].Type,
			ItemNum:  Info.Items[i].Num,
		})
	}
	//判断类型
	//延时
	if Info.EffectTime.Unix() > time.Now().Unix() {
		//取ID
		id, _ := db.CommonEngine.GetDelayMailId()

		isSuc := invoke.MailClient(m).SendDelayMails(int64(id), Info.EffectTime.Unix(), Info.Type, Info.CnTitle, Info.CnContent, "", "", awards, int64(0), int32(0), []string{}, Info.PlayerIds)
		if !isSuc {
			m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
			return
		}
	} else {
		isSuc := invoke.MailClient(m).SendMail(Info.Type, Info.CnTitle, Info.CnContent, "", "", Info.SenderName, awards, Info.PlayerIds, int64(0), int32(0), false, []string{})
		if !isSuc {
			m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
			return
		}
	}
}
