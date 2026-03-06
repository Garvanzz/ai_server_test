package guild

import (
	"errors"
	"xfx/core/model"
	"xfx/proto/proto_guild"
	"xfx/proto/proto_player"
)

// 地图
func (mgr *Manager) SetBuild(ctx *proto_player.Context, id, index int32) (*proto_guild.S2CBuildGuildMap, error) {
	//获取帮派信息
	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return nil, errors.New("no guild")
	}

	if info.GuildId == 0 {
		return nil, errors.New("no guild")
	}

	if info.GuildMap == nil {
		info.GuildMap = make(map[int32]*model.GuildMapItem)
	}

	if _, ok := info.GuildMap[index]; ok {
		return nil, errors.New("guild map has this build")
	}

	info.GuildMap[index] = new(model.GuildMapItem)
	info.GuildMap[index].Index = index
	info.GuildMap[index].Id = id
	info.GuildMap[index].Level = 1

	mgr.cache.Set(ctx.Id, info)
	return &proto_guild.S2CBuildGuildMap{
		Code: proto_guild.GuildCode_ERROR_OK,
		Item: &proto_guild.GuildBuildMapItem{
			Id:    id,
			Index: index,
			Level: 1,
		},
	}, nil
}

// 获取地图
func (mgr *Manager) getGuildMapInfo(ctx *proto_player.Context, playerId int64, guildId int32) (*proto_guild.S2CGetBuildMapInfo, error) {
	//获取帮派信息
	info := mgr.loadPlayerGuildFromCache(playerId)
	if info == nil {
		return nil, errors.New("no guildInfo")
	}

	if info.GuildId == 0 {
		return nil, errors.New("no guildId")
	}

	resp := &proto_guild.S2CGetBuildMapInfo{}
	resp.Items = make(map[int32]*proto_guild.GuildBuildMapItem, 0)
	if info.GuildMap == nil {
		return resp, nil
	}

	for _, v := range info.GuildMap {
		resp.Items[v.Index] = &proto_guild.GuildBuildMapItem{
			Id:    v.Id,
			Index: v.Index,
			Level: v.Level,
		}
	}

	return resp, nil
}
