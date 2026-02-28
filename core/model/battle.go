package model

import (
	"math"
	"sort"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/define"
	"xfx/pkg/log"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"
)

type ChallengeBattleReportBack struct {
	Scene int
	Data  interface{}
}

type BattleReportBack_Player struct {
	PlayerId int64
	Data     interface{}
}

// 竞技场回调
type BattleReportBack_Arena struct {
	PlayerId int64
	ActId    int64
	ActCId   int64
	Fuchou   bool
	Data     interface{}
}

// 天梯回调
type BattleReportBack_Tianti struct {
	PlayerId int64
	ActId    int64
	ActCId   int64
	Data     interface{}
}

// 获取战斗个人数据
func GetBattlePlayerData(pl *proto_player.Context, _hero *Hero, skill *Skill, handIds []int32, magic *Magic, enchant map[int32]*EnchantOption, succinct []int32, Mount *MountOption, weaponry *WeaponryOption, equip []*EquipOption, brace *BraceOption, fashion *Fashion, destiny *Destiny, lineUp []int32) *proto_public.BattleHeroData {
	//获取战斗数据
	batData := new(proto_public.BattleHeroData)
	batData.Pid = pl.Id
	batData.Items = make([]*proto_public.BattleHeroItemData, 0)

	//流派阵容
	liupai := make(map[int32]int32)
	conf := config.CfgMgr.AllJson["Hero"].(map[int64]conf2.Hero)
	for _, v := range lineUp {
		if v <= 0 {
			continue
		}
		_conf := conf[int64(v)]
		if _, ok := liupai[_conf.Job]; !ok {
			liupai[_conf.Job] = 0
		}
		liupai[_conf.Job] += 1
	}
	basicAtkDamage := int32(0)
	basicSkillDamage := int32(0)
	jianyiBuffConTime := int32(0)
	liupaiConfs := config.CfgMgr.AllJson["Liupai"].(map[int64]conf2.Liupai)
	for job, num := range liupai {
		if num < 4 {
			continue
		}

		for _, c := range liupaiConfs {
			if c.Number == num {
				if job == define.PlayerJobYao {
					basicAtkDamage = c.JobAdd_Yao
				} else if job == define.PlayerJobShen {
					basicSkillDamage = c.JobAdd_Shen
				} else if job == define.PlayerJobFo {
					jianyiBuffConTime = (-1) * c.JobAdd_Fo
				}
				break
			}
		}

	}

	//获取布阵
	for index, v := range lineUp {
		if v <= 0 {
			continue
		}

		item := new(proto_public.BattleHeroItemData)
		item.Id = v

		hero := _hero.Hero[v]
		item.Star = hero.Star
		item.Level = hero.Level
		item.Stage = hero.Stage
		item.Cid = hero.Id
		item.Skin = "Default"
		item.LineUpIndex = int32(index)
		item.IsMainHero = v >= define.PlayerHeroID && v <= define.PlayerMainHeroFo

		//技能
		skills := GetHeroSkill(_hero, skill, item.Id)
		item.Skill = skills

		item.Attribute = new(proto_public.BattleAttributeData)

		InitAttribute(item)

		//角色基础
		GetBattleAttribute_Basic(item)

		//修为[主角]
		if item.IsMainHero {
			atkL, defL, hpL, forceL := int32(0), int32(0), int32(0), int32(0)
			if _, ok := hero.Cultivation[1]; ok {
				atkL = hero.Cultivation[1]
			}
			if _, ok := hero.Cultivation[2]; ok {
				defL = hero.Cultivation[2]
			}
			if _, ok := hero.Cultivation[3]; ok {
				hpL = hero.Cultivation[3]
			}
			if _, ok := hero.Cultivation[4]; ok {
				forceL = hero.Cultivation[4]
			}
			GetBattleAttribute_XiuWei(item, atkL, defL, hpL, forceL)
		}

		//图鉴
		GetBattleAttribute_Handbook(item, handIds)

		//功法
		magicIds := make(map[int32]int32)
		for _, v := range magic.LineUp {
			if v <= 0 {
				continue
			}
			magicIds[v] = magic.Ids[v].Level
		}
		GetBattleAttribute_Magic(item, magicIds)

		//附魔
		eids := make(map[int32]int32)
		allLevel := int32(0)
		for _, v := range enchant {
			eids[v.Id] = v.Level
			allLevel += v.Level
		}
		GetBattleAttribute_Enchant(item, eids, allLevel)

		//洗练
		GetBattleAttribute_Succinct(item, succinct)

		//坐骑
		GetBattleAttribute_Mount(item, Mount.Stage, Mount.Star)
		if Mount.UseId > 0 {
			mount_use := Mount.Mount[Mount.UseId]
			GetBattleAttribute_MountFinal(item, Mount.UseId, mount_use.Level)
		}
		GetBattleAttribute_MountEnergy(item, Mount.MountEnergy)
		if item.IsMainHero {
			item.MountId = Mount.UseId
		}

		//神兵
		GetBattleAttribute_Weaponry(item, weaponry.Star)
		if weaponry.UseId > 0 {
			weapon := weaponry.WeaponryItems[weaponry.UseId]
			GetBattleAttribute_WeaponryFinal(item, weaponry.UseId, weapon.Level)
		}
		if item.IsMainHero {
			item.WeaponId = weaponry.UseId
		}

		//装备
		eqids := make(map[int32]int32)
		for _, v := range equip {
			if v.IsUse {
				eqids[v.Id] = v.Level
			}
		}
		GetBattleAttribute_Equip(item, eqids)

		//背饰-灵韵
		aura := make(map[int32]int32)
		if brace.BraceAuraItems != nil {
			for _, v := range brace.BraceAuraItems {
				if v.Type <= 0 || v.Type > 4 {
					continue
				}
				aura[v.Type] = v.Level
			}
		}
		GetBattleAttribute_BraceAura(item, aura)

		//背饰灵韵等级
		GetBattleAttribute_BraceAuraLevel(item, brace.GetAuraStageAward)

		//背饰
		braceMap := make(map[int32]int32)
		if brace.BraceItems != nil {
			for _, v := range brace.BraceItems {
				braceMap[v.Id] = v.Level
			}
		}
		GetBattleAttribute_Brace(item, aura)

		//背饰天赋
		braceTalentMap := make(map[int32]int32)
		if brace.BraceTalentIndexs != nil {
			if talentIndexs, ok := brace.BraceTalentIndexs[int32(brace.BraceTalentIndex)]; ok {
				job := pl.Job
				if jobTalents, ok := talentIndexs.BraceTalentJobs[job]; ok {
					for _, v := range jobTalents.BraceTalentGroups {
						for _, item := range v.BraceTalentItems {
							braceTalentMap[item.Id] = item.Level
						}
					}
				}
			}
		}
		GetBattleAttribute_BraceTalent(item, braceTalentMap, item.IsMainHero)

		//头饰
		headWearMap := make(map[int32]int32)
		if fashion.HeadWear != nil {
			for _, item := range fashion.HeadWear {
				headWearMap[item.Id] = 1
			}
		}
		GetBattleAttribute_HeadWear(item, headWearMap, item.IsMainHero)

		//时装
		fashionMap := make(map[int32]int32)
		if fashion.FashionItems != nil {
			for _, item := range fashion.FashionItems {
				fashionMap[item.Id] = 1
			}
		}
		GetBattleAttribute_Fashion(item, fashionMap, item.IsMainHero)

		//天命
		if destiny.Ids != nil {
			GetBattleAttribute_Destiny(item, destiny.Ids)
		}

		if destiny.SelfIds != nil {
			GetBattleAttribute_DestinyStage(item, destiny.SelfIds)
		}

		//阵容流派
		item.Attribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(basicAtkDamage)
		item.Attribute.BasicSkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(basicSkillDamage)
		item.Attribute.JianyiBuffContinueTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(jianyiBuffConTime)

		batData.Items = append(batData.Items, item)
	}

	return batData
}

