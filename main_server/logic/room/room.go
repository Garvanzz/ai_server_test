package room

import (
	"errors"
	"strconv"
	"time"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"
	"xfx/proto/proto_room"
)

var Module = func() module.Module {
	return new(Manager)
}

type Manager struct {
	modules.BaseModule
	rooms         map[int32]*model.Room     // 房间信息
	roomNames     map[string]int32          // 房间名映射
	playerInvites map[int64][]*model.Invite // 玩家邀请
	playerRoom    map[int64]int32           // 玩家所在房间id
	inviteId      int32                     // 邀请id
	readyStart    map[int32]int64
}

func (mgr *Manager) OnInit(app module.App) {
	mgr.BaseModule.OnInit(app)
	mgr.rooms = make(map[int32]*model.Room)
	mgr.roomNames = make(map[string]int32)
	mgr.playerInvites = make(map[int64][]*model.Invite)
	mgr.readyStart = make(map[int32]int64)
	mgr.playerRoom = make(map[int64]int32)

	mgr.Register("GameServerStartGame", mgr.OnGameServerStartGame)
	mgr.Register("MatchToRoomInfo", mgr.OnMatchToRoomInfo)
	mgr.Register("GetRoomInfo", mgr.OnGetRoomInfo)
	mgr.Register("PlayerInRoom", mgr.OnPlayerInRoom)
	mgr.Register("CreateRoom", mgr.OnCreateRoom)
	mgr.Register("JoinRoom", mgr.OnJoinRoom)
	mgr.Register("ExitRoom", mgr.OnExitRoom)
	mgr.Register("DissolveRoom", mgr.OnDissolveRoom)
	mgr.Register("ReadyGame", mgr.OnPrepareGame)
	mgr.Register("StartGame", mgr.OnStartGame)
	mgr.Register("Disconnect", mgr.OnDisconnect)
	mgr.Register("LineUp", mgr.OnLineUpHero)
	mgr.Register("FindRoom", mgr.OnFindRoom)
	mgr.Register("RangJoinRoom", mgr.OnRangJoinRoom)
	mgr.Register("RefreshList", mgr.OnRefreshRoomList)
	mgr.Register("ChangePassword", mgr.OnChangePassword)
	mgr.Register("PermitLook", mgr.OnPermitLook)
	mgr.Register("SetOpen", mgr.OnSetOpen)
	mgr.Register("ChangeGroup", mgr.OnChangeGroup)
	mgr.Register("SendInviteBack", mgr.OnSendInviteBack)
	mgr.Register("GetPlayerInvite", mgr.OnGetPlayerInvite)
	mgr.Register("GetInviteList", mgr.OnGetInviteList)
	mgr.Register("SendInvite", mgr.OnSendInvite)
	mgr.Register("PlayerIsMatch", mgr.OnPlayerIsMatch)
	mgr.Register("CancelMatch", mgr.OnCancelMatch)
	mgr.Register("StartMatch", mgr.OnStartMatch)

}

func (mgr *Manager) GetType() string { return define.ModuleRoom }

func (mgr *Manager) OnTick(delta time.Duration) {
	//同步开始
	for roomId, times := range mgr.readyStart {
		//通知加载游戏
		if time.Now().Unix()-times >= 10 {
			mgr.OnSyncStatLoadGame(roomId)
			delete(mgr.readyStart, roomId)
		}
	}

	// 去除过期的邀请
	for id, list := range mgr.playerInvites {
		k := 0
		for _, v := range list {
			if time.Now().Unix()-v.InviteTime < 60*1 {
				list[k] = v
				k++
			}
		}
		mgr.playerInvites[id] = list[:k]
	}
}

func (mgr *Manager) OnMessage(msg interface{}) interface{} {
	log.Debug("* room message %v", msg)
	//switch p := msg.(type) {
	//case *proto_public.S2SGameSettleInfo:
	//	mgr.OnSettleGame(p)
	//default:
	//	return nil
	//}
	return nil
}

// OnPlayerInRoom 玩家是否在房间里面
func (mgr *Manager) OnPlayerInRoom(playerId int64) bool {
	_, exist := mgr.playerRoom[playerId]
	return exist
}

