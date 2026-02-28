package model

// 鉴宝
type GemAppraisal struct {
	PoolId       int
	PoolStarTime int64
	Num          int32
	Reward       []int32
	PoolEndTime  int64
	Title        int
}

// 鉴宝月卡
type GemAppraisalMonthCard struct {
	GetTime   int64  //领取时间
	GetDay    int    //领取天数
	EffectDay int    //生效天数
	PID       string //玩家uid
	DbId      int64  // 玩家id
}
