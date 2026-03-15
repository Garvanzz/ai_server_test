package task

import (
	"encoding/json"
	"fmt"
	"sort"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/event"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_task"
)

const (
	taskDataVersion = 2

	claimTypeDaily int32 = 1
	claimTypeWeek  int32 = 2
	claimTypeGuild int32 = 3
)

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
	} else {
		// TODO: 异步存储
	}
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

// Dispatch 任务分发
func Dispatch(ctx global.IPlayer, pl *model.Player, taskType int32, taskCount int32, extraCondition int32, accumulate bool) {
	resetTask(ctx, pl, taskType)

	limit := define.TaskCompleteLimit[taskType]
	if limit != 0 && pl.Task.TaskLimit[taskType] >= limit {
		return
	}

	taskValue := setTaskInfo(pl, taskType, extraCondition, taskCount, accumulate)
	pl.Task.TaskLimit[taskType]++

	pushTasks := new(proto_task.PushTask)
	pushTasks.DailyTasks = pushTaskChange(getBucket(pl, define.TaskTypeDaily), taskType, extraCondition, taskValue)
	pushTasks.WeekTasks = pushTaskChange(getBucket(pl, define.TaskTypeWeek), taskType, extraCondition, taskValue)
	pushTasks.MonthTasks = pushTaskChange(getBucket(pl, define.TaskTypeMonth), taskType, extraCondition, taskValue)
	pushTasks.AchieveTasks = pushTaskChange(getBucket(pl, define.TaskTypeAchieve), taskType, extraCondition, taskValue)
	pushTasks.MainTasks = pushTaskChange(getBucket(pl, define.TaskTypeMain), taskType, extraCondition, taskValue)
	pushTasks.GuildTasks = pushTaskChange(getBucket(pl, define.TaskTypeGuild), taskType, extraCondition, taskValue)
	pushTasks.DrawRankTasks = pushTaskChange(getBucket(pl, define.TaskTypeDrawHeroRank), taskType, extraCondition, taskValue)
	pushTasks.TheCompetitionTasks = pushTaskChange(getBucket(pl, define.TaskTypeTheCompetitionRank), taskType, extraCondition, taskValue)
	pushTasks.PassportDailyTasks = pushTaskChange(getBucket(pl, define.TaskTypePassportDaily), taskType, extraCondition, taskValue)
	pushTasks.PassportWeekTasks = pushTaskChange(getBucket(pl, define.TaskTypePassportWeek), taskType, extraCondition, taskValue)
	pushTasks.PassportSeasonTasks = pushTaskChange(getBucket(pl, define.TaskTypePassportSeason), taskType, extraCondition, taskValue)
	pushTasks.DailyActivePoint = getPoint(pl, define.TaskActivityTypeDaily)
	pushTasks.GuildPoint = getPoint(pl, define.TaskActivityTypeGuild)

	ctx.Send(pushTasks)
}

func IsMainTaskRewarded(pl *model.Player, taskID int32) bool {
	ensureTaskData(pl)
	t, ok := getBucket(pl, define.TaskTypeMain)[taskID]
	if !ok {
		return false
	}
	return t.ReceiveAward
}