// OnGetRoomInfo 获取房间信息
func (mgr *Manager) OnGetRoomInfo(roomId int32) *proto_room.RoomInfo {
	room, ok := mgr.rooms[roomId]
	if !ok {
		return nil
	}

	return global.ToRoomInfoProto(room)
}

// OnCreateRoom 创建房间
func (mgr *Manager) OnCreateRoom(gen *proto_player.Context, msg *proto_room.C2SCreateRoom) (*proto_room.RoomInfo, error) {
	if msg.RoomName == "" {
		return nil, errors.New("room name is empty")
	}

	_, ok := mgr.roomNames[msg.RoomName]
	if ok {
		return nil, errors.New("room name is repeated")
	}

	//获取布阵
	lineup := mgr.OnGetBattleLineUp(gen, msg.RoomType)
	if lineup == nil {
		return nil, errors.New("room is not lineup")
	}
	roomId, _ := db.CommonEngine.GetRoomId()

	room := new(model.Room)
	room.Id = int32(roomId)
	room.Type = msg.RoomType
	room.State = define.RoomStateNormal
	room.Password = "" // TODO:改成string
	room.Name = msg.RoomName
	room.IsPermitLook = msg.GetIsPermitLook() //是否观战
	room.Owner = gen.Id
	room.IsOpen = msg.GetIsOpen()

	player := generateRoomPlayer(gen)
	player.IsReady = true
	player.Group = 1
	player.Heros = lineup
	room.Players = make([]*model.RoomPlayer, 0)
	room.Players = append(room.Players, player)

	mgr.roomNames[msg.RoomName] = int32(roomId)
	mgr.rooms[room.Id] = room
	mgr.playerRoom[gen.Id] = room.Id

	log.Debug("创建房间:%v", room)
	return global.ToRoomInfoProto(room), nil
}

// 获取战斗数据
func (mgr *Manager) OnGetBattleLineUp(gen *proto_player.Context, typ int32) *proto_public.BattleHeroData {
	switch typ {
	case define.MatchModTopPk: //巅峰决斗
		//获取布阵
		lineup := global.GetPlayerLineUpInfo(gen.Id)
		if _, ok := lineup[define.LINEUP_TheCompetition]; !ok {
			return nil
		}
		_lineup := lineup[define.LINEUP_TheCompetition]
		data := global.GetBattlePlayerData(gen, _lineup.HeroId)
		return data
	case define.MatchModArena:
		//获取布阵
		lineup := global.GetPlayerLineUpInfo(gen.Id)
		if _, ok := lineup[define.LINEUP_ARENA]; !ok {
			return nil
		}
		_lineup := lineup[define.LINEUP_ARENA]
		data := global.GetBattlePlayerData(gen, _lineup.HeroId)
		return data
	default:
		return nil
	}
}

// OnJoinRoom 加入房间
func (mgr *Manager) OnJoinRoom(gen *proto_player.Context, msg *proto_room.C2SJoinRoom) proto_room.ERRORCODE {
	//判断房间在不在
	if room, ok := mgr.rooms[msg.RoomId]; ok {
		// 判断公开性
		if room.IsOpen == false && room.Password != msg.Password { // 密码是string
			return proto_room.ERRORCODE_ERR_RoomPassWordErr
		}

		//判断玩家是不是在里面
		if room.IsPlayerInRoom(gen.Id) {
			return proto_room.ERRORCODE_ERR_RoomInErr
		}

		//判断人数
		if len(room.Players) >= 8 {
			return proto_room.ERRORCODE_ERR_RoomPlayerNumErr
		}
		//加入在房间里面去
		player := generateRoomPlayer(gen)
		room.SetPlayer(player)
		mgr.rooms[msg.RoomId] = room
		return proto_room.ERRORCODE_ERR_Ok
	}
	return proto_room.ERRORCODE_ERR_NoRoom
}

// 生成房间玩家信息
func generateRoomPlayer(gen *proto_player.Context) *model.RoomPlayer {
	return &model.RoomPlayer{
		PlayerId: gen.Id,
	}
}

