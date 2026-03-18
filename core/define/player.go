package define

const (
	PlayerPropLevel         = 1  // 等级
	PlayerPropExp           = 2  // 经验
	PlayerPropFaceId        = 5  // 头像id
	PlayerPropFaceSlotId    = 6  // 头像框id
	PlayerPropOfflineTime   = 7  // 玩家上次登录时间
	PlayerPropRank          = 8  //玩家段位
	PlayerPropTitle         = 9  //称号
	PlayerPropJob           = 10 //职业
	PlayerPropSex           = 11 //性别
	PlayerPropClan          = 12 //帮会
	PlayerPropClanId        = 13 // 帮会ID
	PlayerPropHeroId        = 14 // 主角ID
	PlayerPropBubbleId      = 15 // 泡泡ID
	PlayerPropPower         = 16 // 战力
	PlayerPropServerId      = 17 //服务器ID
	PlayerPropEntryServerId = 18 // 入口服ID
	PlayerPropMax           = 24 // 最大Limit
)

// 玩家职业
const (
	PlayerJobYao  = 1
	PlayerJobShen = 2
	PlayerJobFo   = 3
)

// 主角ID
const (
	PlayerMainHeroNull = 3001 //新手
	PlayerMainHeroYao  = 3002 //妖
	PlayerMainHeroShen = 3003 //神
	PlayerMainHeroFo   = 3004 //佛
)

// 默认头像
const (
	PlayerHeadIcon  = 3003 //默认杨过的头像
	PlayerHeadFrame = 2001 //默认头像框
	PlayerHeroID    = 3001 //默认角色
	PlayerTitleID   = 4001 //默认称号
	PlayerBubbleID  = 5001 //默认气泡
)

// 段位
const (
	PlayerRankNull            = 0 //无
	PlayerRankBaiyin          = 1
	PlayerRankHuangjin        = 2
	PlayerRankBojin           = 3
	PlayerRankZuanshi         = 4
	PlayerRankDashi           = 5
	PlayerRankWangzhe         = 6
	PlayerRankZuiQiangWangzhe = 7
)
