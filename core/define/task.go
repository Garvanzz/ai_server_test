package define

const (
	TaskHeroLevel                  = 1  // 角色等级
	TaskClimbTower                 = 2  //爬塔
	TaskDrawCard                   = 3  // 英雄招募
	TaskKillStageMonster           = 4  //击杀关卡小怪
	TaskMainLinePassStage          = 5  //主线通关
	TaskLoginXTimes                = 6  // 登录N次
	TaskLevelUpHeroTime            = 7  // 升级神将次数
	TaskLevelUpMainHeroTime        = 8  // 升级主角次数
	TaskLevelUpMainHeroStage       = 9  // 主角突破阶数
	TaskLevelUpHeroStageTime       = 10 // 神将突破次数
	TaskLevelUpHeroStarTime        = 11 // 神将升星次数
	TaskLevelUpMainHeroXiuweiLevel = 12 // 主角提升修为等级
	TaskOpenBoxTime                = 13 //  开箱子（任意箱子）次数
	TaskDianfengJoinMatch          = 14 //  参与匹配（巅峰决斗）
	TaskDianfengJoinJingcai        = 15 //  参与竞猜（巅峰决斗）
	TaskDianfengJoinZhenying       = 16 //  选择阵营（巅峰决斗）
	TaskDianfengBattleSuc          = 17 //  胜利（巅峰决斗）
	TaskDianfengJingcaiSuc         = 18 //  竞猜成功（巅峰决斗）
	TaskOpenMagicBoxTime           = 19 //  开功法箱子次数
	TaskPassBoxMissionTime         = 20 //  通关宝箱副本次数
	TaskPassLingyuMissionTime      = 21 //  通关灵玉副本次数
	TaskGetGuajiAwardTime          = 22 //  领取挂机奖励次数
	TaskJingjichangChallengeTime   = 23 //  竞技场挑战次数
	TaskDanaotiangongChallengeTime = 24 //  大闹天宫挑战次数
	TaskParadiseTreeWaterTime    = 25 //  乐园种树浇水次数
	TaskMainHeroLevelUpXiuweiTime  = 26 //  主角升级修为次数
)

const (
	TaskTypeDaily              = 1  // 日常任务
	TaskTypeWeek               = 2  // 周任务
	TaskTypeMonth              = 3  // 月任务
	TaskTypeAchieve            = 4  // 成就任务
	TaskTypeMain               = 5  // 主线任务
	TaskTypeGuild              = 6  // 帮派任务
	TaskTypeDrawHeroRank       = 7  // 招募排行任务
	TaskTypeTheCompetitionRank = 8  // 巅峰决斗任务
	TaskTypePassportDaily      = 9  // 通行证每日任务
	TaskTypePassportWeek       = 10 // 通行证每周任务
	TaskTypePassportSeason     = 11 // 通行证赛季任务
)

const (
	TaskRefreshTypeNull     = 0 //无
	TaskRefreshTypeDay      = 1 //每天
	TaskRefreshTypeWeek     = 2 //每周
	TaskRefreshTypeYongJiu  = 3 //永久
	TaskRefreshTypeMonth    = 4 //每月
	TaskRefreshTypeSeason   = 5 //赛季
	TaskRefreshTypeActivity = 6 //活动
)

const (
	TaskActivityTypeDaily = 1 // 日常任务
	TaskActivityTypeGuild = 2 // 帮会任务
)

// TaskCompleteLimit 每日任务完成次数限制
var TaskCompleteLimit = map[int32]int32{
	TaskLoginXTimes: 1,
}
