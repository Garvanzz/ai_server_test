package task

import (
	"encoding/json"
	"fmt"

	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/pkg/log"
)

const taskDataVersion = 2

type legacyPlayerTask struct {
	Status                 map[int32]map[int32]int32
	TaskLimit              map[int32]int32
	ActivePointRecord      map[int32]bool
	ActivePointWeekRecord  map[int32]bool
	ActivePointGuildRecord map[int32]bool
	DailyTask              map[int32]*model.Task
	WeekTask               map[int32]*model.Task
	MonthTask              map[int32]*model.Task
	MainTask               map[int32]*model.Task
	AchieveTask            map[int32]*model.Task
	GuildTask              map[int32]*model.Task
	DrawHeroTask           map[int32]*model.Task
	TheCompetitionTask     map[int32]*model.Task
	PassportDailyTask      map[int32]*model.Task
	PassportWeekTask       map[int32]*model.Task
	PassportSeasonTask     map[int32]*model.Task
	DailyResetTime         int64
	WeekResetTime          int64
	MonthResetTime         int64
	GuildResetTime         int64
	DailyPoint             int32
	GuildPoint             int32
}

func Init(pl *model.Player) {
	pl.Task = &model.PlayerTask{}
	ensureTaskData(pl)
}

func Save(pl *model.Player, isSync bool) {
	ensureTaskData(pl)
	j, err := json.Marshal(pl.Task)
	if err != nil {
		log.Error("player[%v],save task marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		db.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerTask, pl.Id), j)
		return
	}
	// TODO: 异步存储
}

func Load(pl *model.Player) {
	reply, err := db.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerTask, pl.Id))
	if err != nil {
		log.Error("player[%v],load task error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.PlayerTask)
	err = json.Unmarshal(reply.([]byte), &m)
	if err == nil && (m.Version > 0 || m.Buckets != nil || m.Progress != nil) {
		pl.Task = m
		ensureTaskData(pl)
		return
	}

	legacy := new(legacyPlayerTask)
	err = json.Unmarshal(reply.([]byte), &legacy)
	if err != nil {
		log.Error("player[%v],load task unmarshal error:%v", pl.Id, err)
		Init(pl)
		return
	}

	pl.Task = migrateLegacy(legacy)
	ensureTaskData(pl)
}

func ensureTaskData(pl *model.Player) {
	if pl.Task == nil {
		pl.Task = &model.PlayerTask{}
	}
	pl.Task.Version = taskDataVersion
	if pl.Task.Progress == nil {
		pl.Task.Progress = make(map[int32]map[int32]int32)
	}
	if pl.Task.TaskLimit == nil {
		pl.Task.TaskLimit = make(map[int32]int32)
	}
	if pl.Task.Buckets == nil {
		pl.Task.Buckets = make(map[int32]map[int32]*model.Task)
	}
	if pl.Task.Points == nil {
		pl.Task.Points = make(map[int32]int32)
	}
	if pl.Task.ClaimRecord == nil {
		pl.Task.ClaimRecord = make(map[int32]map[int32]bool)
	}
	if pl.Task.ResetAt == nil {
		pl.Task.ResetAt = make(map[int32]int64)
	}
	// Ensure a claim-record map exists for every bucket that has one.
	// Derived from bucketPolicies so new buckets are covered automatically.
	for _, p := range bucketPolicies {
		if p.claimType != 0 {
			if _, ok := pl.Task.ClaimRecord[p.claimType]; !ok {
				pl.Task.ClaimRecord[p.claimType] = make(map[int32]bool)
			}
		}
	}
}

func migrateLegacy(old *legacyPlayerTask) *model.PlayerTask {
	n := &model.PlayerTask{
		Version:     taskDataVersion,
		Progress:    old.Status,
		TaskLimit:   old.TaskLimit,
		Buckets:     make(map[int32]map[int32]*model.Task),
		Points:      make(map[int32]int32),
		ClaimRecord: make(map[int32]map[int32]bool),
		ResetAt:     make(map[int32]int64),
	}
	n.Buckets[define.TaskTypeDaily] = old.DailyTask
	n.Buckets[define.TaskTypeWeek] = old.WeekTask
	n.Buckets[define.TaskTypeMonth] = old.MonthTask
	n.Buckets[define.TaskTypeAchieve] = old.AchieveTask
	n.Buckets[define.TaskTypeMain] = old.MainTask
	n.Buckets[define.TaskTypeGuild] = old.GuildTask
	n.Buckets[define.TaskTypeDrawHeroRank] = old.DrawHeroTask
	n.Buckets[define.TaskTypeTheCompetitionRank] = old.TheCompetitionTask
	n.Buckets[define.TaskTypePassportDaily] = old.PassportDailyTask
	n.Buckets[define.TaskTypePassportWeek] = old.PassportWeekTask
	n.Buckets[define.TaskTypePassportSeason] = old.PassportSeasonTask

	n.Points[define.TaskActivityTypeDaily] = old.DailyPoint
	n.Points[define.TaskActivityTypeGuild] = old.GuildPoint

	n.ClaimRecord[claimTypeDaily] = old.ActivePointRecord
	n.ClaimRecord[claimTypeWeek] = old.ActivePointWeekRecord
	n.ClaimRecord[claimTypeGuild] = old.ActivePointGuildRecord

	n.ResetAt[define.TaskTypeDaily] = old.DailyResetTime
	n.ResetAt[define.TaskTypeWeek] = old.WeekResetTime
	n.ResetAt[define.TaskTypeMonth] = old.MonthResetTime
	n.ResetAt[define.TaskTypeGuild] = old.GuildResetTime
	n.ResetAt[define.TaskTypePassportDaily] = old.DailyResetTime
	n.ResetAt[define.TaskTypePassportWeek] = old.WeekResetTime

	return n
}