// OnMatchToRoomInfo 收到匹配信息
func (mgr *Manager) OnMatchToRoomInfo(roomId1, roomId2 int32) {
	// 整合房间,统一用左边的房间号
	room1 := mgr.rooms[roomId1]
	room2 := mgr.rooms[roomId2]

	for i := 0; i < 2; i++ {
		roomPlayer := room2.Players[i]
		if roomPlayer != nil {
			room1.SetPlayer(roomPlayer)
			mgr.playerRoom[roomPlayer.PlayerId] = roomId1
		}
	}

	delete(mgr.rooms, roomId2)
	delete(mgr.roomNames, room2.Name)

	log.Debug("匹配成功")

	// 同步房间信息
	mgr.SyncRoomInfo(roomId1)

	// 同步匹配信息
	res := &proto_room.PushMatchTeam{
		IsStart: false,
		Time:    int32(10),
	}

	pushList := make([]int64, 0)
	for _, roomPlayer := range room1.Players {
		if roomPlayer != nil {
			pushList = append(pushList, roomPlayer.PlayerId)
		}
	}

	invoke.DispatchPlayers(mgr, pushList, res)

	//10秒后开始执行加载
	mgr.readyStart[roomId1] = time.Now().Unix()
}

// SyncRoomInfo 同步房间信息
func (mgr *Manager) SyncRoomInfo(roomId int32) {
	room := mgr.rooms[roomId]
	l := make([]int64, 0)
	for _, v := range room.Players {
		if v == nil {
			continue
		}
		l = append(l, v.PlayerId)
	}

	invoke.DispatchPlayers(mgr, l, &proto_room.PushRoomIndo{Info: global.ToRoomInfoProto(room)})
}

// OnSyncStatLoadGame 开始加载游戏
func (mgr *Manager) OnSyncStatLoadGame(roomId int32) {
	roomInfo := mgr.rooms[roomId]
	mgr.Cast("launcher", roomInfo)
}

// OnGameServerStartGame 游戏服通知逻辑服开始游戏
func (mgr *Manager) OnGameServerStartGame(info *proto_room.S2CStartGame) {
	mgr.rooms[info.GetRoomInfo().RoomId].State = define.RoomStateGame
	var ids []int64
	for _, v := range info.GetRoomInfo().GetPlayers() {
		//忽略人机
		if v.IsMonster {
			continue
		}
		ids = append(ids, v.CommonPlayerInfo.PlayerId)
	}
	invoke.DispatchPlayers(mgr, ids, info)
}

// OnDissolveRoom 解散房间
func (mgr *Manager) OnDissolveRoom(playerId int64) bool {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		log.Error("")
		return false
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("")
		return false
	}

	// check state
	if room.State != define.RoomStateNormal {
		return false
	}

	if room.Owner != playerId { // 房主
		return false
	}

	// 解散房间
	delete(mgr.rooms, roomId)
	delete(mgr.roomNames, room.Name)
	delete(mgr.playerRoom, playerId)
	return true
}

// OnExitRoom 退出房间
func (mgr *Manager) OnExitRoom(playerId int64) *proto_room.RoomInfo {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		log.Error("OnExitRoom player id error:%v", playerId)
		return nil
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("OnExitRoom room id error:%v", roomId)
		return nil
	}

	// check state
	if room.State != define.RoomStateNormal {
		log.Error("OnExitRoom room state error:%v", room.State)
		return nil
	}

	if room.Owner == playerId { // 房主
		// 看下有没有其他人
		otherPlayer := int64(0)
		for i := 0; i < 2; i++ {
			if room.Players[i] != nil && room.Players[i].PlayerId != playerId {
				otherPlayer = room.Players[i].PlayerId
				break
			}
		}

		if otherPlayer != 0 { // 转移房间所有权
			room.RemovePlayer(playerId)
			room.Owner = otherPlayer
			return global.ToRoomInfoProto(room)
		} else { //解散房间
			delete(mgr.rooms, roomId)
			delete(mgr.roomNames, room.Name)
			return nil
		}
	} else {
		//直接退出
		delete(mgr.playerRoom, playerId)
		room.RemovePlayer(playerId)
		return global.ToRoomInfoProto(room)
	}
}

// OnRangJoinRoom 随机加入
func (mgr *Manager) OnRangJoinRoom(gen *proto_player.Context) int32 {
	roomPlayer := generateRoomPlayer(gen)
	for _, v := range mgr.rooms {
		if v.State == define.RoomStateNormal && v.IsOpen {
			//if v.Type == define.MatchModStage {
			//	continue
			//}

			ok := v.SetPlayer(roomPlayer)
			if !ok {
				continue
			}

			return v.Id
		}
	}
	return -1
}

