package guild

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"time"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/proto/proto_guild"
	"xfx/proto/proto_player"
)

func newGuild(name, noticeBoard string, memLimit, banner, bannerColor int32) *model.Guild {
	guild := new(model.Guild)
	guild.GuildName = name
	guild.NoticeBoard = noticeBoard
	guild.MemberData = make(map[int64]*model.Member)
	guild.LastRefreshTime = time.Now().Unix()
	guild.Props[define.GuildPropMaxMemberCount] = memLimit
	guild.Props[define.GuildPropBanner] = banner
	guild.Props[define.GuildPropBannerColor] = bannerColor
	guild.Props[define.GuildPropLevel] = 1
	guild.Props[define.GuildPropGrowth] = 0
	guild.Props[define.GuildPropAddsucrare] = 0
	guild.Props[define.GuildPropReducetime] = 0
	guild.Props[define.GuildPropTitle] = 1 //默认主题
	guild.Yuanchi = new(model.GuildYuanchi)
	return guild
}

type entity struct {
	guild     *model.Guild
	onlineMap map[int64]struct{}
	saveFlag  bool
}

// 是否是帮会成员
func (ent *entity) isMember(dbId int64) bool {
	_, ok := ent.guild.MemberData[dbId]
	return ok
}

func (ent *entity) getPosition(dbId int64) int32 {
	memInfo, ok := ent.guild.MemberData[dbId]
	if !ok {
		return 0
	}
	return memInfo.Position
}

func (ent *entity) modifyProp(t int, value int32, add bool) {
	before := ent.getProp(t)
	if add {
		ent.setProp(t, before+value)
	} else {
		after := before - value
		if after > 0 {
			ent.setProp(t, after)
		} else {
			ent.setProp(t, 0)
		}
	}
}

func (ent *entity) setProp(t int, value int32) {
	ent.guild.Props[t] = value
}

func (ent *entity) getProp(t int) int32 {
	return ent.guild.Props[t]
}

// 移除玩家信息
func (ent *entity) removeMemberInfo(info *model.PlayerGuild) {
	delete(ent.guild.MemberData, info.Id)
	delete(ent.onlineMap, info.Id)
	ent.setProp(define.GuildPropCurMemberCount, int32(len(ent.guild.MemberData)))

	// TODO:推送成员信息
	//ent.boardCast(&proto_guild.PushMemberInfoChange{Type: 3, MemberInfo: &proto_guild.MemberInfo{Id: info.Id}})
}

// 新增玩家信息
func (ent *entity) addMemberInfo(ctx *proto_player.Context, position int32, info *model.PlayerGuild) {
	memInfo := new(model.Member)
	memInfo.Id = info.Id
	memInfo.Position = position
	memInfo.LastLoginTime = time.Now().Unix()
	memInfo.JoinTime = time.Now().Unix()

	ent.guild.MemberData[memInfo.Id] = memInfo
	ent.setProp(define.GuildPropCurMemberCount, int32(len(ent.guild.MemberData)))

	// 设置会长
	if position == define.GuildMaster {
		ent.setProp(define.GuildPropMaster, int32(info.Id))
	}

	ent.updateMemInfo(ctx)

	if isOnline := invoke.LoginClient(Mgr).IsOnline(memInfo.Id); isOnline {
		ent.onlineMap[memInfo.Id] = struct{}{}
	}

	// TODO:推送成员信息 1/更新 2/新增 3/移除
	//ent.boardCast(&proto_guild.PushMemberInfoChange{Type: 2, MemberInfo: ent.memInfoToProto(memInfo)})
}

// 检查权限
func (ent *entity) checkPermission(dbId int64, action int) bool {
	if memInfo, ok := ent.guild.MemberData[dbId]; !ok {
		return false
	} else {
		if action == 0 {
			return true
		}

		if limit, ok := define.GuildPermission[action]; !ok {
			log.Error("check permission action not found:%v", action)
			return false
		} else {
			return memInfo.Position >= limit
		}
	}
}

