package http

import (
	"github.com/gin-gonic/gin"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_public"
)

// 公告
func (m *HttpModule) GMSendNotice(c *gin.Context) {
	var Info model.NoticeOpt
	if err := c.ShouldBindJSON(&Info); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" main server Info %v", Info)
	msg := &proto_public.S2CNotice{
		Id:         int32(Info.Id),
		Channel:    Info.Channel,
		ServerId:   Info.ServerId,
		Content:    Info.Content,
		ExpireTime: Info.ExpireTime,
		EffectTime: Info.EffectTime,
	}
	invoke.DispatchAllPlayer(m, msg)

	m.httpRetGame(c, SUCCESS, "success")
}

// 跑马灯
func (m *HttpModule) GMSendHorse(c *gin.Context) {
	var Info model.HorseOpt
	if err := c.ShouldBindJSON(&Info); err != nil {
		m.httpRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "params err1")
		return
	}

	log.Debug(" main server horse Info %v", Info)

	msg := &proto_public.S2CHorseOption{
		Id:         int32(0),
		Channel:    Info.Channel,
		ServerId:   Info.ServerId,
		Content:    Info.Content,
		ExpireTime: utils.Now().Unix() + int64(Info.VaildTime),
		EffectTime: utils.Now().Unix(),
		Scene:      Info.Scene,
		Priority:   Info.Priority,
		Type:       define.HorseType_System,
	}
	invoke.DispatchAllPlayer(m, msg)

	m.httpRetGame(c, SUCCESS, "success")
}