// 获取机器人战斗数据
func GetBattleRobotPlayerData(Id int64) *proto_public.BattleHeroData {
	//获取战斗数据
	batData := new(proto_public.BattleHeroData)
	batData.Pid = Id
	batData.Items = make([]*proto_public.BattleHeroItemData, 0)

	//获取机器人组
	robotGroups := config.CfgMgr.AllJson["RobotGroup"].(map[int64]conf2.RobotGroup)
	robotGroup, ok := robotGroups[Id]
	if !ok {
		log.Error("找不到机器人数据:%v", Id)
		return batData
	}

	robots := config.CfgMgr.AllJson["Robot"].(map[int64]conf2.Robot)
	for i := 0; i < len(robotGroup.RobotId); i++ {
		item := new(proto_public.BattleHeroItemData)

		id := robotGroup.RobotId[i]
		robot := robots[int64(id)]
		item.Id = robot.HeroId
		item.Level = robot.Level
		item.Cid = robot.HeroId
		item.Skin = "Default"
		item.LineUpIndex = int32(i)
		item.IsMainHero = robot.HeroId >= define.PlayerHeroID && robot.HeroId <= define.PlayerMainHeroFo
		item.Attribute = new(proto_public.BattleAttributeData)

		InitAttribute(item)
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)] = float32(robot.Atk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)] = float32(robot.Def)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)] = float32(robot.Hp)
		batData.Items = append(batData.Items, item)
	}

	return batData
}

