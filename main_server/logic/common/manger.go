package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/invoke"
	"xfx/pkg/log"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
	"xfx/pkg/utils"

	"github.com/gomodule/redigo/redis"
)

var Module = func() module.Module {
	return new(Manager)
}

type Manager struct {
	modules.BaseModule `json:"-"`
	RobotLock          map[int32]int64 // 过期时间
	LastCheckTime      int64           //上次检测时间
}

func (mgr *Manager) OnInit(app module.App) {
	mgr.BaseModule.OnInit(app)

	mgr.loadData()

	mgr.Register("releaseRobot", mgr.releaseRobot) // 释放机器人
	mgr.Register("matchRobot", mgr.matchRobot)     // 匹配机器人
	mgr.Register("matchRobots", mgr.matchRobots)   // 匹配机器人
}

func (mgr *Manager) loadData() {
	//获取时间

	reply, err := db.RedisExec("GET", define.CommonRedisKey)
	if err != nil {
		log.Error("load TimeEventDayRefresh time error:%v", err)
		return
	}

	if reply == nil {
		mgr.LastCheckTime = utils.Now().Unix()
		mgr.RobotLock = make(map[int32]int64)
		return
	}

	err = json.Unmarshal(reply.([]byte), mgr)
	if err != nil {
		log.Error("unmarshal TimeEventDayRefresh time error:%v", err)
		return
	}
}

func (mgr *Manager) GetType() string { return define.ModuleCommon }

func (mgr *Manager) OnTick(delta time.Duration) {
	now := utils.Now()
	if utils.CheckIsSameDayBySec(utils.GetTodayEndMinUnix(), mgr.LastCheckTime, 0) == false {
		mgr.LastCheckTime = now.Unix()

		//鉴宝月卡
		mgr.MonthCardGamAppraisal()

		//常规月卡
		mgr.NormalMonthCard()
	}

	if len(mgr.RobotLock) == 0 {
		return
	}

	//for robotId, expiration := range mgr.RobotLock {
	//	if now.Unix() >= expiration {
	//		delete(mgr.RobotLock, robotId)
	//	}
	//}
}

func (mgr *Manager) OnDestroy() {

	b, err := json.Marshal(mgr)
	if err != nil {
		log.Error("save TimeEventDayRefresh Init error:%v", err)
		return
	}

	db.RedisExec("SET", define.CommonRedisKey, string(b))
}

func (mgr *Manager) OnMessage(msg interface{}) interface{} {
	log.Debug("* stage message %v", msg)
	return nil
}

// 鉴宝月卡
func (mgr *Manager) MonthCardGamAppraisal() {
	//补发邮件

	reply, err := db.RedisExec("EXISTS", define.GemAppraisal_MonthCard)
	if err != nil {
		log.Error("load MonthCardGamAppraisal error:%v", err)
		return
	}

	if reply == nil {
		log.Error("load MonthCardGamAppraisal is null")
		return
	}

	replys, err := redis.StringMap(db.RedisExec("HGETALL", define.GemAppraisal_MonthCard))
	if err != nil {
		log.Error("load1 MonthCardGamAppraisal error:%v", err)
		return
	}

	//遍历
	for k, value := range replys {
		tab := new(model.GemAppraisalMonthCard)
		err = json.Unmarshal([]byte(value), &tab)
		if err != nil {
			log.Error("load GemAppraisalMonthCard unmarshal error:%v", err)
			continue
		}

		uid, _ := strconv.ParseInt(k, 10, 64)

		//判断时间
		if tab.GetDay >= tab.EffectDay {
			db.RedisExec("HDEL", define.GemAppraisal_MonthCard, uid)
			continue
		}

		if utils.CheckIsSameDayBySec(tab.GetTime, utils.Now().Unix(), 0) {
			continue
		}

		//奖励
		confs := config.MonthCard.All()
		conf := conf2.MonthCard{}
		for _, v := range confs {
			if v.Type == define.MonthCard_GemAppraisal {
				conf = v
				break
			}
		}

		if len(conf.Reward) > 0 {
			//计算天数差
			offseDay := utils.DaysDiff(tab.GetTime, utils.Now().Unix())
			for j := 0; j < len(conf.Reward); j++ {
				conf.Reward[j].ItemNum = conf.Reward[j].ItemNum * offseDay
			}

			ids := make([]int64, 0)
			ids = append(ids, tab.DbId)

			isSuc := invoke.MailClient(mgr).SendMail(define.PlayerMail, "鉴宝月卡", "鉴宝月卡当日未领补发", "", "", "游戏系统", conf.Reward, ids, int64(0), int32(0), false, []string{})
			if !isSuc {
				continue
			}
		}

		tab.GetDay += 1
		tab.GetTime = utils.Now().Unix()

		js, _ := json.Marshal(tab)
		db.RedisExec("HSET", define.GemAppraisal_MonthCard, uid, js)
	}
}

