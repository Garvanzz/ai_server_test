package global

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
	"xfx/proto/proto_game"
	"xfx/proto/proto_player"
	"xfx/proto/proto_public"
)

// 获取战斗数据
func GetBattlePlayerData(pl *proto_player.Context, lineUp []*proto_public.CommonPlayerLineUpItemInfo) *proto_public.BattleHeroData {
	//图鉴ID
	handbookId := make([]int32, 0)
	heroS := GetPlayerHero(pl.Id)
	skillMap := GetPlayerSkill(pl.Id)

	//角色图鉴
	phandbook := GetPlayerHandbook(pl.Id)
	if phandbook.HandbookOption.GetId != nil {
		handbookId = append(handbookId, phandbook.HandbookOption.GetId...)
	}

	magic := GetPlayerMagic(pl.Id)
	equips := GetPlayerEquip(pl.Id)

	//神兵图鉴
	if equips.Weaponry.HandbookIds != nil {
		handbookId = append(handbookId, equips.Weaponry.HandbookIds...)
	}

	//坐骑图鉴
	if equips.Mount.HandbookIds != nil {
		handbookId = append(handbookId, equips.Mount.HandbookIds...)
	}

	//背饰图鉴
	if equips.Brace.HandbookIds != nil {
		handbookId = append(handbookId, equips.Brace.HandbookIds...)
	}

	//时装
	fashion := GetPlayerFashion(pl.Id)
	//头饰图鉴
	if fashion.HeadWearHandbookIds != nil {
		handbookId = append(handbookId, fashion.HeadWearHandbookIds...)
	}

	//时装图鉴
	if fashion.FashionHandbookIds != nil {
		handbookId = append(handbookId, fashion.FashionHandbookIds...)
	}

	var lineUpIds []int32
	for i := 0; i < len(lineUp); i++ {
		lineUpIds = append(lineUpIds, lineUp[i].Id)
	}

	//天命
	destiny := GetPlayerDestiny(pl.Id)

	data := model.GetBattlePlayerData(pl, heroS, skillMap, handbookId, magic, equips.Enchant, equips.Succinct.LevelAward, equips.Mount, equips.Weaponry, equips.Equips, equips.Brace, fashion, destiny, lineUpIds)
	log.Debug("他人战斗数据：%v", data)
	return data
}

// 获取主角属性
func GetMainHeroByBattleHeroData(data *proto_public.BattleHeroData) *proto_public.BattleAttributeData {
	for _, v := range data.Items {
		if v.IsMainHero {
			return v.Attribute
		}
	}
	return nil
}

// 流派克制
func GetLiupaiRestrain(_stagelineup []*proto_public.CommonPlayerLineUpItemInfo, lineUp []int32, selfBattleData *proto_public.BattleHeroData) *proto_public.BattleHeroData {
	//流派克制
	otherLineup := make([]int32, 0)
	for _, v := range _stagelineup {
		otherLineup = append(otherLineup, v.Id)
	}
	restrain, basicAtkDamage, basicSkillDamage, zengyiBuffConTime := GetLineupLiupai(lineUp, otherLineup)
	if restrain {
		for _, data := range selfBattleData.Items {
			val := data.Attribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
			val += float32(basicAtkDamage)
			data.Attribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = val

			val1 := data.Attribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
			val1 += float32(basicSkillDamage)
			data.Attribute.BasicSkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = val1

			val2 := data.Attribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
			val2 += float32(zengyiBuffConTime)
			data.Attribute.ZengyiBuffContinueTime.Sumvalues[int32(proto_public.AttributeType_PerThousand)] = val2
		}
	}
	return selfBattleData
}

