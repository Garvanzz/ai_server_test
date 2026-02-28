package model

type PlayerTask struct {
	Status                 map[int32]map[int32]int32 // 任务信息map
	TaskLimit              map[int32]int32           // 每日完成任务限制
	ActivePointRecord      map[int32]bool            // 每日活跃点领奖记录
	ActivePointWeekRecord  map[int32]bool            // 每周活跃点领奖记录
	ActivePointGuildRecord map[int32]bool            // 帮派活跃点领奖记录
	DailyTask              map[int32]*Task
	WeekTask               map[int32]*Task
	MonthTask              map[int32]*Task
	MainTask               map[int32]*Task //主线
	AchieveTask            map[int32]*Task //成就
	GuildTask              map[int32]*Task //帮会
	DrawHeroTask           map[int32]*Task //招募英雄
	TheCompetitionTask     map[int32]*Task //巅峰决斗
	PassportDailyTask      map[int32]*Task //通行证每日任务
	PassportWeekTask       map[int32]*Task //通行证每周任务
	PassportSeasonTask     map[int32]*Task //通行证赛季任务
	DailyResetTime         int64
	WeekResetTime          int64
	MonthResetTime         int64
	GuildResetTime         int64
	DailyPoint             int32 //每日活跃点
	GuildPoint             int32 //帮派活跃点
}

type Task struct {
	Id             int32
	InitialProcess int32
	TaskType       int32
	Condition      int32
	ExtraCondition int32
	ReceiveAward   bool
}
