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
			if v.ReceiveAward {
				continue
			}
			if !isTaskCompletable(pl, v) {
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
		if !ok || task.ReceiveAward {
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}

		if !isTaskCompletable(pl, task) {
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

	if taskActivityConf.Type == define.TaskActivityTypeDaily {
		resetTask(ctx, pl, define.TaskTypeDaily)
	} else if taskActivityConf.Type == define.TaskActivityTypeGuild {
		resetTask(ctx, pl, define.TaskTypeGuild)
	}

	claimType := claimTypeDaily
	pointType := int32(define.TaskActivityTypeDaily)
	if taskActivityConf.Type == define.TaskActivityTypeGuild {
		claimType = claimTypeGuild
		pointType = int32(define.TaskActivityTypeGuild)
	}

	if taskActivityConf.Value > getPoint(pl, pointType) {
		ctx.Send(&proto_task.S2CGetActivePointReward{Succ: false})
		return
	}

	claims := getClaimMap(pl, claimType)
	if claims[req.RewardId] {
		ctx.Send(&proto_task.S2CGetActivePointReward{Succ: false})
		return
	}
	claims[req.RewardId] = true
	setClaimMap(pl, claimType, claims)

	if len(taskActivityConf.Reward) > 0 {
		bag.AddAward(ctx, pl, taskActivityConf.Reward, true)
	}

	pushTasks := &proto_task.PushTask{}
	applyBucketPush(pushTasks, pointTypeToBucketType(pointType), pl)
	if len(pushTasks.Groups) > 0 {
		ctx.Send(pushTasks)
	}
	ctx.Send(&proto_task.S2CGetActivePointReward{Succ: true})
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

func addActivityValue(ctx global.IPlayer, pl *model.Player, taskType int32, value int32) {
	if taskType == define.TaskTypeDaily {
		setPoint(pl, define.TaskActivityTypeDaily, getPoint(pl, define.TaskActivityTypeDaily)+value)
		return
	}
	if taskType == define.TaskTypeGuild {
		setPoint(pl, define.TaskActivityTypeGuild, getPoint(pl, define.TaskActivityTypeGuild)+value)
		return
	}
	if taskType == define.TaskTypePassportDaily || taskType == define.TaskTypePassportWeek || taskType == define.TaskTypePassportSeason {
		addPassportScore(ctx, pl, value)
	}
}

func applyBucketPush(pushTasks *proto_task.PushTask, taskType int32, pl *model.Player) {
	if taskType == define.TaskTypeAchieve && allRewarded(getBucket(pl, define.TaskTypeAchieve)) {
		setBucket(pl, define.TaskTypeAchieve, loadTaskFromConfig(pl, define.TaskTypeAchieve))
	}
	if taskType == define.TaskTypeMain && allRewarded(getBucket(pl, define.TaskTypeMain)) {
		setBucket(pl, define.TaskTypeMain, loadTaskFromConfig(pl, define.TaskTypeMain))
	}

	meta, ok := findTaskGroupMeta(taskType)
	if !ok {
		return
	}
	pushTasks.Groups = []*proto_task.TaskGroup{buildTaskGroup(pl, meta)}
}
