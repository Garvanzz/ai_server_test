package define

// 背包道具类型
const (
	BagItemTypeNormal          = 1 //普通道具
	BagItemTypeBox             = 2 //宝箱
	BagItemTypeHeroPiece       = 3 //英雄碎片
	BagItemTypeCollectionPiece = 4 //藏品碎片
	BagItemTypeDropBox         = 5 //掉落宝箱
)

// 物品类型
const (
	ItemTypeItem          = 1  //道具
	ItemTypeHero          = 2  //英雄
	ItemTypeSkin          = 3  //皮肤
	ItemTypeEquip         = 4  //装备
	ItemTypeMagic         = 5  //法术
	ItemTypeWeaponry      = 6  //神兵
	ItemTypeMount         = 7  //坐骑
	ItemTypeCollect       = 8  //藏品
	ItemTypeBraces        = 9  //背饰
	ItemTypeLearning      = 10 //心得
	ItemTypePet           = 11 //宠物
	ItemTypePetEquip      = 12 //宠物装备
	ItemTypePetSkill      = 13 //宠物技能
	ItemTypeHeadFrame     = 14 //头像框
	ItemTypeTitle         = 15 //称号
	ItemTypeBubble        = 16 //泡泡
	ItemGuildMaterial     = 17 //帮会材料
	ItemGuildElement      = 18 //帮会元素
	ItemTypeFish          = 19 //钓鱼-鱼类
	ItemTypeFashion       = 20 //时装
	ItemTypeHeadWear      = 21 //头饰
	ItemTypePassportScore = 22 //通行证积分
)

// 道具ID
const (
	ItemIdMoney         = 104 //金币
	ItemIdTupoStore     = 105 //突破石
	ItemIdCultivation   = 101 //修为
	ItemIdDrawCard      = 110 //抽奖券
	ItemIdBoxMoney      = 106 //金币宝箱
	ItemIdBoxStore      = 107 //突破石宝箱
	ItemIdBoxEquip      = 108 //装备宝箱
	ItemIdBoxMagick     = 109 //法术宝箱
	ItemIdBoxLingyu     = 102 //灵玉
	ItemIdTallyAtk      = 113 //攻击符
	ItemIdTallyDef      = 114 //防御符
	ItemIdTallyHp       = 115 //气血符
	ItemIdTallyForce    = 116 //内力符
	ItemIdBraceAura     = 120 //灵光石
	ItemIdShenji        = 123 //神机宝匣
	ItemIdLiLian        = 128 //历练卷轴
	ItemIdGemAppraisal  = 129 //鉴宝镜
	ItemIdXiantao       = 130 //仙桃
	ItemIdYueshi        = 131 //玥钥
	ItemIdShenshi       = 132 //神钥
	ItemIdPetZhaohuan   = 133 //神宠召唤令
	ItemIdPetResetStore = 136 //神宠重置石
	ItemIdFishNormal    = 148 //普通鱼饵
	ItemIdFishAdvance   = 149 //高级鱼饵
	ItemIdXianyu        = 103 //仙玉
)

// 商品id
const (
	ShopIdGemAppraisal = 90005 //鉴宝快捷兑换
)

// 掉落类型
const (
	DropTypeNormal            = 1 // 固定奖励
	DropTypeWeight            = 2 // 权重随机
	DropTypeIndependentRandom = 3 // 概率随机
	DropTypePseudoRandom      = 4 // 伪随机(保底)
)

// 开箱子类型
const (
	OPENBOX_MONEY = 1 //金币
	OPENBOX_STORE = 2 //突破石
	OPENBOX_EQUIP = 3 //装备
	OPENBOX_MAGIC = 4 //法术
)

const (
	LevelMaxLimit    = 6500 //等级限制
	LevelPetMaxLimit = 5000 //宠物等级限制
)

// 卡池类型
const (
	CARDPOOL_HERO         = 1 //角色卡池
	CARDPOOL_SHENJI       = 2 //神机卡池
	CARDPOOL_GEMAPPRAISAL = 3 //鉴宝
	CARDPOOL_PET          = 4 //宠物
	CARDPOOL_PETGIFT      = 5 //宠物天赋
)

// 布阵
const (
	LINEUP_STAGE          = 1 //布阵-关卡
	LINEUP_DANAOTIANGONG  = 2 //布阵-大闹天宫
	LINEUP_CLIMBTOWER     = 3 //爬塔
	LINEUP_TheCompetition = 4 //巅峰对决
	LINEUP_ARENA          = 5 //竞技场
	LINEUP_Tianti         = 6 //天梯
)

