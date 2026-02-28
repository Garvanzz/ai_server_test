package model

type Danaotiangong struct {
	Frequency        int32 //周目
	Stage            int32 //当前层数
	DaychallengeNum  int32 //每天挑战次数
	DaychallengeTime int64 //时间戳
}

type BattleReportBack_Danaotiangong struct {
	Stage int32
	Data  interface{}
}

// 战斗记录
type BattleRecord_Dabaotiangong struct {
	Records []*BattleRecord_DabaotiangongOpt
}

type BattleRecord_DabaotiangongOpt struct {
	Id         int32
	Name       string
	Time       int32
	IsWin      bool
	HeroItem   map[int32]*HeroOption
	CreateTime int64
}
