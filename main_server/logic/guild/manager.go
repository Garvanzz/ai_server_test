package guild

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"xfx/core/cache"
	"xfx/core/config"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/pkg/utils"
	"xfx/pkg/utils/sensitive"
	"xfx/proto/proto_guild"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"
	"xfx/proto/proto_rank"

	"github.com/golang/protobuf/proto"
)

var Module = func() module.Module {
	return Mgr
}

var Mgr *Manager

func init() {
	Mgr = new(Manager)
	Mgr.guilds = make(map[int64]*entity)
	Mgr.guildList = make([]int64, 0)
	Mgr.saveFlag = true
	Mgr.playerSaveTime = time.Now().Unix()
	Mgr.GuildApplication = make(map[int64][]*model.GuildApply)
}

type Manager struct {
	modules.BaseModule `json:"-"`
	guilds             map[int64]*entity
	GuildApplication   map[int64][]*model.GuildApply
	guildList          []int64
	lastTickTime       int64
	playerSaveTime     int64
	saveFlag           bool
	cache              *cache.WriteBackCache[int64, *model.PlayerGuild]
}

func (mgr *Manager) OnInit(app module.App) {
	mgr.BaseModule.OnInit(app)

	mgr.loadGuildData()
	mgr.cache = cache.New[int64, *model.PlayerGuild](cache.Options[int64, *model.PlayerGuild]{
		Capacity:      10000,
		DefaultTTL:    5 * time.Minute,
		FlushInterval: 30 * time.Second,
		SaveFunc:      savePlayerData,
	})

	mgr.Register("searchGuildByName", mgr.OnSearchGuildByName)                 // 根据名字搜索帮会
	mgr.Register("getGuildListByPage", mgr.OnGetGuildListByPage)               // 获取页数获取公会
	mgr.Register("createGuild", mgr.OnCreateGuild)                             // 创建帮会
	mgr.Register("joinGuild", mgr.OnJoinGuild)                                 // 直接加入帮会
	mgr.Register("dealApply", mgr.OnDealApply)                                 // 处理帮会申请
	mgr.Register("getGuildApplyList", mgr.OnGetGuildApplyList)                 // 获取帮会申请列表
	mgr.Register("getEvents", mgr.OnGetEvents)                                 // 获取帮会日志
	mgr.Register("getMemberList", mgr.OnGetMemberList)                         // 获取帮会成员列表
	mgr.Register("kickOutMember", mgr.OnKickOutMember)                         // 踢出帮会
	mgr.Register("leaveGuild", mgr.OnLeaveGuild)                               // 离开帮会
	mgr.Register("impeachMaster", mgr.OnImpeachMaster)                         // 弹劾会长
	mgr.Register("assignPosition", mgr.OnAssignPosition)                       // 任命职位
	mgr.Register("setGuildInfo", mgr.OnSetGuildInfo)                           // 设置帮会信息
	mgr.Register("playerGuildDetail", mgr.OnPlayerGuildDetail)                 // 玩家公会信息
	mgr.Register("updateMemInfo", mgr.OnUpdateMemInfo)                         // 更新玩家信息
	mgr.Register("changeGuildName", mgr.OnChangeGuildName)                     // 更改公会名
	mgr.Register("onlineBoardCast", mgr.onlineBoardCast)                       // 帮会广播
	mgr.Register("getGuildInfoByPlayerId", mgr.getGuildInfoByPlayerId)         // 根据玩家id获取帮会信息
	mgr.Register("getAllGuildId", mgr.getAllGuildId)                           // 获取所有帮会id
	mgr.Register("guildSign", mgr.GuildSign)                                   // 帮会签到
	mgr.Register("getGuildInfoByIds", mgr.getGuildInfoByIds)                   // 获取帮会信息
	mgr.Register("getPlayerInfoByIds", mgr.getPlayerInfoByIds)                 // 获取玩家帮会信息
	mgr.Register("getGuildRankInfoByIds", mgr.getGuildRankInfoByIds)           // 获取帮会排行榜信息
	mgr.Register("getGuildRankInfoByPlayerId", mgr.getGuildRankInfoByPlayerId) // 根据玩家id获取帮会排行榜信息
	mgr.Register("getPlayerGuildId", mgr.getPlayerGuildId)                     // 根据玩家帮会id
	mgr.Register("getYuanchiData", mgr.getYuanchiData)                         // 获取初始元池
	mgr.Register("yuanchiAddMaterials", mgr.addYuanchiMaterials)               // 元池-材料增加
	mgr.Register("SetBuild", mgr.SetBuild)                                     // 设置建筑
	mgr.Register("getGuildMapInfo", mgr.getGuildMapInfo)                       // 获取地图
	mgr.Register("guildPray", mgr.guildPray)                                   // 祈福
}

func (mgr *Manager) GetType() string { return define.ModuleGuild }