// 加载帮会数据
func (mgr *Manager) loadGuildData() {
	cdb, _ := db.GetEngine(mgr.GetApp().GetEnv().ID)
	reply, err := cdb.RedisExec("GET", define.GuildRedisKey)
	if err != nil {
		log.Error("guild mgr load data err", err)
		return
	}

	if reply != nil {
		err = json.Unmarshal(reply.([]byte), mgr)
		if err != nil {
			log.Error("guild mgr OnSave unmarshal err", err)
			return
		}
	}

	guildDBs := make([]*model.GuildDB, 0)
	rdb, _ := db.GetEngine(mgr.App.GetEnv().ID)
	err = rdb.Mysql.Table(define.TableGuild).Find(&guildDBs)
	if err != nil {
		log.Error("load all guild info error:%v", err)
		return
	}

	for _, guildDB := range guildDBs {
		ent := new(entity)
		ent.guild = guildDbToGuild(guildDB)
		ent.onlineMap = make(map[int64]struct{})
		mgr.guilds[ent.guild.Id] = ent
		mgr.guildList = append(mgr.guildList, ent.guild.Id)
	}
}

func guildDbToGuild(guildDB *model.GuildDB) *model.Guild {
	guild := new(model.Guild)
	guild.Id = guildDB.Id
	guild.NoticeBoard = guildDB.NoticeBoard
	guild.GuildName = guildDB.GuildName
	guild.MemberData = guildDB.MemberData
	guild.Props[define.GuildPropBanner] = guildDB.Banner
	guild.Props[define.GuildPropBannerColor] = guildDB.BannerColor
	guild.Props[define.GuildPropLevelLimit] = guildDB.LevelLimit
	guild.Props[define.GuildPropMaster] = guildDB.Master
	guild.Props[define.GuildPropIgnoreLevelLimit] = guildDB.IgnoreLevelLimit
	guild.Props[define.GuildPropMaxMemberCount] = guildDB.MaxMemberCount
	guild.Props[define.GuildPropCurMemberCount] = guildDB.CurMemberCount
	guild.Props[define.GuildPropApplyNeedApproval] = guildDB.ApplyNeedApproval
	guild.Props[define.GuildPropLevel] = guildDB.Level
	guild.Props[define.GuildPropExp] = guildDB.Exp
	guild.Props[define.GuildPropGrowth] = guildDB.Growth
	guild.Props[define.GuildPropReducetime] = guildDB.ReduceTime
	guild.Props[define.GuildPropAddsucrare] = guildDB.AddSucRare
	guild.Props[define.GuildPropTitle] = guildDB.Title

	//反序列化
	if guildDB.Yuanchi != "" {
		var yuanchi *model.GuildYuanchi
		err := json.Unmarshal([]byte(guildDB.Yuanchi), &yuanchi)
		if err != nil {
			log.Error("guild mgr OnSave unmarshal err", err)
		}
		guild.Yuanchi = yuanchi
	} else {
		guild.Yuanchi = new(model.GuildYuanchi)
	}

	return guild
}

func guildToGuildDb(guild *model.Guild) *model.GuildDB {
	guildDB := new(model.GuildDB)
	guildDB.Id = guild.Id
	guildDB.NoticeBoard = guild.NoticeBoard
	guildDB.GuildName = guild.GuildName
	guildDB.MemberData = guild.MemberData
	guildDB.Banner = guild.Props[define.GuildPropBanner]
	guildDB.BannerColor = guild.Props[define.GuildPropBannerColor]
	guildDB.LevelLimit = guild.Props[define.GuildPropLevelLimit]
	guildDB.Master = guild.Props[define.GuildPropMaster]
	guildDB.IgnoreLevelLimit = guild.Props[define.GuildPropIgnoreLevelLimit]
	guildDB.MaxMemberCount = guild.Props[define.GuildPropMaxMemberCount]
	guildDB.CurMemberCount = guild.Props[define.GuildPropCurMemberCount]
	guildDB.ApplyNeedApproval = guild.Props[define.GuildPropApplyNeedApproval]
	guildDB.Level = guild.Props[define.GuildPropLevel]
	guildDB.Exp = guild.Props[define.GuildPropExp]
	guildDB.Growth = guild.Props[define.GuildPropGrowth]
	guildDB.ReduceTime = guild.Props[define.GuildPropReducetime]
	guildDB.AddSucRare = guild.Props[define.GuildPropAddsucrare]
	guildDB.Title = guild.Props[define.GuildPropTitle]

	//序列化
	js, _ := json.Marshal(guild.Yuanchi)
	guildDB.Yuanchi = string(js)
	return guildDB
}

