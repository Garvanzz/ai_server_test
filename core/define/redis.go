package define

const (
	LoginToken            = "LoginToken"
	Account               = "Account"
	AccountRole           = "AccountRole"
	PlayerGuildKey        = "player_guild"
	PlayerLastLoginServer = "LastLoginServer"
)

const (
	Player             = "Player"
	PlayerBase         = "Base"
	PlayerBag          = "Bag"
	PlayerShop         = "Shop"
	PlayerHandbook     = "Handbook"
	PlayerTask         = "Task"
	PlayerHero         = "Hero"
	PlayerDraw         = "Draw"
	PlayerStage        = "Stage"
	PlayerLineUp       = "LineUp"
	PlayerEquip        = "Equip"
	PlayerWelfare      = "Welfare"
	PlayerOpenBox      = "OpenBox"
	PlayerIdleBox      = "IdleBox"
	PlayerSkill        = "Skill"
	PlayerMagic        = "Magic"
	PlayerDestiny      = "Destiny"
	PlayerShenjiDraw   = "ShenjiDraw"
	PlayerCollection   = "Collection"
	PlayerDivine       = "Divine"
	PlayerGemAppraisal = "GemAppraisal"
	PlayerPet          = "Pet"
	PlayerPetEquip     = "PetEquip"
	PlayerPetHandbook  = "PetHandBook"
	PlayerPetDraw      = "PetDraw"
	PlayerProp         = "PlayerProp"
	Danaotiangong      = "Danaotiangong"
	BRDanaotiangong    = "BRDanaotiangong"
	PlayerMission      = "Mission"
	PlayerFashion      = "Fashion"
	PlayerParadise     = "Paradise"
	PlayerCdkey        = "Cdkey"
)

// 独立的
const (
	GemAppraisal_MonthCard = "GemMonthcard" //鉴宝月卡
)

const (
	CommonRedisKey = "Common" //每日时间
)

// 聊天
const (
	ChatKuafu    = "ChatKuafu"
	ChatWorld    = "ChatWorld"
	ChatGuild    = "ChatGuild"
	ChatPrivate  = "ChatPrivate"
	ChatZudui    = "ChatZudui"
	ChatChuanwen = "ChatChuanwen"
)

// 好友
const (
	Friend           = "Friend"
	Friend_Gift      = "FriendGift"
	Friend_Recommend = "FriendRecommend"
)

// 支付
const (
	PayChache = "PayCache"
)

const (
	ActivityRedisKey       = "Activity"
	ActivityPlayerRedisKey = "ActivityPlayer"
	GuildRedisKey          = "GuildManager"
	CommonCdkey            = "CommonCdkey" // 通用兑换码全局计数
)

// 交易所
const (
	TransactionOrder   = "TransactionOrder"
	Transaction        = "Transaction"
	TransactionRecords = "TransactionRecords"
)

// 花果山
const (
	ParadisePartner       = "ParadisePartner"
	ParadisePartnerInvite = "ParadisePartnerInvite"
)

const (
	RedisRetNone = iota + 1
	RedisRetBag
	RedisRetRank
	RedisRetStage
)
