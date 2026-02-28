package invoke

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_player"
	"xfx/proto/proto_room"
)

type RoomModClient struct {
	invoke Invoker
	Type   string
}

func RoomClient(invoker Invoker) RoomModClient {
	return RoomModClient{
		invoke: invoker,
		Type:   define.ModuleRoom,
	}
}

// GameServerStartGame 游戏服通知逻辑服开始游戏
func (m RoomModClient) GameServerStartGame(info *proto_room.S2CStartGame) {
	_, err := m.invoke.Invoke(m.Type, "GameServerStartGame", info)
	if err != nil {
		log.Error("GameServerStartGame error:%v", err)
	}
}

// MatchToRoomInfo 收到匹配信息
func (m RoomModClient) MatchToRoomInfo(roomId1, roomId2 int32) {
	_, err := m.invoke.Invoke(m.Type, "MatchToRoomInfo", roomId1, roomId2)
	if err != nil {
		log.Error("MatchToRoomInfo error:%v", err)
	}
}

// GetRoomInfo 获取房间信息
func (m RoomModClient) GetRoomInfo(roomId int32) *proto_room.RoomInfo {
	result, err := m.invoke.Invoke(m.Type, "GetRoomInfo", roomId)
	if err != nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_room.RoomInfo)
}

// PlayerInRoom 玩家是否在房间里面
func (m RoomModClient) PlayerInRoom(playerId int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "PlayerInRoom", playerId))
	return result
}

// CreateRoom 创建房间
func (m RoomModClient) CreateRoom(gen *proto_player.Context, msg *proto_room.C2SCreateRoom) (*proto_room.RoomInfo, error) {
	result, err := m.invoke.Invoke(m.Type, "CreateRoom", gen, msg)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*proto_room.RoomInfo), nil
}

// JoinRoom 加入房间
func (m RoomModClient) JoinRoom(gen *proto_player.Context, msg *proto_room.C2SJoinRoom) proto_room.ERRORCODE {
	result, err := m.invoke.Invoke(m.Type, "JoinRoom", gen, msg)
	if err != nil {
		log.Error("JoinRoom error:%v", err)
		return proto_room.ERRORCODE_ERR_RoomNoConfig
	}

	return result.(proto_room.ERRORCODE)
}

// ExitRoom 退出房间
func (m RoomModClient) ExitRoom(playerId int64) *proto_room.RoomInfo {
	result, err := m.invoke.Invoke(m.Type, "ExitRoom", playerId)
	if err != nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_room.RoomInfo)
}

// DissolveRoom 解散房间
func (m RoomModClient) DissolveRoom(playerId int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "DissolveRoom", playerId))
	return result
}

// ReadyGame 准备开始\取消准备
func (m RoomModClient) ReadyGame(_roomId int32, playerId int64, isReady bool) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "ReadyGame", _roomId, playerId, isReady))
	return result
}

// StartGame 开始游戏
func (m RoomModClient) StartGame(gen *proto_player.Context, msg *proto_room.C2SStartGame) (*proto_room.RoomInfo, error) {
	result, err := m.invoke.Invoke(m.Type, "StartGame", gen, msg)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*proto_room.RoomInfo), nil
}

// Disconnect 断开连接
func (m RoomModClient) Disconnect(pl *model.Player) {
	_, err := m.invoke.Invoke(m.Type, "Disconnect", pl)
	if err != nil {
		log.Error("Disconnect error:%v", err)
	}
}

// LineUp 随机加入
func (m RoomModClient) LineUp(roomId int32, playerId int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "LineUp", roomId, playerId))
	return result
}

// FindRoom 搜索房间
func (m RoomModClient) FindRoom(roomName string) *proto_room.RoomInfo {
	result, err := m.invoke.Invoke(m.Type, "FindRoom", roomName)
	if err != nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return result.(*proto_room.RoomInfo)
}

