package task

import (
	"sort"

	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
)

// setTaskInfo updates the raw progress counter for (taskType, extraCondition).
// If accumulate is true the amount is added; otherwise only a higher value replaces.
// Returns the new value and whether it actually changed.
func setTaskInfo(pl *model.Player, taskType, extraCondition, amount int32, accumulate bool) (int32, bool) {
	ensureTaskData(pl)
	m, ok := pl.Task.Progress[taskType]
	if !ok {
		m = make(map[int32]int32)
	}
	old := m[extraCondition]
	if accumulate {
		m[extraCondition] += amount
	} else if m[extraCondition] < amount {
		m[extraCondition] = amount
	}
	pl.Task.Progress[taskType] = m
	return m[extraCondition], m[extraCondition] != old
}

// getTaskProgress returns the raw accumulated progress for (taskType, extraCondition).
func getTaskProgress(pl *model.Player, taskType, extraCondition int32) int32 {
	ensureTaskData(pl)
	m, ok := pl.Task.Progress[taskType]
	if !ok {
		return 0
	}
	return m[extraCondition]
}

// getVisibleTaskProgress returns displayed progress (raw - initialProcess, floored at 0).
func getVisibleTaskProgress(pl *model.Player, task *model.Task) int32 {
	v := getTaskProgress(pl, task.TaskType, task.ExtraCondition) - task.InitialProcess
	if v < 0 {
		return 0
	}
	return v
}

// isTaskCompletable reports whether the task has reached its completion threshold.
func isTaskCompletable(pl *model.Player, task *model.Task) bool {
	return getVisibleTaskProgress(pl, task) >= task.Condition
}

// buildTaskState creates a fresh model.Task from a config row, snapshotting the
// player's current progress as the baseline (InitialProcess) when Reset is true.
func buildTaskState(pl *model.Player, taskConf conf.Task) *model.Task {
	seedProgressByTaskType(pl, taskConf.TaskType)
	initProcess := int32(0)
	if taskConf.Reset {
		initProcess = getTaskProgress(pl, taskConf.TaskType, taskConf.Condition2)
	}
	return &model.Task{
		Id:             taskConf.Id,
		InitialProcess: initProcess,
		TaskType:       taskConf.TaskType,
		Condition:      taskConf.Condition1,
		ExtraCondition: taskConf.Condition2,
	}
}

// seedProgressByTaskType pre-populates progress counters that must be derived
// from the player's current state rather than being driven by Dispatch events.
func seedProgressByTaskType(pl *model.Player, taskType int32) {
	switch taskType {
	case define.TaskHeroLevel:
		heroId := int32(pl.GetProp(define.PlayerPropHeroId))
		if heroData := pl.Hero.Hero[heroId]; heroData != nil {
			_, _ = setTaskInfo(pl, taskType, 0, heroData.Level, false)
		}
	}
}

// sortedTaskConfigIDs returns task config keys in ascending order.
func sortedTaskConfigIDs(m map[int64]conf.Task) []int32 {
	ids := make([]int32, 0, len(m))
	for id := range m {
		ids = append(ids, int32(id))
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// sortedIDs returns model.Task map keys in ascending order.
func sortedIDs(m map[int32]*model.Task) []int32 {
	ids := make([]int32, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
