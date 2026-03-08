package guild

import (
	"xfx/core/model"
	"xfx/pkg/utils"
)

func newPlayerGuild(dbId int64) *model.PlayerGuild {
	playerGuild := &model.PlayerGuild{
		Id:              dbId,
		LastRefreshTime: utils.Now().Unix(),
		ToDaySign:       false,
		SignDay:         0,
		GuildMap:        make(map[int32]*model.GuildMapItem),
		GuildPray:       new(model.GuildPrayItem),
	}
	return playerGuild
}

// 离开帮会清除玩家帮会信息
func clearGuildPlayer(info *model.PlayerGuild) {
	info.GuildId = 0
	info.LastQuitTime = utils.Now().Unix()
	info.GuildMap = make(map[int32]*model.GuildMapItem)
}