func (mgr *Manager) OnTick(delta time.Duration) {
	now := time.Now()

	if mgr.lastTickTime == 0 {
		mgr.lastTickTime = now.Unix()
		return
	}

	// 30分钟落库一次(分段落库)
	if now.Unix()-mgr.lastTickTime > 1800 {
		n := 0
		for _, ent := range mgr.guilds {
			if n >= define.SaveMaxCount {
				break
			}

			if ent.saveFlag == mgr.saveFlag {
				continue
			}

			ent.onSave()

			n++
			ent.saveFlag = !ent.saveFlag
		}

		if n == 0 {
			mgr.saveFlag = !mgr.saveFlag
			mgr.lastTickTime = now.Unix()
		}
	}

	// 玩家数据定时落库
	if now.Unix()-mgr.playerSaveTime >= 1700 && mgr.cache != nil {
		// 保存玩家数据
		mgr.cache.Iterate(savePlayerData)
		mgr.playerSaveTime = now.Unix()

		mgr.OnSave()
	}

	for _, ent := range mgr.guilds {
		ent.update()
	}
}

func (mgr *Manager) OnMessage(msg any) any {
	switch v := msg.(type) {
	case *event.Event:
		mgr.OnEvent(v)
	default:
		return nil
	}
	return nil
}

func (mgr *Manager) OnEvent(event *event.Event) {
	if event == nil {
		return
	}

	// 玩家基础信息
	ctx, ok := event.M["player"].(*proto_player.Context)
	if !ok {
		log.Error("guild on event player is nil")
		return
	}

	switch event.Type {
	case define.EventTypePlayerOnline:
		mgr.online(ctx)
	case define.EventTypePlayerOffline:
		mgr.offline(ctx)
	default:
	}
}

func (mgr *Manager) OnDestroy() {
	for _, ent := range mgr.guilds {
		ent.onSave()
	}

	// 缓存中的数据落库
	if mgr.cache != nil {
		mgr.cache.Close()
	}

	mgr.OnSave()
}

func (mgr *Manager) OnSave() {
	mgrData, err := json.Marshal(mgr)
	if err != nil {
		log.Error("guild mgr OnSave marshal err", err)
		return
	}

	cdb, _ := db.GetEngine(mgr.GetApp().GetEnv().ID)
	_, err = cdb.RedisExec("SET", define.GuildRedisKey, string(mgrData))
	if err != nil {
		log.Error("guild mgr OnSave db err", err)
		return
	}
}

// 玩家下线
func (mgr *Manager) offline(ctx *proto_player.Context) {
	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		log.Error("guildMgr offline error:%v", info.GuildId)
		return
	}

	delete(ent.onlineMap, info.Id)

	ent.updateMemInfo(ctx)
}

// 玩家上线
func (mgr *Manager) online(ctx *proto_player.Context) {
	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		log.Error("guildMgr online error:%v", info.GuildId)
		return
	}

	ent.onlineMap[info.Id] = struct{}{}
	ent.updateMemInfo(ctx)
}

// OnCreateGuild 创建帮会
func (mgr *Manager) OnCreateGuild(ctx *proto_player.Context, req *proto_guild.C2SCreateGuild) bool {
	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		log.Error("create guild info is null")
		return false
	}

	if info.GuildId != 0 {
		log.Error("guild guildId is not equip 0")
		return false
	}

	// 检查是否一小时内退出过帮会
	//now := time.Now()
	//if now.Unix() < info.LastQuitTime+define.JoinGuildCD && info.LastQuitTime != 0 {
	//	log.Error("create guild time limit")
	//	return false
	//}

	// 检查是否有重名的帮会
	rdb, _ := db.GetEngine(mgr.App.GetEnv().ID)
	exist, err := rdb.Mysql.Table(define.TableGuild).Where("guild_name = ?", req.Name).Exist()
	if err != nil {
		log.Error("guild name err:%v", err)
		return false
	}

	if exist {
		log.Error("exist this guild name:%v", req.Name)
		return false
	}

	guildConf, ok := config.Guild.Find(1)
	if !ok {
		log.Error("get guild config error")
		return false
	}

	ent := &entity{
		guild:     newGuild(req.Name, req.NoticeBoard, guildConf.Num, req.Banner, req.BannerColor),
		onlineMap: make(map[int64]struct{}),
	}

	// 插入到数据库
	ok = ent.onSave()
	if !ok {
		//obj.GetConnection().Send(&proto_guild.S2CCreate{Result: false})
		log.Error("guild save:%v", req.Name)
		return false
	}

	// 添加成员
	ent.addMemberInfo(ctx, define.GuildMaster, info)

	// 添加帮会日志
	ent.addGuildLog(define.GuildEventCreate, []int64{ctx.Id}, ctx.Name, req.Name)

	log.Debug("create guild id:%v,name:%v", ent.guild.Id, req.Name)

	// 更改玩家个人帮会信息
	info.GuildId = ent.guild.Id

	mgr.guilds[ent.guild.Id] = ent
	mgr.guildList = append(mgr.guildList, ent.guild.Id)

	// TODO:推送帮会信息
	//pushGuild(ent.guild.Id, 4, obj.GetDBId())

	return true
}

