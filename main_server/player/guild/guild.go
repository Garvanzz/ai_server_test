package guild

import (
	"xfx/core/config"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_guild"
)

// ReqSetGuildRule 设置帮会信息
func ReqSetGuildRule(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SSetGuildRule) {
	resp := invoke.GuildClient(ctx).SetGuildInfo(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqImpeachMaster 弹劾会长
func ReqImpeachMaster(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SImpeachMaster) {
	resp := invoke.GuildClient(ctx).ImpeachMaster(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqLeaveGuild 离开帮会
func ReqLeaveGuild(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SLeaveGuild) {
	resp := invoke.GuildClient(ctx).LeaveGuild(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqKickOutMember 帮会踢人
func ReqKickOutMember(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SKickOutMember) {
	resp := invoke.GuildClient(ctx).KickOutMember(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqDealApply 处理请求
func ReqDealApply(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SDealApply) {
	resp := invoke.GuildClient(ctx).DealApply(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqAssignPosition 任命职位
func ReqAssignPosition(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SAssignPosition) {
	resp := invoke.GuildClient(ctx).AssignPosition(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqGuildEvents 获取帮会日志
func ReqGuildEvents(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SGuildEvent) {
	resp := invoke.GuildClient(ctx).GetEvents(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqJoinGuild 加入帮会
func ReqJoinGuild(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SJoinGuild) {

	if req.Id == 0 {
		log.Error("join guild id is 0")
		return
	}

	resp := invoke.GuildClient(ctx).JoinGuild(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqMemberList 请求帮会成员列表
func ReqMemberList(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SGetMemberList) {
	resp := invoke.GuildClient(ctx).GetMemberList(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqGuildApplyList 获取帮会申请列表
func ReqGuildApplyList(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SGetApply) {
	resp := invoke.GuildClient(ctx).GetGuildApplyList(pl.ToContext())
	ctx.Send(resp)
}

// ReqSearchGuildByName  根据名字搜索帮会
func ReqSearchGuildByName(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SSearchByName) {
	if req.Name == "" {
		log.Error("searchGuildByName name is empty")
		return
	}

	resp := invoke.GuildClient(ctx).SearchGuildByName(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqGuildListByPage 根据页数获取帮会信息
func ReqGuildListByPage(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SGuildByPage) {
	if req.Page <= 0 {
		log.Error("mgr get guild by page,page error")
		return
	}

	resp := invoke.GuildClient(ctx).GetGuildListByPage(pl.ToContext(), req)
	ctx.Send(resp)
}

// ReqGuildData 获取帮会数据
func ReqGuildData(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SGetGuildInfo) {
	resp := new(proto_guild.S2CGetGuildInfo)

	infos := invoke.GuildClient(ctx).GetGuildInfoByIds([]int64{int64(req.Id)})
	resp.Guild = infos[int64(req.Id)]
	ctx.Send(resp)
}

// ReqGuildChangeName 帮会改名
func ReqGuildChangeName(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SChangeGuildName) {
	resp := new(proto_guild.S2CChangeGuildName)

	conf := config.Global.Get().GuildRename
	costs := make(map[int32]int32)
	costs[conf[0].ItemId] = conf[0].ItemNum
	if !internal.CheckItemsEnough(pl, costs) {
		log.Error("ReqGuildChangeName is faild item not enough")
		ctx.Send(resp)
		return
	}
	resp = invoke.GuildClient(ctx).ChangeGuildName(pl.ToContext(), req.Name)

	ctx.Send(resp)
}

// ReqCreateGuild 创建帮会
func ReqCreateGuild(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SCreateGuild) {
	resp := new(proto_guild.S2CCreateGuild)

	// 检查消耗是否足够
	consumeItems := config.Global.Get().CreateGuildConsume
	consume := global.MergeItemE(consumeItems)
	log.Debug("创建公会消耗:%v", consume)
	if !internal.CheckItemsEnough(pl, consume) {
		log.Error("create guild item not enough")
		ctx.Send(resp)
		return
	}

	// 检查是否有对应的配置
	resp.Result = invoke.GuildClient(ctx).CreateGuild(pl.ToContext(), req)
	// 扣除道具
	if resp.Result {
		internal.SubItems(ctx, pl, consume)
	}

	if !resp.Result {
		log.Error("ReqCreateGuild error")
		return
	}

	ctx.Send(resp)
}

// ReqDissolveGuild TODO:解散帮会
func ReqDissolveGuild(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SDissolveGuild) {
	//resp := new(proto_guild.S2CDissolveGuild)
	//
	//reply, err := ctx.Invoke(define.ModuleGuild, "getGuildListByPage", pl.ToContext(), req)
	//if err != nil {
	//	log.Error("invoke guild error:%v", err)
	//	return
	//}
	//
	//resp = reply.(*proto_guild.S2CDissolveGuild)
	//ctx.Send(resp)}
}

// ReqPlayerGuildDetail 获取玩家帮会数据
func ReqPlayerGuildDetail(ctx global.IPlayer, pl *model.Player, req *proto_guild.C2SPlayerGuildDetail) {
	resp := invoke.GuildClient(ctx).PlayerGuildDetail(pl.ToContext(), req)
	ctx.Send(resp)
}
