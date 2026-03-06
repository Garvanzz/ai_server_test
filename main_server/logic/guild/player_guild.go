package guild

import (
	"time"
	"xfx/core/model"
)

func newPlayerGuild(dbId int64) *model.PlayerGuild {
	playerGuild := &model.PlayerGuild{
		Id:              dbId,
		LastRefreshTime: time.Now().Unix(),
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
	info.LastQuitTime = time.Now().Unix()
	info.GuildMap = make(map[int32]*model.GuildMapItem)
}