// 常规月卡
func (mgr *Manager) NormalMonthCard() {
	//获取活动是否开启
	reply, err := invoke.ActivityClient(mgr).GetActivityStatusByType(define.ActivityTypeNormalMonthCard)
	if err != nil || reply == nil {
		log.Error("get activity data id error:%v", err)
		return
	}

	if reply.ActivityId <= 0 || !reply.IsOpen {
		return
	}

	//处理数据
}

// 匹配机器人
func (mgr *Manager) matchRobots(mode int, startPower int64, endPower int64, count int32) ([]*model.Robot, error) {
	robots := make([]*model.Robot, 0)
	for i := int32(0); i < count; i++ {
		robot, err := mgr.matchRobot(mode, startPower, endPower)
		if err != nil {
			log.Debug("匹配机器失败:%v", err)
			continue
		}
		log.Debug("匹配机器成功:%v", robot)
		robots = append(robots, robot)
	}
	if len(robots) <= 0 {
		return nil, errors.New("robotIds is null")
	}
	return robots, nil
}

// 匹配机器人
func (mgr *Manager) matchRobot(mode int, startPower int64, endPower int64) (*model.Robot, error) {
	robotGroupConfs := config.RobotGroup.All()

	var robots []conf2.RobotGroup
	for _, robotGroupConf := range robotGroupConfs {
		if endPower > robotGroupConf.Power && startPower <= robotGroupConf.Power && robotGroupConf.Mode == int32(mode) {
			if _, ok := mgr.RobotLock[robotGroupConf.Id]; ok {
				continue
			}
			robots = append(robots, robotGroupConf)
			break
		}
	}

	if len(robots) == 0 {
		return nil, errors.New("robotIds is null")
	}

	//robotIds = make([]int, len(robotGroupConf.RobotId)) // 创建新切片
	//copy(robotIds, robotGroupConf.RobotId)

	robotGroup := robots[0]

	mgr.RobotLock[robotGroup.Id] = 0

	return &model.Robot{
		Id:       robotGroup.Id,
		RobotIds: robotGroup.RobotId,
		Power:    robotGroup.Power,
	}, nil
}

// 释放机器人
func (mgr *Manager) releaseRobot(robotId int32) {
	delete(mgr.RobotLock, robotId)
}

func (mgr *Manager) createRobot() {
	reply, err := db.RedisExec("keys", "robot:1")
	if err != nil {
		log.Error("createRobots %v", err)
		return
	}

	if len(reply.([]interface{})) == 0 {
		robotConfs := config.RobotGroup.All()

		for _, v := range robotConfs {
			robot := model.Robot{
				Id:       v.Id,
				RobotIds: v.RobotId,
				Power:    v.Power,
			}

			b, _ := json.Marshal(robot)

			db.RedisExec("set", fmt.Sprintf("robot:%d", robot.Id), string(b))
		}
	}
}
