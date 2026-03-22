package room

import (
	"fmt"
	"strconv"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_room"
)

// CreateRoom 创建房间
func CreateRoom(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SCreateRoom) {
	data := &proto_room.S2CCreateRoom{}
	if pl.Cache.RoomId > 0 {
		//获取房间类型
		info := invoke.RoomClient(ctx).PlayerInRoom(pl.Id)
		if !info {
			data.Code = proto_room.ERRORCODE_ERR_RoomInErr
			ctx.Send(data)
			return
		}

		//同步房间信息
		infos := invoke.RoomClient(ctx).GetRoomInfo(pl.Cache.RoomId)

		SyncRoomInfo(ctx, infos)
		data.Code = proto_room.ERRORCODE_ERR_Ok
		data.RoomType = infos.GetType()
		ctx.Send(data)
		return
	}

	roomInfo, err := invoke.RoomClient(ctx).CreateRoom(pl.ToContext(), msg)
	if err != nil {
		log.Error("CreateRoom error: %v", err)
		data.Code = proto_room.ERRORCODE_ERR_RoomState
		ctx.Send(data)
		return
	}

	SyncRoomInfo(ctx, roomInfo)

	//玩家标记
	pl.Cache.RoomId = roomInfo.GetRoomId()
	pl.Cache.RoomType = roomInfo.GetType()

	data.Code = proto_room.ERRORCODE_ERR_Ok
	data.RoomType = roomInfo.GetType()
	ctx.Send(data)
}

// JoinRoom 加入房间
func JoinRoom(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SJoinRoom) {
	code := invoke.RoomClient(ctx).JoinRoom(pl.ToContext(), msg)
	var data = &proto_room.S2CJoinRoom{}

	if code == proto_room.ERRORCODE_ERR_Ok {
		infos := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())

		SyncRoomInfo(ctx, infos)

		//玩家标记
		pl.Cache.RoomId = msg.GetRoomId()
		pl.Cache.RoomType = infos.GetType()
	}

	data.Code = code
	ctx.Send(data)
}