// OnFindRoom 搜索房间
func (mgr *Manager) OnFindRoom(roomName string) *proto_room.RoomInfo {
	num, err := strconv.ParseInt(roomName, 10, 32)
	if err != nil {
		//判断名字
		if len(roomName) > 0 {
			id := mgr.roomNames[roomName]
			if id <= 0 {
				return new(proto_room.RoomInfo)
			} else {
				return global.ToRoomInfoProto(mgr.rooms[id])
			}
		}
	} else {
		if num > 0 {
			if _, ok := mgr.rooms[int32(num)]; ok {
				return global.ToRoomInfoProto(mgr.rooms[int32(num)])
			}
		}

		if len(roomName) > 0 {
			id := mgr.roomNames[roomName]
			if id <= 0 {
				return new(proto_room.RoomInfo)
			} else {
				return global.ToRoomInfoProto(mgr.rooms[id])
			}
		}
	}

	return new(proto_room.RoomInfo)
}

// OnChangePassword 修改密码
func (mgr *Manager) OnChangePassword(roomId int32, password string) proto_room.ERRORCODE {
	if room, ok := mgr.rooms[roomId]; ok {
		room.Password = password
		mgr.rooms[roomId] = room
		return proto_room.ERRORCODE_ERR_Ok
	} else {
		return proto_room.ERRORCODE_ERR_NoRoom
	}
}

// OnPermitLook 允许观看
func (mgr *Manager) OnPermitLook(roomId int32, isPermit bool) proto_room.ERRORCODE {
	if room, ok := mgr.rooms[roomId]; ok {
		room.IsPermitLook = isPermit
		mgr.rooms[roomId] = room
		return proto_room.ERRORCODE_ERR_Ok
	} else {
		return proto_room.ERRORCODE_ERR_NoRoom
	}
}

// OnSetOpen 设置公开
func (mgr *Manager) OnSetOpen(roomId int32, open bool) proto_room.ERRORCODE {
	if room, ok := mgr.rooms[roomId]; ok {
		room.IsOpen = open
		mgr.rooms[roomId] = room
		return proto_room.ERRORCODE_ERR_Ok
	} else {
		return proto_room.ERRORCODE_ERR_NoRoom
	}
}

// OnChangeGroup 改变阵容
func (mgr *Manager) OnChangeGroup(playerId int64) proto_room.ERRORCODE {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		return proto_room.ERRORCODE_ERR_NoRoom
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("OnChangeGroupFunc room id error:%v")
		return proto_room.ERRORCODE_ERR_NoRoom
	}

	if room.State != define.RoomStateNormal {
		log.Error("OnChangeGroupFunc room state error:%v", room.State)
		return proto_room.ERRORCODE_ERR_RoomInErr
	}

	index := 0
	for i, v := range room.Players {
		if v.PlayerId == playerId {
			index = i + 1
		}
	}

	if index == 0 {
		return proto_room.ERRORCODE_ERR_RoomInErr
	}

	if index == 1 || index == 2 {
		room.Players[0], room.Players[1] = room.Players[1], room.Players[0]
	} else {
		room.Players[2], room.Players[3] = room.Players[3], room.Players[2]
	}

	return proto_room.ERRORCODE_ERR_Ok
}

// OnRefreshRoomList 刷新房间列表
func (mgr *Manager) OnRefreshRoomList(msg *proto_room.C2SRefreshRoom) []*proto_room.RoomInfo {
	num := 0
	list := make([]*proto_room.RoomInfo, 0)
	for _, v := range mgr.rooms {

		//根据条件刷新
		if v.Type == msg.Type {
			continue
		}

		if msg.Open == 1 && !v.IsOpen {
			continue
		} else if msg.Open == 2 && v.IsOpen {
			continue
		}

		if msg.Look == 1 && !v.IsPermitLook {
			continue
		} else if msg.Look == 2 && v.IsPermitLook {
			continue
		}

		num++
		list = append(list, global.ToRoomInfoProto(v))
		if num > 5 {
			break
		}
	}
	return list
}