// RangJoinRoom 随机加入
func (m RoomModClient) RangJoinRoom(gen *proto_player.Context) int32 {
	result, err := Int32(m.invoke.Invoke(m.Type, "RangJoinRoom", gen))
	if err != nil {
		log.Error("RangJoinRoom error:%v", err)
		return 0
	}
	return result
}

// RefreshList 刷新房间列表
func (m RoomModClient) RefreshList(msg *proto_room.C2SRefreshRoom) []*proto_room.RoomInfo {
	result, err := m.invoke.Invoke(m.Type, "RefreshList", msg)
	if err != nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return result.([]*proto_room.RoomInfo)
}

// ChangePassword 修改密码
func (m RoomModClient) ChangePassword(roomId int32, password string) proto_room.ERRORCODE {
	result, err := m.invoke.Invoke(m.Type, "ChangePassword", roomId, password)
	if err != nil {
		log.Error("ChangePassword error:%v", err)
		return proto_room.ERRORCODE_ERR_RoomNoConfig
	}

	return result.(proto_room.ERRORCODE)
}

// PermitLook 允许观看
func (m RoomModClient) PermitLook(roomId int32, isPermit bool) proto_room.ERRORCODE {
	result, err := m.invoke.Invoke(m.Type, "PermitLook", roomId, isPermit)
	if err != nil {
		log.Error("PermitLook error:%v", err)
		return proto_room.ERRORCODE_ERR_RoomNoConfig
	}

	return result.(proto_room.ERRORCODE)
}

// SetOpen 设置公开
func (m RoomModClient) SetOpen(roomId int32, open bool) proto_room.ERRORCODE {
	result, err := m.invoke.Invoke(m.Type, "SetOpen", roomId, open)
	if err != nil {
		log.Error("SetOpen error:%v", err)
		return proto_room.ERRORCODE_ERR_RoomNoConfig
	}

	return result.(proto_room.ERRORCODE)
}

// ChangeGroup 改变阵容
func (m RoomModClient) ChangeGroup(playerId int64) proto_room.ERRORCODE {
	result, err := m.invoke.Invoke(m.Type, "ChangeGroup", playerId)
	if err != nil {
		log.Error("ChangeGroup error:%v", err)
		return proto_room.ERRORCODE_ERR_RoomNoConfig
	}

	return result.(proto_room.ERRORCODE)
}

// SendInviteBack 邀请反馈
func (m RoomModClient) SendInviteBack(gen *proto_player.Context, msg *proto_room.C2SInviteBack) proto_room.ERRORCODE {
	result, err := m.invoke.Invoke(m.Type, "SendInviteBack", gen, msg)
	if err != nil {
		log.Error("SendInviteBack error:%v", err)
		return proto_room.ERRORCODE_ERR_RoomNoConfig
	}

	return result.(proto_room.ERRORCODE)
}

// GetPlayerInvite 判断玩家是否被另一个玩家邀请
func (m RoomModClient) GetPlayerInvite(sendId, recId int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "GetPlayerInvite", sendId, recId))
	return result
}

// GetInviteList 获得邀请列表
func (m RoomModClient) GetInviteList(playerId int64) []*model.Invite {
	result, err := m.invoke.Invoke(m.Type, "GetInviteList", playerId)
	if err != nil {
		return nil
	}

	if result == nil {
		return nil
	}

	return result.([]*model.Invite)
}

// SendInvite 发出邀请
func (m RoomModClient) SendInvite(gen *proto_player.Context, msg *proto_room.C2SRoomInvite) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "SendInvite", gen, msg))
	return result
}

// PlayerIsMatch 玩家是否在匹配中
func (m RoomModClient) PlayerIsMatch(playerId int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "PlayerIsMatch", playerId))
	return result
}

// CancelMatch 取消匹配
func (m RoomModClient) CancelMatch(playerId int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "CancelMatch", playerId))
	return result
}

// StartMatch 开始匹配
func (m RoomModClient) StartMatch(playerId int64) bool {
	result, _ := Bool(m.invoke.Invoke(m.Type, "StartMatch", playerId))
	return result
}
