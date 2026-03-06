package guild

import (
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_guild"
)

// ReqGuildSign 帮会建造
func ReqGuildBuild(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SBuildGuildMap) {
	resp := new(proto_guild.S2CBuildGuildMap)

	if req.Id <= 0 {
		log.Error("ID <=0")
		resp.Code = proto_guild.GuildCode_ERROR_INDEXHAD
		ctx.Send(resp)
		return
	}

	//获取自己的帮派数据
	info, err := invoke.GuildClient(ctx).SetBuild(pl.ToContext(), req.Id, req.Index)
	if err != nil {
		log.Error("invoke guild error:%v", err)
		resp.Code = proto_guild.GuildCode_ERROR_INDEXHAD
		ctx.Send(resp)
		return
	}

	if info == nil {
		log.Error("invoke guild error is nil")
		resp.Code = proto_guild.GuildCode_ERROR_INDEXHAD
		ctx.Send(resp)
		return
	}

	log.Debug("info:%v", info)
	resp = info
	ctx.Send(resp)
}

// ReqGuildMap 获取帮会信息
func ReqGuildBuildMapInfo(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SGetBuildMapInfo) {
	resp := new(proto_guild.S2CGetBuildMapInfo)

	if req.GuildId <= 0 {
		log.Error("GuildId <=0")
		resp.Items = make(map[int32]*proto_guild.GuildBuildMapItem)
		ctx.Send(resp)
		return
	}

	if req.PlayerId <= 0 {
		log.Error("PlayerId <=0")
		resp.Items = make(map[int32]*proto_guild.GuildBuildMapItem)
		ctx.Send(resp)
		return
	}

	//获取帮派地图数据
	info, err := invoke.GuildClient(ctx).GetGuildMapInfo(pl.ToContext(), req.PlayerId, req.GuildId)
	if err != nil {
		log.Error("invoke guild error:%v", err)
		resp.Items = make(map[int32]*proto_guild.GuildBuildMapItem)
		ctx.Send(resp)
		return
	}

	if info == nil {
		log.Error("invoke guild error is nil")
		resp.Items = make(map[int32]*proto_guild.GuildBuildMapItem)
		ctx.Send(resp)
		return
	}

	log.Debug("info:%v", info)
	resp = info
	ctx.Send(resp)
}
