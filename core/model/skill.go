package model

import (
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/proto/proto_magic"
)

type Skill struct {
	Ids map[int32]int32
}

type Magic struct {
	LineUp []int32
	Ids    map[int32]*MagicItem
}

type MagicItem struct {
	Id    int32
	Num   int32
	Level int32
}

func ToMagicProto(opt *Magic) *proto_magic.MagicOption {
	maps := make(map[int32]*proto_magic.MagicItem)
	for k, v := range opt.Ids {
		item := new(proto_magic.MagicItem)
		item.Id = v.Id
		item.Num = v.Num
		item.Level = v.Level
		maps[k] = item
	}
	return &proto_magic.MagicOption{
		Magics: maps,
		LineUp: opt.LineUp,
	}
}

// 获取角色解锁的技能
func GetHeroSkill(pl *Hero, skill *Skill, heroId int32) map[int32]int32 {
	heroConf := config.CfgMgr.AllJson()["Hero"].(map[int64]conf2.Hero)[int64(heroId)]
	heroItem := pl.Hero[heroId]
	unlockSkill := heroConf.SkillUnlock
	skills := make(map[int32]int32)
	for i := 0; i < len(unlockSkill); i++ {
		if unlockSkill[i][0] > heroItem.Stage {
			continue
		}

		if _, ok := skill.Ids[unlockSkill[i][1]]; ok {
			level := skill.Ids[unlockSkill[i][1]]
			skills[unlockSkill[i][1]] = level
		} else {
			skills[unlockSkill[i][1]] = 1
		}
	}
	return skills
}
