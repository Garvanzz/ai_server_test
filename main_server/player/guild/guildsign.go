package guild

import (
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_guild"
)

// ReqGuildSign 帮会签到数据
func ReqGuildSign(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SGuildSign) {
	resp := new(proto_guild.S2CGuildSign)

	//获取自己的帮派数据
	info, err := invoke.GuildClient(ctx).GuildSign(pl.Id)
	if err != nil {
		log.Error("invoke guild error:%v", err)
		resp.Result = false
		ctx.Send(resp)
		return
	}

	if info == nil {
		log.Error("invoke guild error is nil")
		resp.Result = false
		ctx.Send(resp)
		return
	}

	log.Debug("info:%v", info)

	resp.SignItem = &proto_guild.GuildSignItem{
		Sign:        info.SignDay,
		IsTodaySign: info.ToDaySign,
	}
	ctx.Send(resp)
}