func InitAttribute(item *proto_public.BattleHeroItemData) *proto_public.BattleHeroItemData {
	atk := new(proto_public.BattleItemAttributeData)
	atk.Sumvalues = make(map[int32]float32)
	atk.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	atk.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	atk.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.Atk = atk

	def := new(proto_public.BattleItemAttributeData)
	def.Sumvalues = make(map[int32]float32)
	def.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	def.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	def.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.Def = def

	hp := new(proto_public.BattleItemAttributeData)
	hp.Sumvalues = make(map[int32]float32)
	hp.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	hp.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	hp.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.Hp = hp

	force := new(proto_public.BattleItemAttributeData)
	force.Sumvalues = make(map[int32]float32)
	force.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	force.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	force.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.Force = force

	addCrit := new(proto_public.BattleItemAttributeData)
	addCrit.Sumvalues = make(map[int32]float32)
	//addCrit.Sumvalues[	int32(proto_public.AttributeType_Basic)] = 0
	//addCrit.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	addCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//addCrit.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.AddCrit = addCrit

	ignoreCrit := new(proto_public.BattleItemAttributeData)
	ignoreCrit.Sumvalues = make(map[int32]float32)
	//ignoreCrit.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//ignoreCrit.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	ignoreCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//ignoreCrit.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.IgnoreCrit = ignoreCrit

	addDamage := new(proto_public.BattleItemAttributeData)
	addDamage.Sumvalues = make(map[int32]float32)
	//addDamage.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	addDamage.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	addDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//addDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.AddDamage = addDamage

	damageRes := new(proto_public.BattleItemAttributeData)
	damageRes.Sumvalues = make(map[int32]float32)
	//damageRes.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	damageRes.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	damageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//damageRes.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.DamageRes = damageRes

	anger := new(proto_public.BattleItemAttributeData)
	anger.Sumvalues = make(map[int32]float32)
	//anger.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//anger.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	anger.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//anger.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.Anger = anger

	ignoreDodge := new(proto_public.BattleItemAttributeData)
	ignoreDodge.Sumvalues = make(map[int32]float32)
	//ignoreDodge.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//ignoreDodge.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	ignoreDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//ignoreDodge.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.IgnoreDodge = ignoreDodge

	addDodge := new(proto_public.BattleItemAttributeData)
	addDodge.Sumvalues = make(map[int32]float32)
	//addDodge.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//addDodge.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	addDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//addDodge.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.AddDodge = addDodge

	critDamage := new(proto_public.BattleItemAttributeData)
	critDamage.Sumvalues = make(map[int32]float32)
	//critDamage.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//critDamage.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	critDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//critDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.CritDamage = critDamage

	ignoreCritDamage := new(proto_public.BattleItemAttributeData)
	ignoreCritDamage.Sumvalues = make(map[int32]float32)
	//ignoreCritDamage.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//ignoreCritDamage.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	ignoreCritDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//ignoreCritDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.IgnoreCriteDamage = ignoreCritDamage

	addAtkSpeed := new(proto_public.BattleItemAttributeData)
	addAtkSpeed.Sumvalues = make(map[int32]float32)
	//addAtkSpeed.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//addAtkSpeed.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	addAtkSpeed.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//addAtkSpeed.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.AddAtkSpeed = addAtkSpeed

	conDamageRes := new(proto_public.BattleItemAttributeData)
	conDamageRes.Sumvalues = make(map[int32]float32)
	//conDamageRes.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//conDamageRes.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	conDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//conDamageRes.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.ContDamageRes = conDamageRes

	conDamage := new(proto_public.BattleItemAttributeData)
	conDamage.Sumvalues = make(map[int32]float32)
	//conDamage.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//conDamage.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	conDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//conDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.ContDamage = conDamage

	normalAtkDamage := new(proto_public.BattleItemAttributeData)
	normalAtkDamage.Sumvalues = make(map[int32]float32)
	//normalAtkDamage.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//normalAtkDamage.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	normalAtkDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//normalAtkDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.NormalAtkDamage = normalAtkDamage

	normalAtkDamageRes := new(proto_public.BattleItemAttributeData)
	normalAtkDamageRes.Sumvalues = make(map[int32]float32)
	//normalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//normalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	normalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//normalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.NormalAtkDamageRes = normalAtkDamageRes

	skillDamage := new(proto_public.BattleItemAttributeData)
	skillDamage.Sumvalues = make(map[int32]float32)
	//skillDamage.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//skillDamage.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	skillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//skillDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.SkillDamage = skillDamage

	skillDamageRes := new(proto_public.BattleItemAttributeData)
	skillDamageRes.Sumvalues = make(map[int32]float32)
	//skillDamageRes.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//skillDamageRes.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	skillDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//skillDamageRes.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.SkillDamageRes = skillDamageRes

	basicAttackDamage := new(proto_public.BattleItemAttributeData)
	basicAttackDamage.Sumvalues = make(map[int32]float32)
	//basicAttackDamage.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//basicAttackDamage.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	basicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//basicAttackDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.BasicAttackDamage = basicAttackDamage

	basicSkillDamage := new(proto_public.BattleItemAttributeData)
	basicSkillDamage.Sumvalues = make(map[int32]float32)
	//basicSkillDamage.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//basicSkillDamage.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	basicSkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//basicSkillDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.BasicSkillDamage = basicSkillDamage

	zengyiBuffConTime := new(proto_public.BattleItemAttributeData)
	zengyiBuffConTime.Sumvalues = make(map[int32]float32)
	//zengyiBuffConTime.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//zengyiBuffConTime.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	zengyiBuffConTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//zengyiBuffConTime.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.ZengyiBuffContinueTime = zengyiBuffConTime

	jianyiBuffConTime := new(proto_public.BattleItemAttributeData)
	jianyiBuffConTime.Sumvalues = make(map[int32]float32)
	//jianyiBuffConTime.Sumvalues[int32(proto_public.AttributeType_Basic)] = 0
	//jianyiBuffConTime.Sumvalues[int32(proto_public.AttributeType_Final)] = 0
	jianyiBuffConTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = 0
	//jianyiBuffConTime.Sumvalues[int32(proto_public.AttributeType_HandBook)] = 0
	item.Attribute.JianyiBuffContinueTime = jianyiBuffConTime

	return item
}

