package task

import (
	"xfx/core/define"
	"xfx/core/model"
)

// claimType constants identify which activity-point claim bucket a TaskGroup uses.
const (
	claimTypeDaily int32 = 1
	claimTypeWeek  int32 = 2
	claimTypeGuild int32 = 3
)

// resetKind describes the periodic reset cadence of a task bucket.
type resetKind int8

const (
	resetNone     resetKind = 0 // never resets (achieve, main, passport season)
	resetDaily    resetKind = 1 // resets once per day
	resetWeekly   resetKind = 2 // resets once per week
	resetMonthly  resetKind = 3 // resets once per month
	resetActivity resetKind = 4 // enabled/disabled by a global activity
)

// bucketPolicy is the single source of truth for every task bucket's behaviour.
// To add a new bucket type: add one entry here and, if its loading logic is
// non-standard, add a case to loadTaskFromConfig in loader.go.
type bucketPolicy struct {
	bucketType     int32
	claimType      int32     // 0 → no activity-point claim system
	pointType      int32     // 0 → no local activity points
	reset          resetKind
	activityKind   string // non-empty when reset == resetActivity
	autoAdvance    bool   // reload bucket when every task in it is rewarded (achieve / main)
	notifyPassport bool   // fire passport-score event on task reward instead of local points
	clearTaskLimit bool   // reset pl.Task.TaskLimit when this bucket resets
}

// bucketPolicies is the authoritative, ordered registry of all task buckets.
var bucketPolicies = []bucketPolicy{
	{
		bucketType:     define.TaskTypeDaily,
		claimType:      claimTypeDaily,
		pointType:      define.TaskActivityTypeDaily,
		reset:          resetDaily,
		clearTaskLimit: true,
	},
	{
		bucketType: define.TaskTypeWeek,
		claimType:  claimTypeWeek,
		reset:      resetWeekly,
	},
	{
		bucketType: define.TaskTypeMonth,
		reset:      resetMonthly,
	},
	{
		bucketType:  define.TaskTypeAchieve,
		autoAdvance: true,
	},
	{
		bucketType:  define.TaskTypeMain,
		autoAdvance: true,
	},
	{
		bucketType: define.TaskTypeGuild,
		claimType:  claimTypeGuild,
		pointType:  define.TaskActivityTypeGuild,
		reset:      resetDaily,
	},
	{
		bucketType:   define.TaskTypeDrawHeroRank,
		reset:        resetActivity,
		activityKind: define.ActivityTypeDrawHeroRank,
	},
	{
		bucketType:   define.TaskTypeTheCompetitionRank,
		reset:        resetActivity,
		activityKind: define.ActivityTypeTheCompetition,
	},
	{
		bucketType:     define.TaskTypePassportDaily,
		reset:          resetDaily,
		notifyPassport: true,
	},
	{
		bucketType:     define.TaskTypePassportWeek,
		reset:          resetWeekly,
		notifyPassport: true,
	},
	{
		bucketType:     define.TaskTypePassportSeason,
		notifyPassport: true,
	},
}

// findPolicy returns the policy for the given bucketType, or false if unknown.
func findPolicy(bucketType int32) (bucketPolicy, bool) {
	for _, p := range bucketPolicies {
		if p.bucketType == bucketType {
			return p, true
		}
	}
	return bucketPolicy{}, false
}

// ---- bucket / point / claim store helpers ----

func getBucket(pl *model.Player, bucketType int32) map[int32]*model.Task {
	ensureTaskData(pl)
	return pl.Task.Buckets[bucketType]
}

func setBucket(pl *model.Player, bucketType int32, tasks map[int32]*model.Task) {
	ensureTaskData(pl)
	pl.Task.Buckets[bucketType] = tasks
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
