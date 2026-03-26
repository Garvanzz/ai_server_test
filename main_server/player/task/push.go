package task

import (
	"sort"

	"xfx/core/model"
	"xfx/proto/proto_task"
)

// buildTaskSnapshot builds a full S2CGetTasks snapshot (all buckets, all tasks).
func buildTaskSnapshot(pl *model.Player) *proto_task.S2CGetTasks {
	groups := make([]*proto_task.TaskGroup, 0, len(bucketPolicies))
	for _, p := range bucketPolicies {
		groups = append(groups, buildTaskGroup(pl, p))
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].BucketType < groups[j].BucketType })
	return &proto_task.S2CGetTasks{Groups: groups}
}

// buildTaskGroup converts one bucket to its proto representation using the policy
// to decide whether to include claim records and activity points.
func buildTaskGroup(pl *model.Player, p bucketPolicy) *proto_task.TaskGroup {
	group := &proto_task.TaskGroup{
		BucketType: p.bucketType,
		Tasks:      taskToProto(getBucket(pl, p.bucketType), pl),
	}
	if p.claimType != 0 {
		group.ClaimRecord = copyClaimMap(getClaimMap(pl, p.claimType))
	}
	if p.pointType != 0 {
		group.Point = getPoint(pl, p.pointType)
	}
	return group
}

// buildIncrementalPush scans every bucket and returns a PushTask containing only
// the buckets that have tasks matching (taskType, extraCondition).
func buildIncrementalPush(pl *model.Player, taskType, extraCondition, taskValue int32) *proto_task.PushTask {
	groups := make([]*proto_task.TaskGroup, 0)
	for _, p := range bucketPolicies {
		changed := changedTaskStates(getBucket(pl, p.bucketType), taskType, extraCondition, taskValue)
		if len(changed) == 0 {
			continue
		}
		group := &proto_task.TaskGroup{
			BucketType: p.bucketType,
			Tasks:      changed,
		}
		if p.claimType != 0 {
			group.ClaimRecord = copyClaimMap(getClaimMap(pl, p.claimType))
		}
		if p.pointType != 0 {
			group.Point = getPoint(pl, p.pointType)
		}
		groups = append(groups, group)
	}
	return &proto_task.PushTask{Groups: groups}
}

// taskToProto converts a bucket map to its proto TaskState representation.
func taskToProto(m map[int32]*model.Task, pl *model.Player) map[int32]*proto_task.TaskState {
	ret := make(map[int32]*proto_task.TaskState, len(m))
	for k, v := range m {
		ret[k] = &proto_task.TaskState{
			Progress: getVisibleTaskProgress(pl, v),
			Rewarded: v.ReceiveAward,
		}
	}
	return ret
}

// changedTaskStates returns the incremental TaskState entries for tasks in bucket m
// whose (TaskType, ExtraCondition) matches the given pair and are not yet rewarded.
// Returns nil when nothing changed (avoids empty group pushes).
func changedTaskStates(m map[int32]*model.Task, taskType, extraCondition, taskValue int32) map[int32]*proto_task.TaskState {
	ret := make(map[int32]*proto_task.TaskState)
	for k, v := range m {
		if v.TaskType != taskType || v.ExtraCondition != extraCondition || v.ReceiveAward {
			continue
		}
		progress := taskValue - v.InitialProcess
		if progress < 0 {
			progress = 0
		}
		ret[k] = &proto_task.TaskState{Progress: progress}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}