// 战力计算
func GetBattlePower(battleAttribute *proto_public.BattleAttributeData) int64 {
	//基础战力
	_atk := battleAttribute.Atk.Sumvalues[int32(proto_public.AttributeType_Basic)]
	_def := battleAttribute.Def.Sumvalues[int32(proto_public.AttributeType_Basic)]
	_hp := battleAttribute.Hp.Sumvalues[int32(proto_public.AttributeType_Basic)]
	_force := battleAttribute.Force.Sumvalues[int32(proto_public.AttributeType_Basic)]

	basePower := _hp*0.1 + _atk*0.4 + _def*0.3 + _force*0.2

	// 进阶加成系数
	var bound float32
	atkt := battleAttribute.Atk.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	deft := battleAttribute.Def.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	hpt := battleAttribute.Hp.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	forcet := battleAttribute.Force.Sumvalues[int32(proto_public.AttributeType_PerThousand)]

	atkf := battleAttribute.Atk.Sumvalues[int32(proto_public.AttributeType_Final)]
	deff := battleAttribute.Def.Sumvalues[int32(proto_public.AttributeType_Final)]
	hpf := battleAttribute.Hp.Sumvalues[int32(proto_public.AttributeType_Final)]
	forcef := battleAttribute.Force.Sumvalues[int32(proto_public.AttributeType_Final)]

	atkh := battleAttribute.Atk.Sumvalues[int32(proto_public.AttributeType_HandBook)]
	defh := battleAttribute.Def.Sumvalues[int32(proto_public.AttributeType_HandBook)]
	hph := battleAttribute.Hp.Sumvalues[int32(proto_public.AttributeType_HandBook)]
	forceh := battleAttribute.Force.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	bound += ((atkt + atkf + atkh + deft + deff + defh + hpt + hpf + hph + forcet + forcef + forceh) / float32(1000))

	// 伤害加成与抗性
	addDamaget := battleAttribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	addDamagef := battleAttribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_Final)]
	addDamageh := battleAttribute.AddDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	damageRest := battleAttribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	damageResf := battleAttribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_Final)]
	damageResh := battleAttribute.DamageRes.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	bound += ((addDamaget + addDamagef + addDamageh) - (damageRest + damageResf + damageResh)) / float32(1000)

	// 暴击系统
	addCritt := battleAttribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	addCritf := battleAttribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_Final)]
	addCrith := battleAttribute.AddCrit.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	ignoreCritt := battleAttribute.IgnoreCrit.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	ignoreCritf := battleAttribute.IgnoreCrit.Sumvalues[int32(proto_public.AttributeType_Final)]
	ignoreCritdh := battleAttribute.IgnoreCrit.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	bound += ((addCritt+addCritf+addCrith)*0.7 - (ignoreCritt+ignoreCritf+ignoreCritdh)*0.3) / float32(1000)

	// 暴伤系统
	critDamaget := battleAttribute.CritDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	critDamagef := battleAttribute.CritDamage.Sumvalues[int32(proto_public.AttributeType_Final)]
	critDamageh := battleAttribute.CritDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	ignoreCritDamaget := battleAttribute.IgnoreCriteDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	ignoreCritDamagef := battleAttribute.IgnoreCriteDamage.Sumvalues[int32(proto_public.AttributeType_Final)]
	ignoreCritDamageh := battleAttribute.IgnoreCriteDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	bound += ((critDamaget + critDamagef + critDamageh) - (ignoreCritDamaget + ignoreCritDamagef + ignoreCritDamageh)) / float32(1500)

	// 闪避系统
	addDodget := battleAttribute.AddDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	addDodgef := battleAttribute.AddDodge.Sumvalues[int32(proto_public.AttributeType_Final)]
	addDodgeh := battleAttribute.AddDodge.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	ignoreDodget := battleAttribute.IgnoreDodge.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	ignoreDodgef := battleAttribute.IgnoreDodge.Sumvalues[int32(proto_public.AttributeType_Final)]
	ignoreDodgeh := battleAttribute.IgnoreDodge.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	bound += (addDodget + addDodgef + addDodgeh + ignoreDodget + ignoreDodgef + ignoreDodgeh) / float32(2000)

	// 攻击类型加成
	addAtkSpeedt := battleAttribute.AddAtkSpeed.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	addAtkSpeedf := battleAttribute.AddAtkSpeed.Sumvalues[int32(proto_public.AttributeType_Final)]
	addAtkSpeedh := battleAttribute.AddAtkSpeed.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	normalAtkDamaget := battleAttribute.NormalAtkDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	normalAtkDamagef := battleAttribute.NormalAtkDamage.Sumvalues[int32(proto_public.AttributeType_Final)]
	normalAtkDamageh := battleAttribute.NormalAtkDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	skillDamaget := battleAttribute.SkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	skillDamagef := battleAttribute.SkillDamage.Sumvalues[int32(proto_public.AttributeType_Final)]
	skillDamageh := battleAttribute.SkillDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	contDamaget := battleAttribute.ContDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	contDamagef := battleAttribute.ContDamage.Sumvalues[int32(proto_public.AttributeType_Final)]
	contDamageh := battleAttribute.ContDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	basicAttackDamaget := battleAttribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	basicAttackDamagef := battleAttribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_Final)]
	basicAttackDamageh := battleAttribute.BasicAttackDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	basicSkillDamaget := battleAttribute.BasicSkillDamage.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	basicSkillDamagef := battleAttribute.BasicSkillDamage.Sumvalues[int32(proto_public.AttributeType_Final)]
	basicSkillDamageh := battleAttribute.BasicSkillDamage.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	bound += (addAtkSpeedt + addAtkSpeedf + addAtkSpeedh +
		normalAtkDamaget + normalAtkDamagef + normalAtkDamageh +
		skillDamaget + skillDamagef + skillDamageh +
		contDamaget + contDamagef + contDamageh + basicAttackDamaget + basicAttackDamagef +
		basicAttackDamageh + basicSkillDamaget + basicSkillDamagef + basicSkillDamageh) / float32(2500)

	// 伤害抗性减免
	normalAtkDamageRest := battleAttribute.NormalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	normalAtkDamageResf := battleAttribute.NormalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_Final)]
	normalAtkDamageResh := battleAttribute.NormalAtkDamageRes.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	skillDamageRest := battleAttribute.SkillDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	skillDamageResf := battleAttribute.SkillDamageRes.Sumvalues[int32(proto_public.AttributeType_Final)]
	skillDamageResh := battleAttribute.SkillDamageRes.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	contDamageRest := battleAttribute.ContDamageRes.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	contDamageResf := battleAttribute.ContDamageRes.Sumvalues[int32(proto_public.AttributeType_Final)]
	contDamageResh := battleAttribute.ContDamageRes.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	bound -= (normalAtkDamageRest + normalAtkDamageResf + normalAtkDamageResh +
		skillDamageRest + skillDamageResf + skillDamageResh +
		contDamageRest + contDamageResf + contDamageResh) / float32(3000)

	//特殊属性战力计算
	angerp := battleAttribute.Anger.Sumvalues[int32(proto_public.AttributeType_PerThousand)]
	angerf := battleAttribute.Anger.Sumvalues[int32(proto_public.AttributeType_Final)]
	angerh := battleAttribute.Anger.Sumvalues[int32(proto_public.AttributeType_HandBook)]

	specialCombat := (angerp + angerf + angerh) * 0.15
	power := basePower*(1+bound) + specialCombat
	log.Debug("战力计算为: %v", power)
	return int64(power)
}

// 通用玩家战斗返回
func BattleBackPlayer(ctx IPlayer, pl *model.Player, port *proto_game.C2SChallengeBattleReport, award []*proto_public.Item) {
	res := new(proto_game.S2CBattleSettle)
	res.WinId = port.WinId
	res.Items = make(map[int64]*proto_game.ChallengeBattlePlayerReportBack)
	for _, v := range port.Items {
		res.Items[v.PlayerId] = &proto_game.ChallengeBattlePlayerReportBack{
			Items:  v.Items,
			IsLeft: v.PlayerId == pl.Id,
		}

		if v.PlayerId < define.PlayerIdBase {
			res.Items[v.PlayerId].Hero = ToCommonPlayerByRobot(v.PlayerId)
		} else {
			res.Items[v.PlayerId].Hero = GetPlayerInfo(v.PlayerId).ToCommonPlayer()
		}
	}
	if award != nil {
		res.Awards = award
	}
	ctx.Send(res)
}
