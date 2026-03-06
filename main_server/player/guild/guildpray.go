package guild

import (
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_guild"
)

// ReqGuildPray 帮会祈福
func ReqGuildPray(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SGuildPray) {
	resp := new(proto_guild.S2CGuildPray)

	//获取自己的帮派数据
	info, err := invoke.GuildClient(ctx).GuildPray(pl.ToContext(), req.Index)
	if err != nil {
		log.Error("invoke guild error:%v", err)
		resp.Code = proto_guild.GuildCode_ERROR_ALGET
		ctx.Send(resp)
		return
	}

	if info == nil {
		log.Error("invoke guild error is nil")
		resp.Code = proto_guild.GuildCode_ERROR_ALGET
		ctx.Send(resp)
		return
	}

	resp.Code = proto_guild.GuildCode_ERROR_OK
	resp.PrayItem = &proto_guild.GuildPrayItem{
		IsTodayPray: info.GuildPray.IsTodayPray,
		Index:       info.GuildPray.PrayType,
		RangType:    info.GuildPray.RangeType,
		RangValue:   info.GuildPray.RangeValue,
	}
	ctx.Send(resp)
}