// 角色基础
func GetBattleAttribute_Basic(item *proto_public.BattleHeroItemData) *proto_public.BattleHeroItemData {
	attConf := config.CfgMgr.AllJson["HeroBasicAttribute"].(map[int64]conf2.HeroBasicAttribute)[int64(item.Id)]
	upStarConf := config.CfgMgr.AllJson["HeroUpStar"].(map[int64]conf2.HeroUpStar)[int64(item.Id)]
	upLevelConf := config.CfgMgr.AllJson["HeroUpLevel"].(map[int64]conf2.HeroUpLevel)[int64(item.Id)]
	upStageConf := config.CfgMgr.AllJson["HeroUpStage"].(map[int64]conf2.HeroUpStage)[int64(item.Id)]
	_atkBasic := float32(attConf.BasicAtk) *
		(float32(item.Level) * float32(upLevelConf.AtkRatio) / define.AttributeRate) *
		(1 + float32(item.Stage)*float32(upStageConf.AtkRatio)/define.AttributeRate) *
		(1 + float32(item.Star)*float32(upStarConf.AtkRatio)/define.AttributeRate)
	item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)] += _atkBasic

	_defBasic := float32(attConf.BasicDef) *
		(float32(item.Level) * float32(upLevelConf.DefRatio) / define.AttributeRate) *
		(1 + float32(item.Stage)*float32(upStageConf.DefRatio)/define.AttributeRate) *
		(1 + float32(item.Star)*float32(upStarConf.DefRatio)/define.AttributeRate)
	item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)] += _defBasic

	_hpBasic := float32(attConf.BasicDef) *
		(float32(item.Level) * float32(upLevelConf.HpRatio) / define.AttributeRate) *
		(1 + float32(item.Stage)*float32(upStageConf.HpRatio)/define.AttributeRate) *
		(1 + float32(item.Star)*float32(upStarConf.HpRatio)/define.AttributeRate)
	item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)] += _hpBasic

	_forceBasic := float32(attConf.BasicForce) *
		(float32(item.Level) * float32(upLevelConf.ForceRatio) / define.AttributeRate) *
		(1 + float32(item.Stage)*float32(upStageConf.ForceRatio)/define.AttributeRate) *
		(1 + float32(item.Star)*float32(upStarConf.ForceRatio)/define.AttributeRate)
	item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Basic)] += _forceBasic
	return item
}

// 修为
func GetBattleAttribute_XiuWei(item *proto_public.BattleHeroItemData, AtkLevel, DefLevel, HpLevel, ForceLevel int32) *proto_public.BattleHeroItemData {
	heroConf := config.CfgMgr.AllJson["Hero"].(map[int64]conf2.Hero)[int64(item.Id)]

	culConfs := config.CfgMgr.AllJson["HeroCultivation"].(map[int64]conf2.HeroCultivation)
	var cultivation conf2.HeroCultivation
	for _, v := range culConfs {
		if v.Job == heroConf.Job && v.Stage == item.Stage {
			cultivation = v
			break
		}
	}

	if cultivation.Id <= 0 {
		return item
	}

	_Hp :=
		float32(cultivation.BasicHp) +
			float32(cultivation.BasicHp)*float32(math.Pow(
				float64(HpLevel)/float64(cultivation.Cultivation),
				float64(cultivation.AttibuteRatio),
			),
			)
	item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)] += _Hp

	_Atk :=
		float32(cultivation.BasicAtk) +
			float32(cultivation.BasicAtk)*float32(math.Pow(
				float64(AtkLevel)/float64(cultivation.Cultivation),
				float64(cultivation.AttibuteRatio),
			),
			)
	item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)] += _Atk

	_Def :=
		float32(cultivation.BasicDef) +
			float32(cultivation.BasicDef)*float32(math.Pow(
				float64(DefLevel)/float64(cultivation.Cultivation),
				float64(cultivation.AttibuteRatio),
			),
			)
	item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)] += _Def

	_Force :=
		float32(cultivation.BasicForce) +
			float32(cultivation.BasicForce)*float32(math.Pow(
				float64(ForceLevel)/float64(cultivation.Cultivation),
				float64(cultivation.AttibuteRatio),
			),
			)
	item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Basic)] += _Force

	return item
}

