package internal

import (
	"xfx/core/config"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/proto/proto_hero"
)

// SyncHeroChange 同步角色
func SyncHeroChange(ctx global.IPlayer, pl *model.Player, id int32) {
	res := &proto_hero.PushHeroChange{}
	if _, ok := pl.Hero.Hero[id]; !ok {
		return
	}
	hero := model.ToBagHeroProtoByHero(pl.Hero.Hero[id])
	res.HeroOption = hero
	ctx.Send(res)
}

// 检测角色解锁技能
func CheckHeroSkill(ctx global.IPlayer, pl *model.Player, heroId, level int32) {
	heroConf, _ := config.Hero.Find(int64(heroId))
	heroItem := pl.Hero.Hero[heroId]
	unlockSkill := heroConf.SkillUnlock
	for i := 0; i < len(unlockSkill); i++ {
		if unlockSkill[i][0] > heroItem.Stage {
			continue
		}

		UpdateSkill(ctx, pl, unlockSkill[i][1], level)
	}
}
