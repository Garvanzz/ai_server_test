package define

const (
	MonthCard_Week         = 1 //周卡
	MonthCard_Month        = 2 //月卡
	MonthCard_GemAppraisal = 3 //鉴宝月卡
)

const (
	SignType_Normal    = 1 //常规签到
	SignType_BuSignIn  = 2 //补签到
	SignType_AccSignIn = 3 //累计签到
)

const (
	FuncOpenCondition_None          = 0
	FuncOpenCondition_MainHeroLevel = 1 //主角等级
	FuncOpenCondition_MainHeroStage = 2 //主角阶数
	FuncOpenCondition_MainHeroStar  = 3 //主角星数
	FuncOpenCondition_Stage         = 4 //关卡
	FuncOpenCondition_Forward       = 5 //前置
)
