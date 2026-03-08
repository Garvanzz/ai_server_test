package model

import "xfx/core/config/conf"

// ActivityInfo GM/Invoke 用：活动实例概要（含状态），放在 core/model 避免 invoke 依赖 activity 包导致循环引用
type ActivityInfo struct {
	ActId     int64  `json:"act_id"`
	CfgId     int64  `json:"cfg_id"`
	Type      string `json:"type"`
	State     string `json:"state"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	CloseTime int64  `json:"close_time"`
	TimeType  int32  `json:"time_type"`
	Season    int32  `json:"season"`
}

type ActivityData struct {
	Id        int64
	CfgId     int64
	Type      string
	State     string
	StartTime int64
	EndTime   int64
	TimeType  int32
	CloseTime int64
	TimeValue int32
	Data      any
}

type ActDataDailyAccumulateRecharge struct {
	ids []int32
}

// ActDataMonthCard 月卡
type ActDataMonthCard struct {
	Day      int32 //月卡剩余天数
	LastTime int64 //这个服务器使用
	Count    int32 //购买次数
}

// ActDataTheCompetition 巅峰对决
type ActDataTheCompetition struct {
	Score      map[int32]int64 //双方总积分
	StakeGroup map[int32]int64 //双方押注情况
}

// ActDataLadderRace 天梯
type ActDataLadderRace struct {
	Season     int32 //赛季
	SeasonTime int64 //这一赛季结束时间
	RankPlayer map[int64]*ActDataLadderRaceRankPlayer
}

// ActDataLadderRaceRankPlayer 天梯排位玩家
type ActDataLadderRaceRankPlayer struct {
	Score      int64
	Rank       int32
	LittleRank int32
}

// ActDataGoFish 钓鱼
type ActDataGoFish struct {
	Pool                 map[int32]map[int32]int32 // 鱼池数量
	PoolRefreshTime      int64                     // 鱼池刷新时间
	StartTime            int32                     //每天赛事开启的时间
	EndTime              int32                     //每天赛事关闭的时间
	PoolRefreshOffseTime int32                     //池子刷新时间
	FireRankAward        bool                      //派发奖励
}

// ActDataPassport 通行证
type ActDataPassport struct {
	Season int32 // 当前赛季
}

// ===========================玩家个人数据===================================

type DailyAccumulateRechargePd struct {
	Money   int32
	GetList []int32
}

// MonthCardPd 月卡
type MonthCardPd struct {
	Day      int32 // 月卡剩余天数
	Count    int32 // 购买次数
	LastTime int64 // 这个服务器使用
}

// TheCompetitionPd 巅峰对决
type TheCompetitionPd struct {
	IsChoose     bool
	ChooseId     int32
	IsStake      bool  // 是否押注
	StakeCount   int32 // 押注数量
	StageGroupId int32 // 押注方ID
	Score        int32 // 积分
}

// FundOptionPd 基金
type FundOptionPd struct {
	Type       int32
	NormalIds  []int32
	AdvanceIds []int32
	IsBuy      bool
}

// ArenaOptionPd 竞技场
type ArenaOptionPd struct {
	ChallengeTime     int32            // 挑战次数
	PlayerIds         []int64          // 展示的6位
	RefreshTime       int32            // 刷新次数
	RefreshCD         int64            // 刷新冷却
	LineUp            []ArenaLineUpIds // 布阵
	LastRefreshTime   int64            //上一次刷新时间
	LastChallengeTime int64            //上一次挑战刷新时间
}

// ArenaLineUpIds 竞技场布阵
type ArenaLineUpIds struct {
	Index int32
	Id    []int32
}

// 竞技场记录
type BattleReportRecord_Arena struct {
	TargetId int64
	IsAttack bool
	Rank     int32
	Time     int64
	ActId    int64
	IsFuchou bool
}

// LadderRacePd 天梯
type LadderRacePd struct {
	Score             int32           // 当前积分
	LineUp            []LadderRaceIds // 布阵
	ChallengeTime     int32           // 挑战次数
	LastChallengeTime int64           //上一次挑战刷新时间
}

// LadderRaceIds 天梯布阵
type LadderRaceIds struct {
	Index int32
	Id    []int32
}

// 天梯记录
type BattleReportRecord_LadderRace struct {
	TargetId int64
	IsAttack bool
	Rank     int32
	Time     int64
	ActId    int64
}

// GoFishPd 钓鱼
type GoFishPd struct {
	Fish         map[int32]int32 // 拥有的鱼
	SignDay      int32           // 签到天数
	LastSignTime int64           // 上一次签到时间
	Exp          int32           // 经验
	GetList      []int32         // 领取等级奖励记录
}

// PassportPd 通行证
type PassportPd struct {
	Score      int32   // 当前积分
	Level      int32   // 当前等级
	NormalIds  []int32 // 已领取的普通奖励ID列表
	AdvanceIds []int32 // 已领取的高级奖励ID列表
	IsBuy      bool    // 是否购买高级通行证
}

// ------------其他数据----------------
type GoFishBack struct {
	Code int     //1没有鱼了 2 空军 3 成功
	Ids  []int32 //鱼的id
}

// 通用活动奖励回调
type CommonActivityAwardBack struct {
	Award []conf.ItemE
	Code  int32 //0成功 1已经完成 2不满条件
}
