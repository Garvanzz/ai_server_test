package conf

// Activity 活动配置 "Year-Month-Day Hour:Minute:Second"
type Activity struct {
	Id           int64  `json:"Id"`
	Type         string `json:"Type"`
	StartTime    string `json:"StartTime"`
	EndTime      string `json:"EndTime"`
	CloseTime    string `json:"CloseTime"`
	LastTime     int32  `json:"LastTime"`
	ActTime      int32  `json:"ActTime"` // 0 一直开启\1 读取配置表时间\2 关闭活动
	DurationTime int32  `json:"DurationTime"`
	CanClear     int32  `json:"CanClear"`
	RefreshType  int32  `json:"RefreshType"`
	Param1       int32  `json:"Param1"`
	Param2       int32  `json:"Param2"`
	Param3       int32  `json:"Param3"`
}

type ActivityBase struct {
	Id       int64 `json:"id"`
	Activity int64 `json:"activity"`
}

func (b ActivityBase) GetActivityId() int64 {
	return b.Activity
}

// 每日累充
type ActDailyAccumulateRecharge struct {
	ActivityBase
	Progress int32   `json:"Progress"`
	Award    []ItemE `json:"Award"`
}

// 巅峰对决
type ActTheCompetition struct {
	ActivityBase
	GroupIds         []int32 `json:"GroupIds"`
	Time             int32   `json:"Time"`
	SuccessGiveAward []ItemE `json:"SuccessGiveAward"`
	FaildGiveAward   []ItemE `json:"FaildGiveAward"`
	Stake            []int32 `json:"Stake"`
	StakeOpenTime    string  `json:"StakeOpenTime"`
	StakeRate        []int32 `json:"StakeRate"`
}

// 主线基金
type ActMainLineFund struct {
	ActivityBase
	Stage        int32   `json:"Stage"`
	NormalAward  []ItemE `json:"NormalAward"`
	AdvanceAward []ItemE `json:"AdvanceAward"`
}

// 成长基金
type ActLevelFund struct {
	ActivityBase
	Level        int32   `json:"Level"`
	NormalAward  []ItemE `json:"NormalAward"`
	AdvanceAward []ItemE `json:"AdvanceAward"`
}

// 宝箱基金
type ActBoxFund struct {
	ActivityBase
	Level        int32   `json:"Level"`
	NormalAward  []ItemE `json:"NormalAward"`
	AdvanceAward []ItemE `json:"AdvanceAward"`
}

// 竞技场
type ActArena struct {
	ActivityBase
	Time          int32   `json:"Time"`
	ChallengeTime int32   `json:"ChallengeTime"`
	RefreshTime   int32   `json:"RefreshTime"`
	RefreshCD     int32   `json:"RefreshCD"`
	RobotPower    []int32 `json:"RobotPower"`
	PowerCount    int32   `json:"PowerCount"`
	Awards        []ItemE `json:"Awards"`
	DropAwardId   int32   `json:"DropAwardId"`
}

// 天梯
type ActLadderRace struct {
	ActivityBase
	Season        int32   `json:"Season"`
	SeasonTime    int32   `json:"SeasonTime"`
	RobotPower    []int32 `json:"RobotPower"`
	PowerCount    int32   `json:"PowerCount"`
	BasicScore    int32   `json:"BasicScore"`
	Awards        []ItemE `json:"Awards"`
	DropAwardId   int32   `json:"DropAwardId"`
	ChallengeTime int32   `json:"ChallengeTime"`
}

type ActLadderRaceScore struct {
	ActivityBase
	LittleRank  int32   `json:"LittleRank"`
	Rank        int32   `json:"Rank"`
	Score       int32   `json:"Score"`
	SettleScore []int32 `json:"SettleScore"`
}

// 钓鱼
type ActGoFish struct {
	ActivityBase
	Type         int32           `json:"Type"`
	Fish         map[int32]int32 `json:"Fish"`
	Exp          int32           `json:"Exp"`
	StartTime    int32           `json:"Time"`
	RefreshTime  int32           `json:"RefreshTime"`
	EndTime      int32           `json:"EndTime"`
	PointAddRare int32           `json:"PointAddRare"`
	NeedMinCost  int32           `json:"NeedMinCost"`
	DoubleRate   int32           `json:"DoubleRate"`
}

type Fish struct {
	Id   int64 `json:"Id"`
	Type int32 `json:"Type"`
	Rate int32 `json:"Rate"`
	Exp  int32 `json:"Exp"`
}

type FishSign struct {
	Id        int64   `json:"Id"`
	Day       int32   `json:"Day"`
	Reward    []ItemE `json:"Reward"`
	AccSign   bool    `json:"AccSign"`
	CumReward []ItemE `json:"CumReward"`
}

type FishLevelAward struct {
	Id     int64   `json:"Id"`
	Level  int32   `json:"Level"`
	Exp    int32   `json:"Exp"`
	Reward []ItemE `json:"Reward"`
}

// 通行证
type ActPassport struct {
	Id           int64   `json:"Id"`
	Level        int32   `json:"Level"`
	Score        int32   `json:"Score"`
	Season       int32   `json:"Season"`
	NormalAward  []ItemE `json:"NormalAward"`
	AdvanceAward []ItemE `json:"AdvanceAward"`
}