// OnDealApply 处理帮会申请 1同意 2拒绝
func (mgr *Manager) OnDealApply(ctx *proto_player.Context, req *proto_guild.C2SDealApply) (resp *proto_guild.S2CDealApply) {
	resp = new(proto_guild.S2CDealApply)

	dbId := ctx.Id
	info := mgr.loadPlayerGuildFromCache(dbId)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	if (req.Action != 1 && req.Action != 2) || req.PlayerId == 0 {
		log.Error("deal guild apply params error:%v,%v", req.Action, req.PlayerId)
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	// 权限判断
	if !ent.checkPermission(info.Id, define.ActionDealApply) {
		return
	}

	// 找到是申请
	application := mgr.getApplicationAndDel(req.PlayerId, ent.guild.Id)
	if application == nil {
		log.Error("get application error:%v", req.PlayerId)
		return
	}

	if req.Action == 1 { // 同意申请
		tarInfo := mgr.loadPlayerGuildFromCache(req.PlayerId)
		if tarInfo == nil {
			return
		}

		if ent.getProp(define.GuildPropCurMemberCount) >= ent.getProp(define.GuildPropMaxMemberCount) {
			resp.Result = true
			return
		}

		// 判断玩家是否已经有帮会
		if tarInfo.GuildId != 0 {
			resp.Result = true
			return
		}

		tarInfo.GuildId = ent.guild.Id

		memInfo := new(model.Member)
		memInfo.Id = tarInfo.Id
		memInfo.Account = tarInfo.Account
		memInfo.Name = tarInfo.Name
		memInfo.FaceId = tarInfo.FaceId
		memInfo.FaceSlotId = tarInfo.FaceSlotId
		memInfo.Level = tarInfo.Level
		memInfo.Position = define.GuildOrdinary
		memInfo.LastLoginTime = time.Now().Unix()
		memInfo.JoinTime = time.Now().Unix()

		ent.guild.MemberData[memInfo.Id] = memInfo
		ent.setProp(define.GuildPropCurMemberCount, int32(len(ent.guild.MemberData)))

		// 推送成员信息 1/更新 2/新增 3/移除
		ent.boardCast(&proto_guild.PushMemberInfoChange{Type: 2, MemberInfo: ent.memInfoToProto(memInfo)})

		//添加日志
		ent.addGuildLog(define.GuildEventJoin, []int64{info.Id, tarInfo.Id}, tarInfo.Name)

		// TODO：检测是否在线
		//if global.ServerG.GetObjectMgr().PlayerIsOnline(memInfo.Id) {
		//	ent.onlineMap[apply.Originator] = struct{}{}
		//}

		// TODO:推送
		//pushGuild(info.GuildId, 1, apply.Originator)
	}

	resp.Result = true
	return
}

// OnAssignPosition TODO:任命职位
func (mgr *Manager) OnAssignPosition(ctx *proto_player.Context, req *proto_guild.C2SAssignPosition) (resp *proto_guild.S2CAssignPosition) {
	resp = new(proto_guild.S2CAssignPosition)

	dbId := ctx.Id
	info := mgr.loadPlayerGuildFromCache(dbId)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	// 权限判断
	if !ent.checkPermission(ctx.Id, 0) {
		log.Error("appoint position 权限不够")
		return
	}

	// 判断是否有权限设置这个职位
	selfPosition := ent.guild.MemberData[dbId].Position
	if selfPosition != define.GuildMaster && selfPosition <= req.Position && selfPosition <= ent.guild.MemberData[req.PlayerId].Position {
		log.Error("appoint position 权限不够")
		return
	}

	// 设置的职位和目标当前职位一样
	if ent.guild.MemberData[req.PlayerId].Position == req.Position {
		log.Error("appoint same position")
		return
	}

	var logType int32

	if req.Position == define.GuildMaster { // 帮主
		ent.guild.MemberData[dbId].Position = ent.guild.MemberData[req.PlayerId].Position
		ent.guild.MemberData[req.PlayerId].Position = define.GuildMaster
		ent.setProp(define.GuildPropMaster, int32(req.PlayerId))
		logType = define.GuildEventAssignMaster
	} else if req.Position == define.GuildOrdinary { // 卸任
		ent.guild.MemberData[req.PlayerId].Position = req.Position
		logType = define.GuildEventAssignElder
	} else { // 其他职位
		num := 0
		for _, v := range ent.guild.MemberData {
			if v.Position == req.Position {
				num++
			}
		}

		guildConf, ok := config.Guild.Find(int64(ent.getProp(define.GuildPropLevel)))
		if !ok {
			log.Error("appoint position 配置错误")
			return
		}

		limit := 0
		if req.Position == define.GuildElder {
			logType = define.GuildEventAssignElder
			limit = int(guildConf.Elder)
		} else {
			logType = define.GuildEventAssignViceMaster
			limit = int(guildConf.ViceMaster)
		}

		// 判断剩余职位数量
		if num >= limit {
			log.Error("appoint position 职位人数满了")
			return
		}

		ent.guild.MemberData[req.PlayerId].Position = req.Position
	}

	// 新增日志
	ent.addGuildLog(logType, []int64{info.Id, req.PlayerId}, ent.guild.MemberData[req.PlayerId].Name)

	resp.Result = true
	return
}

// OnImpeachMaster 弹劾会长
func (mgr *Manager) OnImpeachMaster(ctx *proto_player.Context, req *proto_guild.C2SImpeachMaster) (resp *proto_guild.S2CImpeachMaster) {
	resp = new(proto_guild.S2CImpeachMaster)

	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	if !ent.checkPermission(ctx.Id, 0) {
		return
	}

	ownerInfo := ent.guild.MemberData[int64(ent.getProp(define.GuildPropMaster))]
	_, isOnline := ent.onlineMap[ownerInfo.Id]

	impeachLimit := config.Global.Get().GuildImpeachOfflineTime
	// 会长7天未上线
	if time.Now().Unix()-ownerInfo.LastLoginTime < impeachLimit*86400 && isOnline {
		return
	}

	// 帮会贡献超过1000点
	//impeachContributionLimit := config.CfgMgr.AllJson["Global"].(conf.Global).GuildImpeachNeedActivity
	//if ent.guild.MemberData[info.Id] < impeachContributionLimit {
	//	return
	//}

	// 交换职位
	ownerInfo.Position = ent.guild.MemberData[info.Id].Position
	ent.guild.MemberData[info.Id].Position = define.GuildMaster
	ent.setProp(define.GuildPropMaster, int32(info.Id))

	ent.addGuildLog(define.GuildEventImpeach, []int64{info.Id, ownerInfo.Id}, ent.guild.MemberData[ctx.Id].Name)

	resp.Result = true
	return
}

// OnJoinGuild 直接加入帮会 0/加入失败 1/直接加入帮会 2/发送帮会申请
func (mgr *Manager) OnJoinGuild(ctx *proto_player.Context, req *proto_guild.C2SJoinGuild) (resp *proto_guild.S2CJoinGuild) {
	resp = new(proto_guild.S2CJoinGuild)

	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId != 0 {
		return
	}

	// 检查是否一小时内退出过帮会
	now := time.Now()
	if now.Unix() < info.LastQuitTime+define.JoinGuildCD && info.LastQuitTime != 0 {
		log.Error("join guild time limit")
		return
	}

	ent, ok := mgr.guilds[req.Id]
	if !ok {
		log.Error("join guild id error:%v", req.Id)
		return
	}

	// 帮会人数到达上限
	if ent.getProp(define.GuildPropCurMemberCount) >= ent.getProp(define.GuildPropMaxMemberCount) {
		log.Error("join guild member num limit:%v:%v", ent.getProp(define.GuildPropCurMemberCount), ent.getProp(define.GuildPropMaxMemberCount))
		return
	}

	// TODO:是否达到帮会门槛
	//if int32(ctx.Level) < ent.getProp(define.GuildPropLevelLimit) {
	//	log.Error("join guild level limit:%v,%v", ctx.Level, ent.getProp(define.GuildPropLevelLimit))
	//	return
	//}

	if ent.getProp(define.GuildPropApplyNeedApproval) == 0 { // 不需要审批
		info.GuildId = req.Id
		ent.addMemberInfo(ctx, define.GuildOrdinary, info)
		ent.onlineMap[info.Id] = struct{}{}

		ent.addGuildLog(define.GuildEventJoin, []int64{info.Id}, ctx.Name)

		// 推送
		//pushGuild(guildId, 4, obj.GetDBId())
		resp.Result = true
		resp.Pass = true
		return
	} else { //需要审批
		// 发送申请
		ent.sendGuildApply(req.Id, ctx)

		// 推送红点
		//mgr.pushReddot(guildId, 1)
		resp.Result = true
		resp.Pass = false
		return
	}
}

// OnLeaveGuild 离开帮会
func (mgr *Manager) OnLeaveGuild(ctx *proto_player.Context, req *proto_guild.C2SLeaveGuild) (resp *proto_guild.S2CLeaveGuild) {
	resp = new(proto_guild.S2CLeaveGuild)

	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	if !ent.checkPermission(ctx.Id, 0) {
		log.Error("leave guild permission error")
		return
	}

	// 判断是否是会长 只有会长就解散帮会
	if ent.getProp(define.GuildPropMaster) == int32(info.Id) {
		// 帮会还有其他人
		if ent.getProp(define.GuildPropCurMemberCount) > 1 {
			log.Error("guild member num > 1")
			return
		}

		rdb, _ := db.GetEngine(mgr.App.GetEnv().ID)

		n, err := rdb.Mysql.Table(define.TableGuild).Where("id = ?", ent.guild.Id).Delete()
		if err != nil {
			log.Error("delete guild error:", err)
			return
		}

		if n != 1 {
			log.Error("delete guild n!=1 ,n:%v,id:%v", n, ent.guild.Id)
			return
		}

		// 清空玩家的帮会信息
		clearGuildPlayer(info)

		delete(mgr.guilds, ent.guild.Id)
		for index, v := range mgr.guildList {
			if v == ent.guild.Id {
				mgr.guildList = append(mgr.guildList[:index], mgr.guildList[index+1:]...)
			}
		}

		// 删除帮会数据
		deleteGuildData(ent.guild.Id)

		// 删除帮会战力排行
		_, err = rdb.RedisExec("zrem", "rank_guild_cup_point", ent.guild.Id)
		if err != nil {
			log.Error("delete guild rank error:%v", err)
		}

		// 删除帮会聊天记录
		_, err = rdb.RedisExec("del", fmt.Sprintf("guild_chat_history:%d", ent.guild.Id))
		if err != nil {
			log.Error("delete guild chan history error:%v", err)
		}

		log.Debug("删除帮会成功:%v", ent.guild.Id)
	} else {
		// 添加日志
		ent.addGuildLog(define.GuildEventQuit, []int64{info.Id}, ctx.Name)

		// 清空信息
		ent.removeMemberInfo(info)

		// 清空玩家帮会信息
		clearGuildPlayer(info)
	}

	log.Debug("退出帮会成功")
	resp.Result = true
	return
}

// OnKickOutMember 踢出帮会
func (mgr *Manager) OnKickOutMember(ctx *proto_player.Context, req *proto_guild.C2SKickOutMember) (resp *proto_guild.S2CKickOutMember) {
	resp = new(proto_guild.S2CKickOutMember)

	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	tarInfo := mgr.loadPlayerGuildFromCache(req.PlayerId)
	if tarInfo == nil {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	// 权限判断
	if !ent.checkPermission(info.Id, define.ActionKickOut) {
		return
	}

	// 确保对方在当前帮会
	tarMemInfo, ok := ent.guild.MemberData[req.PlayerId]
	if !ok {
		log.Error("kick out member target error:%v", req.PlayerId)
		return
	}

	// 判断是否自己的职位比对方高
	if ent.guild.MemberData[ctx.Id].Position <= tarMemInfo.Position {
		return
	}

	// 添加日志
	ent.addGuildLog(define.GuildEventKickOut, []int64{info.Id, tarInfo.Id}, ctx.Name, ent.guild.MemberData[req.PlayerId].Name)

	// 清空信息
	ent.removeMemberInfo(tarInfo)

	//清空玩家信息
	clearGuildPlayer(tarInfo)

	// 推送被踢出帮会玩家
	pushGuild(0, 2, req.PlayerId)

	resp.Result = true
	return
}

// OnGetEvents 获取帮会事件
func (mgr *Manager) OnGetEvents(ctx *proto_player.Context, req *proto_guild.C2SGuildEvent) (resp *proto_guild.S2CGuildEvent) {
	resp = new(proto_guild.S2CGuildEvent)

	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		log.Error("is not guild member1")
		return
	}

	logs := make([]*model.GuildLog, 0)
	rdb, _ := db.GetEngine(mgr.App.GetEnv().ID)

	err := rdb.Mysql.Table(define.TableGuildLog).Where("guild_id = ?", info.GuildId).Desc("timestamp").Limit(define.LogCountMax).Find(&logs)
	if err != nil {
		log.Error("load guild log error:%v", err)
		return
	}

	ret := make([]*proto_guild.Event, 0, len(logs))
	for _, v := range logs {
		ret = append(ret, &proto_guild.Event{
			Timestamp: v.Timestamp,
			Action:    v.Action,
			Params:    v.Params,
		})
	}

	log.Debug("获取帮会日志:%v", ret)
	resp.Events = ret
	return
}

// OnGetMemberList 获取帮会成员列表
func (mgr *Manager) OnGetMemberList(ctx *proto_player.Context, req *proto_guild.C2SGetMemberList) (resp *proto_guild.S2CGetMemberList) {
	resp = new(proto_guild.S2CGetMemberList)

	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		log.Error("get member list player guild id is 0")
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	resp.List = ent.getMemberList()
	return
}

func (ent *entity) getMemberList() []*proto_guild.MemberInfo {
	l := make([]*proto_guild.MemberInfo, 0)
	for _, memInfo := range ent.guild.MemberData {
		l = append(l, ent.memInfoToProto(memInfo))
	}
	return l
}

// getGuildInfoById 根据id获取帮会信息
func (mgr *Manager) getGuildInfoById(guildId int64) *model.Guild {
	if guildId == 0 {
		return nil
	}

	if ent, ok := mgr.guilds[guildId]; !ok {
		return nil
	} else {
		guild := new(model.Guild)
		*guild = *ent.guild
		return guild
	}
}

// getGuildInfoByPlayerId 根据玩家id获取帮会信息
func (mgr *Manager) getGuildInfoByPlayerId(dbId int64) *model.Guild {
	info := mgr.loadPlayerGuildFromCache(dbId)
	if info == nil {
		return nil
	}

	return mgr.getGuildInfoById(info.GuildId)
}

// onlineBoardCast 帮会广播
func (mgr *Manager) onlineBoardCast(guildId int64, message proto.Message) {
	ent, ok := mgr.guilds[guildId]
	if !ok {
		return
	}

	ent.boardCast(message)
}

func (mgr *Manager) joinGuild(ctx *proto_player.Context, guildId int64, info *model.PlayerGuild) bool {
	ent, ok := mgr.guilds[guildId]
	if !ok {
		log.Error("random join guild get entity error")
		return false
	}

	// 帮会人数到达上限
	if ent.getProp(define.GuildPropCurMemberCount) >= ent.getProp(define.GuildPropMaxMemberCount) {
		return false
	}

	// 是否达到帮会门槛
	level := int32(ctx.Level)
	ignoreCondition := ent.getProp(define.GuildPropIgnoreLevelLimit) == 1
	if level < ent.getProp(define.GuildPropLevelLimit) && ignoreCondition {
		return false
	}

	// 是否需要审批
	if ent.getProp(define.GuildPropApplyNeedApproval) == 1 {
		return false
	}

	info.GuildId = guildId
	// 添加成员信息
	ent.addMemberInfo(ctx, define.GuildOrdinary, info)

	ent.addGuildLog(define.GuildEventJoin, []int64{info.Id}, ctx.Name)

	// 推送 TODO:
	//pushGuild(guildId, 4, obj.GetDBId())
	return true
}

// OnGetGuildApplyList 获取帮会申请列表
func (mgr *Manager) OnGetGuildApplyList(ctx *proto_player.Context) (resp *proto_guild.S2CGetApply) {
	resp = new(proto_guild.S2CGetApply)

	dbId := ctx.Id
	info := mgr.loadPlayerGuildFromCache(dbId)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	now := time.Now().Unix()

	applications := mgr.GuildApplication[info.GuildId]
	list := make([]*proto_guild.Apply, 0)
	for _, v := range applications {
		if v.Expiration < now {
			continue
		}

		list = append(list, &proto_guild.Apply{
			PlayerId:   v.PlayerId,
			Name:       v.Name,
			Level:      v.Level,
			FaceId:     v.FaceId,
			FaceSlotId: v.FaceSlotId,
		})
	}

	resp.Applys = list
	return
}

// 推送帮会信息变化 (通知类型 1/同意申请 2/被踢出帮会 4/加入帮会)
func pushGuild(guildId int64, t int32, dbId int64) {
	//invoke.CastAgent()
	//obj := global.ServerG.GetObjectMgr().GetPlayer(dbId)
	//if obj != nil && obj.IsOnline() {
	//	obj.GetConnection().Send(&proto_guild.PushGuild{GuildId: int32(guildId), Type: t})
	//}
}

// 推送红点
func (mgr *Manager) pushReddot(guildId int64, reddotType int32) {
	ent, ok := mgr.guilds[guildId]
	if !ok {
		return
	}

	switch reddotType {
	case 1: // 新申请
		pushList := make([]int64, 0)
		for k := range ent.onlineMap {
			if ent.guild.MemberData[k].Position >= define.GuildViceMaster {
				pushList = append(pushList, k)
			}
		}

		//for _, id := range pushList {
		//isOnline := global.ServerG.GetObjectMgr().PlayerIsOnline(id)
		//if isOnline {
		//	obj := global.ServerG.GetObjectMgr().GetPlayer(id)
		//	if obj != nil {
		//		obj.GetConnection().Send(&proto_guild.PushReddot{ReddotType: reddotType})
		//	}
		//}
		//}
	default:
	}
}

// OnSetGuildInfo 设置帮会信息
func (mgr *Manager) OnSetGuildInfo(ctx *proto_player.Context, req *proto_guild.C2SSetGuildRule) (resp *proto_guild.S2CSetGuildRule) {
	resp = new(proto_guild.S2CSetGuildRule)

	dbId := ctx.Id
	info := mgr.loadPlayerGuildFromCache(dbId)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	//检查权限
	if !ent.checkPermission(info.Id, define.ActionSetGuildRule) {
		return
	}

	if req.Banner != 0 {
		ent.setProp(define.GuildPropBanner, req.Banner)
	}
	if req.BannerColor != 0 {
		ent.setProp(define.GuildPropBannerColor, req.BannerColor)
	}

	if req.NoticeBoard != "" {
		ent.guild.NoticeBoard = req.NoticeBoard
	}

	ent.setProp(define.GuildPropLevelLimit, req.LevelLimit)
	ent.setProp(define.GuildPropApplyNeedApproval, utils.BoolToInt(req.ApplyNeedApproval))
	ent.setProp(define.GuildPropIgnoreLevelLimit, utils.BoolToInt(req.IgnoreLevelLimit))

	resp.Result = true
	return
}

// OnPlayerGuildDetail 玩家公会信息
func (mgr *Manager) OnPlayerGuildDetail(ctx *proto_player.Context, req *proto_guild.C2SPlayerGuildDetail) (resp *proto_guild.S2CPlayerGuildDetail) {
	resp = new(proto_guild.S2CPlayerGuildDetail)

	dbId := ctx.Id
	info := mgr.loadPlayerGuildFromCache(dbId)
	if info == nil {
		return
	}

	resp.JoinGuildCd = info.LastQuitTime
	resp.GuildId = info.GuildId
	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	resp.Guild = model.GuildToProto(ent.guild)
	//签到
	if info.ToDaySign {
		//判定时间
		if !utils.CheckIsSameDayBySec(info.SignTime, time.Now().Unix(), 0) {
			info.ToDaySign = false
			if info.SignDay >= 7 {
				info.SignDay = 0
			}
		}
	}

	resp.SignItem = &proto_guild.GuildSignItem{
		Sign:        info.SignDay,
		IsTodaySign: info.ToDaySign,
	}

	//祈福
	if info.GuildPray == nil {
		info.GuildPray = new(model.GuildPrayItem)
	}

	if !utils.CheckIsSameDayBySec(info.GuildPray.TodayPrayTime, time.Now().Unix(), 0) {
		info.GuildPray.IsTodayPray = false
	}

	resp.Pray = &proto_guild.GuildPrayItem{
		IsTodayPray: info.GuildPray.IsTodayPray,
		Index:       info.GuildPray.PrayType,
		RangType:    info.GuildPray.RangeType,
		RangValue:   info.GuildPray.RangeValue,
	}

	//场景地图
	if info.GuildMap == nil {
		info.GuildMap = make(map[int32]*model.GuildMapItem, 0)
	}

	resp.Items = make(map[int32]*proto_guild.GuildBuildMapItem)
	for _, v := range info.GuildMap {
		resp.Items[v.Index] = &proto_guild.GuildBuildMapItem{
			Id:    v.Id,
			Index: v.Index,
			Level: v.Level,
		}
	}

	return
}

// OnUpdateMemInfo 更新玩家信息
func (mgr *Manager) OnUpdateMemInfo(ctx *proto_player.Context) {
	info := mgr.loadPlayerGuildFromCache(ctx.Id)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	ent.updateMemInfo(ctx)
}

// OnChangeGuildName 更改公会名
func (mgr *Manager) OnChangeGuildName(ctx *proto_player.Context, name string) (resp *proto_guild.S2CChangeGuildName) {
	resp = new(proto_guild.S2CChangeGuildName)

	dbId := ctx.Id
	info := mgr.loadPlayerGuildFromCache(dbId)
	if info == nil {
		return
	}

	if info.GuildId == 0 {
		return
	}

	ent, ok := mgr.guilds[info.GuildId]
	if !ok {
		return
	}

	//检查权限
	if !ent.checkPermission(info.Id, define.ActionSetGuildRule) {
		return
	}

	if name == "" {
		return
	}

	// TODO:检查是否有重复的公会名
	for _, v := range mgr.guilds {
		if v.guild.GuildName == name {
			resp.Result = false
			return
		}
	}

	//名字过滤
	if sensitive.Filter.IsSensitive(name) {
		resp.Result = false
		return
	}

	// 修改公会名
	ent.guild.GuildName = name
	mgr.guilds[info.GuildId] = ent

	resp.Result = true
	return
}

// OnGetGuildListByPage 根据页数获取帮会列表
func (mgr *Manager) OnGetGuildListByPage(ctx *proto_player.Context, req *proto_guild.C2SGuildByPage) *proto_guild.S2CGuildByPage {
	resp := new(proto_guild.S2CGuildByPage)

	start := (req.Page-1)*define.PageMax + 1

	//rdb, _ := db.GetEngine(mgr.App.GetEnv().ID)
	//reply, err := rdb.RedisExec("zrevrange", "rank_guild_cup_point", 0, -1)
	//if err != nil {
	//	log.Error("get guild cup point rank error:", err)
	//	return resp
	//}

	ret := make([]*proto_guild.Guild, 0)

	dbId := ctx.Id

	count := int32(0)
	//res, _ := reply.([]any)
	//for i := 0; i < len(res); i++ {
	for _, guildId := range mgr.guildList {
		//str := string(res[i].([]byte))
		//guildId, _ := strconv.ParseInt(str, 10, 64)

		ent, ok := mgr.guilds[guildId]
		if !ok {
			log.Error("mgr guild data error2:%v", guildId)
			continue
		}

		count++

		if count < start || count > start+define.PageMax {
			continue
		}

		guildProto := model.GuildToProto(ent.guild)
		guildProto.HasSendApply = mgr.hasSendApplication(dbId, ent.guild.Id)
		ret = append(ret, guildProto)
	}

	pageMax := count / define.PageMax
	if count%define.PageMax != 0 {
		pageMax++
	}

	resp.Guilds = ret

	resp.Count = count
	resp.CurPage = req.Page
	resp.PageMax = pageMax
	return resp
}

// OnSearchGuildByName 根据帮会名搜索帮会
func (mgr *Manager) OnSearchGuildByName(ctx *proto_player.Context, req *proto_guild.C2SSearchByName) *proto_guild.S2CSearchByName {
	resp := new(proto_guild.S2CSearchByName)

	dbId := ctx.Id

	ret := make([]*proto_guild.Guild, 0)
	for _, id := range mgr.guildList {
		ent, ok := mgr.guilds[id]
		if !ok {
			log.Error("mgr guild data error:%v", id)
			continue
		}

		if !strings.Contains(ent.guild.GuildName, req.Name) {
			continue
		}

		guildProto := model.GuildToProto(ent.guild)
		guildProto.HasSendApply = mgr.hasSendApplication(dbId, ent.guild.Id)
		ret = append(ret, guildProto)
	}

	resp.Guilds = ret
	return resp
}

// getAllGuildId 获取所有帮会id
func (mgr *Manager) getAllGuildId() []int64 {
	list := make([]int64, len(mgr.guildList))
	copy(list, mgr.guildList)
	return list
}

// 是否加入公会申请
func (mgr *Manager) hasSendApplication(playerId, guildId int64) bool {
	ent, ok := mgr.guilds[guildId]
	if !ok {
		return false
	}

	now := time.Now().Unix()
	applications := mgr.GuildApplication[ent.guild.Id]
	for _, application := range applications {
		if application.PlayerId == playerId && now < application.Expiration {
			return true
		}
	}
	return false
}

// 获取公会申请并删除
func (mgr *Manager) getApplicationAndDel(playerId, guildId int64) *model.GuildApply {
	ent, ok := mgr.guilds[guildId]
	if !ok {
		return nil
	}

	applications := mgr.GuildApplication[ent.guild.Id]
	for i, application := range applications {
		if application.PlayerId == playerId {
			// 删除申请
			mgr.GuildApplication[ent.guild.Id] = append(applications[:i], applications[i+1:]...)

			return application
		}
	}
	return nil
}

// 签到
func (mgr *Manager) GuildSign(playerId int64) (*model.PlayerGuild, error) {
	info := mgr.loadPlayerGuildFromCache(playerId)
	if info == nil {
		return nil, errors.New("sign faild: no player")
	}

	if info.GuildId == 0 {
		return nil, errors.New("sign faild: no guild")
	}

	if info.ToDaySign {
		return info, errors.New("sign faild: had sign")
	}

	if info.SignDay >= 7 {
		info.SignDay = 0
	}
	info.SignDay += 1
	info.ToDaySign = true
	info.SignTime = time.Now().Unix()

	return info, nil
}

// 获取帮会信息
func (mgr *Manager) getGuildInfoByIds(ids []int64) map[int64]*proto_guild.Guild {
	ret := make(map[int64]*proto_guild.Guild)

	for _, id := range ids {
		ent, ok := mgr.guilds[id]
		if !ok {
			continue
		}

		ret[id] = model.GuildToProto(ent.guild)
	}
	return ret
}

// 获取玩家工会信息
func (mgr *Manager) getPlayerInfoByIds(ids []int64) map[int64]*proto_public.CommonPlayerInfo {
	ret := make(map[int64]*proto_public.CommonPlayerInfo)
	for _, id := range ids {
		info := mgr.loadPlayerGuildFromCache(id)
		if info == nil {
			continue
		}

		if info.GuildId == 0 {
			log.Error("get guild PlayerInfoByIds info failed: %v", id)
			continue
		}

		ent, ok := mgr.guilds[info.GuildId]
		if !ok {
			continue
		}

		model.GuildToProto(ent.guild)
		ret[id] = &proto_public.CommonPlayerInfo{
			PlayerId:   id,
			Name:       info.Name,
			FaceId:     info.FaceId,
			FaceSlotId: info.FaceSlotId,
			Level:      info.Level,
			Position:   ent.getPosition(id),
		}
	}

	return ret
}

// 获取帮会排行榜信息
func (mgr *Manager) getGuildRankInfoByIds(ids []int64) map[int64]*proto_rank.GuildRankItem {
	log.Debug("getGuildRankInfoByIds start:%v", ids)
	ret := make(map[int64]*proto_rank.GuildRankItem)

	for _, id := range ids {
		ent, ok := mgr.guilds[id]
		if !ok {
			continue
		}

		master := ent.guild.MemberData[int64(ent.getProp(define.GuildPropMaster))]
		ret[id] = &proto_rank.GuildRankItem{
			Id:               ent.guild.Id,
			Name:             ent.guild.GuildName,
			Level:            ent.getProp(define.GuildPropLevel),
			Banner:           ent.getProp(define.GuildPropBanner),
			BannerColor:      ent.getProp(define.GuildPropBannerColor),
			Master:           int64(ent.getProp(define.GuildPropMaster)),
			CurMemberNum:     ent.getProp(define.GuildPropCurMemberCount),
			MasterFaceId:     master.FaceId,
			MasterFaceSlotId: master.FaceSlotId,
			MasterLevel:      master.Level,
			MasterName:       master.Name,
		}
	}
	log.Debug("getGuildRankInfoByIds end:%v", ret)
	return ret
}

// 获取帮会排行榜信息
func (mgr *Manager) getGuildRankInfoByPlayerId(playerId int64) *proto_rank.GuildRankItem {
	info := mgr.loadPlayerGuildFromCache(playerId)
	if info == nil || info.GuildId == 0 {
		return nil
	}

	return mgr.getGuildRankInfoByIds([]int64{info.GuildId})[info.GuildId]
}

// 获取玩家帮会id
func (mgr *Manager) getPlayerGuildId(playerId int64) int64 {
	info := mgr.loadPlayerGuildFromCache(playerId)
	if info == nil {
		return 0
	}

	return info.GuildId
}
