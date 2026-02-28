package define

// 活动类型活动类型和配置表名一致
const (
	ActivityTypeDailyAccRecharge = "ActDailyAccumulateRecharge" // 每日累充
	ActivityTypeDrawHeroRank     = "ActDrawRank"                // 招募角色排行榜
	ActivityTypeRechargeRank     = "ActRechargeRank"            // 充值排行榜
	ActivityTypeNormalMonthCard  = "ActMonthCard"               // 常规月卡
	ActivityTypeTheCompetition   = "ActTheCompetition"          // 巅峰决斗
	ActivityTypeMainLineFund     = "ActMainLineFund"            // 主线基金
	ActivityTypeLevelFund        = "ActLevelFund"               // 成长基金
	ActivityTypeBoxFund          = "ActBoxFund"                 // 宝箱基金
	ActivityTypeArena            = "ActArena"                   // 竞技场
	ActivityTypeLadderRace       = "ActLadderRace"              // 天梯
	ActivityTypeGoFish           = "ActGoFish"                  // 钓鱼
	ActivityTypePassport         = "ActPassport"                // 通行证
	ActivityTypeSeason           = "ActSeason"                  // 赛季
)

// 配置表时间类型
const (
	ActTimeClose            = 0 // 活动关闭
	ActTimeAlwaysOpen       = 1 // 活动常驻
	ActTimeConfigured       = 2 // 配置时间
	ActTimeServerConfigured = 3 // 按照服务器开启时间
	ActTimeSeason           = 4 //赛季类型
)

const (
	ActivityPlayerDataBase = 10000000000
)

const (
	ActivityFund_MainLine = 1 //主线基金
	ActivityFund_Level    = 2 //成长基金
	ActivityFund_Box      = 3 //宝箱基金
)
