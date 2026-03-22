package logic

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/dto"
	"xfx/pkg/log"
)

// GmGetGuildList 获取帮会列表（Guild 与路由 /gm/guild 一致）
func GmGetGuildList(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	log.Debug("获取帮会列表, %v", result)

	serverIdAny, ok := result["serverId"]
	if !ok {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId required")
		return
	}
	serverId, ok := serverIdAny.(float64)
	if !ok || int(serverId) <= 0 {
		HTTPRetGame(c, ERR_ACCOUNT_PARAMS_ERROR, "serverId invalid")
		return
	}

	guildId := int64(0)
	if idAny, ok := result["id"]; ok {
		if idFloat, ok := idAny.(float64); ok {
			guildId = int64(idFloat)
		}
	}

	var guids []model.GuildDB
	query := db.AccountDb.Table(define.GuildTable).Where("server_id = ?", int(serverId))
	if guildId > 0 {
		query = query.Where("id = ?", guildId)
	}
	err := query.Find(&guids)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		HTTPRetGame(c, ERR_DB, err.Error())
		return
	}

	opts := make([]*dto.GMGuidOption, 0)
	for i := 0; i < len(guids); i++ {
		opts = append(opts, &dto.GMGuidOption{
			ServerId:        int32(guids[i].ServerId),
			GuidId:          int32(guids[i].Id),
			GuidName:        guids[i].GuildName,
			GuidGrowth:      int32(0),
			GuidBanner:      guids[i].Banner,
			GuidBannerColor: guids[i].BannerColor,
			GuidNumber:      guids[i].CurMemberCount,
			GuidMaxNumber:   guids[i].MaxMemberCount,
			GuidNotice:      guids[i].NoticeBoard,
			GuidMansterId:   guids[i].Master,
			GuidNeed:        strconv.FormatBool(guids[i].ApplyNeedApproval == 1),
			GuidLimit:       strconv.FormatInt(int64(guids[i].Level), 10),
			GuiLimitLevel:   guids[i].Level,
		})
	}

	HTTPRetGameData(c, SUCCESS, "success", opts, map[string]any{"totalCount": len(opts)})
}
