package guild

import (
	"errors"
	"xfx/core/config"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/utils"
	"xfx/proto/proto_player"
)

// 祈福
func (mgr *Manager) guildPray(ctx *proto_player.Context, index int32) (*model.PlayerGuild, error) {
	//获取帮派信息
	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return nil, errors.New("no guild")
	}

	if info.GuildId == 0 {
		return nil, errors.New("no guildId")
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return nil, errors.New("no guildId")
	}

	if info.GuildPray == nil {
		info.GuildPray = new(model.GuildPrayItem)
	}

	if !utils.CheckIsSameDayBySec(info.GuildPray.TodayPrayTime, utils.Now().Unix(), 0) {
		info.GuildPray.IsTodayPray = false
	}

	if info.GuildPray.IsTodayPray {
		return nil, errors.New("today is pray")
	}

	info.GuildPray.TodayPrayTime = utils.Now().Unix()
	info.GuildPray.IsTodayPray = true
	info.GuildPray.PrayType = index
	if index == 3 {
		conf, _ := config.GuildTitle.Find(int64(ent.guild.Props[define.GuildPropTitle]))
		rangVal := utils.RandInt(1, len(conf.PrayRangeType)+1)
		info.GuildPray.RangeType = int32(rangVal)
		info.GuildPray.RangeValue = conf.PrayRangeValue[rangVal-1]
	}
	return info, nil
}
