package invoke

import (
	"github.com/golang/protobuf/proto"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_guild"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"
	"xfx/proto/proto_rank"
)

type GuildModClient struct {
	invoke Invoker
	Type   string
}

func GuildClient(invoker Invoker) GuildModClient {
	return GuildModClient{
		invoke: invoker,
		Type:   define.ModuleGuild,
	}
}

// SearchGuildByName 根据帮会名搜索帮会
func (m GuildModClient) SearchGuildByName(ctx *proto_player.Context, req *proto_guild.C2SSearchByName) *proto_guild.S2CSearchByName {
	result, err := m.invoke.Invoke(m.Type, "searchGuildByName", ctx, req)
	if err != nil {
		log.Error("SearchGuildByName error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CSearchByName)
}

// GetGuildListByPage 根据页数获取帮会列表
func (m GuildModClient) GetGuildListByPage(ctx *proto_player.Context, req *proto_guild.C2SGuildByPage) *proto_guild.S2CGuildByPage {
	result, err := m.invoke.Invoke(m.Type, "getGuildListByPage", ctx, req)
	if err != nil {
		log.Error("GetGuildListByPage error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CGuildByPage)
}

// CreateGuild 创建帮会
func (m GuildModClient) CreateGuild(ctx *proto_player.Context, req *proto_guild.C2SCreateGuild) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "createGuild", ctx, req))
	return result
}

// JoinGuild 直接加入帮会 0/加入失败 1/直接加入帮会 2/发送帮会申请
func (m GuildModClient) JoinGuild(ctx *proto_player.Context, req *proto_guild.C2SJoinGuild) *proto_guild.S2CJoinGuild {
	result, err := m.invoke.Invoke(m.Type, "joinGuild", ctx, req)
	if err != nil {
		log.Error("GetGuildListByPage error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CJoinGuild)
}

// DealApply 处理帮会申请 1同意 2拒绝
func (m GuildModClient) DealApply(ctx *proto_player.Context, req *proto_guild.C2SDealApply) *proto_guild.S2CDealApply {
	result, err := m.invoke.Invoke(m.Type, "DealApply", ctx, req)
	if err != nil {
		log.Error("DealApply error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CDealApply)
}

// GetGuildApplyList 获取帮会申请列表
func (m GuildModClient) GetGuildApplyList(ctx *proto_player.Context) *proto_guild.S2CGetApply {
	result, err := m.invoke.Invoke(m.Type, "getGuildApplyList", ctx)
	if err != nil {
		log.Error("GetGuildApplyList error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CGetApply)
}

// GetEvents 获取帮会事件
func (m GuildModClient) GetEvents(ctx *proto_player.Context, req *proto_guild.C2SGuildEvent) *proto_guild.S2CGuildEvent {
	result, err := m.invoke.Invoke(m.Type, "getEvents", ctx, req)
	if err != nil {
		log.Error("GetEvents error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CGuildEvent)
}

// GetMemberList 获取帮会成员列表
func (m GuildModClient) GetMemberList(ctx *proto_player.Context, req *proto_guild.C2SGetMemberList) *proto_guild.S2CGetMemberList {
	result, err := m.invoke.Invoke(m.Type, "getMemberList", ctx, req)
	if err != nil {
		log.Error("GetMemberList error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CGetMemberList)
}

// KickOutMember 踢出帮会
func (m GuildModClient) KickOutMember(ctx *proto_player.Context, req *proto_guild.C2SKickOutMember) *proto_guild.S2CKickOutMember {
	result, err := m.invoke.Invoke(m.Type, "kickOutMember", ctx, req)
	if err != nil {
		log.Error("KickOutMember error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CKickOutMember)
}

// LeaveGuild 离开帮会
func (m GuildModClient) LeaveGuild(ctx *proto_player.Context, req *proto_guild.C2SLeaveGuild) *proto_guild.S2CLeaveGuild {
	result, err := m.invoke.Invoke(m.Type, "leaveGuild", ctx, req)
	if err != nil {
		log.Error("LeaveGuild error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CLeaveGuild)
}

// ImpeachMaster 弹劾会长
func (m GuildModClient) ImpeachMaster(ctx *proto_player.Context, req *proto_guild.C2SImpeachMaster) *proto_guild.S2CImpeachMaster {
	result, err := m.invoke.Invoke(m.Type, "impeachMaster", ctx, req)
	if err != nil {
		log.Error("ImpeachMaster error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CImpeachMaster)
}

// AssignPosition 任命职位
func (m GuildModClient) AssignPosition(ctx *proto_player.Context, req *proto_guild.C2SAssignPosition) *proto_guild.S2CAssignPosition {
	result, err := m.invoke.Invoke(m.Type, "assignPosition", ctx, req)
	if err != nil {
		log.Error("AssignPosition error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CAssignPosition)
}

// SetGuildInfo 设置帮会信息
func (m GuildModClient) SetGuildInfo(ctx *proto_player.Context, req *proto_guild.C2SSetGuildRule) *proto_guild.S2CSetGuildRule {
	result, err := m.invoke.Invoke(m.Type, "setGuildInfo", ctx, req)
	if err != nil {
		log.Error("SetGuildInfo error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CSetGuildRule)
}

// PlayerGuildDetail 玩家公会信息
func (m GuildModClient) PlayerGuildDetail(ctx *proto_player.Context, req *proto_guild.C2SPlayerGuildDetail) *proto_guild.S2CPlayerGuildDetail {
	result, err := m.invoke.Invoke(m.Type, "playerGuildDetail", ctx, req)
	if err != nil {
		log.Error("PlayerGuildDetail error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CPlayerGuildDetail)
}

// UpdateMemInfo 更新玩家信息
func (m GuildModClient) UpdateMemInfo(ctx *proto_player.Context) {
	_, err := m.invoke.Invoke(m.Type, "updateMemInfo", ctx)
	if err != nil {
		log.Error("UpdateMemInfo error:%v", err)
	}
}

// ChangeGuildName 更改公会名
func (m GuildModClient) ChangeGuildName(ctx *proto_player.Context, name string) *proto_guild.S2CChangeGuildName {
	result, err := m.invoke.Invoke(m.Type, "changeGuildName", ctx, name)
	if err != nil {
		log.Error("ChangeGuildName error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_guild.S2CChangeGuildName)
}

// OnlineBoardCast 帮会广播
func (m GuildModClient) OnlineBoardCast(guildId int64, message proto.Message) {
	_, err := m.invoke.Invoke(m.Type, "onlineBoardCast", guildId, message)
	if err != nil {
		log.Error("OnlineBoardCast error:%v", err)
		return
	}
}

// GetGuildInfoByPlayerId 根据玩家id获取帮会信息
func (m GuildModClient) GetGuildInfoByPlayerId(dbId int64) *model.Guild {
	result, err := m.invoke.Invoke(m.Type, "getGuildInfoByPlayerId", dbId)
	if err != nil {
		log.Error("GetGuildInfoByPlayerId error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*model.Guild)
}

// GetAllGuildId 获取所有帮会id
func (m GuildModClient) GetAllGuildId() []int64 {
	result, err := m.invoke.Invoke(m.Type, "getAllGuildId")
	if err != nil {
		log.Error("GetAllGuildId error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.([]int64)
}

// GuildSign 帮会签到
func (m GuildModClient) GuildSign(playerId int64) (*model.PlayerGuild, error) {
	result, err := m.invoke.Invoke(m.Type, "guildSign", playerId)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*model.PlayerGuild), nil
}

// GetGuildInfoByIds 获取帮会信息
func (m GuildModClient) GetGuildInfoByIds(ids []int64) map[int64]*proto_guild.Guild {
	result, err := m.invoke.Invoke(m.Type, "getGuildInfoByIds", ids)
	if err != nil {
		log.Error("GetAllGuildId error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(map[int64]*proto_guild.Guild)
}

// GetPlayerInfoByIds 获取玩家工会信息
func (m GuildModClient) GetPlayerInfoByIds(ids []int64) map[int64]*proto_public.CommonPlayerInfo {
	result, err := m.invoke.Invoke(m.Type, "getPlayerInfoByIds", ids)
	if err != nil {
		log.Error("GetPlayerInfoByIds error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(map[int64]*proto_public.CommonPlayerInfo)
}

// GetGuildRankInfoByIds 获取帮会排行榜信息
func (m GuildModClient) GetGuildRankInfoByIds(ids []int64) map[int64]*proto_rank.GuildRankItem {
	result, err := m.invoke.Invoke(m.Type, "getGuildRankInfoByIds", ids)
	if err != nil {
		log.Error("GetGuildRankInfoByIds error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(map[int64]*proto_rank.GuildRankItem)
}

// GetGuildRankInfoByPlayerId 根据玩家id获取帮会排行榜信息
func (m GuildModClient) GetGuildRankInfoByPlayerId(playerId int64) *proto_rank.GuildRankItem {
	result, err := m.invoke.Invoke(m.Type, "getGuildRankInfoByPlayerId", playerId)
	if err != nil {
		log.Error("GetGuildRankInfoByPlayerId error:%v", err)
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_rank.GuildRankItem)
}

// GetPlayerGuildId 根据玩家帮会id
func (m GuildModClient) GetPlayerGuildId(playerId int64) int64 {
	result, _ := Int64(m.invoke.Invoke(m.Type, "getPlayerGuildId", playerId))
	return result
}

// GetYuanchiData 获取初始元池
func (m GuildModClient) GetYuanchiData(ctx *proto_player.Context) (*proto_guild.S2CInitYuanchi, error) {
	result, err := m.invoke.Invoke(m.Type, "getYuanchiData", ctx)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}
	return result.(*proto_guild.S2CInitYuanchi), nil
}

// YuanchiAddMaterials 元池-材料增加
func (m GuildModClient) YuanchiAddMaterials(ctx *proto_player.Context, materials map[int32]int32) error {
	_, err := m.invoke.Invoke(m.Type, "yuanchiAddMaterials", ctx, materials)
	return err
}

// SetBuild 设置建筑
func (m GuildModClient) SetBuild(ctx *proto_player.Context, id, index int32) (*proto_guild.S2CBuildGuildMap, error) {
	result, err := m.invoke.Invoke(m.Type, "SetBuild", ctx, id, index)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}
	return result.(*proto_guild.S2CBuildGuildMap), nil
}

// GetGuildMapInfo 获取地图
func (m GuildModClient) GetGuildMapInfo(ctx *proto_player.Context, playerId int64, guildId int32) (*proto_guild.S2CGetBuildMapInfo, error) {
	result, err := m.invoke.Invoke(m.Type, "getGuildMapInfo", ctx, playerId, guildId)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}
	return result.(*proto_guild.S2CGetBuildMapInfo), nil
}

// GuildPray 祈福
func (m GuildModClient) GuildPray(ctx *proto_player.Context, index int32) (*model.PlayerGuild, error) {
	result, err := m.invoke.Invoke(m.Type, "guildPray", ctx, index)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}
	return result.(*model.PlayerGuild), nil
}