// 更新帮会信息
func (ent *entity) onSave() bool {

	rdb, _ := db.GetEngine(Mgr.App.GetEnv().ID)

	guildDB := guildToGuildDb(ent.guild)
	if guildDB.Id == 0 { // 插入联盟信息
		n, err := rdb.Mysql.Table(define.TableGuild).Insert(guildDB)
		if err != nil {
			log.Error("insert guild db error: %v", err)
			return false
		}

		if n == 0 {
			log.Error("insert guild db failed: %v", guildDB.GuildName)
			return false
		}

		ent.guild.Id = guildDB.Id
	} else {
		_, err := rdb.Mysql.Table(define.TableGuild).Where("id = ?", guildDB.Id).AllCols().Update(guildDB)
		if err != nil {
			log.Error("update guild db error: %v", err)
			return false
		}
	}
	return true
}

// tick
func (ent *entity) update() {
	now := time.Now()
	if ent.guild.LastRefreshTime == 0 {
		ent.guild.LastRefreshTime = now.Unix()
		return
	}

	//元池tick
	ent.updateYuanchi()
}

// 发送帮会申请
func (ent *entity) sendGuildApply(guildId int64, ctx *proto_player.Context) {
	Mgr.getApplicationAndDel(guildId, ent.guild.Id)

	applications := Mgr.GuildApplication[ent.guild.Id]
	if applications == nil {
		applications = make([]*model.GuildApply, 0)
	}

	apply := new(model.GuildApply)
	apply.PlayerId = ctx.Id
	apply.Account = ctx.Uid
	apply.Name = ctx.Name
	apply.Level = int32(ctx.Level)
	apply.FaceId = int32(ctx.FaceId)
	apply.FaceSlotId = int32(ctx.FaceSlotId)
	apply.Expiration = time.Now().Unix() + define.ApplyKeepTime

	applications = append(applications, apply)

	// 判断联盟申请的条数是否超过上限
	if len(applications) > define.ApplyCountMax {
		applications = applications[len(applications)-define.ApplyCountMax:]
	}

	Mgr.GuildApplication[ent.guild.Id] = applications
}

// 添加帮会日志到mysql
func (ent *entity) addGuildLog(action int32, dbId []int64, params ...string) {
	guildLog := new(model.GuildLog)
	guildLog.GuildId = ent.guild.Id
	guildLog.Action = action
	guildLog.Timestamp = time.Now().Unix()
	guildLog.Params = params
	guildLog.DbId = dbId

	rdb, _ := db.GetEngine(Mgr.App.GetEnv().ID)
	n, err := rdb.Mysql.Table(define.TableGuildLog).Insert(guildLog)
	if err != nil {
		log.Error("insert guild log error:%v,%v", err, guildLog)
		return
	}

	if n == 0 {
		log.Error("insert guild log failed:%v,%v", guildLog)
		return
	}
}

// TODO:更新普通信息  // 写入下线时间
func (ent *entity) updateMemInfo(ctx *proto_player.Context) {
	dbId := ctx.Id
	memInfo, ok := ent.guild.MemberData[dbId]
	if !ok {
		log.Error("get member data error:%v", dbId)
		return
	}

	memInfo.FaceId = int32(ctx.FaceId)
	memInfo.FaceSlotId = int32(ctx.FaceSlotId)
	memInfo.Name = ctx.Name
	memInfo.Level = int32(ctx.Level)
	//memInfo.LastLoginTime =utils.Convert[int64](gen, define.Gen) TODO:上次登陆时间
}