func pushTaskChange(m map[int32]*model.Task, taskType, extraCondition, taskValue int32) map[int32]*proto_task.TaskState {
	ret := make(map[int32]*proto_task.TaskState)
	for k, v := range m {
		if taskType == v.TaskType && v.ExtraCondition == extraCondition && !v.ReceiveAward {
			progress := taskValue - v.InitialProcess
			ret[k] = &proto_task.TaskState{Progress: progress, GetReward: false}
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func setTaskInfo(pl *model.Player, taskType int32, extraCondition int32, amount int32, accumulate bool) int32 {
	ensureTaskData(pl)
	m, ok := pl.Task.Progress[taskType]
	if !ok {
		m = make(map[int32]int32)
	}

	if accumulate {
		m[extraCondition] += amount
	} else if m[extraCondition] < amount {
		m[extraCondition] = amount
	}

	pl.Task.Progress[taskType] = m
	return m[extraCondition]
}

func resetTask(ctx global.IPlayer, pl *model.Player, taskType int32) {
	ensureTaskData(pl)
	now := utils.Now().Unix()

	if pl.Task.ResetAt[define.TaskTypeDaily] == 0 {
		initBucketsOnFirstTouch(ctx, pl, now)
		return
	}

	if (taskType == define.TaskTypeDaily || taskType == 0) && !utils.IsSameDayBySecWithHour(now, pl.Task.ResetAt[define.TaskTypeDaily], 0) {
		setBucket(pl, define.TaskTypeDaily, loadTaskFromConfig(pl, define.TaskTypeDaily))
		setPoint(pl, define.TaskActivityTypeDaily, 0)
		setClaimMap(pl, claimTypeDaily, make(map[int32]bool))
		pl.Task.TaskLimit = make(map[int32]int32)
		pl.Task.ResetAt[define.TaskTypeDaily] = now
	}

	if (taskType == define.TaskTypePassportDaily || taskType == 0) && !utils.IsSameDayBySecWithHour(now, pl.Task.ResetAt[define.TaskTypePassportDaily], 0) {
		setBucket(pl, define.TaskTypePassportDaily, loadTaskFromConfig(pl, define.TaskTypePassportDaily))
		pl.Task.ResetAt[define.TaskTypePassportDaily] = now
	}

	if (taskType == define.TaskTypeGuild || taskType == 0) && !utils.IsSameDayBySecWithHour(now, pl.Task.ResetAt[define.TaskTypeGuild], 0) {
		setBucket(pl, define.TaskTypeGuild, loadTaskFromConfig(pl, define.TaskTypeGuild))
		setPoint(pl, define.TaskActivityTypeGuild, 0)
		setClaimMap(pl, claimTypeGuild, make(map[int32]bool))
		pl.Task.ResetAt[define.TaskTypeGuild] = now
	}

	if (taskType == define.TaskTypeWeek || taskType == 0) && !utils.IsSameWeekBySec(now, pl.Task.ResetAt[define.TaskTypeWeek]) {
		setClaimMap(pl, claimTypeWeek, make(map[int32]bool))
		setBucket(pl, define.TaskTypeWeek, loadTaskFromConfig(pl, define.TaskTypeWeek))
		pl.Task.ResetAt[define.TaskTypeWeek] = now
	}

	if (taskType == define.TaskTypePassportWeek || taskType == 0) && !utils.IsSameWeekBySec(now, pl.Task.ResetAt[define.TaskTypePassportWeek]) {
		setBucket(pl, define.TaskTypePassportWeek, loadTaskFromConfig(pl, define.TaskTypePassportWeek))
		pl.Task.ResetAt[define.TaskTypePassportWeek] = now
	}

	if (taskType == define.TaskTypeMonth || taskType == 0) && !utils.IsSameMonthBySec(now, pl.Task.ResetAt[define.TaskTypeMonth]) {
		setBucket(pl, define.TaskTypeMonth, loadTaskFromConfig(pl, define.TaskTypeMonth))
		pl.Task.ResetAt[define.TaskTypeMonth] = now
	}

	if taskType == define.TaskTypeDrawHeroRank || taskType == 0 {
		reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeDrawHeroRank)
		if err != nil {
			log.Error("get activity data id error:%v", err)
		} else if reply.ActivityId > 0 && getBucket(pl, define.TaskTypeDrawHeroRank) == nil {
			setBucket(pl, define.TaskTypeDrawHeroRank, loadTaskFromConfig(pl, define.TaskTypeDrawHeroRank))
		}
	}

	if taskType == define.TaskTypeTheCompetitionRank || taskType == 0 {
		act, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeTheCompetition)
		if err != nil {
			log.Error("get activity data id error:%v", err)
		} else if act.ActivityId > 0 && getBucket(pl, define.TaskTypeTheCompetitionRank) == nil {
			setBucket(pl, define.TaskTypeTheCompetitionRank, loadTaskFromConfig(pl, define.TaskTypeTheCompetitionRank))
		}
	}
}

func loadTaskFromConfig(pl *model.Player, tp int32) map[int32]*model.Task {
	ret := make(map[int32]*model.Task)
	taskConfs := config.Task.All()

	if tp == define.TaskTypeAchieve {
		current := getBucket(pl, define.TaskTypeAchieve)
		if len(current) == 0 {
			id := firstAchieveTaskID(taskConfs)
			if id > 0 {
				conf := taskConfs[int64(id)]
				ret[id] = buildTaskState(pl, conf)
			}
			return ret
		}

		ids := sortedIDs(current)
		for _, id := range ids {
			conf, ok := taskConfs[int64(id)]
			if !ok || conf.BackTask == 0 {
				continue
			}
			nextConf, ok := taskConfs[int64(conf.BackTask)]
			if !ok {
				continue
			}
			ret[nextConf.Id] = buildTaskState(pl, nextConf)
			return ret
		}
		return ret
	}

	for _, id := range sortedTaskConfigIDs(taskConfs) {
		taskConf := taskConfs[int64(id)]
		if taskConf.Type == define.TaskTypeMain {
			heroId := int32(pl.GetProp(define.PlayerPropHeroId))
			heroData := pl.Hero.Hero[heroId]
			if len(taskConf.Param) > 0 && heroData.Stage == taskConf.Param[0] {
				ret[id] = buildTaskState(pl, taskConf)
			}
			continue
		}

		if taskConf.Type == tp {
			ret[id] = buildTaskState(pl, taskConf)
		}
	}

	return ret
}

func ReqTaskData(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetTasks) {
	resetTask(ctx, pl, 0)

	resp := &proto_task.S2CGetTasks{}
	resp.DailyTasks = taskToProto(getBucket(pl, define.TaskTypeDaily), pl)
	resp.WeekTasks = taskToProto(getBucket(pl, define.TaskTypeWeek), pl)
	resp.MonthTasks = taskToProto(getBucket(pl, define.TaskTypeMonth), pl)
	resp.AchieveTasks = taskToProto(getBucket(pl, define.TaskTypeAchieve), pl)
	resp.MainTasks = taskToProto(getBucket(pl, define.TaskTypeMain), pl)
	resp.GuildTasks = taskToProto(getBucket(pl, define.TaskTypeGuild), pl)
	resp.DrawRankTasks = taskToProto(getBucket(pl, define.TaskTypeDrawHeroRank), pl)
	resp.TheCompetitionTasks = taskToProto(getBucket(pl, define.TaskTypeTheCompetitionRank), pl)
	resp.PassportDailyTasks = taskToProto(getBucket(pl, define.TaskTypePassportDaily), pl)
	resp.PassportWeekTasks = taskToProto(getBucket(pl, define.TaskTypePassportWeek), pl)
	resp.PassportSeasonTasks = taskToProto(getBucket(pl, define.TaskTypePassportSeason), pl)
	resp.ActivePointRecord = copyClaimMap(getClaimMap(pl, claimTypeDaily))
	resp.ActivePointWeekRecord = copyClaimMap(getClaimMap(pl, claimTypeWeek))
	resp.GuildPointRecord = copyClaimMap(getClaimMap(pl, claimTypeGuild))
	resp.DailyActivePoint = getPoint(pl, define.TaskActivityTypeDaily)
	resp.GuildPoint = getPoint(pl, define.TaskActivityTypeGuild)
	ctx.Send(resp)
}

func taskToProto(m map[int32]*model.Task, pl *model.Player) map[int32]*proto_task.TaskState {
	ret := make(map[int32]*proto_task.TaskState)
	for k, v := range m {
		ret[k] = &proto_task.TaskState{
			Progress:  getTaskProgress(pl, v.TaskType, v.ExtraCondition) - v.InitialProcess,
			GetReward: v.ReceiveAward,
		}
	}
	return ret
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

func ReqReceiveReward(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetReward) {
	resetTask(ctx, pl, req.Type)

	tasks := getBucket(pl, req.Type)
	if tasks == nil {
		log.Error("ReqReceiveReward type error:%v", req.Type)
		ctx.Send(&proto_task.S2CGetReward{Succ: false})
		return
	}

	awards := make([]conf.ItemE, 0)
	taskConfs := config.Task.All()
	if req.Id == 0 {
		for id, v := range tasks {
			if v.ReceiveAward {
				continue
			}
			progress := getTaskProgress(pl, v.TaskType, v.ExtraCondition) - v.InitialProcess
			if progress < v.Condition {
				continue
			}
			taskConf, ok := taskConfs[int64(id)]
			if !ok {
				continue
			}
			awards = append(awards, taskConf.Reward...)
			v.ReceiveAward = true
			addActivityValue(ctx, pl, req.Type, taskConf.ActivityValue)
		}
	} else {
		task, ok := tasks[req.Id]
		if !ok || task.ReceiveAward {
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}

		progress := getTaskProgress(pl, task.TaskType, task.ExtraCondition) - task.InitialProcess
		if progress < task.Condition {
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}

		taskConf, ok := taskConfs[int64(req.Id)]
		if !ok {
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}

		task.ReceiveAward = true
		awards = append(awards, taskConf.Reward...)
		addActivityValue(ctx, pl, req.Type, taskConf.ActivityValue)
	}

	if len(awards) > 0 {
		bag.AddAward(ctx, pl, awards, true)
		internal.PushPlayerData(ctx, pl)
	}

	pushTasks := &proto_task.PushTask{}
	applyBucketPush(pushTasks, req.Type, pl)
	ctx.Send(pushTasks)
	ctx.Send(&proto_task.S2CGetReward{Succ: true, Id: req.Id})
}

func ReqReceiveActivePointReward(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetActivePointReward) {
	taskActivityConf, ok := config.TaskActivity.Find(int64(req.Id))
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
	if claims[req.Id] {
		ctx.Send(&proto_task.S2CGetActivePointReward{Succ: false})
		return
	}
	claims[req.Id] = true
	setClaimMap(pl, claimType, claims)

	if len(taskActivityConf.Reward) > 0 {
		bag.AddAward(ctx, pl, taskActivityConf.Reward, true)
	}

	pushTasks := &proto_task.PushTask{}
	if claimType == claimTypeDaily {
		pushTasks.ActivePointRecord = copyClaimMap(claims)
	} else {
		pushTasks.GuildPointRecord = copyClaimMap(claims)
	}
	ctx.Send(pushTasks)
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
	if _, ok := pl.Task.ClaimRecord[claimTypeDaily]; !ok {
		pl.Task.ClaimRecord[claimTypeDaily] = make(map[int32]bool)
	}
	if _, ok := pl.Task.ClaimRecord[claimTypeWeek]; !ok {
		pl.Task.ClaimRecord[claimTypeWeek] = make(map[int32]bool)
	}
	if _, ok := pl.Task.ClaimRecord[claimTypeGuild]; !ok {
		pl.Task.ClaimRecord[claimTypeGuild] = make(map[int32]bool)
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

	reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeDrawHeroRank)
	if err == nil && reply.ActivityId > 0 {
		setBucket(pl, define.TaskTypeDrawHeroRank, loadTaskFromConfig(pl, define.TaskTypeDrawHeroRank))
	}
	act, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeTheCompetition)
	if err == nil && act.ActivityId > 0 {
		setBucket(pl, define.TaskTypeTheCompetitionRank, loadTaskFromConfig(pl, define.TaskTypeTheCompetitionRank))
	}
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
		setTaskInfo(pl, taskType, 0, heroData.Level, false)
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
	if taskType == define.TaskTypeDaily {
		pushTasks.DailyTasks = taskToProto(getBucket(pl, define.TaskTypeDaily), pl)
		pushTasks.DailyActivePoint = getPoint(pl, define.TaskActivityTypeDaily)
		return
	}
	if taskType == define.TaskTypeWeek {
		pushTasks.WeekTasks = taskToProto(getBucket(pl, define.TaskTypeWeek), pl)
		return
	}
	if taskType == define.TaskTypeMonth {
		pushTasks.MonthTasks = taskToProto(getBucket(pl, define.TaskTypeMonth), pl)
		return
	}
	if taskType == define.TaskTypeAchieve {
		if allRewarded(getBucket(pl, define.TaskTypeAchieve)) {
			setBucket(pl, define.TaskTypeAchieve, loadTaskFromConfig(pl, define.TaskTypeAchieve))
		}
		pushTasks.AchieveTasks = taskToProto(getBucket(pl, define.TaskTypeAchieve), pl)
		return
	}
	if taskType == define.TaskTypeMain {
		if allRewarded(getBucket(pl, define.TaskTypeMain)) {
			setBucket(pl, define.TaskTypeMain, loadTaskFromConfig(pl, define.TaskTypeMain))
		}
		pushTasks.MainTasks = taskToProto(getBucket(pl, define.TaskTypeMain), pl)
		return
	}
	if taskType == define.TaskTypeGuild {
		pushTasks.GuildTasks = taskToProto(getBucket(pl, define.TaskTypeGuild), pl)
		pushTasks.GuildPoint = getPoint(pl, define.TaskActivityTypeGuild)
		return
	}
	if taskType == define.TaskTypeDrawHeroRank {
		pushTasks.DrawRankTasks = taskToProto(getBucket(pl, define.TaskTypeDrawHeroRank), pl)
		return
	}
	if taskType == define.TaskTypeTheCompetitionRank {
		pushTasks.TheCompetitionTasks = taskToProto(getBucket(pl, define.TaskTypeTheCompetitionRank), pl)
		return
	}
	if taskType == define.TaskTypePassportDaily {
		pushTasks.PassportDailyTasks = taskToProto(getBucket(pl, define.TaskTypePassportDaily), pl)
		return
	}
	if taskType == define.TaskTypePassportWeek {
		pushTasks.PassportWeekTasks = taskToProto(getBucket(pl, define.TaskTypePassportWeek), pl)
		return
	}
	if taskType == define.TaskTypePassportSeason {
		pushTasks.PassportSeasonTasks = taskToProto(getBucket(pl, define.TaskTypePassportSeason), pl)
	}
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
