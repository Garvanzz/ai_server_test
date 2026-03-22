package task

import (
	"sort"

	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/proto/proto_task"
)

func pushTaskChange(m map[int32]*model.Task, taskType, extraCondition, taskValue int32) map[int32]*proto_task.TaskState {
	ret := make(map[int32]*proto_task.TaskState)
	for k, v := range m {
		if taskType == v.TaskType && v.ExtraCondition == extraCondition && !v.ReceiveAward {
			progress := taskValue - v.InitialProcess
			if progress < 0 {
				progress = 0
			}
			ret[k] = &proto_task.TaskState{Progress: progress, Rewarded: false}
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func taskToProto(m map[int32]*model.Task, pl *model.Player) map[int32]*proto_task.TaskState {
	ret := make(map[int32]*proto_task.TaskState)
	for k, v := range m {
		progress := getVisibleTaskProgress(pl, v)
		ret[k] = &proto_task.TaskState{Progress: progress, Rewarded: v.ReceiveAward}
	}
	return ret
}

func setTaskInfo(pl *model.Player, taskType int32, extraCondition int32, amount int32, accumulate bool) (int32, bool) {
	ensureTaskData(pl)
	m, ok := pl.Task.Progress[taskType]
	if !ok {
		m = make(map[int32]int32)
	}
	oldValue := m[extraCondition]

	if accumulate {
		m[extraCondition] += amount
	} else if m[extraCondition] < amount {
		m[extraCondition] = amount
	}

	pl.Task.Progress[taskType] = m
	return m[extraCondition], m[extraCondition] != oldValue
}

func getTaskInitProcess(pl *model.Player, taskConf conf.Task) int32 {
	if taskConf.Reset {
		return getTaskProgress(pl, taskConf.TaskType, taskConf.Condition2)
	}
	return 0
}

func getTaskProgress(pl *model.Player, taskType int32, extraCondition int32) int32 {
	ensureTaskData(pl)
	m, ok := pl.Task.Progress[taskType]
	if !ok {
		return 0
	}
	return m[extraCondition]
}

func getVisibleTaskProgress(pl *model.Player, task *model.Task) int32 {
	progress := getTaskProgress(pl, task.TaskType, task.ExtraCondition) - task.InitialProcess
	if progress < 0 {
		return 0
	}
	return progress
}

func isTaskCompletable(pl *model.Player, task *model.Task) bool {
	return getVisibleTaskProgress(pl, task) >= task.Condition
}

func buildTaskState(pl *model.Player, taskConf conf.Task) *model.Task {
	seedProgressByTaskType(pl, taskConf.TaskType)
	return &model.Task{
		Id:             taskConf.Id,
		InitialProcess: getTaskInitProcess(pl, taskConf),
		TaskType:       taskConf.TaskType,
		Condition:      taskConf.Condition1,
		ExtraCondition: taskConf.Condition2,
	}
}

func seedProgressByTaskType(pl *model.Player, taskType int32) {
	switch taskType {
	case define.TaskHeroLevel:
		heroId := int32(pl.GetProp(define.PlayerPropHeroId))
		heroData := pl.Hero.Hero[heroId]
		if heroData != nil {
			_, _ = setTaskInfo(pl, taskType, 0, heroData.Level, false)
		}
	}
}

func sortedTaskConfigIDs(m map[int64]conf.Task) []int32 {
	ids := make([]int32, 0, len(m))
	for id := range m {
		ids = append(ids, int32(id))
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func sortedIDs(m map[int32]*model.Task) []int32 {
	ids := make([]int32, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func firstAchieveTaskID(taskConfs map[int64]conf.Task) int32 {
	ids := sortedTaskConfigIDs(taskConfs)
	for _, id := range ids {
		c := taskConfs[int64(id)]
		if c.Type == define.TaskTypeAchieve && c.FrontTask == 0 {
			return id
		}
	}
	return 0
}

func getBucket(pl *model.Player, taskType int32) map[int32]*model.Task {
	ensureTaskData(pl)
	return pl.Task.Buckets[taskType]
}

func setBucket(pl *model.Player, taskType int32, tasks map[int32]*model.Task) {
	ensureTaskData(pl)
	pl.Task.Buckets[taskType] = tasks
}

func getPoint(pl *model.Player, pointType int32) int32 {
	ensureTaskData(pl)
	return pl.Task.Points[pointType]
}

func setPoint(pl *model.Player, pointType, value int32) {
	ensureTaskData(pl)
	pl.Task.Points[pointType] = value
}

func getClaimMap(pl *model.Player, claimType int32) map[int32]bool {
	ensureTaskData(pl)
	m, ok := pl.Task.ClaimRecord[claimType]
	if !ok {
		m = make(map[int32]bool)
		pl.Task.ClaimRecord[claimType] = m
	}
	return m
}

func setClaimMap(pl *model.Player, claimType int32, m map[int32]bool) {
	ensureTaskData(pl)
	pl.Task.ClaimRecord[claimType] = m
}

func copyClaimMap(src map[int32]bool) map[int32]bool {
	dst := make(map[int32]bool, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func allRewarded(tasks map[int32]*model.Task) bool {
	if len(tasks) == 0 {
		return false
	}
	for _, t := range tasks {
		if !t.ReceiveAward {
			return false
		}
	}
	return true
}
