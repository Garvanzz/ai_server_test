package config

import "xfx/core/config/conf"

// ---------------------------------------------------------------------------
// 所有配置表声明：每行 = 一张配置表。
// 新增配置只需：1) conf/ 写 struct  2) 这里加一行。
// ---------------------------------------------------------------------------

//// === 英雄 ===
//var (
//	Hero            = NewTable[conf.Hero]("Hero")
//	HeroBasicAttr   = NewTable[conf.HeroBasicAttribute]("HeroBasicAttribute")
//	HeroUpStar      = NewTable[conf.HeroUpStar]("HeroUpStar")
//	HeroUpStage     = NewTable[conf.HeroUpStage]("HeroUpStage")
//	HeroUpLevel     = NewTable[conf.HeroUpLevel]("HeroUpLevel")
//	HeroCultivation = NewTable[conf.HeroCultivation]("HeroCultivation")
//	HeroMagic       = NewTable[conf.HeroMagic]("HeroMagic")
//	HeroMagicLevel  = NewTable[conf.HeroMagicLevel]("HeroMagicLevel")
//	HeroPool        = NewTable[conf.HeroPool]("HeroPool")
//)
//
//// === 物品 / 商店 / 充值 ===
//var (
//	Item     = NewTable[conf.Item]("Item")
//	Shop     = NewTable[conf.Shop]("Shop")
//	Recharge = NewTable[conf.Recharge]("Recharge")
//)
//
//// === 任务 ===
//var (
//	Task         = NewTable[conf.Task]("Task")
//	TaskActivity = NewTable[conf.TaskActivity]("TaskActivity")
//	Mission      = NewTable[conf.Mission]("Mission")
//)

// === 全局配置 (单对象) ===
var Global = NewSingle[conf.Global]("Global")

//// === 战斗 / 关卡 ===
//var (
//	Skill             = NewTable[conf.Skill]("Skill")
//	Buff              = NewTable[conf.Buff]("Buff")
//	Monster           = NewTable[conf.Monster]("Monster")
//	StageMonsterGroup = NewTable[conf.StageMonsterGroup]("StageMonsterGroup")
//	MonsterGroup      = NewTable[conf.MonsterGroup]("MonsterGroup")
//	Stage             = NewTable[conf.Stage]("Stage")
//	Chapter           = NewTable[conf.Chapter]("Chapter")
//	ClimbTower        = NewTable[conf.ClimbTower]("ClimbTower")
//	Robot             = NewTable[conf.Robot]("Robot")
//	RobotGroup        = NewTable[conf.RobotGroup]("RobotGroup")
//)
//
//// === 属性 / 装备 ===
//var (
//	AttributeId = NewTable[conf.AttributeId]("AttributeId")
//	Equip       = NewTable[conf.Equip]("Equip")
//	EquipSell   = NewTable[conf.EquipSell]("EquipSell")
//)
//
//// === 掉落 ===
//var (
//	Drop         = NewTable[conf.Drop]("Drop")
//	BoxLevelDrop = NewTable[conf.BoxLevelDrop]("BoxLevelDrop")
//)
//
//// === 玩家 ===
//var (
//	PlayerLevel = NewTable[conf.PlayerLevel]("PlayerLevel")
//	DaySign     = NewTable[conf.DaySign]("DaySign")
//)
//
//// === 图鉴 ===
//var (
//	Handbook      = NewTable[conf.HandBook]("Handbook")
//	HandbookAward = NewTable[conf.HandBookAward]("HandbookAward")
//)
//
//// === 抽卡 ===
//var (
//	DrawPool        = NewTable[conf.DrawPool]("DrawPool")
//	DrawStageAward  = NewTable[conf.DrawStageAward]("DrawStageAward")
//	ShenjiPool      = NewTable[conf.ShenjiPool]("ShenjiPool")
//	TreasurePool    = NewTable[conf.TreasurePool]("TreasurePool")
//	RecruitLvAward  = NewTable[conf.RecruitLevelAward]("RecruitLevelAward")
//)
//
//// === 挂机 ===
//var IdleBox = NewTable[conf.IdleBox]("IdleBox")
//
//// === 坐骑 ===
//var (
//	Mount              = NewTable[conf.Mount]("Mount")
//	MountStage         = NewTable[conf.MountStage]("MountStage")
//	MountLevel         = NewTable[conf.MountLevel]("MountLevel")
//	MountEnergy        = NewTable[conf.MountEnergy]("MountEnergy")
//	MountEnergyAttr    = NewTable[conf.MountEnergyAttribute]("MountEnergyAttribute")
//)
//
//// === 神兵 ===
//var (
//	Weaponry      = NewTable[conf.Weaponry]("Weaponry")
//	WeaponryStar  = NewTable[conf.WeaponryStar]("WeaponryStar")
//	WeaponryLevel = NewTable[conf.WeaponryLevel]("WeaponryLevel")
//)
//
//// === 附魔 / 精炼 ===
//var (
//	Enchant      = NewTable[conf.Enchant]("Enchant")
//	EnchantStage = NewTable[conf.EnchantStage]("EnchantStage")
//	Succinct     = NewTable[conf.Succinct]("Succinct")
//	SuccinctSkill = NewTable[conf.SuccinctSkill]("SuccinctSkill")
//)
//
//// === 护符 ===
//var (
//	BraceAura         = NewTable[conf.BraceAura]("BraceAura")
//	BraceAuraStage    = NewTable[conf.BraceAuraStage]("BraceAuraStage")
//	Braces            = NewTable[conf.Braces]("Braces")
//	BraceTalent       = NewTable[conf.BraceTalent]("BraceTalent")
//	BraceTalentLevel  = NewTable[conf.BraceTalentLevel]("BraceTalentLevel")
//	BracesLevel       = NewTable[conf.BracesLevel]("BracesLevel")
//)
//
//// === 命运 ===
//var (
//	DestinyStage = NewTable[conf.DestinyStage]("DestinyStage")
//	DestinyLevel = NewTable[conf.DestinyLevel]("DestinyLevel")
//)
//
//// === 图鉴收集 ===
//var (
//	Collection       = NewTable[conf.Collection]("Collection")
//	CollectionUpStar = NewTable[conf.CollectionUpStar]("CollectionUpStar")
//)
//
//// === 占卜 / 学问 ===
//var (
//	Divine          = NewTable[conf.Divine]("Divine")
//	Learning        = NewTable[conf.Learning]("Learning")
//	LearningCompose = NewTable[conf.LearningCompose]("LearningCompose")
//)
//
//// === 活动 ===
//var (
//	Activity                = NewTable[conf.Activity]("Activity")
//	ActDailyAccRecharge     = NewTable[conf.ActDailyAccumulateRecharge]("ActDailyAccumulateRecharge")
//	ActTheCompetition       = NewTable[conf.ActTheCompetition]("ActTheCompetition")
//	ActBoxFund              = NewTable[conf.ActBoxFund]("ActBoxFund")
//	ActMainLineFund         = NewTable[conf.ActMainLineFund]("ActMainLineFund")
//	ActLevelFund            = NewTable[conf.ActLevelFund]("ActLevelFund")
//	ActArena                = NewTable[conf.ActArena]("ActArena")
//	ActGoFish               = NewTable[conf.ActGoFish]("ActGoFish")
//	ActLadderRace           = NewTable[conf.ActLadderRace]("ActLadderRace")
//	ActLadderRaceScore      = NewTable[conf.ActLadderRaceScore]("ActLadderRaceScore")
//	ActPassport             = NewTable[conf.ActPassport]("ActPassport")
//)

