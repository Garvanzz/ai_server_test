package task

import (
	"xfx/core/config"
	"xfx/core/define"
	"xfx/core/model"
)

func loadTaskFromConfig(pl *model.Player, tp int32) map[int32]*model.Task {
	ret := make(map[int32]*model.Task)
	taskConfs := config.Task.All()

	if tp == define.TaskTypeAchieve {
		current := getBucket(pl, define.TaskTypeAchieve)
		if len(current) == 0 {
			id := firstAchieveTaskID(taskConfs)
			if id > 0 {
				conf := taskConfs[int64(id)]
				ret[id] = buildTaskState(pl, conf)
			}
			return ret
		}

		ids := sortedIDs(current)
		for _, id := range ids {
			conf, ok := taskConfs[int64(id)]
			if !ok || conf.BackTask == 0 {
				continue
			}
			nextConf, ok := taskConfs[int64(conf.BackTask)]
			if !ok {
				continue
			}
			ret[nextConf.Id] = buildTaskState(pl, nextConf)
			return ret
		}
		return ret
	}

	for _, id := range sortedTaskConfigIDs(taskConfs) {
		taskConf := taskConfs[int64(id)]
		if taskConf.Type == define.TaskTypeMain {
			heroId := int32(pl.GetProp(define.PlayerPropHeroId))
			heroData := pl.Hero.Hero[heroId]
			if heroData != nil && len(taskConf.Param) > 0 && heroData.Stage == taskConf.Param[0] {
				ret[id] = buildTaskState(pl, taskConf)
			}
			continue
		}

		if taskConf.Type == tp {
			ret[id] = buildTaskState(pl, taskConf)
		}
	}

	return ret
}
