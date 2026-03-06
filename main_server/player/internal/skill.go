package internal

import (
	"xfx/core/config"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/proto/proto_skill"
)

// UpdateSkill 更新技能 key = id, value = 等级
func UpdateSkill(ctx global.IPlayer, pl *model.Player, Id, level int32) {
	pl.Skill.Ids[Id] = level

	ids := make(map[int32]int32)
	ids[Id] = level
	ctx.Send(&proto_skill.PushSkillInfo{SkillIds: ids})
}

// 获取角色解锁的技能
func GetHeroSkill(pl *model.Player, heroId int32) map[int32]int32 {
	heroConf, _ := config.Hero.Find(int64(heroId))
	heroItem := pl.Hero.Hero[heroId]
	unlockSkill := heroConf.SkillUnlock
	skills := make(map[int32]int32)
	for i := 0; i < len(unlockSkill); i++ {
		if unlockSkill[i][0] > heroItem.Stage {
			continue
		}

		if _, ok := pl.Skill.Ids[unlockSkill[i][1]]; ok {
			level := pl.Skill.Ids[unlockSkill[i][1]]
			skills[unlockSkill[i][1]] = level
		} else {
			skills[unlockSkill[i][1]] = 1
		}
	}
	return skills
}