// 图鉴
func GetBattleAttribute_Handbook(item *proto_public.BattleHeroItemData, ids []int32) *proto_public.BattleHeroItemData {
	for _, v := range ids {
		if v <= 0 {
			continue
		}
		//基础
		conf := config.CfgMgr.AllJson["HandbookAward"].(map[int64]conf2.HandBookAward)[int64(v)]
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicHp)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicForce)

		//千分比
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddHp)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddForce)

		//主角
		if item.IsMainHero {
			item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_HandBook)] += float32(conf.AddHeroAtk)
			item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_HandBook)] += float32(conf.AddHeroDef)
			item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_HandBook)] += float32(conf.AddHeroHp)
			item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_HandBook)] += float32(conf.AddHeroForce)
		}
	}
	return item
}

// 功法
func GetBattleAttribute_Magic(item *proto_public.BattleHeroItemData, ids map[int32]int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["HeroMagicLevel"].(map[int64]conf2.HeroMagicLevel)

	for k, v := range ids {
		var conf conf2.HeroMagicLevel
		for _, n := range confs {
			if n.MagicId == k && n.Level == v {
				conf = n
				break
			}
		}

		if conf.Id <= 0 {
			continue
		}

		//千分比
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.PerHeroAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.PerHeroDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.PerHeroHP)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.PerHeroForce)

		//主角
		if item.IsMainHero {
			item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.HeroAtk)
			item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.HeroDef)
			item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.HeroHP)
			item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.HeroForce)
		}
	}
	return item
}

// 附魔
func GetBattleAttribute_Enchant(item *proto_public.BattleHeroItemData, ids map[int32]int32, level int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["Enchant"].(map[int64]conf2.Enchant)

	for _key, _level := range ids {
		var conf conf2.Enchant
		for _, n := range confs {
			if n.Level == _level {
				conf = n
				break
			}
		}

		if conf.Id <= 0 {
			continue
		}

		//主角
		if item.IsMainHero {
			if _key == define.ItemIdTallyAtk {
				item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicAtk)
			} else if _key == define.ItemIdTallyDef {
				item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicDef)
			} else if _key == define.ItemIdTallyHp {
				item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicHp)
			} else if _key == define.ItemIdTallyForce {
				item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicForce)
			}
		}
	}

	//总等级
	stageConfs := config.CfgMgr.AllJson["EnchantStage"].(map[int64]conf2.EnchantStage)
	confList := make([]conf2.EnchantStage, 0)
	for _, v := range stageConfs {
		confList = append(confList, v)
	}

	//排序
	sort.Slice(confList, func(i, j int) bool {
		return confList[i].Id < confList[j].Id
	})

	var confStage conf2.EnchantStage
	for k := 0; k < len(confList); k++ {
		if confList[k].Level >= level {
			if k == 0 {
				confStage = confList[k]
			} else {
				confStage = confList[k-1]
			}
			break
		}
	}

	if confStage.Id > 0 {
		//千分比
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(confStage.AddAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(confStage.AddDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(confStage.AddHp)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(confStage.AddForce)
	}

	return item
}

// 洗练
func GetBattleAttribute_Succinct(item *proto_public.BattleHeroItemData, ids []int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["Succinct"].(map[int64]conf2.Succinct)

	for _, _level := range ids {
		var conf conf2.Succinct
		for _, n := range confs {
			if n.Level == _level {
				conf = n
				break
			}
		}

		if conf.Id <= 0 {
			continue
		}

		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddHp)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddForce)
	}

	return item
}

// 坐骑
func GetBattleAttribute_Mount(item *proto_public.BattleHeroItemData, stage, star int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["MountStage"].(map[int64]conf2.MountStage)

	var conf conf2.MountStage
	for _, v := range confs {
		if v.Star == star && v.Stage == stage {
			conf = v
			break
		}
	}

	if conf.Id > 0 {
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddHp)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddForce)
	}

	return item
}

// 坐骑最终加成
func GetBattleAttribute_MountFinal(item *proto_public.BattleHeroItemData, id, level int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["MountLevel"].(map[int64]conf2.MountLevel)

	var conf conf2.MountLevel
	for _, v := range confs {
		if v.Level == level && v.MountId == id {
			conf = v
			break
		}
	}

	if conf.Id > 0 {
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumHp)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumForce)

		item.Attribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddCrit)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
	}

	return item
}

// 坐骑赋能
func GetBattleAttribute_MountEnergy(item *proto_public.BattleHeroItemData, ids map[int32]int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["MountEnergyAttribute"].(map[int64]conf2.MountEnergyAttribute)

	for _typ, _level := range ids {
		var conf conf2.MountEnergyAttribute
		for _, v := range confs {
			if v.Level == _level && v.AttributeType == _typ {
				conf = v
				break
			}
		}

		if conf.Id <= 0 {
			continue
		}

		if conf.Id > 0 {
			if _typ == 1 {
				item.Attribute.SkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAttribute1)
				item.Attribute.SkillDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAttribute2)
			} else if _typ == 2 {
				item.Attribute.NormalAtkDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAttribute1)
				item.Attribute.NormalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAttribute2)
			} else if _typ == 3 {
				item.Attribute.ContDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAttribute1)
				item.Attribute.ContDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAttribute2)
			}
		}
	}

	return item
}

