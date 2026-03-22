package task

import (
	"sort"

	"xfx/core/define"
	"xfx/core/model"
	"xfx/proto/proto_task"
)

type taskGroupMeta struct {
	bucketType int32
	claimType  int32
	pointType  int32
}

var taskGroupMetas = []taskGroupMeta{
	{bucketType: define.TaskTypeDaily, claimType: claimTypeDaily, pointType: define.TaskActivityTypeDaily},
	{bucketType: define.TaskTypeWeek, claimType: claimTypeWeek},
	{bucketType: define.TaskTypeMonth},
	{bucketType: define.TaskTypeAchieve},
	{bucketType: define.TaskTypeMain},
	{bucketType: define.TaskTypeGuild, claimType: claimTypeGuild, pointType: define.TaskActivityTypeGuild},
	{bucketType: define.TaskTypeDrawHeroRank},
	{bucketType: define.TaskTypeTheCompetitionRank},
	{bucketType: define.TaskTypePassportDaily},
	{bucketType: define.TaskTypePassportWeek},
	{bucketType: define.TaskTypePassportSeason},
}

func buildTaskSnapshot(pl *model.Player) *proto_task.S2CGetTasks {
	return &proto_task.S2CGetTasks{Groups: buildAllTaskGroups(pl)}
}

func buildAllTaskGroups(pl *model.Player) []*proto_task.TaskGroup {
	groups := make([]*proto_task.TaskGroup, 0, len(taskGroupMetas))
	for _, meta := range taskGroupMetas {
		groups = append(groups, buildTaskGroup(pl, meta))
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].BucketType < groups[j].BucketType })
	return groups
}

func buildTaskGroup(pl *model.Player, meta taskGroupMeta) *proto_task.TaskGroup {
	group := &proto_task.TaskGroup{
		BucketType: meta.bucketType,
		Tasks:      taskToProto(getBucket(pl, meta.bucketType), pl),
	}
	if meta.claimType != 0 {
		group.ClaimRecord = copyClaimMap(getClaimMap(pl, meta.claimType))
	}
	if meta.pointType != 0 {
		group.Point = getPoint(pl, meta.pointType)
	}
	return group
}

func buildTaskGroupPush(pl *model.Player, bucketType int32) *proto_task.PushTask {
	meta, ok := findTaskGroupMeta(bucketType)
	if !ok {
		return &proto_task.PushTask{}
	}
	return &proto_task.PushTask{Groups: []*proto_task.TaskGroup{buildTaskGroup(pl, meta)}}
}

func buildIncrementalPush(pl *model.Player, taskType, extraCondition, taskValue int32) *proto_task.PushTask {
	groups := make([]*proto_task.TaskGroup, 0, len(taskGroupMetas))
	for _, meta := range taskGroupMetas {
		changed := pushTaskChange(getBucket(pl, meta.bucketType), taskType, extraCondition, taskValue)
		if len(changed) == 0 {
			continue
		}
		group := &proto_task.TaskGroup{
			BucketType: meta.bucketType,
			Tasks:      changed,
		}
		if meta.claimType != 0 {
			group.ClaimRecord = copyClaimMap(getClaimMap(pl, meta.claimType))
		}
		if meta.pointType != 0 {
			group.Point = getPoint(pl, meta.pointType)
		}
		groups = append(groups, group)
	}
	return &proto_task.PushTask{Groups: groups}
}

func pointTypeToBucketType(pointType int32) int32 {
	switch pointType {
	case define.TaskActivityTypeDaily:
		return define.TaskTypeDaily
	case define.TaskActivityTypeGuild:
		return define.TaskTypeGuild
	default:
		return 0
	}
}

func findTaskGroupMeta(bucketType int32) (taskGroupMeta, bool) {
	for _, meta := range taskGroupMetas {
		if meta.bucketType == bucketType {
			return meta, true
		}
	}
	return taskGroupMeta{}, false
}
