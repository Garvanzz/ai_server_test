package internal

import (
	"strconv"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
)

// 判定功能开启
func FuncOpenJudgeLogic(conf conf2.FunctionOpen, pl *model.Player) bool {
	cond := conf.Condition
	for i, v := range cond {
		switch v {
		case define.FuncOpenCondition_None:
			return true
		case define.FuncOpenCondition_MainHeroLevel:
			level, _ := strconv.ParseInt(conf.Param[i], 16, 64)
			return pl.Hero.Hero[int32(pl.GetProp(define.PlayerPropHeroId))].Level >= int32(level)
		case define.FuncOpenCondition_MainHeroStage:
			stage, _ := strconv.ParseInt(conf.Param[i], 16, 64)
			return pl.Hero.Hero[int32(pl.GetProp(define.PlayerPropHeroId))].Stage >= int32(stage)
		case define.FuncOpenCondition_MainHeroStar:
			star, _ := strconv.ParseInt(conf.Param[i], 16, 64)
			return pl.Hero.Hero[int32(pl.GetProp(define.PlayerPropHeroId))].Star >= int32(star)
		case define.FuncOpenCondition_Forward:
			var conf conf2.FunctionOpen
			confs := config.FunctionOpen.All()
			for _, v := range confs {
				if v.Type == v.Param[i] {
					conf = v
					break
				}
			}

			if conf.Id <= 0 {
				return false
			}
			return FuncOpenJudgeLogic(conf, pl)
		}
	}
	return false
}