// OnGetInviteList 获得邀请列表
func (mgr *Manager) OnGetInviteList(playerId int64) []*model.Invite {
	return mgr.playerInvites[playerId]
}

// OnGetPlayerInvite TODO:判断玩家是否被另一个玩家邀请
func (mgr *Manager) OnGetPlayerInvite(sendId, recId int64) bool {
	//if mgr.roomInvites[sendId] != nil {
	//	for k := 0; k < len(mgr.roomInvites[sendId]); k++ {
	//		if mgr.roomInvites[sendId][k].RecUid == recId {
	//			return true
	//		}
	//	}
	//}
	return false
}

// OnSendInvite 发出邀请
func (mgr *Manager) OnSendInvite(gen *proto_player.Context, msg *proto_room.C2SRoomInvite) bool {
	dbId := gen.Id
	invite := &model.Invite{
		RoomName:   gen.Name,
		RoomId:     msg.RoomId,
		Group:      msg.Group,
		Receiver:   msg.Id,
		Sender:     dbId,
		SenderName: "",
		InviteTime: time.Now().Unix(),
	}

	inviteList, ok := mgr.playerInvites[msg.Id]
	if !ok {
		mgr.playerInvites[msg.Id] = []*model.Invite{invite}
		return true
	}

	for _, v := range inviteList {
		if v.Sender == dbId {
			return false
		}
	}

	inviteList = append(inviteList, invite)
	mgr.playerInvites[msg.Id] = inviteList
	return true
}

// OnSendInviteBack 邀请反馈
func (mgr *Manager) OnSendInviteBack(gen *proto_player.Context, msg *proto_room.C2SInviteBack) proto_room.ERRORCODE {
	// 判断过期没有
	inviteList := mgr.playerInvites[msg.Id]
	var invite *model.Invite
	for index, v := range inviteList {
		if v.Id == int32(msg.Id) {
			invite = v
			mgr.playerInvites[msg.Id] = append(inviteList[0:index], inviteList[index+1:]...)
			break
		}
	}

	if invite == nil {
		return proto_room.ERRORCODE_ERR_RoomIviteOutTime
	}

	if msg.Agree == true {
		//判断房间存不存在
		room, ok := mgr.rooms[invite.RoomId]
		if !ok {
			return proto_room.ERRORCODE_ERR_NoRoom
		}

		roomPlayer := generateRoomPlayer(gen)
		ok = room.SetPlayer(roomPlayer)

		if !ok {
			return proto_room.ERRORCODE_ERR_RoomPlayerNumErr
		}
	}

	return proto_room.ERRORCODE_ERR_Ok
}

// OnLineUpHero 上阵英雄
func (mgr *Manager) OnLineUpHero(roomId int32, playerId int64) bool {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		return false
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("OnLineUpHero room id error:%v")
		return false
	}

	if room.State != define.RoomStateNormal {
		log.Error("OnLineUpHero room state error:%v", room.State)
		return false
	}

	roomPlayer := room.GetPlayer(playerId)
	if roomPlayer == nil {
		log.Error("OnLineUpHero player is not exist:%v", playerId)
		return false
	}

	return true
}

// OnPrepareGame 准备开始\取消准备
func (mgr *Manager) OnPrepareGame(_roomId int32, playerId int64, isReady bool) bool {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		log.Error("Ready is error no room : roomId:%v, %v", roomId, _roomId)
		return false
	}

	if roomId != _roomId {
		log.Error("Ready is error : roomId:%v, %v", roomId, _roomId)
		return false
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("ReadyGame room id error:%v")
		return false
	}

	if room.State != define.RoomStateNormal {
		log.Error("ReadyGame room state error:%v", room.State)
		return false
	}

	roomPlayer := room.GetPlayer(playerId)
	if roomPlayer == nil {
		log.Error("OnCancelReadyGame player is not exist:%v", playerId)
		return false
	}

	roomPlayer.IsReady = isReady

	return true
}

// OnGetRoomByPlayerId 根据玩家id查找房间信息
func (mgr *Manager) OnGetRoomByPlayerId(playerId int64) (*proto_room.RoomInfo, bool) {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		return nil, false
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("OnGetRoomByPlayerId id error:%v", roomId)
		return nil, false
	}

	return global.ToRoomInfoProto(room), true
}

