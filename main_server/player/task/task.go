package task

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/proto/proto_task"
)

// Dispatch 任务分发
func Dispatch(ctx global.IPlayer, pl *model.Player, taskType int32, taskCount int32, extraCondition int32, accumulate bool) {
	resetTask(ctx, pl, taskType)

	limit := define.TaskCompleteLimit[taskType]
	if limit != 0 && pl.Task.TaskLimit[taskType] >= limit {
		return
	}

	taskValue, changed := setTaskInfo(pl, taskType, extraCondition, taskCount, accumulate)
	if !changed {
		return
	}
	pl.Task.TaskLimit[taskType]++

	pushTasks := buildIncrementalPush(pl, taskType, extraCondition, taskValue)
	if len(pushTasks.Groups) > 0 {
		ctx.Send(pushTasks)
	}
}

func IsMainTaskRewarded(pl *model.Player, taskID int32) bool {
	ensureTaskData(pl)
	t, ok := getBucket(pl, define.TaskTypeMain)[taskID]
	if !ok {
		return false
	}
	return t.ReceiveAward
}

func ReqTaskData(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetTasks) {
	resetTask(ctx, pl, 0)
	ctx.Send(buildTaskSnapshot(pl))
}