// === 月卡 ===
var MonthCard = NewTable[conf.MonthCard]("MonthCard")

//// === 宠物 ===
//var (
//	Pet                 = NewTable[conf.Pet]("Pet")
//	PetUpLevel          = NewTable[conf.PetUpLevel]("PetUpLevel")
//	PetUpStage          = NewTable[conf.PetUpStage]("PetUpStage")
//	PetUpStar           = NewTable[conf.PetUpStar]("PetUpStar")
//	PetDrawPool         = NewTable[conf.PetDrawPool]("PetDrawPool")
//	PetEquipHandbook    = NewTable[conf.PetEquipHandbook]("PetEquipHandbook")
//	PetEquipHbAward     = NewTable[conf.PetEquipHandbookAward]("PetEquipHandbookAward")
//	PetGiftCost         = NewTable[conf.PetGiftCost]("PetGiftCost")
//	PetGift             = NewTable[conf.PetGift]("PetGift")
//	PetSkill            = NewTable[conf.PetSkill]("PetSkill")
//	PetEquip            = NewTable[conf.PetEquip]("PetEquip")
//	PetEquipLevel       = NewTable[conf.PetEquipLevel]("PetEquipLevel")
//)
//
//// === 帮会 ===
//var (
//	Guild         = NewTable[conf.Guild]("Guild")
//	GuildTitle    = NewTable[conf.GuildTitle]("GuildTitle")
//	GuildElement  = NewTable[conf.GuildElement]("GuildElement")
//	GuildMaterial = NewTable[conf.GuildMaterial]("GuildMaterial")
//)
//
//// === 大闹天宫 ===
//var (
//	Uproar          = NewTable[conf.Uproar]("Uproar")
//	UproarFrequency = NewTable[conf.UproarFrequency]("UproarFrequency")
//)
//
//// === 排行 ===
//var RankAward = NewTable[conf.RankAward]("RankAward")
//
//// === 广播 / 公告 ===
//var (
//	BroadCast   = NewTable[conf.BroadCast]("BroadCast")
//	ChatChuanWen = NewTable[conf.ChatChuanWen]("ChatChuanWen")
//)
//
//// === 功能开放 ===
//var FunctionOpen = NewTable[conf.FunctionOpen]("FunctionOpen")
//
//// === 流派 ===
//var (
//	Liupai         = NewTable[conf.Liupai]("Liupai")
//	LiupaiRestrain = NewTable[conf.LiupaiRestrain]("LiupaiRestrain")
//)
//
//// === 鱼 ===
//var (
//	Fish           = NewTable[conf.Fish]("Fish")
//	FishSign       = NewTable[conf.FishSign]("FishSign")
//	FishLevelAward = NewTable[conf.FishLevelAward]("FishLevelAward")
//)
//
//// === 时装 / 头饰 ===
//var (
//	Fashion  = NewTable[conf.Fashion]("Fashion")
//	Headwear = NewTable[conf.Headwear]("Headwear")
//)
//
//// === 伴侣 ===
//var (
//	ParternerIntimacy = NewTable[conf.ParternerIntimacy]("ParternerIntimacy")
//	ParternerMission  = NewTable[conf.ParternerMission]("ParternerMission")
//	WineRack          = NewTable[conf.WineRack]("WineRack")
//	PeachTree         = NewTable[conf.PeachTree]("PeachTree")
//)
//
//// === 兑换码 ===
//var Cdkey = NewTable[conf.Cdkey]("Cdkey")

// === 福利 (CollectionSuit 等 JSON 存在但暂无 struct 的，按需在此注册) ===
// var CollectionSuit = NewTable[conf.CollectionSuit]("CollectionSuit")
