package internal

import (
	"xfx/core/config"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/proto/proto_public"
)

// 获取战斗个人数据
func GetBattleSelfPlayerData(pl *model.Player, lineUp []int32) *proto_public.BattleHeroData {

	//图鉴
	handbookId := make([]int32, 0)

	//角色图鉴
	if pl.Handbook.HandbookOption.GetId != nil {
		handbookId = append(handbookId, pl.Handbook.HandbookOption.GetId...)
	}

	//神兵图鉴
	if pl.Equip.Weaponry.HandbookIds != nil {
		handbookId = append(handbookId, pl.Equip.Weaponry.HandbookIds...)
	}

	//坐骑图鉴
	if pl.Equip.Mount.HandbookIds != nil {
		handbookId = append(handbookId, pl.Equip.Mount.HandbookIds...)
	}

	//背饰图鉴
	if pl.Equip.Brace.HandbookIds != nil {
		handbookId = append(handbookId, pl.Equip.Brace.HandbookIds...)
	}

	//头饰图鉴
	if pl.Fashion.HeadWearHandbookIds != nil {
		handbookId = append(handbookId, pl.Fashion.HeadWearHandbookIds...)
	}

	//时装图鉴
	if pl.Fashion.FashionHandbookIds != nil {
		handbookId = append(handbookId, pl.Fashion.FashionHandbookIds...)
	}

	data := model.GetBattlePlayerData(pl.ToContext(), pl.Hero, pl.Skill, handbookId, pl.Magic, pl.Equip.Enchant, pl.Equip.Succinct.LevelAward, pl.Equip.Mount, pl.Equip.Weaponry, pl.Equip.Equips, pl.Equip.Brace, pl.Fashion, pl.Destiny, lineUp)
	log.Debug("个人战斗数据：%v", data)
	return data
}

// 获取战斗数据-他人
func GetBattleOtherPlayerData(pl *model.PlayerInfo, lineUp []*proto_public.CommonPlayerLineUpItemInfo) *proto_public.BattleHeroData {
	heroS := global.GetPlayerHero(pl.Id)
	skillMap := global.GetPlayerSkill(pl.Id)
	magic := global.GetPlayerMagic(pl.Id)
	equips := global.GetPlayerEquip(pl.Id)
	//时装
	fashion := global.GetPlayerFashion(pl.Id)
	var lineUpIds []int32
	for i := 0; i < len(lineUp); i++ {
		lineUpIds = append(lineUpIds, lineUp[i].Id)
	}

	//图鉴
	handbookId := make([]int32, 0)
	//角色图鉴
	phandbook := global.GetPlayerHandbook(pl.Id)
	if phandbook.HandbookOption.GetId != nil {
		handbookId = append(handbookId, phandbook.HandbookOption.GetId...)
	}

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

	//头饰图鉴
	if fashion.HeadWearHandbookIds != nil {
		handbookId = append(handbookId, fashion.HeadWearHandbookIds...)
	}

	//时装图鉴
	if fashion.FashionHandbookIds != nil {
		handbookId = append(handbookId, fashion.FashionHandbookIds...)
	}

	//天命
	destiny := global.GetPlayerDestiny(pl.Id)

	data := model.GetBattlePlayerData(pl.ToToContext(), heroS, skillMap, handbookId, magic, equips.Enchant, equips.Succinct.LevelAward, equips.Mount, equips.Weaponry, equips.Equips, equips.Brace, fashion, destiny, lineUpIds)
	log.Debug("他人战斗数据：%v", data)
	return data
}

// 获取机器人
func GetRobotBattleData(id int64) *proto_public.BattleHeroData {
	data := model.GetBattleRobotPlayerData(id)
	log.Debug("机器人战斗数据：%v", data)
	return data
}

// 获取机器人阵容
func GetRobotLineUp(id int64) []*proto_public.CommonPlayerLineUpItemInfo {
	lineup := make([]*proto_public.CommonPlayerLineUpItemInfo, 0)
	robotGroups := config.RobotGroup.All()
	robotGroup, ok := robotGroups[id]
	if !ok {
		return lineup
	}

	robots := config.Robot.All()
	for _, v := range robotGroup.RobotId {
		_lineup := new(proto_public.CommonPlayerLineUpItemInfo)
		robot := robots[int64(v)]
		_lineup.Id = v
		_lineup.Level = robot.Level
		_lineup.Star = robot.Star
		lineup = append(lineup, _lineup)
	}

	return lineup
}
