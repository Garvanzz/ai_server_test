package logic

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strconv"
	"xfx/core/model"
	"xfx/gm_server/db"
	"xfx/gm_server/define"
	"xfx/gm_server/gm_model"
	"xfx/pkg/log"
)

// 获取帮会列表
func GmGetGuidList(c *gin.Context) {
	rawData, _ := c.GetRawData()
	var result map[string]interface{}
	if err := json.Unmarshal(rawData, &result); err != nil {
		log.Error("GmStartServer find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	log.Debug("获取帮会列表, %v", result)

	//serverId, ok := result["serverId"].(string)
	//if !ok {
	//	log.Error("GmGetGuidList find serverId err")
	//	httpRetGame(c, ERR_DB, "GmGetGuidList find serverId err")
	//	return
	//}
	//
	//id, ok := result["id"].(string)
	//if !ok {
	//	log.Error("GmGetGuidList find id err")
	//	httpRetGame(c, ERR_DB, "GmGetGuidList find id err")
	//	return
	//}

	var guids []model.GuildDB
	err := db.AccountDb.Table(define.TableGuild).Find(&guids)
	if err != nil {
		log.Error("getserverlist2 find err :%v", err.Error())
		httpRetGame(c, ERR_DB, err.Error())
		return
	}

	opts := make([]*gm_model.GMGuidOption, 0)
	for i := 0; i < len(guids); i++ {
		opts = append(opts, &gm_model.GMGuidOption{
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

	js, _ := json.Marshal(opts)
	httpRetGame(c, SUCCESS, "success", map[string]any{
		"data":       string(js),
		"totalCount": len(opts),
	})
}