func (ent *entity) memInfoToProto(memInfo *model.Member) *proto_guild.MemberInfo {
	var lastLoginTime int64
	_, isOnline := ent.onlineMap[memInfo.Id]
	if isOnline {
		lastLoginTime = 0
	} else {
		lastLoginTime = time.Now().Unix() - memInfo.LastLoginTime
	}

	return &proto_guild.MemberInfo{
		Id:            memInfo.Id,
		Name:          memInfo.Name,
		FaceId:        memInfo.FaceId,
		Position:      memInfo.Position,
		LastLoginTime: lastLoginTime,
		Level:         memInfo.Level,
		FaceSlotId:    memInfo.FaceSlotId,
	}
}

func (ent *entity) ProtoTomemInfo(info *proto_guild.MemberInfo) *model.Member {
	var lastLoginTime int64
	_, isOnline := ent.onlineMap[info.Id]
	if isOnline {
		lastLoginTime = 0
	} else {
		lastLoginTime = time.Now().Unix() - info.LastLoginTime
	}

	return &model.Member{
		Id:            info.Id,
		Name:          info.Name,
		FaceId:        info.FaceId,
		Position:      info.Position,
		LastLoginTime: lastLoginTime,
		Level:         info.Level,
		FaceSlotId:    info.FaceSlotId,
	}
}

// 删除帮会数据
func deleteGuildData(guildId int64) {
	rdb, _ := db.GetEngine(Mgr.App.GetEnv().ID)
	_, err := rdb.Mysql.Table(define.TableGuildLog).Where("guild_id = ?", guildId).Delete()
	if err != nil {
		log.Error("delete guild log error:%,n:%v", err)
		return
	}

	// 删除帮会申请
	delete(Mgr.GuildApplication, guildId)
}

// TODO:帮会广播
func (ent *entity) boardCast(message proto.Message) {
	for id := range ent.onlineMap {
		invoke.Dispatch(Mgr, id, message)
	}
}

// TODO:发送公告
//func (ent *entity) sendAnnouncement(messageId int64, cn, en string, params []*proto_talk.ParamUnit) {
//	push := new(proto_talk.PushTalk)
//	push.Talk = new(proto_talk.TalkData)
//	push.Talk.Header = new(proto_talk.TalkHeader)
//	push.Talk.Header.Columns = 1 //消息分栏 联盟
//	push.Talk.Header.Type = 2    //类型 2=系统
//	push.Talk.Player = nil
//	push.Talk.Msg = ""
//	push.Talk.Timestamp = time.Now().Unix()
//	push.Talk.Goods = nil
//	push.Talk.SysMsgId = messageId
//	push.Talk.SysMsgCN = cn
//	push.Talk.SysMsgEN = en
//	push.Talk.SysParams = params
//
//	new redis存
//temp, err := json.Marshal(push.Talk)
//if err != nil {
//	log.Error("marshal system chat message error : %v", err)
//	return
//} else {
//	Mgr.AddGuildChatHistory(string(temp), ent.guild.Id)
//	ent.boardCast(push)
//}
//}

// AddGuildChatHistory 添加帮会聊天记录缓存
func (mgr *Manager) AddGuildChatHistory(message string, guildId int64) {
	if guildId == 0 {
		return
	}

	key := fmt.Sprintf("guild_chat_history:%d", guildId)

	rdb, _ := db.GetEngine(mgr.App.GetEnv().ID)

	//尾部添
	rdb.RedisExec("RPUSH", key, message)

	//然后整理长度 不超过指定存储数量(ChatMsgLen)
	rdb.RedisExec("LTRIM", key, 0-define.ChatMsgLen, -1)
}
