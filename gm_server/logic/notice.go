package logic

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/pkg/log"
)

// 发送公告
func GmSendNotice(c *gin.Context) {
	var Info model.NoticeItem
	if err := c.ShouldBindJSON(&Info); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" Info %v", Info)

	if len(Info.Content) <= 0 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	//入库
	item := model.NoticeOpt{
		Channel:  Info.Channel,
		ServerId: Info.ServerId,
		Content:  Info.Content,
		//Title:      Info.Title,
		ExpireTime: Info.ExpireTime,
		EffectTime: time.Now().Unix(),
	}

	_, err := db.AccountDb.Table(define.NoticeTable).Insert(item)
	if err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}

	// TODO:
	//if Info.IsImmediately {
	//	// 通知游戏服（main_server）
	//	js, _ := json.Marshal(item)
	//	err, _ := HttpRequest(js, "/gm/notice")
	//	if err != nil {
	//		httpRetGame(c, ERR_SERVER_INTERNAL, "fail")
	//		return
	//	} else {
	//		httpRetGame(c, SUCCESS, "success")
	//		return
	//	}
	//}

	httpRetGame(c, SUCCESS, "success")
}

// 发送跑马灯
func GmSendHorse(c *gin.Context) {
	var Info model.HorseItem
	if err := c.ShouldBindJSON(&Info); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" horse Info %v", Info)

	if len(Info.Content) <= 0 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	js, _ := json.Marshal(Info)
	// 通知游戏服（main_server）
	err, _ := HttpRequest(js, "/gm/horse")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, "fail")
		return
	} else {
		httpRetGame(c, SUCCESS, "success")
		return
	}
}
