package task

import (
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
)

// loadTaskFromConfig builds the initial task map for a bucket from the game config.
// Achievement chain and main-task loading have custom logic; all other bucket
// types are loaded with a simple type-filter pass.
func loadTaskFromConfig(pl *model.Player, tp int32) map[int32]*model.Task {
	taskConfs := config.Task.All()
	switch tp {
	case define.TaskTypeAchieve:
		return loadAchieveTasks(pl, taskConfs)
	case define.TaskTypeMain:
		return loadMainTasks(pl, taskConfs)
	default:
		ret := make(map[int32]*model.Task)
		for _, id := range sortedTaskConfigIDs(taskConfs) {
			c := taskConfs[int64(id)]
			if c.Type == tp {
				ret[id] = buildTaskState(pl, c)
			}
		}
		return ret
	}
}

// loadAchieveTasks returns the next task in the achievement chain.
//   - Empty bucket → start at the root task (FrontTask == 0).
//   - Non-empty bucket → advance to the BackTask of the last task.
func loadAchieveTasks(pl *model.Player, taskConfs map[int64]conf.Task) map[int32]*model.Task {
	ret := make(map[int32]*model.Task)
	current := getBucket(pl, define.TaskTypeAchieve)

	if len(current) == 0 {
		if id := firstAchieveTaskID(taskConfs); id > 0 {
			ret[id] = buildTaskState(pl, taskConfs[int64(id)])
		}
		return ret
	}

	for _, id := range sortedIDs(current) {
		c, ok := taskConfs[int64(id)]
		if !ok || c.BackTask == 0 {
			continue
		}
		next, ok := taskConfs[int64(c.BackTask)]
		if !ok {
			continue
		}
		ret[next.Id] = buildTaskState(pl, next)
		return ret
	}
	return ret
}

// loadMainTasks returns main-story tasks whose stage requirement matches the hero.
func loadMainTasks(pl *model.Player, taskConfs map[int64]conf.Task) map[int32]*model.Task {
	ret := make(map[int32]*model.Task)
	heroId := int32(pl.GetProp(define.PlayerPropHeroId))
	heroData := pl.Hero.Hero[heroId]
	if heroData == nil {
		return ret
	}
	for _, id := range sortedTaskConfigIDs(taskConfs) {
		c := taskConfs[int64(id)]
		if c.Type != define.TaskTypeMain {
			continue
		}
		if len(c.Param) > 0 && heroData.Stage == c.Param[0] {
			ret[id] = buildTaskState(pl, c)
		}
	}
	return ret
}

// firstAchieveTaskID returns the ID of the root achievement task (no predecessor).
func firstAchieveTaskID(taskConfs map[int64]conf.Task) int32 {
	for _, id := range sortedTaskConfigIDs(taskConfs) {
		c := taskConfs[int64(id)]
		if c.Type == define.TaskTypeAchieve && c.FrontTask == 0 {
			return id
		}
	}
	return 0
}
