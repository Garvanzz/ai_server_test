package config

import (
	"xfx/core/config/conf"
)

func Parse(data map[string]any) {
	var parseData any
	for jsonName, v := range data {
		//log.Debug("parse json name:%v", jsonName)
		switch jsonName {
		case "Hero":
			parseData = ParseToStruct[conf.Hero](v, jsonName)
		case "HeroBasicAttribute":
			parseData = ParseToStruct[conf.HeroBasicAttribute](v, jsonName)
		case "HeroUpStar":
			parseData = ParseToStruct[conf.HeroUpStar](v, jsonName)
		case "HeroUpStage":
			parseData = ParseToStruct[conf.HeroUpStage](v, jsonName)
		case "HeroUpLevel":
			parseData = ParseToStruct[conf.HeroUpLevel](v, jsonName)
		case "HeroCultivation":
			parseData = ParseToStruct[conf.HeroCultivation](v, jsonName)
		case "Item":
			parseData = ParseToStruct[conf.Item](v, jsonName)
		case "Shop":
			parseData = ParseToStruct[conf.Shop](v, jsonName)
		case "Recharge":
			parseData = ParseToStruct[conf.Recharge](v, jsonName)
		case "Guild":
			parseData = ParseToStruct[conf.Guild](v, jsonName)
		case "Task":
			parseData = ParseToStruct[conf.Task](v, jsonName)
		case "TaskActivity":
			parseData = ParseToStruct[conf.TaskActivity](v, jsonName)
		case "Global":
			parseData = AttachToStruct[conf.Global](v, jsonName)
		case "DrawPool":
			parseData = ParseToStruct[conf.DrawPool](v, jsonName)
		case "HeroPool":
			parseData = ParseToStruct[conf.HeroPool](v, jsonName)
		case "Skill":
			parseData = ParseToStruct[conf.Skill](v, jsonName)
		case "Monster":
			parseData = ParseToStruct[conf.Monster](v, jsonName)
		case "StageMonsterGroup":
			parseData = ParseToStruct[conf.StageMonsterGroup](v, jsonName)
		case "MonsterGroup":
			parseData = ParseToStruct[conf.MonsterGroup](v, jsonName)
		case "Stage":
			parseData = ParseToStruct[conf.Stage](v, jsonName)
		case "Chapter":
			parseData = ParseToStruct[conf.Chapter](v, jsonName)
		case "Buff":
			parseData = ParseToStruct[conf.Buff](v, jsonName)
		case "AttributeId":
			parseData = ParseToStruct[conf.AttributeId](v, jsonName)
		case "Equip":
			parseData = ParseToStruct[conf.Equip](v, jsonName)
		case "Drop":
			parseData = ParseToStruct[conf.Drop](v, jsonName)
		case "BoxLevelDrop":
			parseData = ParseToStruct[conf.BoxLevelDrop](v, jsonName)
		case "PlayerLevel":
			parseData = ParseToStruct[conf.PlayerLevel](v, jsonName)
		case "DaySign":
			parseData = ParseToStruct[conf.DaySign](v, jsonName)
		case "EquipSell":
			parseData = ParseToStruct[conf.EquipSell](v, jsonName)
		case "HandbookAward":
			parseData = ParseToStruct[conf.HandBookAward](v, jsonName)
		case "HeroMagicLevel":
			parseData = ParseToStruct[conf.HeroMagicLevel](v, jsonName)
		case "HeroMagic":
			parseData = ParseToStruct[conf.HeroMagic](v, jsonName)
		case "IdleBox":
			parseData = ParseToStruct[conf.IdleBox](v, jsonName)
		case "Mount":
			parseData = ParseToStruct[conf.Mount](v, jsonName)
		case "MountStage":
			parseData = ParseToStruct[conf.MountStage](v, jsonName)
		case "MountEnergy":
			parseData = ParseToStruct[conf.MountEnergy](v, jsonName)
		case "MountEnergyAttribute":
			parseData = ParseToStruct[conf.MountEnergyAttribute](v, jsonName)
		case "WeaponryStar":
			parseData = ParseToStruct[conf.WeaponryStar](v, jsonName)
		case "DrawStageAward":
			parseData = ParseToStruct[conf.DrawStageAward](v, jsonName)
		case "RecruitLevelAward":
			parseData = ParseToStruct[conf.RecruitLevelAward](v, jsonName)
		case "Enchant":
			parseData = ParseToStruct[conf.Enchant](v, jsonName)
		case "Succinct":
			parseData = ParseToStruct[conf.Succinct](v, jsonName)
		case "SuccinctSkill":
			parseData = ParseToStruct[conf.SuccinctSkill](v, jsonName)
		case "Handbook":
			parseData = ParseToStruct[conf.HandBook](v, jsonName)
		case "MountLevel":
			parseData = ParseToStruct[conf.MountLevel](v, jsonName)
		case "WeaponryLevel":
			parseData = ParseToStruct[conf.WeaponryLevel](v, jsonName)
		case "BraceAura":
			parseData = ParseToStruct[conf.BraceAura](v, jsonName)
		case "BraceAuraStage":
			parseData = ParseToStruct[conf.BraceAuraStage](v, jsonName)
		case "ShenjiPool":
			parseData = ParseToStruct[conf.ShenjiPool](v, jsonName)
		case "DestinyStage":
			parseData = ParseToStruct[conf.DestinyStage](v, jsonName)
		case "DestinyLevel":
			parseData = ParseToStruct[conf.DestinyLevel](v, jsonName)
		case "Braces":
			parseData = ParseToStruct[conf.Braces](v, jsonName)
		case "BraceTalent":
			parseData = ParseToStruct[conf.BraceTalent](v, jsonName)
		case "BraceTalentLevel":
			parseData = ParseToStruct[conf.BraceTalentLevel](v, jsonName)
		case "BracesLevel":
			parseData = ParseToStruct[conf.BracesLevel](v, jsonName)
		case "CollectionUpStar":
			parseData = ParseToStruct[conf.CollectionUpStar](v, jsonName)
		case "Collection":
			parseData = ParseToStruct[conf.Collection](v, jsonName)
		case "Divine":
			parseData = ParseToStruct[conf.Divine](v, jsonName)
		case "Learning":
			parseData = ParseToStruct[conf.Learning](v, jsonName)
		case "LearningCompose":
			parseData = ParseToStruct[conf.LearningCompose](v, jsonName)
		case "Activity":
			parseData = ParseToStruct[conf.Activity](v, jsonName)
		case "ActDailyAccumulateRecharge":
			parseData = ParseToStruct[conf.ActDailyAccumulateRecharge](v, jsonName)
		case "TreasurePool":
			parseData = ParseToStruct[conf.TreasurePool](v, jsonName)
		case "MonthCard":
			parseData = ParseToStruct[conf.MonthCard](v, jsonName)
		case "Pet":
			parseData = ParseToStruct[conf.Pet](v, jsonName)
		case "PetUpLevel":
			parseData = ParseToStruct[conf.PetUpLevel](v, jsonName)
		case "PetUpStage":
			parseData = ParseToStruct[conf.PetUpStage](v, jsonName)
		case "PetUpStar":
			parseData = ParseToStruct[conf.PetUpStar](v, jsonName)
		case "PetDrawPool":
			parseData = ParseToStruct[conf.PetDrawPool](v, jsonName)
		case "PetEquipHandbook":
			parseData = ParseToStruct[conf.PetEquipHandbook](v, jsonName)
		case "PetEquipHandbookAward":
			parseData = ParseToStruct[conf.PetEquipHandbookAward](v, jsonName)
		case "PetGiftCost":
			parseData = ParseToStruct[conf.PetGiftCost](v, jsonName)
		case "PetGift":
			parseData = ParseToStruct[conf.PetGift](v, jsonName)
		case "PetSkill":
			parseData = ParseToStruct[conf.PetSkill](v, jsonName)
		case "PetEquip":
			parseData = ParseToStruct[conf.PetEquip](v, jsonName)
		case "PetEquipLevel":
			parseData = ParseToStruct[conf.PetEquipLevel](v, jsonName)
		case "GuildTitle":
			parseData = ParseToStruct[conf.GuildTitle](v, jsonName)
		case "GuildElement":
			parseData = ParseToStruct[conf.GuildElement](v, jsonName)
		case "GuildMaterial":
			parseData = ParseToStruct[conf.GuildMaterial](v, jsonName)
		case "Uproar":
			parseData = ParseToStruct[conf.Uproar](v, jsonName)
		case "UproarFrequency":
			parseData = ParseToStruct[conf.UproarFrequency](v, jsonName)
		case "Mission":
			parseData = ParseToStruct[conf.Mission](v, jsonName)
		case "EnchantStage":
			parseData = ParseToStruct[conf.EnchantStage](v, jsonName)
		case "ClimbTower":
			parseData = ParseToStruct[conf.ClimbTower](v, jsonName)
		case "Robot":
			parseData = ParseToStruct[conf.Robot](v, jsonName)
		case "RobotGroup":
			parseData = ParseToStruct[conf.RobotGroup](v, jsonName)
		case "BroadCast":
			parseData = ParseToStruct[conf.BroadCast](v, jsonName)
		case "ActTheCompetition":
			parseData = ParseToStruct[conf.ActTheCompetition](v, jsonName)
		case "ChatChuanWen":
			parseData = ParseToStruct[conf.ChatChuanWen](v, jsonName)
		case "ActBoxFund":
			parseData = ParseToStruct[conf.ActBoxFund](v, jsonName)
		case "ActMainLineFund":
			parseData = ParseToStruct[conf.ActMainLineFund](v, jsonName)
		case "ActLevelFund":
			parseData = ParseToStruct[conf.ActLevelFund](v, jsonName)
		case "ActArena":
			parseData = ParseToStruct[conf.ActArena](v, jsonName)
		case "RankAward":
			parseData = ParseToStruct[conf.RankAward](v, jsonName)
		case "FunctionOpen":
			parseData = ParseToStruct[conf.FunctionOpen](v, jsonName)
		case "Liupai":
			parseData = ParseToStruct[conf.Liupai](v, jsonName)
		case "LiupaiRestrain":
			parseData = ParseToStruct[conf.LiupaiRestrain](v, jsonName)
		case "ActGoFish":
			parseData = ParseToStruct[conf.ActGoFish](v, jsonName)
		case "Fish":
			parseData = ParseToStruct[conf.Fish](v, jsonName)
		case "FishSign":
			parseData = ParseToStruct[conf.FishSign](v, jsonName)
		case "FishLevelAward":
			parseData = ParseToStruct[conf.FishLevelAward](v, jsonName)
		case "Weaponry":
			parseData = ParseToStruct[conf.Weaponry](v, jsonName)
		case "Fashion":
			parseData = ParseToStruct[conf.Fashion](v, jsonName)
		case "Headwear":
			parseData = ParseToStruct[conf.Headwear](v, jsonName)
		case "ActLadderRace":
			parseData = ParseToStruct[conf.ActLadderRace](v, jsonName)
		case "ActLadderRaceScore":
			parseData = ParseToStruct[conf.ActLadderRaceScore](v, jsonName)
		case "ActPassport":
			parseData = ParseToStruct[conf.ActPassport](v, jsonName)
		case "ParternerIntimacy":
			parseData = ParseToStruct[conf.ParternerIntimacy](v, jsonName)
		case "WineRack":
			parseData = ParseToStruct[conf.WineRack](v, jsonName)
		case "ParternerMission":
			parseData = ParseToStruct[conf.ParternerMission](v, jsonName)
		case "PeachTree":
			parseData = ParseToStruct[conf.PeachTree](v, jsonName)
		case "Cdkey":
			parseData = ParseToStruct[conf.Cdkey](v, jsonName)
		}

		CfgMgr.AllJson[jsonName] = parseData
	}
}
