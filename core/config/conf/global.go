package conf

type Global struct {
	CreateGuildConsume        []ItemE // 创建帮会消耗
	GuildImpeachOfflineTime   int64   // 弹劾会长时间
	GuildImpeachNeedActivity  int32   // 弹劾所需活跃度
	UseItemMaxCount           int32   // 物品单次使用数量上限
	TransformJob              int32   //切换职业的花费
	MoneyBoxRange             int32   //金币宝箱浮动百分比，基于基础值来
	IdleBoxTime               int32   // 挂机宝箱累计间隔时间
	IdleBoxMaxTime            int32   // 挂机宝箱累计最大时长(小时)
	HeroUpLevelRange          int32   //神将升级区间
	MainHeroLevelLimit        int32   //主角等级限制
	MountChangeName           []ItemE //坐骑改名消耗
	EnchantLimit              string  //附魔限制
	FriendGift                []ItemE //好友赠礼
	MaxFriendNum              int32   //最大好友数
	PetDrawCardCostNum        int32   //神宠召唤消耗的召唤令数量
	PetDrawCardDikouNum       int32   //神宠召唤神宠抵扣的数量
	PetDrawCardCostLingPetNum int32   //神宠召唤消耗的灵宠碎片数量
	PetDrawCardCostShenPetNum int32   //神宠召唤消耗的神宠碎片数量
	PetXilianCost             []ItemE //宠物洗礼消耗
	PetSkillCost              []ItemE //宠物定向技能消耗
	PetSkillRemoveCost        []ItemE //宠物技能移除消耗
	GuildRename               []ItemE //帮会改名消耗
	PlayerRename              []ItemE //角色改名消耗
	SupplementSign            []ItemE //补签消耗
	GofishWeightNormal        []int32 //普通鱼饵的品质概率
	GofishWeightAdvance       []int32 //高级鱼饵的品质概率
	GofishBasicRate           int32   //钓鱼基础概率
	ArenaFuchouCost           []ItemE //竞技场复仇消耗
	PaternerCoolDown          int32   //重新邀请冷却时间：天
	PaternerInviteCost        []ItemE //邀请作为伴侣的消耗
	PaternerInviteEffectTime  int32   //伴侣邀请有效时间
	PaternerGiveCount         int32   //伴侣赠送次数
	GiftRaceAddIntimacy       []int32 //赠送品质获得亲密度
}