// ExitRoom 退出房间
func ExitRoom(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SExitRoom) {
	res := &proto_room.S2CExitRoom{}
	//判断是否在匹配中
	//TODO:roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
	isMatch := invoke.RoomClient(ctx).PlayerIsMatch(pl.Id)
	if isMatch {

		// TODO:逻辑问题
		//ctx.Invoke("Match", "CancelMatch", pl.Id, roomInfo.Type)
	}

	roomInfo1 := invoke.RoomClient(ctx).ExitRoom(pl.Id)
	if roomInfo1 != nil {
		SyncRoomInfo(ctx, roomInfo1)
	}

	//标记
	pl.Cache.RoomId = 0
	pl.Cache.RoomType = 0

	res.Code = proto_room.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// ExitRoom 解散房间
func DissolveRoom(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SDissolveRoom) {
	res := &proto_room.S2CDissolveRoom{}
	//判断是否在匹配中
	roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
	isMatch := invoke.RoomClient(ctx).PlayerIsMatch(pl.Id)
	if isMatch {

		// TODO: 逻辑问题
		//invoke.MatchClient(ctx).CancelMatch()
		//ctx.Invoke("Match", "CancelMatch", pl.Id, roomInfo.Type)
	}

	isLeader := invoke.RoomClient(ctx).DissolveRoom(pl.Id)
	if !isLeader {
		res.Code = proto_room.ERRORCODE_ERR_RoomNotIsLeader
		ctx.Send(res)
		return
	}

	//标记
	pl.Cache.RoomId = 0
	pl.Cache.RoomType = 0

	res.Code = proto_room.ERRORCODE_ERR_Ok
	log.Debug("房间解散:%v", roomInfo.RoomId)
	//通知房间的其他人
	var ids []int64
	for _, v := range roomInfo.GetPlayers() {
		ids = append(ids, v.CommonPlayerInfo.PlayerId)
	}
	invoke.DispatchPlayers(ctx, ids, res)
}

// SyncRoomInfo 同步房间内的消息
func SyncRoomInfo(ctx global.IPlayer, roomInfo *proto_room.RoomInfo) {
	var _info = &proto_room.PushRoomIndo{}
	_info.Info = roomInfo

	//通知房间内的其他人
	var ids []int64
	for _, v := range roomInfo.GetPlayers() {
		ids = append(ids, v.CommonPlayerInfo.PlayerId)
	}
	invoke.DispatchPlayers(ctx, ids, _info)
}

// StartGame 开始游戏
func StartGame(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SStartGame) {
	_, err := invoke.RoomClient(ctx).StartGame(pl.ToContext(), msg)
	var res = &proto_room.S2CStartGame{}
	if err != nil {
		log.Error("StartGame error : %v", err)
		res.State = int32(proto_room.ERRORCODE_ERR_RoomState)
		ctx.Send(res)
		return
	}
}

// 允许观战
func PermitLook(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SPermitLook) {
	res := &proto_room.S2CPermitLook{}
	state := invoke.RoomClient(ctx).PermitLook(msg.GetRoomId(), msg.GetIsPermit())

	roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
	SyncRoomInfo(ctx, roomInfo)

	res.Code = state
	ctx.Send(res)
}

// 准备
func ReadyGame(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SGameReady) {
	res := &proto_room.S2CGameReady{}
	suc := invoke.RoomClient(ctx).ReadyGame(msg.GetRoomId(), pl.Id, true)
	if suc {
		//在推房间信息
		roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
		SyncRoomInfo(ctx, roomInfo)
		res.Code = proto_room.ERRORCODE_ERR_Ok
		ctx.Send(res)
	} else {
		res.Code = proto_room.ERRORCODE_ERR_NoRoom
		ctx.Send(res)
	}
}

// 取消准备
func CancelReadyGame(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SGameReadyCancel) {
	res := &proto_room.S2CGameReadyCancel{}
	suc := invoke.RoomClient(ctx).ReadyGame(msg.GetRoomId(), pl.Id, false)
	if suc {
		roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
		SyncRoomInfo(ctx, roomInfo)

		res.Code = proto_room.ERRORCODE_ERR_Ok
		ctx.Send(res)
	} else {
		res.Code = proto_room.ERRORCODE_ERR_NoRoom
		ctx.Send(res)
	}
}

// 上阵
func LineUpGame(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SRoomLineUp) {

}

// 搜索房间
func FindRoom(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SFindRoom) {
	res := &proto_room.S2CFindRoom{}
	info := invoke.RoomClient(ctx).FindRoom(msg.GetRoomNameOrId())
	if info.RoomId == 0 {
		res.Code = proto_room.ERRORCODE_ERR_NoRoom
		ctx.Send(res)
		return
	}

	res.RoomInfo = info
	res.Code = proto_room.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// 随机加入房间
func RangleJoinRoom(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SRangeJoinRoom) {
	res := &proto_room.S2CRangeJoinRoom{}
	roomId := invoke.RoomClient(ctx).RangJoinRoom(pl.ToContext())
	if roomId < 0 {
		log.Error("RangleJoinRoom error")
		res.Code = proto_room.ERRORCODE_ERR_NoRoom
		ctx.Send(res)
		return
	}

	roomInfo := invoke.RoomClient(ctx).GetRoomInfo(roomId)
	SyncRoomInfo(ctx, roomInfo)

	res.Code = proto_room.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// 刷新列表
func RefreshRoomList(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SRefreshRoom) {
	res := &proto_room.S2CRefreshRoom{}
	infos := invoke.RoomClient(ctx).RefreshList(msg)
	res.Info = infos
	res.Code = proto_room.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// 修改密码
func ChangePassword(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SChangePassword) {
	res := &proto_room.S2CChangePassword{}

	// TODO: 密码要改成string
	//state := invoke.RoomClient(ctx).ChangePassword(msg.GetRoomId(), msg.GetPassWord())
	state := invoke.RoomClient(ctx).ChangePassword(msg.GetRoomId(), "123456")

	roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
	res.Info = roomInfo
	res.Code = state
	ctx.Send(res)
}

// 设置公开
func SetOpen(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SRoomOpen) {
	res := &proto_room.S2CRoomOpen{}
	state := invoke.RoomClient(ctx).SetOpen(msg.GetRoomId(), msg.GetIsOpen())

	roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
	SyncRoomInfo(ctx, roomInfo)

	res.Code = state
	ctx.Send(res)
}

// 更换阵容
func SetGroup(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SSetGroup) {
	res := &proto_room.S2CSetGroup{}
	// TODO:参数有问题
	//state := invoke.RoomClient(ctx).ChangeGroup(msg.GetRoomId(), pl.Id, msg.GetGroup(), msg.GetIndex())
	state := proto_room.ERRORCODE_ERR_Ok

	roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
	SyncRoomInfo(ctx, roomInfo)

	res.Code = state
	ctx.Send(res)
}

// GetInviteList 获得邀请列表
func GetInviteList(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SGetInviteList) {
	res := &proto_room.S2CGetInviteList{}
	//获取好友
	if msg.Type == 1 {
		reply, err := db.RedisExec("smembers", fmt.Sprintf("%s:%d", define.Friend, pl.Id))
		if err != nil {
			log.Error("ReqFriendList err: %v", err)
			ctx.Send(res)
			return
		}
		vals := reply.([]interface{})

		data := make([]*proto_room.InviteOption, 0)
		for i := 0; i < len(vals); i++ {
			str := string(vals[i].([]byte))
			dbId, _ := strconv.ParseInt(str, 10, 64)
			r := new(proto_room.InviteOption)
			playerInfo := global.GetPlayerInfo(dbId)
			r.CommonPlayerInfo = playerInfo.ToCommonPlayer()

			isOnline := invoke.LoginClient(ctx).IsOnline(dbId)
			if isOnline {
				r.PlayerState = 1
			} else {
				r.PlayerState = 3
			}
			inv := invoke.RoomClient(ctx).GetPlayerInvite(pl.Id, dbId)
			if inv {
				r.State = 1
			} else {
				// TODO: 没有这个方法
				fus, err := ctx.Invoke("Room", "GetPlayerRefuseVisite", pl.Id, dbId)
				if err != nil {
					log.Error("GetPlayerRefuseVisite err : %v", err)
					continue
				}
				if fus.(bool) == true {
					r.State = 2
				} else {
					r.State = 3
				}
			}
			data = append(data, r)
		}

		if len(data) > 0 {
			ret := &proto_room.PushInviteList{}
			ret.Type = msg.Type
			ret.Opts = data
			ctx.Send(ret)
		}
	}
	res.Code = proto_room.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// RoomInvite 邀请
func RoomInvite(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SRoomInvite) {
	res := &proto_room.S2CRoomInvite{}
	//不能邀请自己
	if pl.Id == msg.Id {
		res.Code = proto_room.ERRORCODE_ERR_RoomState
		ctx.Send(res)
		return
	}

	//发出邀请
	succ := invoke.RoomClient(ctx).SendInvite(pl.ToContext(), msg)
	if !succ {
		res.Code = proto_room.ERRORCODE_ERR_RoomState
		ctx.Send(res)
		return
	}

	da := &proto_room.PushNewInvite{}
	da.RoomId = msg.RoomId
	da.Name = pl.Base.Name
	da.SendUid = pl.Id
	da.Group = msg.Group
	da.InviteTime = 60
	da.RecUid = msg.Id
	invoke.Dispatch(ctx, da.RecUid, da)

	res.Code = proto_room.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// 邀请反馈
func RoomInviteBack(ctx global.IPlayer, pl *model.Player, msg *proto_room.C2SInviteBack) {
	res := &proto_room.S2CInviteBack{}
	//获取邀请
	state := invoke.RoomClient(ctx).SendInviteBack(pl.ToContext(), msg)

	if state == proto_room.ERRORCODE_ERR_Ok {
		roomInfo := invoke.RoomClient(ctx).GetRoomInfo(msg.GetRoomId())
		SyncRoomInfo(ctx, roomInfo)
	}

	//同步列表
	res.Code = state
	ctx.Send(res)
}

// 玩家断线
func Logout(ctx global.IPlayer, pl *model.Player) {
	pl.Cache.RoomId = 0
	pl.Cache.RoomType = 0
}

// StartGameMatch 开始游戏匹配
func StartGameMatch(ctx global.IPlayer, pl *model.Player) {
	res := &proto_room.S2CRoomMatchGame{}
	if pl.Cache.RoomId <= 0 {
		res.Code = proto_room.ERRORCODE_ERR_NoRoom
		ctx.Send(res)
		return
	}

	//是否正在匹配
	isMatch := invoke.RoomClient(ctx).PlayerIsMatch(pl.Id)
	if isMatch {
		res.Code = proto_room.ERRORCODE_ERR_RoomMatcing
		ctx.Send(res)
		return
	}

	//是否正在匹配
	state := invoke.RoomClient(ctx).StartMatch(pl.Id)
	if !state {
		log.Error("StartGameMatch state:%v", state)
		res.Code = proto_room.ERRORCODE_ERR_RoomMatcing
		ctx.Send(res)
		return
	}

	res.Code = proto_room.ERRORCODE_ERR_Ok
	ctx.Send(res)
}

// CancelGameMatch 取消游戏匹配
func CancelGameMatch(ctx global.IPlayer, pl *model.Player) {
	res := &proto_room.S2CMatchCancel{}
	if pl.Cache.RoomId <= 0 {
		res.Code = proto_room.ERRORCODE_ERR_NoRoom
		ctx.Send(res)
		return
	}

	//是否正在匹配
	isMatch := invoke.RoomClient(ctx).PlayerIsMatch(pl.Id)
	if !isMatch {
		log.Error("CancelGameMatch not in match")
		res.Code = proto_room.ERRORCODE_ERR_RoomMatcing
		ctx.Send(res)
		return
	}

	//取消匹配
	ok := invoke.RoomClient(ctx).CancelMatch(pl.Id)
	if !ok {
		log.Error("CancelGameMatch error")
		res.Code = proto_room.ERRORCODE_ERR_RoomMatcing
		ctx.Send(res)
		return
	}

	res.Code = proto_room.ERRORCODE_ERR_Ok
	ctx.Send(res)
}
