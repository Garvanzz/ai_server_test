package define

const (
	GuildPropBanner            = 1  // 旗帜
	GuildPropBannerColor       = 2  // 旗帜颜色
	GuildPropLevelLimit        = 3  // 进入帮会等级限制
	GuildPropMaster            = 4  // 会长
	GuildPropIgnoreLevelLimit  = 5  // 无视帮会门槛
	GuildPropMaxMemberCount    = 6  // 最大成员数量
	GuildPropCurMemberCount    = 7  // 当前成员数量
	GuildPropApplyNeedApproval = 8  // 申请是否需要审批
	GuildPropLevel             = 9  // 帮会等级
	GuildPropExp               = 10 // 帮会经验
	GuildPropGrowth            = 11 // 帮会成长值
	GuildPropReducetime        = 12 // 帮会减少时长
	GuildPropAddsucrare        = 13 // 帮会增加成功率
	GuildPropTitle             = 14 //主题
	GuildPropMax               = 20 // prop最大长度
)

const (
	PageMax       = 10        // 帮会列表每页最大数量
	LogCountMax   = 30        // 帮会日志最大条数
	ApplyKeepTime = 3600 * 72 // 帮会申请默认保留时间
	JoinGuildCD   = 3600      // 重新加入帮会冷却时间
	ApplyCountMax = 30        // 帮会申请最大数量
	SaveMaxCount  = 1         // 每次tick 最大帮会落库数量
)

const (
	TableGuild      = "guild"
	TableGuildApply = "guild_apply"
	TableGuildLog   = "guild_log"
)

// 帮会职位
const (
	GuildOrdinary   = 1 // 成员
	GuildElder      = 2 // 长老
	GuildViceMaster = 3 // 副帮主
	GuildMaster     = 4 // 帮主
)

// GuildPermission 行为对应需要权限
var GuildPermission = map[int]int32{
	ActionKickOut:         GuildMaster,
	ActionDealApply:       GuildMaster,
	ActionSetGuildRule:    GuildMaster,
	ActionChangeGuildName: GuildMaster,
}

// 行为
const (
	ActionKickOut = iota + 1
	ActionDealApply
	ActionSetGuildRule
	ActionChangeGuildName
)

// 帮会日志类型
const (
	GuildEventCreate           = iota + 1 // 创建帮会
	GuildEventJoin                        // 加入帮会
	GuildEventAssignMaster                // 任命帮主
	GuildEventAssignViceMaster            // 任命副帮主
	GuildEventAssignElder                 // 任命长老
	GuildEventKickOut                     // 踢出帮会
	GuildEventQuit                        // 离开帮会
	GuildEventImpeach                     // 弹劾帮主
)
