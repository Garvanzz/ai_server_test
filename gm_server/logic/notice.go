package logic

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"time"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
)

// 发送公告
func GmSendNotice(c *gin.Context) {
	var Info gm_model.NoticeItem
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
	item := gm_model.NoticeOpt{
		Channel:    Info.Channel,
		ServerId:   Info.ServerId,
		Content:    Info.Content,
		Title:      Info.Title,
		ExpireTime: Info.ExpireTime,
		EffectTime: time.Now().Unix(),
	}

	_, err := db.AccountDb.Table(define.Notice).Insert(item)
	if err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, err.Error())
		return
	}

	if Info.IsImmediately {
		//通知大厅
		js, _ := json.Marshal(item)
		//通知游戏服
		err, _ := HttpRequest(js, "GMSendNotice")
		if err != nil {
			httpRetGame(c, ERR_SERVER_INTERNAL, "fail")
			return
		} else {
			httpRetGame(c, SUCCESS, "success")
			return
		}
	}

	httpRetGame(c, SUCCESS, "success")
}

// 发送跑马灯
func GmSendHorse(c *gin.Context) {
	var Info gm_model.HorseItem
	if err := c.ShouldBindJSON(&Info); err != nil {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" horse Info %v", Info)

	if len(Info.Content) <= 0 {
		httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	//通知大厅
	js, _ := json.Marshal(Info)
	//通知游戏服
	err, _ := HttpRequest(js, "GMSendHorse")
	if err != nil {
		httpRetGame(c, ERR_SERVER_INTERNAL, "fail")
		return
	} else {
		httpRetGame(c, SUCCESS, "success")
		return
	}
}
