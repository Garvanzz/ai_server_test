package task

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/utils"
)

func resetTask(ctx global.IPlayer, pl *model.Player, taskType int32) {
	ensureTaskData(pl)
	now := utils.Now().Unix()

	if pl.Task.ResetAt[define.TaskTypeDaily] == 0 {
		initBucketsOnFirstTouch(ctx, pl, now)
		return
	}

	resetDailyTasks(pl, now, taskType)
	resetPassportDailyTasks(pl, now, taskType)
	resetGuildTasks(pl, now, taskType)
	resetWeekTasks(pl, now, taskType)
	resetPassportWeekTasks(pl, now, taskType)
	resetMonthTasks(pl, now, taskType)
	refreshActivityTaskBuckets(ctx, pl, taskType)
}

func resetDailyTasks(pl *model.Player, now int64, taskType int32) {
	if (taskType != define.TaskTypeDaily && taskType != 0) || utils.IsSameDayBySecWithHour(now, pl.Task.ResetAt[define.TaskTypeDaily], 0) {
		return
	}
	setBucket(pl, define.TaskTypeDaily, loadTaskFromConfig(pl, define.TaskTypeDaily))
	setPoint(pl, define.TaskActivityTypeDaily, 0)
	setClaimMap(pl, claimTypeDaily, make(map[int32]bool))
	pl.Task.TaskLimit = make(map[int32]int32)
	pl.Task.ResetAt[define.TaskTypeDaily] = now
}

func resetPassportDailyTasks(pl *model.Player, now int64, taskType int32) {
	if (taskType != define.TaskTypePassportDaily && taskType != 0) || utils.IsSameDayBySecWithHour(now, pl.Task.ResetAt[define.TaskTypePassportDaily], 0) {
		return
	}
	setBucket(pl, define.TaskTypePassportDaily, loadTaskFromConfig(pl, define.TaskTypePassportDaily))
	pl.Task.ResetAt[define.TaskTypePassportDaily] = now
}

func resetGuildTasks(pl *model.Player, now int64, taskType int32) {
	if (taskType != define.TaskTypeGuild && taskType != 0) || utils.IsSameDayBySecWithHour(now, pl.Task.ResetAt[define.TaskTypeGuild], 0) {
		return
	}
	setBucket(pl, define.TaskTypeGuild, loadTaskFromConfig(pl, define.TaskTypeGuild))
	setPoint(pl, define.TaskActivityTypeGuild, 0)
	setClaimMap(pl, claimTypeGuild, make(map[int32]bool))
	pl.Task.ResetAt[define.TaskTypeGuild] = now
}

func resetWeekTasks(pl *model.Player, now int64, taskType int32) {
	if (taskType != define.TaskTypeWeek && taskType != 0) || utils.IsSameWeekBySec(now, pl.Task.ResetAt[define.TaskTypeWeek]) {
		return
	}
	setClaimMap(pl, claimTypeWeek, make(map[int32]bool))
	setBucket(pl, define.TaskTypeWeek, loadTaskFromConfig(pl, define.TaskTypeWeek))
	pl.Task.ResetAt[define.TaskTypeWeek] = now
}

func resetPassportWeekTasks(pl *model.Player, now int64, taskType int32) {
	if (taskType != define.TaskTypePassportWeek && taskType != 0) || utils.IsSameWeekBySec(now, pl.Task.ResetAt[define.TaskTypePassportWeek]) {
		return
	}
	setBucket(pl, define.TaskTypePassportWeek, loadTaskFromConfig(pl, define.TaskTypePassportWeek))
	pl.Task.ResetAt[define.TaskTypePassportWeek] = now
}

func resetMonthTasks(pl *model.Player, now int64, taskType int32) {
	if (taskType != define.TaskTypeMonth && taskType != 0) || utils.IsSameMonthBySec(now, pl.Task.ResetAt[define.TaskTypeMonth]) {
		return
	}
	setBucket(pl, define.TaskTypeMonth, loadTaskFromConfig(pl, define.TaskTypeMonth))
	pl.Task.ResetAt[define.TaskTypeMonth] = now
}

func refreshActivityTaskBuckets(ctx global.IPlayer, pl *model.Player, taskType int32) {
	refreshOneActivityBucket(ctx, pl, taskType, define.TaskTypeDrawHeroRank, define.ActivityTypeDrawHeroRank)
	refreshOneActivityBucket(ctx, pl, taskType, define.TaskTypeTheCompetitionRank, define.ActivityTypeTheCompetition)
}

func refreshOneActivityBucket(ctx global.IPlayer, pl *model.Player, taskType, bucketType int32, activityType string) {
	if taskType != bucketType && taskType != 0 {
		return
	}
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(activityType)
	if err != nil {
		log.Error("get activity data id error:%v", err)
		return
	}
	if reply.ActivityId <= 0 {
		setBucket(pl, bucketType, nil)
		return
	}
	if getBucket(pl, bucketType) == nil {
		setBucket(pl, bucketType, loadTaskFromConfig(pl, bucketType))
	}
}

func initBucketsOnFirstTouch(ctx global.IPlayer, pl *model.Player, now int64) {
	setBucket(pl, define.TaskTypeDaily, loadTaskFromConfig(pl, define.TaskTypeDaily))
	setBucket(pl, define.TaskTypeWeek, loadTaskFromConfig(pl, define.TaskTypeWeek))
	setBucket(pl, define.TaskTypeMonth, loadTaskFromConfig(pl, define.TaskTypeMonth))
	setBucket(pl, define.TaskTypeMain, loadTaskFromConfig(pl, define.TaskTypeMain))
	setBucket(pl, define.TaskTypeAchieve, loadTaskFromConfig(pl, define.TaskTypeAchieve))
	setBucket(pl, define.TaskTypeGuild, loadTaskFromConfig(pl, define.TaskTypeGuild))
	setBucket(pl, define.TaskTypePassportDaily, loadTaskFromConfig(pl, define.TaskTypePassportDaily))
	setBucket(pl, define.TaskTypePassportWeek, loadTaskFromConfig(pl, define.TaskTypePassportWeek))
	setBucket(pl, define.TaskTypePassportSeason, loadTaskFromConfig(pl, define.TaskTypePassportSeason))
	setPoint(pl, define.TaskActivityTypeDaily, 0)
	setPoint(pl, define.TaskActivityTypeGuild, 0)

	pl.Task.ResetAt[define.TaskTypeDaily] = now
	pl.Task.ResetAt[define.TaskTypeWeek] = now
	pl.Task.ResetAt[define.TaskTypeMonth] = now
	pl.Task.ResetAt[define.TaskTypeGuild] = now
	pl.Task.ResetAt[define.TaskTypePassportDaily] = now
	pl.Task.ResetAt[define.TaskTypePassportWeek] = now

	refreshActivityTaskBuckets(ctx, pl, 0)
}
