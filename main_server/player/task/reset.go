package task

import (
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/utils"
)

// resetTask checks every bucket against the current time and resets any that
// have passed their cadence boundary. taskType == 0 checks all buckets.
func resetTask(ctx global.IPlayer, pl *model.Player, taskType int32) {
	ensureTaskData(pl)
	now := utils.Now().Unix()

	// First ever touch: populate all buckets from scratch.
	if pl.Task.ResetAt[define.TaskTypeDaily] == 0 {
		initBucketsOnFirstTouch(ctx, pl, now)
		return
	}

	for _, p := range bucketPolicies {
		if taskType != 0 && taskType != p.bucketType {
			continue
		}
		applyReset(ctx, pl, p, now)
	}
}

// applyReset resets a single bucket according to its policy if the cadence has elapsed.
func applyReset(ctx global.IPlayer, pl *model.Player, p bucketPolicy, now int64) {
	switch p.reset {
	case resetNone:
		return
	case resetActivity:
		refreshActivityBucket(ctx, pl, p)
		return
	case resetDaily:
		if utils.IsSameDayBySecWithHour(now, pl.Task.ResetAt[p.bucketType], 0) {
			return
		}
	case resetWeekly:
		if utils.IsSameWeekBySec(now, pl.Task.ResetAt[p.bucketType]) {
			return
		}
	case resetMonthly:
		if utils.IsSameMonthBySec(now, pl.Task.ResetAt[p.bucketType]) {
			return
		}
	}

	setBucket(pl, p.bucketType, loadTaskFromConfig(pl, p.bucketType))
	if p.claimType != 0 {
		setClaimMap(pl, p.claimType, make(map[int32]bool))
	}
	if p.pointType != 0 {
		setPoint(pl, p.pointType, 0)
	}
	if p.clearTaskLimit {
		pl.Task.TaskLimit = make(map[int32]int32)
	}
	pl.Task.ResetAt[p.bucketType] = now
}

// refreshActivityBucket enables or disables an activity-gated bucket based on
// whether the upstream activity is currently running.
func refreshActivityBucket(ctx global.IPlayer, pl *model.Player, p bucketPolicy) {
	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(p.activityKind)
	if err != nil {
		log.Error("get activity status error (kind=%s): %v", p.activityKind, err)
		return
	}
	if reply.ActivityId <= 0 {
		setBucket(pl, p.bucketType, nil)
		return
	}
	if getBucket(pl, p.bucketType) == nil {
		setBucket(pl, p.bucketType, loadTaskFromConfig(pl, p.bucketType))
	}
}

// initBucketsOnFirstTouch populates every bucket the first time a player touches
// the task system, and seeds the reset-time cursors.
func initBucketsOnFirstTouch(ctx global.IPlayer, pl *model.Player, now int64) {
	for _, p := range bucketPolicies {
		if p.reset == resetActivity {
			refreshActivityBucket(ctx, pl, p)
		} else {
			setBucket(pl, p.bucketType, loadTaskFromConfig(pl, p.bucketType))
		}
		if p.pointType != 0 {
			setPoint(pl, p.pointType, 0)
		}
		// Seed the reset cursor for periodic buckets so subsequent logins compare correctly.
		if p.reset != resetNone && p.reset != resetActivity {
			pl.Task.ResetAt[p.bucketType] = now
		}
	}
}