// 神兵
func GetBattleAttribute_Weaponry(item *proto_public.BattleHeroItemData, star int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["WeaponryStar"].(map[int64]conf2.WeaponryStar)

	var conf conf2.WeaponryStar
	for _, v := range confs {
		if v.Star == star {
			conf = v
			break
		}
	}

	if conf.Id > 0 {
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddHp)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddForce)
		item.Attribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddCrit)
	}

	return item
}

// 神兵最终
func GetBattleAttribute_WeaponryFinal(item *proto_public.BattleHeroItemData, id, level int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["WeaponryLevel"].(map[int64]conf2.WeaponryLevel)

	var conf conf2.WeaponryLevel
	for _, v := range confs {
		if v.Level == level && v.WeaponryId == id {
			conf = v
			break
		}
	}

	if conf.Id > 0 {
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumAtk)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumDef)
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumHp)
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumForce)
		item.Attribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddCrit)
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
	}

	return item
}

// 背饰灵韵
func GetBattleAttribute_BraceAura(item *proto_public.BattleHeroItemData, levels map[int32]int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["BraceAura"].(map[int64]conf2.BraceAura)

	for k, v := range levels {
		var conf conf2.BraceAura
		for _, _conf := range confs {
			if _conf.Level == v {
				conf = _conf
				break
			}
		}
		if conf.Id > 0 {
			//攻击
			if k == 1 {
				item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAtk)
			} else if k == 2 { //防御
				item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
			} else if k == 3 { //气血
				item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddHp)
			} else if k == 4 { //内力
				item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddForce)
			}
		}
	}

	return item
}

// 背饰灵韵等级
func GetBattleAttribute_BraceAuraLevel(item *proto_public.BattleHeroItemData, levels []int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["BraceAuraStage"].(map[int64]conf2.BraceAuraStage)
	for _, v := range levels {
		var conf conf2.BraceAuraStage
		for _, _conf := range confs {
			if _conf.Id == v {
				conf = _conf
				break
			}
		}
		if conf.Id > 0 {
			item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAtk)
			item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
			item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddHp)
			item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddForce)
		}
	}

	return item
}

// 背饰
func GetBattleAttribute_Brace(item *proto_public.BattleHeroItemData, levels map[int32]int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["BracesLevel"].(map[int64]conf2.BracesLevel)

	for _, v := range levels {
		var conf conf2.BracesLevel
		for _, _conf := range confs {
			if _conf.Level == v {
				conf = _conf
				break
			}
		}
		if conf.Id > 0 {
			item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumAtk)
			item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumDef)
			item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumHp)
			item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(conf.AddSumForce)
			item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDef)
			item.Attribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddCrit)
		}
	}

	return item
}

