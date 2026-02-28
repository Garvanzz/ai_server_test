package model

import (
	"xfx/core/define"
	"xfx/pkg/log"
	"xfx/proto/proto_guild"
)

// PlayerGuild 玩家帮会信息
type PlayerGuild struct {
	Id              int64                   // 玩家dbId
	Account         string                  // 玩家account // TODO:
	GuildId         int64                   // 所属帮会
	Name            string                  // 玩家名
	FaceId          int32                   // 头像
	FaceSlotId      int32                   // 头像框
	Level           int32                   // 等级
	LastQuitTime    int64                   // 上次退出帮会的时间
	LastRefreshTime int64                   // 上次刷新时间
	ToDaySign       bool                    // 今日是否签到
	SignTime        int64                   // 签到时间
	SignDay         int32                   // 签到天数
	GuildMap        map[int32]*GuildMapItem //地图元素
	GuildPray       *GuildPrayItem          //祈福
}

// Guild 帮会信息
type Guild struct {
	Id              int64                      // 帮会唯一id
	NoticeBoard     string                     // 公告栏信息
	GuildName       string                     // 帮会名
	MemberData      map[int64]*Member          // 帮会成员
	Props           [define.GuildPropMax]int32 // 帮会信息
	LastRefreshTime int64                      // 上次刷新时间
	Yuanchi         *GuildYuanchi              //元池
}

// 元池
type GuildYuanchi struct {
	Materials map[int32]int32            //材料
	Refinings map[int32]*YuanchiRefining //炼制中
	Elements  map[int32]int32            //元素
}

type YuanchiRefining struct {
	Id          int32 //炼制Id
	YuanchiItem map[int32]*YuanchiItem
	AllTime     int32 //总时长
	Time        int32 //炼制时长
	Rate        int32 //成功率
	LastTime    int64
}

type YuanchiItem struct {
	Id  int32
	Num int32
}

// GuildLog 帮会炼制记录
type GuildRefiningLog struct {
	Data  *YuanchiRefining
	State bool
	Time  int64
}

type GuildApply struct {
	PlayerId   int64
	Account    string // account id
	Name       string // 玩家名
	Level      int32  // 玩家等级
	FaceId     int32  // 玩家头像
	FaceSlotId int32  // 玩家头像框
	Expiration int64  // 过期时间
}

// GuildLog 帮会日志
type GuildLog struct {
	Id        int64    // 自增id
	GuildId   int64    // 帮会id
	Timestamp int64    // 发生时间
	Action    int32    // 事件类型
	DbId      []int64  // 玩家id
	Params    []string // 参数
}

// Member 帮会成员信息
type Member struct {
	Id                 int64  // 玩家dbId
	Account            string // account id
	Name               string // 玩家名
	FaceId             int32  // 玩家头像
	FaceSlotId         int32  // 头像框
	Level              int32  // 玩家等级
	Position           int32  // 帮会职位
	JoinTime           int64  // 入会时间
	Power              int32  // 战力
	WeeklyContribution int32  // 周贡献
	LastLoginTime      int64  // 上一次登录时间
}

func GuildToProto(guild *Guild) *proto_guild.Guild {
	if guild == nil {
		return nil
	}

	// 会长信息
	masterInfo, ok := guild.MemberData[int64(guild.Props[define.GuildPropMaster])]
	if !ok {
		log.Error("guild to proto master error,guildId:%v,masterId:%v", guild.Id, guild.Props[define.GuildPropMaster])
		return &proto_guild.Guild{}
	}

	return &proto_guild.Guild{
		Id:                int32(guild.Id),
		Name:              guild.GuildName,
		Banner:            guild.Props[define.GuildPropBanner],
		BannerColor:       guild.Props[define.GuildPropBannerColor],
		LevelLimit:        guild.Props[define.GuildPropLevelLimit],
		CurMemberNum:      guild.Props[define.GuildPropCurMemberCount],
		MaxMemberNum:      guild.Props[define.GuildPropMaxMemberCount],
		MasterId:          masterInfo.Id,
		NoticeBoard:       guild.NoticeBoard,
		ApplyNeedApproval: guild.Props[define.GuildPropApplyNeedApproval] == 1,
		IgnoreLevelLimit:  guild.Props[define.GuildPropIgnoreLevelLimit] == 1,
		Level:             guild.Props[define.GuildPropLevel],
		Exp:               guild.Props[define.GuildPropExp],
		Growth:            guild.Props[define.GuildPropGrowth],
		Reducetime:        guild.Props[define.GuildPropReducetime],
		Addsucrare:        guild.Props[define.GuildPropAddsucrare],
		Title:             guild.Props[define.GuildPropTitle],
	}
}

// GuildDB 帮会信息DB结构(落库使用)
type GuildDB struct {
	Id                int64             // 帮会唯一id
	NoticeBoard       string            // 公告栏信息
	GuildName         string            // 帮会名
	Banner            int32             // 旗帜
	BannerColor       int32             // 旗帜颜色
	LevelLimit        int32             // 帮会门槛
	Master            int32             // 会长
	IgnoreLevelLimit  int32             // 无视帮会门槛
	MaxMemberCount    int32             // 最大成员数量
	CurMemberCount    int32             // 当前成员数量
	ApplyNeedApproval int32             // 申请是否需要审批
	Level             int32             // 帮会等级
	Exp               int32             // 帮会经验
	MemberData        map[int64]*Member // 帮会成员
	Growth            int32             // 成长值
	ReduceTime        int32             // 减少时长
	AddSucRare        int32             // 增加成功率
	Yuanchi           string            // 元池
	Title             int32             // 主题
}

// 地图结构体
type GuildMapItem struct {
	Id    int32
	Level int32
	Index int32
}

// 祈福
type GuildPrayItem struct {
	IsTodayPray   bool
	TodayPrayTime int64
	PrayType      int32
	RangeType     int32
	RangeValue    int32
}
