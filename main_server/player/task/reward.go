package task

import (
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_task"
)

func ReqReceiveReward(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetReward) {
	resetTask(ctx, pl, req.BucketType)

	tasks := getBucket(pl, req.BucketType)
	if tasks == nil {
		log.Error("ReqReceiveReward type error:%v", req.BucketType)
		ctx.Send(&proto_task.S2CGetReward{Succ: false})
		return
	}

	awards := make([]conf.ItemE, 0)
	rewarded := false
	taskConfs := config.Task.All()
	if req.TaskId == 0 {
		for id, v := range tasks {
			if v.ReceiveAward || !isTaskCompletable(pl, v) {
				continue
			}
			taskConf, ok := taskConfs[int64(id)]
			if !ok {
				continue
			}
			awards = append(awards, taskConf.Reward...)
			v.ReceiveAward = true
			rewarded = true
			addActivityValue(ctx, pl, req.BucketType, taskConf.ActivityValue)
		}
	} else {
		task, ok := tasks[req.TaskId]
		if !ok || task.ReceiveAward || !isTaskCompletable(pl, task) {
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}
		taskConf, ok := taskConfs[int64(req.TaskId)]
		if !ok {
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}
		task.ReceiveAward = true
		awards = append(awards, taskConf.Reward...)
		rewarded = true
		addActivityValue(ctx, pl, req.BucketType, taskConf.ActivityValue)
	}

	if !rewarded {
		ctx.Send(&proto_task.S2CGetReward{Succ: false, TaskId: req.TaskId})
		return
	}
	if len(awards) > 0 {
		bag.AddAward(ctx, pl, awards, true)
		internal.PushPlayerData(ctx, pl)
	}
	pushTasks := &proto_task.PushTask{}
	applyBucketPush(pushTasks, req.BucketType, pl)
	if len(pushTasks.Groups) > 0 {
		ctx.Send(pushTasks)
	}
	ctx.Send(&proto_task.S2CGetReward{Succ: true, TaskId: req.TaskId})
}

func ReqReceiveActivePointReward(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetActivePointReward) {
	taskActivityConf, ok := config.TaskActivity.Find(int64(req.RewardId))
	if !ok {
		ctx.Send(&proto_task.S2CGetActivePointReward{Succ: false})
		return
	}

	// Find the bucket policy whose local point type matches the activity reward type.
	var matchPolicy bucketPolicy
	found := false
	for _, p := range bucketPolicies {
		if p.pointType != 0 && p.pointType == taskActivityConf.Type {
			matchPolicy = p
			found = true
			resetTask(ctx, pl, p.bucketType)
			break
		}
	}
	if !found {
		ctx.Send(&proto_task.S2CGetActivePointReward{Succ: false})
		return
	}

	if taskActivityConf.Value > getPoint(pl, matchPolicy.pointType) {
		ctx.Send(&proto_task.S2CGetActivePointReward{Succ: false})
		return
	}
	claims := getClaimMap(pl, matchPolicy.claimType)
	if claims[req.RewardId] {
		ctx.Send(&proto_task.S2CGetActivePointReward{Succ: false})
		return
	}
	claims[req.RewardId] = true
	setClaimMap(pl, matchPolicy.claimType, claims)

	if len(taskActivityConf.Reward) > 0 {
		bag.AddAward(ctx, pl, taskActivityConf.Reward, true)
	}
	pushTasks := &proto_task.PushTask{}
	applyBucketPush(pushTasks, matchPolicy.bucketType, pl)
	if len(pushTasks.Groups) > 0 {
		ctx.Send(pushTasks)
	}
	ctx.Send(&proto_task.S2CGetActivePointReward{Succ: true})
}

// addActivityValue credits activity value for a completed task.
// Buckets with a local pointType accumulate points; passport buckets fire an external event.
func addActivityValue(ctx global.IPlayer, pl *model.Player, taskType int32, value int32) {
	if value == 0 {
		return
	}
	p, ok := findPolicy(taskType)
	if !ok {
		return
	}
	if p.pointType != 0 {
		setPoint(pl, p.pointType, getPoint(pl, p.pointType)+value)
		return
	}
	if p.notifyPassport {
		addPassportScore(ctx, pl, value)
	}
}

func addPassportScore(ctx global.IPlayer, pl *model.Player, score int32) {
	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":         "passport_task_score",
		"player":      pl.ToContext(),
		"score":       score,
		"playermodel": pl,
		"IPlayer":     ctx,
	})
}

// applyBucketPush finalises a reward and pushes the updated bucket to the client.
// For auto-advance buckets (achieve, main) it reloads the next task set when all are done.
func applyBucketPush(pushTasks *proto_task.PushTask, taskType int32, pl *model.Player) {
	p, ok := findPolicy(taskType)
	if !ok {
		return
	}
	if p.autoAdvance && allRewarded(getBucket(pl, taskType)) {
		setBucket(pl, taskType, loadTaskFromConfig(pl, taskType))
	}
	pushTasks.Groups = []*proto_task.TaskGroup{buildTaskGroup(pl, p)}
}