// 背饰天赋
func GetBattleAttribute_BraceTalent(item *proto_public.BattleHeroItemData, ids map[int32]int32, isMainHero bool) *proto_public.BattleHeroItemData {
	confs_talentLevels := config.CfgMgr.AllJson["BraceTalentLevel"].(map[int64]conf2.BraceTalentLevel)
	confs_talents := config.CfgMgr.AllJson["BraceTalent"].(map[int64]conf2.BraceTalent)
	for k, v := range ids {
		if v <= 0 {
			continue
		}

		var conf conf2.BraceTalent
		for _, _conf := range confs_talents {
			if _conf.Id == k {
				conf = _conf
				break
			}
		}
		if conf.Id > 0 {
			var _conf conf2.BraceTalentLevel
			for _, level := range confs_talentLevels {
				if level.TalentLevelId == conf.TalentLevelId && level.Level == v {
					_conf = level
					break
				}
			}

			if _conf.Id > 0 {
				switch _conf.AttId {
				case define.AttributeIdHp:
					item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
				case define.AttributeIdAttack:
					item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
				case define.AttributeIdDef:
					item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
				case define.AttributeIdForce:
					item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
				case define.AttributeIdAddDamage:
					if isMainHero {
						item.Attribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdDamageRes:
					if isMainHero {
						item.Attribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdAnger:
					if isMainHero {
						item.Attribute.Anger.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdIgnoreDodge:
					if isMainHero {
						item.Attribute.IgnoreDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdAddDodge:
					if isMainHero {
						item.Attribute.AddDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdCrit:
					if isMainHero {
						item.Attribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdIgnoreCrit:
					if isMainHero {
						item.Attribute.IgnoreCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdCritDamage:
					if isMainHero {
						item.Attribute.CritDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdIgnoreCritDamage:
					if isMainHero {
						item.Attribute.IgnoreCriteDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdAtkSpeed:
					if isMainHero {
						item.Attribute.AddAtkSpeed.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdContDamageRes:
					if isMainHero {
						item.Attribute.ContDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdNormalAtkDamage:
					if isMainHero {
						item.Attribute.NormalAtkDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdNormalAtkDamageRes:
					if isMainHero {
						item.Attribute.NormalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdSkillDamage:
					if isMainHero {
						item.Attribute.SkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdSkillDamageRes:
					if isMainHero {
						item.Attribute.SkillDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdContDamage:
					if isMainHero {
						item.Attribute.ContDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdFinalDef:
					if isMainHero {
						item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(_conf.AttValue)
					}
				case define.AttributeIdFinalAtk:
					if isMainHero {
						item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(_conf.AttValue)
					}
				case define.AttributeIdFinalHp:
					if isMainHero {
						item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(_conf.AttValue)
					}
				case define.AttributeIdFinalForce:
					if isMainHero {
						item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(_conf.AttValue)
					}
				case define.AttributeIdFinalAddDamage:
					if isMainHero {
						item.Attribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(_conf.AttValue)
					}
				case define.AttributeIdFinalDamageRes:
					if isMainHero {
						item.Attribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(_conf.AttValue)
					}
				case define.AttributeIdBasicAttackDamage:
					if isMainHero {
						item.Attribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdBasicSkillDamage:
					if isMainHero {
						item.Attribute.BasicSkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdZengyiBuffContinueTime:
					if isMainHero {
						item.Attribute.ZengyiBuffContinueTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				case define.AttributeIdJianyiBuffContinueTime:
					if isMainHero {
						item.Attribute.JianyiBuffContinueTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(_conf.AttValue)
					}
				}
			}
		}
	}

	return item
}

// 头饰
func GetBattleAttribute_HeadWear(item *proto_public.BattleHeroItemData, ids map[int32]int32, isMainHero bool) *proto_public.BattleHeroItemData {
	Headwears := config.CfgMgr.AllJson["Headwear"].(map[int64]conf2.Headwear)
	for k, v := range ids {
		if v <= 0 {
			continue
		}

		AttId := int32(0)
		AttValue := int32(0)
		for _, _conf := range Headwears {
			if _conf.Id == k {
				for attrId, attrVal := range _conf.AttributeId {
					AttId = attrId
					AttValue = attrVal
				}
				break
			}
		}

		if AttId > 0 {
			switch AttId {
			case define.AttributeIdHp:
				item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
			case define.AttributeIdAttack:
				item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
			case define.AttributeIdDef:
				item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
			case define.AttributeIdForce:
				item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
			case define.AttributeIdAddDamage:
				if isMainHero {
					item.Attribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdDamageRes:
				if isMainHero {
					item.Attribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdAnger:
				if isMainHero {
					item.Attribute.Anger.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdIgnoreDodge:
				if isMainHero {
					item.Attribute.IgnoreDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdAddDodge:
				if isMainHero {
					item.Attribute.AddDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdCrit:
				if isMainHero {
					item.Attribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdIgnoreCrit:
				if isMainHero {
					item.Attribute.IgnoreCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdCritDamage:
				if isMainHero {
					item.Attribute.CritDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdIgnoreCritDamage:
				if isMainHero {
					item.Attribute.IgnoreCriteDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdAtkSpeed:
				if isMainHero {
					item.Attribute.AddAtkSpeed.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdContDamageRes:
				if isMainHero {
					item.Attribute.ContDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdNormalAtkDamage:
				if isMainHero {
					item.Attribute.NormalAtkDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdNormalAtkDamageRes:
				if isMainHero {
					item.Attribute.NormalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdSkillDamage:
				if isMainHero {
					item.Attribute.SkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdSkillDamageRes:
				if isMainHero {
					item.Attribute.SkillDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdContDamage:
				if isMainHero {
					item.Attribute.ContDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdFinalDef:
				if isMainHero {
					item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalAtk:
				if isMainHero {
					item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalHp:
				if isMainHero {
					item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalForce:
				if isMainHero {
					item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalAddDamage:
				if isMainHero {
					item.Attribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalDamageRes:
				if isMainHero {
					item.Attribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdBasicAttackDamage:
				if isMainHero {
					item.Attribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdBasicSkillDamage:
				if isMainHero {
					item.Attribute.BasicSkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdZengyiBuffContinueTime:
				if isMainHero {
					item.Attribute.ZengyiBuffContinueTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdJianyiBuffContinueTime:
				if isMainHero {
					item.Attribute.JianyiBuffContinueTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			}
		}
	}

	return item
}

// 时装
func GetBattleAttribute_Fashion(item *proto_public.BattleHeroItemData, ids map[int32]int32, isMainHero bool) *proto_public.BattleHeroItemData {
	Fashions := config.CfgMgr.AllJson["Fashion"].(map[int64]conf2.Fashion)
	for k, v := range ids {
		if v <= 0 {
			continue
		}

		AttId := int32(0)
		AttValue := int32(0)
		for _, _conf := range Fashions {
			if _conf.Id == k {
				for attrId, attrVal := range _conf.AttributeId {
					AttId = attrId
					AttValue = attrVal
				}
				break
			}
		}

		if AttId > 0 {
			switch AttId {
			case define.AttributeIdHp:
				item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
			case define.AttributeIdAttack:
				item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
			case define.AttributeIdDef:
				item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
			case define.AttributeIdForce:
				item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
			case define.AttributeIdAddDamage:
				if isMainHero {
					item.Attribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdDamageRes:
				if isMainHero {
					item.Attribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdAnger:
				if isMainHero {
					item.Attribute.Anger.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdIgnoreDodge:
				if isMainHero {
					item.Attribute.IgnoreDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdAddDodge:
				if isMainHero {
					item.Attribute.AddDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdCrit:
				if isMainHero {
					item.Attribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdIgnoreCrit:
				if isMainHero {
					item.Attribute.IgnoreCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdCritDamage:
				if isMainHero {
					item.Attribute.CritDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdIgnoreCritDamage:
				if isMainHero {
					item.Attribute.IgnoreCriteDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdAtkSpeed:
				if isMainHero {
					item.Attribute.AddAtkSpeed.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdContDamageRes:
				if isMainHero {
					item.Attribute.ContDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdNormalAtkDamage:
				if isMainHero {
					item.Attribute.NormalAtkDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdNormalAtkDamageRes:
				if isMainHero {
					item.Attribute.NormalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdSkillDamage:
				if isMainHero {
					item.Attribute.SkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdSkillDamageRes:
				if isMainHero {
					item.Attribute.SkillDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdContDamage:
				if isMainHero {
					item.Attribute.ContDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdFinalDef:
				if isMainHero {
					item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalAtk:
				if isMainHero {
					item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalHp:
				if isMainHero {
					item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalForce:
				if isMainHero {
					item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalAddDamage:
				if isMainHero {
					item.Attribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdFinalDamageRes:
				if isMainHero {
					item.Attribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_Final)] += float32(AttValue)
				}
			case define.AttributeIdBasicAttackDamage:
				if isMainHero {
					item.Attribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdBasicSkillDamage:
				if isMainHero {
					item.Attribute.BasicSkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdZengyiBuffContinueTime:
				if isMainHero {
					item.Attribute.ZengyiBuffContinueTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			case define.AttributeIdJianyiBuffContinueTime:
				if isMainHero {
					item.Attribute.JianyiBuffContinueTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(AttValue)
				}
			}
		}
	}

	return item
}

// 装备
func GetBattleAttribute_Equip(item *proto_public.BattleHeroItemData, equip map[int32]int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["Equip"].(map[int64]conf2.Equip)

	for id, _level := range equip {
		var conf conf2.Equip
		conf = confs[int64(id)]
		if conf.Id <= 0 {
			continue
		}

		level := float32(_level)
		item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicAtk) * level
		item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicDef) * level
		item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicHp) * level
		item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.BasicForce) * level

		//主角
		if item.IsMainHero {
			item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.RAtk) * level
			item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.Rdef) * level
			item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.RHp) * level
			item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_Basic)] += float32(conf.RForce) * level

			item.Attribute.SkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddSkilDamage) * level
			item.Attribute.SkillDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAniSkilDamage) * level
			item.Attribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDamage) * level
			item.Attribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAniDamage) * level
			item.Attribute.Anger.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddMp) * level
			item.Attribute.IgnoreDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddIgnoreDodge) * level
			item.Attribute.AddDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddDodge) * level
			item.Attribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddCrit) * level
			item.Attribute.IgnoreCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddIgnoreCrit) * level
			//爆伤和忽略爆伤没有
			item.Attribute.AddAtkSpeed.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAtkSpeed) * level
			item.Attribute.ContDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddTimeDmage) * level
			item.Attribute.ContDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddTimeAniDamage) * level
			item.Attribute.NormalAtkDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddNormalAtk) * level
			item.Attribute.NormalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AddAniNormalAtk) * level
		}
	}

	return item
}

// 天命
func GetBattleAttribute_Destiny(item *proto_public.BattleHeroItemData, ids []int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["DestinyLevel"].(map[int64]conf2.DestinyLevel)

	for _, v := range ids {
		var conf conf2.DestinyLevel
		for _, _conf := range confs {
			if _conf.Id == v {
				conf = _conf
				break
			}
		}
		if conf.Id > 0 {
			switch conf.TeamAttribute {
			case define.AttributeIdHp:
				item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AttributeValue)
			case define.AttributeIdAttack:
				item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AttributeValue)
			case define.AttributeIdDef:
				item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AttributeValue)
			case define.AttributeIdForce:
				item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.AttributeValue)
			}
		}
	}

	return item
}

// 天高阶
func GetBattleAttribute_DestinyStage(item *proto_public.BattleHeroItemData, ids []int32) *proto_public.BattleHeroItemData {
	confs := config.CfgMgr.AllJson["DestinyStage"].(map[int64]conf2.DestinyStage)

	for _, v := range ids {
		var conf conf2.DestinyStage
		for _, _conf := range confs {
			if _conf.Id == v {
				conf = _conf
				break
			}
		}
		if conf.Id > 0 {
			switch conf.SelfAttribute {
			case define.AttributeIdHp:
				item.Attribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.SelfAttributeValue)
			case define.AttributeIdAttack:
				item.Attribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.SelfAttributeValue)
			case define.AttributeIdDef:
				item.Attribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.SelfAttributeValue)
			case define.AttributeIdForce:
				item.Attribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)] += float32(conf.SelfAttributeValue)
			}
		}
	}

	return item
}