const (
	BattleScene_None          = 0
	BattleScene_Danaotiangong = 1 //大闹天宫
	BattleScene_Mission       = 2 //副本
	BattleScene_StageBoss     = 3 //主关卡Boss
	BattleScene_Player        = 4 //玩家
	BattleScene_Arena         = 5 //竞技场
	BattleScene_Tianti        = 6 //天梯
)

const (
	AttributeRate = 1000.0 //属性倍率
)

const (
	HorseType_System        = 0  //系统
	HorseType_DrawCard      = 1  //抽卡
	HorseType_Activity      = 2  //活动
	HorseType_Pet           = 5  //宠物
	HorseType_ShopBuy       = 6  //商城购买
	HorseType_CharperChange = 7  //章节变化
	HorseType_RankUpdate    = 10 //排名更新
)

const (
	HorseType_Condition_PlayerName   = 1  //玩家名字
	HorseType_Condition_Rate         = 2  //品质
	HorseType_Condition_ItemId       = 3  //道具ID
	HorseType_Condition_HeroId       = 4  //角色ID
	HorseType_Condition_EquipId      = 5  //装备ID
	HorseType_Condition_CardPoolType = 6  //卡池类型
	HorseType_Condition_ActType      = 7  //活动类型
	HorseType_Condition_MountId      = 8  //坐骑ID
	HorseType_Condition_PetId        = 9  //宠物ID
	HorseType_Condition_TitleId      = 10 //称号ID
	HorseType_Condition_ShopId       = 12 //商城ID
	HorseType_Condition_CharperId    = 13 //章节ID
	HorseType_Condition_RankIndex    = 14 //排名
	HorseType_Condition_RankType     = 15 //排行类型
	HorseType_Condition_CollectId    = 16 //藏品ID
)

const (
	ChatChuanwenType_System        = 0 //系统
	ChatChuanwenType_AddMagic      = 1 //获得功法
	ChatChuanwenType_CharperChange = 3 //章节变化
	ChatChuanwenType_DrawCard      = 5 //抽卡
	ChatChuanwenType_DrawCardPet   = 7 //宠物
)

const (
	ChatChuanwenType_Condition_PlayerName   = 1  //玩家名字
	ChatChuanwenType_Condition_Rate         = 2  //品质
	ChatChuanwenType_Condition_CardPool     = 3  //卡池类型
	ChatChuanwenType_Condition_HeroId       = 4  //角色ID
	ChatChuanwenType_Condition_EquipId      = 5  //装备ID
	ChatChuanwenType_Condition_ItemId       = 7  //道具ID
	ChatChuanwenType_Condition_CharperIndex = 6  //章节数
	ChatChuanwenType_Condition_MountId      = 8  //坐骑ID
	ChatChuanwenType_Condition_PetId        = 9  //宠物ID
	ChatChuanwenType_Condition_Rank         = 10 //排行榜类型
	ChatChuanwenType_Condition_RankIndex    = 11 //排名
	ChatChuanwenType_Condition_MagicId      = 12 //功法ID
)

const (
	HandbookAwardType_Hero     = 1
	HandbookAwardType_Fashion  = 2
	HandbookAwardType_HeadWear = 3
	HandbookAwardType_Mount    = 4
	HandbookAwardType_Brace    = 5
	HandbookAwardType_Weapon   = 6
)

// 机器人模块
const (
	RobotMode_Arena  = 1
	RobotMode_Tianti = 2
)

// 花果山-伴侣系统
const (
	PartnerInviteStatusPending  = 1 // 邀请状态:待处理
	PartnerInviteStatusAccepted = 2 // 邀请状态:已同意
	PartnerInviteStatusRejected = 3 // 邀请状态:已拒绝

	// 奖励类型常量(用于区分不同类型的解锁奖励)
	PartnerRewardTypeBuff      = 1 // BUFF
	PartnerRewardTypeHeadFrame = 2 // 头像框
	PartnerRewardTypeHeadWear  = 3 // 头饰
	PartnerRewardTypeBrace     = 4 // 背饰
	PartnerRewardTypeMount     = 5 // 坐骑
	PartnerRewardTypeSkill     = 6 // 技能
)
