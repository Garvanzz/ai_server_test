package task

import (
	"encoding/json"
	"fmt"
	"time"
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

func Init(pl *model.Player) {
	pl.Task = new(model.PlayerTask)
	pl.Task.Status = make(map[int32]map[int32]int32)
	pl.Task.TaskLimit = make(map[int32]int32)
	pl.Task.ActivePointRecord = make(map[int32]bool)
	pl.Task.ActivePointWeekRecord = make(map[int32]bool)
	pl.Task.ActivePointGuildRecord = make(map[int32]bool)
	pl.Task.GuildPoint = 0
	pl.Task.DailyPoint = 0
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Task)
	if err != nil {
		log.Error("player[%v],save task marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save task error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerTask, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save task error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerTask, pl.Id))
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
	if err != nil {
		log.Error("player[%v],load task unmarshal error:%v", pl.Id, err)
	}
	pl.Task = m

	// TODO:load new tasks
}

// Dispatch  任务分发
func Dispatch(ctx global.IPlayer, pl *model.Player, taskType int32, taskCount int32, extraCondition int32, accumulate bool) {
	resetTask(ctx, pl, taskType)

	// 检查任务完成次数限制
	limit := define.TaskCompleteLimit[taskType]
	if limit != 0 && pl.Task.TaskLimit[taskType] >= limit {
		return
	}

	// 通用任务记录
	taskValue := setTaskInfo(pl, taskType, extraCondition, taskCount, accumulate)
	pl.Task.TaskLimit[taskType]++

	// 推送
	pushTasks := new(proto_task.PushTask)
	pushTasks.DailyTasks = pushTaskChange(pl.Task.DailyTask, taskType, extraCondition, taskValue)
	pushTasks.WeekTasks = pushTaskChange(pl.Task.WeekTask, taskType, extraCondition, taskValue)
	pushTasks.MonthTasks = pushTaskChange(pl.Task.MonthTask, taskType, extraCondition, taskValue)
	pushTasks.AchieveTasks = pushTaskChange(pl.Task.AchieveTask, taskType, extraCondition, taskValue)
	pushTasks.MainTasks = pushTaskChange(pl.Task.MainTask, taskType, extraCondition, taskValue)
	pushTasks.GuildTasks = pushTaskChange(pl.Task.GuildTask, taskType, extraCondition, taskValue)
	pushTasks.DrawRankTasks = pushTaskChange(pl.Task.DrawHeroTask, taskType, extraCondition, taskValue)
	pushTasks.TheCompetitionTasks = pushTaskChange(pl.Task.TheCompetitionTask, taskType, extraCondition, taskValue)
	pushTasks.PassportDailyTasks = pushTaskChange(pl.Task.PassportDailyTask, taskType, extraCondition, taskValue)
	pushTasks.PassportWeekTasks = pushTaskChange(pl.Task.PassportWeekTask, taskType, extraCondition, taskValue)
	pushTasks.PassportSeasonTasks = pushTaskChange(pl.Task.PassportSeasonTask, taskType, extraCondition, taskValue)
	pushTasks.DailyActivePoint = pl.Task.DailyPoint
	pushTasks.GuildPoint = pl.Task.GuildPoint

	ctx.Send(pushTasks)
}

// 推送任务变化
func pushTaskChange(m map[int32]*model.Task, taskType, extraCondition, taskValue int32) map[int32]*proto_task.TaskState {
	ret := make(map[int32]*proto_task.TaskState)
	for k, v := range m {
		if taskType == v.TaskType && v.ExtraCondition == extraCondition && !v.ReceiveAward {
			progress := taskValue - v.InitialProcess
			ret[k] = new(proto_task.TaskState)
			ret[k].Progress = progress
			ret[k].GetReward = false

			//progress := taskValue - v.InitialProcess
			//if progress >= v.Condition {
			//	ret[k] = new(proto_task.TaskState)
			//	ret[k].Progress = progress
			//}
		}
	}

	if len(ret) > 0 {
		return ret
	} else {
		return nil
	}
}

// 通用统计任务信息
func setTaskInfo(pl *model.Player, taskType int32, extraCondition int32, amount int32, accumulate bool) int32 {
	m, ok := pl.Task.Status[taskType]
	if !ok {
		m = make(map[int32]int32)
	}

	if accumulate {
		m[extraCondition] += amount
	} else {
		if m[extraCondition] < amount {
			m[extraCondition] = amount
		}
	}

	pl.Task.Status[taskType] = m
	return m[extraCondition]
}

// 重置任务
func resetTask(ctx global.IPlayer, pl *model.Player, taskType int32) {
	now := time.Now().Unix()

	// 初始化任务
	if pl.Task.DailyResetTime == 0 {
		pl.Task.DailyTask = loadTaskFromConfig(pl, define.TaskTypeDaily)
		pl.Task.WeekTask = loadTaskFromConfig(pl, define.TaskTypeWeek)
		pl.Task.MonthTask = loadTaskFromConfig(pl, define.TaskTypeMonth)
		pl.Task.MainTask = loadTaskFromConfig(pl, define.TaskTypeMain)
		pl.Task.AchieveTask = loadTaskFromConfig(pl, define.TaskTypeAchieve)
		pl.Task.GuildTask = loadTaskFromConfig(pl, define.TaskTypeGuild)
		pl.Task.TheCompetitionTask = loadTaskFromConfig(pl, define.TaskTypeTheCompetitionRank)
		pl.Task.PassportDailyTask = loadTaskFromConfig(pl, define.TaskTypePassportDaily)
		pl.Task.PassportWeekTask = loadTaskFromConfig(pl, define.TaskTypePassportWeek)
		pl.Task.PassportSeasonTask = loadTaskFromConfig(pl, define.TaskTypePassportSeason)
		pl.Task.DailyPoint = 0
		pl.Task.GuildPoint = 0

		pl.Task.DailyResetTime = now
		pl.Task.WeekResetTime = now
		pl.Task.MonthResetTime = now
		pl.Task.GuildResetTime = now
		return
	}

	// 检查是否跨天
	if (taskType == define.TaskTypeDaily || taskType == 0) && !utils.CheckIsSameDayBySec(now, pl.Task.DailyResetTime, 0) {
		//每日任务
		pl.Task.DailyTask = loadTaskFromConfig(pl, define.TaskTypeDaily)

		// 重置活跃点
		pl.Task.DailyPoint = 0
		pl.Task.ActivePointRecord = make(map[int32]bool)

		// 重置任务完成次数限制
		pl.Task.TaskLimit = make(map[int32]int32)

		pl.Task.DailyResetTime = now
	}

	// 检查是否跨天 - 通行证每日任务
	if (taskType == define.TaskTypePassportDaily || taskType == 0) && !utils.CheckIsSameDayBySec(now, pl.Task.DailyResetTime, 0) {
		pl.Task.PassportDailyTask = loadTaskFromConfig(pl, define.TaskTypePassportDaily)
	}

	// 检查是否跨天
	if (taskType == define.TaskTypeGuild || taskType == 0) && !utils.CheckIsSameDayBySec(now, pl.Task.GuildResetTime, 0) {
		//帮派任务
		pl.Task.GuildTask = loadTaskFromConfig(pl, define.TaskTypeGuild)

		// 重置活跃点
		pl.Task.GuildPoint = 0
		pl.Task.ActivePointGuildRecord = make(map[int32]bool)

		pl.Task.GuildResetTime = now
	}

	// 检查是否跨周
	if (taskType == define.TaskTypeWeek || taskType == 0) && !utils.IsSameWeekBySec(now, pl.Task.DailyResetTime) {
		// 重置活跃点
		//pl.SetProp(define.PlayerPropActiveWeekPoint, 0, false)
		pl.Task.ActivePointWeekRecord = make(map[int32]bool)
		pl.Task.WeekTask = loadTaskFromConfig(pl, define.TaskTypeWeek)
		pl.Task.WeekResetTime = now
	}

	// 检查是否跨周 - 通行证每周任务
	if (taskType == define.TaskTypePassportWeek || taskType == 0) && !utils.IsSameWeekBySec(now, pl.Task.WeekResetTime) {
		pl.Task.PassportWeekTask = loadTaskFromConfig(pl, define.TaskTypePassportWeek)
	}

	// 检查是否跨月
	if (taskType == define.TaskTypeMonth || taskType == 0) && !utils.IsSameMonthBySec(now, pl.Task.DailyResetTime) {
		pl.Task.MonthTask = loadTaskFromConfig(pl, define.TaskTypeMonth)
		pl.Task.MonthResetTime = now
	}

	//检查活动任务 - 招募排行榜
	if taskType == define.TaskTypeDrawHeroRank || taskType == 0 {
		reply, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeDrawHeroRank)
		if err != nil {
			log.Error("get activity data id error:%v", err)
		} else {
			if reply.ActivityId > 0 && pl.Task.DrawHeroTask == nil {
				pl.Task.DrawHeroTask = loadTaskFromConfig(pl, define.TaskTypeDrawHeroRank)
			}
		}
	}

	//活动-巅峰决斗
	if taskType == define.TaskTypeTheCompetitionRank || taskType == 0 {

		act, err := invoke.ActivityClient(ctx).GetActivityStatusByType(define.ActivityTypeTheCompetition)
		if err != nil {
			log.Error("get activity data id error:%v", err)
		} else {
			if act.ActivityId > 0 && pl.Task.TheCompetitionTask == nil {
				pl.Task.TheCompetitionTask = loadTaskFromConfig(pl, define.TaskTypeTheCompetitionRank)
			}
		}
	}

	return
}

func loadTaskFromConfig(pl *model.Player, tp int32) map[int32]*model.Task {
	taskConfs := config.Task.All()

	ret := make(map[int32]*model.Task)

	// 成就任务特殊处理
	if tp == define.TaskTypeAchieve {
		if pl.Task.AchieveTask == nil {
			for id, taskConf := range taskConfs {
				if taskConf.Type == tp && taskConf.FrontTask == 0 {
					ret[int32(id)] = &model.Task{
						Id:             int32(id),
						InitialProcess: getTaskInitProcess(pl, taskConf),
						TaskType:       taskConf.TaskType,
						Condition:      taskConf.Condition1,
						ExtraCondition: taskConf.Condition2,
					}

					switch taskConf.TaskType {
					//主角等级
					case define.TaskHeroLevel:
						heroId := int32(pl.GetProp(define.PlayerPropHeroId))
						heroData := pl.Hero.Hero[heroId]
						setTaskInfo(pl, taskConf.TaskType, 0, heroData.Level, false)
					default:
					}
					return ret
				}
			}
		} else {
			// 切换下一个
			for id := range pl.Task.AchieveTask {
				taskConf := taskConfs[int64(id)]
				if taskConf.BackTask != 0 {
					nextTaskConf := taskConfs[int64(taskConf.BackTask)]

					ret[nextTaskConf.Id] = &model.Task{
						Id:             nextTaskConf.Id,
						InitialProcess: getTaskInitProcess(pl, nextTaskConf),
						TaskType:       nextTaskConf.TaskType,
						Condition:      nextTaskConf.Condition1,
						ExtraCondition: nextTaskConf.Condition2,
					}

					switch taskConf.TaskType {
					//主角等级
					case define.TaskHeroLevel:
						heroId := int32(pl.GetProp(define.PlayerPropHeroId))
						heroData := pl.Hero.Hero[heroId]
						setTaskInfo(pl, taskConf.TaskType, 0, heroData.Level, false)
					default:
					}
					return ret
				}
			}
		}
		return ret
	}

	for id, taskConf := range taskConfs {
		if taskConf.Type == define.TaskTypeMain {
			//获取阶数
			heroId := int32(pl.GetProp(define.PlayerPropHeroId))
			heroData := pl.Hero.Hero[heroId]
			if heroData.Stage == taskConf.Param[0] {
				ret[int32(id)] = &model.Task{
					Id:             int32(id),
					InitialProcess: getTaskInitProcess(pl, taskConf),
					TaskType:       taskConf.TaskType,
					Condition:      taskConf.Condition1,
					ExtraCondition: taskConf.Condition2,
				}

				switch taskConf.TaskType {
				//主角等级
				case define.TaskHeroLevel:
					setTaskInfo(pl, taskConf.TaskType, 0, heroData.Level, false)
				default:
				}
			}
		} else if taskConf.Type == tp {
			ret[int32(id)] = &model.Task{
				Id:             int32(id),
				InitialProcess: getTaskInitProcess(pl, taskConf),
				TaskType:       taskConf.TaskType,
				Condition:      taskConf.Condition1,
				ExtraCondition: taskConf.Condition2,
			}

			switch taskConf.TaskType {
			//主角等级
			case define.TaskHeroLevel:
				heroId := int32(pl.GetProp(define.PlayerPropHeroId))
				heroData := pl.Hero.Hero[heroId]
				setTaskInfo(pl, taskConf.TaskType, 0, heroData.Level, false)
			default:
			}
		}
	}
	return ret
}

// ReqTaskData 请求任务数据
func ReqTaskData(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetTasks) {
	resetTask(ctx, pl, 0)

	resp := new(proto_task.S2CGetTasks)
	resp.DailyTasks = taskToProto(pl.Task.DailyTask, pl)
	resp.WeekTasks = taskToProto(pl.Task.WeekTask, pl)
	resp.MonthTasks = taskToProto(pl.Task.MonthTask, pl)
	resp.AchieveTasks = taskToProto(pl.Task.AchieveTask, pl)
	resp.MainTasks = taskToProto(pl.Task.MainTask, pl)
	resp.GuildTasks = taskToProto(pl.Task.GuildTask, pl)
	resp.DrawRankTasks = taskToProto(pl.Task.DrawHeroTask, pl)
	resp.TheCompetitionTasks = taskToProto(pl.Task.TheCompetitionTask, pl)
	resp.PassportDailyTasks = taskToProto(pl.Task.PassportDailyTask, pl)
	resp.PassportWeekTasks = taskToProto(pl.Task.PassportWeekTask, pl)
	resp.PassportSeasonTasks = taskToProto(pl.Task.PassportSeasonTask, pl)
	resp.ActivePointRecord = pl.Task.ActivePointRecord
	resp.ActivePointWeekRecord = pl.Task.ActivePointWeekRecord
	resp.GuildPointRecord = pl.Task.ActivePointGuildRecord
	resp.DailyActivePoint = pl.Task.DailyPoint
	resp.GuildPoint = pl.Task.GuildPoint
	ctx.Send(resp)
}

func taskToProto(m map[int32]*model.Task, pl *model.Player) map[int32]*proto_task.TaskState {
	ret := make(map[int32]*proto_task.TaskState)
	for k, v := range m {
		taskState := new(proto_task.TaskState)
		taskState.Progress = getTaskProgress(pl, v.TaskType, v.ExtraCondition) - v.InitialProcess
		taskState.GetReward = v.ReceiveAward
		ret[k] = taskState
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
	m, ok := pl.Task.Status[taskType]
	if !ok {
		return 0
	}

	return m[extraCondition]
}

// ReqReceiveReward TODO:领奖任务奖励
func ReqReceiveReward(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetReward) {
	resetTask(ctx, pl, req.Type)

	var tasks map[int32]*model.Task
	if req.Type == define.TaskTypeDaily {
		tasks = pl.Task.DailyTask
	} else if req.Type == define.TaskTypeWeek {
		tasks = pl.Task.WeekTask
	} else if req.Type == define.TaskTypeMonth {
		tasks = pl.Task.MonthTask
	} else if req.Type == define.TaskTypeAchieve {
		tasks = pl.Task.AchieveTask
	} else if req.Type == define.TaskTypeMain {
		tasks = pl.Task.MainTask
	} else if req.Type == define.TaskTypeGuild {
		tasks = pl.Task.GuildTask
	} else if req.Type == define.TaskTypeDrawHeroRank {
		tasks = pl.Task.DrawHeroTask
	} else if req.Type == define.TaskTypeTheCompetitionRank {
		tasks = pl.Task.TheCompetitionTask
	} else if req.Type == define.TaskTypePassportDaily {
		tasks = pl.Task.PassportDailyTask
	} else if req.Type == define.TaskTypePassportWeek {
		tasks = pl.Task.PassportWeekTask
	} else if req.Type == define.TaskTypePassportSeason {
		tasks = pl.Task.PassportSeasonTask
	} else {
		log.Error("ReqReceiveReward type error:%v", req.Type)
		return
	}

	awards := make([]conf.ItemE, 0)
	taskConfs := config.Task.All()
	if req.Id == 0 { // 一键领取
		for id, v := range tasks {
			if v.ReceiveAward {
				continue
			}

			progress := getTaskProgress(pl, v.TaskType, v.ExtraCondition) - v.InitialProcess
			if progress >= v.Condition {
				// TODO:融合奖励
				taskConf := taskConfs[(int64)(id)]
				awards = append(awards, taskConf.Reward...)
				v.ReceiveAward = true

				//活跃点
				if req.Type == define.TaskTypeDaily {
					pl.Task.DailyPoint += taskConf.ActivityValue
				} else if req.Type == define.TaskTypeGuild {
					pl.Task.GuildPoint += taskConf.ActivityValue
				} else if req.Type == define.TaskTypePassportDaily || req.Type == define.TaskTypePassportWeek || req.Type == define.TaskTypePassportSeason {
					// 通行证任务: 增加积分而不是活跃点
					addPassportScore(ctx, pl, taskConf.ActivityValue)
				}
			}
		}

		if len(awards) > 0 {
			bag.AddAward(ctx, pl, awards, true)
			internal.PushPlayerData(ctx, pl)
		}
	} else {
		task, ok := tasks[req.Id]
		if !ok {
			log.Error("任务不存在")
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}

		if task.ReceiveAward {
			log.Error("已经领取奖励")
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}

		progress := getTaskProgress(pl, task.TaskType, task.ExtraCondition) - task.InitialProcess
		if progress < task.Condition {
			log.Error("不满足条件")
			ctx.Send(&proto_task.S2CGetReward{Succ: false})
			return
		}

		task.ReceiveAward = true

		// TODO:发奖
		taskConf := taskConfs[(int64)(req.Id)]
		awards = append(awards, taskConf.Reward...)

		//活跃点
		if req.Type == define.TaskTypeDaily {
			pl.Task.DailyPoint += taskConf.ActivityValue
		} else if req.Type == define.TaskTypeGuild {
			pl.Task.GuildPoint += taskConf.ActivityValue
		} else if req.Type == define.TaskTypePassportDaily || req.Type == define.TaskTypePassportWeek || req.Type == define.TaskTypePassportSeason {
			// 通行证任务: 增加积分而不是活跃点
			addPassportScore(ctx, pl, taskConf.ActivityValue)
		}

		if len(awards) > 0 {
			bag.AddAward(ctx, pl, awards, true)
			internal.PushPlayerData(ctx, pl)
		}
	}

	// 推送
	pushTasks := new(proto_task.PushTask)
	if req.Type == define.TaskTypeDaily {
		pushTasks.DailyTasks = taskToProto(pl.Task.DailyTask, pl)
		pushTasks.DailyActivePoint = pl.Task.DailyPoint
	} else if req.Type == define.TaskTypeWeek {
		pushTasks.WeekTasks = taskToProto(pl.Task.WeekTask, pl)
	} else if req.Type == define.TaskTypeMonth {
		pushTasks.MonthTasks = taskToProto(pl.Task.MonthTask, pl)
	} else if req.Type == define.TaskTypeAchieve {
		// 检查是否完成了所有任务
		finish := true
		for _, task := range pl.Task.AchieveTask {
			if !task.ReceiveAward {
				finish = false
				break
			}
		}

		log.Debug("成就任务:%v, %v", finish, pl.Task.AchieveTask)
		if finish {
			pl.Task.AchieveTask = loadTaskFromConfig(pl, define.TaskTypeAchieve)
			log.Debug("切换成就任务:%v", pl.Task.AchieveTask)
		}
		pushTasks.AchieveTasks = taskToProto(pl.Task.AchieveTask, pl)
	} else if req.Type == define.TaskTypeMain {
		// 检查是否完成了所有任务 解锁下一章任务
		finish := true
		for _, task := range pl.Task.MainTask {
			if !task.ReceiveAward {
				finish = false
				break
			}
		}

		if finish {
			pl.Task.MainTask = loadTaskFromConfig(pl, define.TaskTypeMain)
		}
		pushTasks.MainTasks = taskToProto(pl.Task.MainTask, pl)
	} else if req.Type == define.TaskTypeGuild {
		pushTasks.GuildTasks = taskToProto(pl.Task.GuildTask, pl)
		pushTasks.GuildPoint = pl.Task.GuildPoint
	} else if req.Type == define.TaskTypeDrawHeroRank {
		pushTasks.DrawRankTasks = taskToProto(pl.Task.DrawHeroTask, pl)
	} else if req.Type == define.TaskTypeTheCompetitionRank {
		pushTasks.TheCompetitionTasks = taskToProto(pl.Task.TheCompetitionTask, pl)
	} else if req.Type == define.TaskTypePassportDaily {
		pushTasks.PassportDailyTasks = taskToProto(pl.Task.PassportDailyTask, pl)
	} else if req.Type == define.TaskTypePassportWeek {
		pushTasks.PassportWeekTasks = taskToProto(pl.Task.PassportWeekTask, pl)
	} else if req.Type == define.TaskTypePassportSeason {
		pushTasks.PassportSeasonTasks = taskToProto(pl.Task.PassportSeasonTask, pl)
	}

	ctx.Send(pushTasks)

	ctx.Send(&proto_task.S2CGetReward{Succ: true, Id: req.Id})
}

// ReqReceiveActivePointReward 领奖活跃点奖励
func ReqReceiveActivePointReward(ctx global.IPlayer, pl *model.Player, req *proto_task.C2SGetActivePointReward) {
	// 是否有配置
	taskActivityConf, ok := config.TaskActivity.Find(int64(req.Id))
	if !ok {
		return
	}

	resetTask(ctx, pl, taskActivityConf.Type)

	if taskActivityConf.Type == define.TaskActivityTypeDaily {
		// 是否满足条件
		point := pl.Task.DailyPoint
		if taskActivityConf.Value > point {
			return
		}

		// 检查是否领取奖励
		_, ok = pl.Task.ActivePointRecord[req.Id]
		if ok {
			return
		}
		pl.Task.ActivePointRecord[req.Id] = true
	} else if taskActivityConf.Type == define.TaskActivityTypeGuild {
		// 是否满足条件
		point := pl.Task.GuildPoint
		if taskActivityConf.Value > point {
			return
		}

		// 检查是否领取奖励
		_, ok = pl.Task.ActivePointGuildRecord[req.Id]
		if ok {
			return
		}
		pl.Task.ActivePointGuildRecord[req.Id] = true
	}

	// TODO:发奖
	if len(taskActivityConf.Reward) > 0 {
		bag.AddAward(ctx, pl, taskActivityConf.Reward, true)
	}

	pushTasks := new(proto_task.PushTask)
	if taskActivityConf.Type == define.TaskActivityTypeDaily {
		pushTasks.ActivePointRecord = pl.Task.ActivePointRecord
	} else if taskActivityConf.Type == define.TaskActivityTypeGuild {
		pushTasks.GuildPointRecord = pl.Task.ActivePointGuildRecord
	}
	ctx.Send(pushTasks)

	ctx.Send(&proto_task.S2CGetActivePointReward{Succ: true})
}

// addPassportScore 为通行证活动增加积分
func addPassportScore(ctx global.IPlayer, pl *model.Player, score int32) {
	// 通过全局事件系统触发通行证积分事件
	event.DoEvent(define.EventTypeActivity, map[string]any{
		"key":         "passport_task_score",
		"player":      pl.ToContext(),
		"score":       score,
		"playermodel": pl,
		"IPlayer":     ctx,
	})

	log.Debug("addPassportScore dispatch event: playerId=%d, score=%d", pl.Id, score)
}