// OnStartGame 开始游戏
func (mgr *Manager) OnStartGame(gen *proto_player.Context, msg *proto_room.C2SStartGame) (*proto_room.RoomInfo, error) {
	log.Debug("马上执行开始游戏")
	//判断是不是房主
	if roomInfo, ok := mgr.OnGetRoomByPlayerId(gen.Id); ok {
		for _, v := range roomInfo.GetPlayers() {
			if v.GetIsleader() == true && v.GetCommonPlayerInfo().GetPlayerId() == gen.Id {
				log.Debug("正在处理开始游戏")
				mgr.Cast("launcher", roomInfo)
				return nil, nil
			}
		}
		return nil, errors.New("start game you are not room leader")
	} else {
		return nil, errors.New("start game no room")
	}
}

// OnDisconnect TODO:断开连接
func (mgr *Manager) OnDisconnect(pl *model.Player) {
	if roomInfo, ok := mgr.OnGetRoomByPlayerId(pl.Id); ok {
		var ids []int64
		for _, v := range roomInfo.GetPlayers() {
			ids = append(ids, v.CommonPlayerInfo.PlayerId)
		}
		invoke.DispatchPlayers(mgr, ids, &proto_room.S2CDissolveRoom{
			Code: proto_room.ERRORCODE_ERR_Ok,
		})
		delete(mgr.rooms, roomInfo.GetRoomId())
		delete(mgr.roomNames, roomInfo.Name)
	}
}

// OnStartMatch 开始匹配
func (mgr *Manager) OnStartMatch(playerId int64) bool {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		return false
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("OnStartMatch room id error:%v")
		return false
	}

	if room.State != define.RoomStateNormal {
		log.Error("OnStartMatch room state error:%v", room.State)
		return false
	}

	if !room.IsPlayerInRoom(playerId) {
		log.Error("OnStartMatch player is not exist:%v", playerId)
		return false
	}

	//判断是不是房主
	if room.Owner != playerId {
		log.Error("OnStartMatch player is not owner:%v", playerId)
		return false
	}

	if !room.CanMatch() {
		log.Error("OnStartMatch room cant match:%v", playerId)
		return false
	}

	//进入匹配队列
	ok = invoke.MatchClient(mgr).StartMatch(&model.MatchTeam{
		Id:          roomId,
		AverageRank: room.AverageRank(),
		Type:        room.Type,
		IsGroup:     true,
	})

	if !ok {
		log.Error("OnStartMatch room error:%v", playerId)
		return false
	}

	// 开始匹配成功
	room.State = define.RoomStateMatch
	push := &proto_room.PushMatchTeam{
		IsStart: true,
	}

	pushList := make([]int64, 0)
	for _, v := range room.Players {
		pushList = append(pushList, v.PlayerId)
	}

	//通知进入匹配
	invoke.DispatchPlayers(mgr, pushList, push)

	return true
}

// OnCancelMatch 取消匹配
func (mgr *Manager) OnCancelMatch(playerId int64) bool {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		return false
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("OnCancelMatch room id error:%v")
		return false
	}

	if room.State != define.RoomStateMatch {
		log.Error("OnCancelMatch room state error:%v", room.State)
		return false
	}

	if !room.IsPlayerInRoom(playerId) {
		log.Error("OnCancelMatch player is not exist:%v", playerId)
		return false
	}

	// 取消匹配
	ok = invoke.MatchClient(mgr).CancelMatch(room.Type, room.Id)
	if !ok {
		log.Error("OnCancelMatch error:%v", playerId)
		return false
	}

	room.State = define.RoomStateNormal
	return true
}

// OnPlayerIsMatch 玩家是否在匹配中
func (mgr *Manager) OnPlayerIsMatch(playerId int64) bool {
	roomId, ok := mgr.playerRoom[playerId]
	if !ok {
		return false
	}

	room, ok := mgr.rooms[roomId]
	if !ok {
		log.Error("OnPlayerIsMatch room id error:%v")
		return false
	}

	if !room.IsPlayerInRoom(playerId) {
		log.Error("OnPlayerIsMatch player is not exist:%v", playerId)
		return false
	}

	return room.State == define.RoomStateMatch
}
