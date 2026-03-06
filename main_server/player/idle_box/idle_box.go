package idle_box

import (
	"encoding/json"
	"fmt"
	"time"
	"xfx/core/config"
	conf2 "xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/proto/proto_idlebox"
)

func Init(pl *model.Player) {
	pl.IdleBox = new(model.IdleBox)

	now := time.Now().Unix()
	pl.IdleBox.StartTime = now
	pl.IdleBox.EndTime = now + int64(config.Global.Get().IdleBoxMaxTime*3600)
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.IdleBox)
	if err != nil {
		log.Error("player[%v],save IdleBox marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("Load IdleBox error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerIdleBox, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("Load IdleBox error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerIdleBox, pl.Id))
	if err != nil {
		log.Error("player[%v],load bag error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.IdleBox)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load IdleBox unmarshal error:%v", pl.Id, err)
	}

	pl.IdleBox = m
}

// ReqAddTime TODO：请求加时
func ReqAddTime(ctx global.IPlayer, pl *model.Player, req *proto_idlebox.C2SAddTime) {
	res := &proto_idlebox.S2CAddTime{}
	ctx.Send(res)
}

// ReqReceiveReward 领取奖励
func ReqReceiveReward(ctx global.IPlayer, pl *model.Player, req *proto_idlebox.C2SReceiveAward) {
	res := &proto_idlebox.S2CReceiveAward{}

	log.Debug("请求领取挂机奖励")
	idleBoxConfs := config.IdleBox.All()

	//获取关卡
	stageId := pl.Stage.CurStage
	//关卡
	stageConf_stage := conf2.IdleBox{}
	for _, v := range idleBoxConfs {
		if v.StageRange[0] <= stageId && stageId < v.StageRange[1] {
			stageConf_stage = v
			break
		}
	}

	//爬塔
	climb := pl.Mission.ClimbTower.Stage
	stageConf_climb := conf2.IdleBox{}
	for _, v := range idleBoxConfs {
		if v.TowerRange[0] <= climb && climb < v.TowerRange[1] {
			stageConf_climb = v
			break
		}
	}

	// TODO:计算奖励
	var reward []conf2.ItemE
	reward = calcReward(pl.IdleBox.StartTime, pl.IdleBox.EndTime, stageConf_stage, stageConf_climb)
	log.Debug("请求领取挂机奖励解析:%v", reward)
	if reward == nil {
		res.StartTime = pl.IdleBox.StartTime
		res.EndTime = pl.IdleBox.EndTime
		ctx.Send(res)
		return
	}

	bag.AddAward(ctx, pl, reward, true)

	now := time.Now().Unix()
	pl.IdleBox.StartTime = now
	pl.IdleBox.EndTime = now + int64(config.Global.Get().IdleBoxMaxTime*3600)

	res.StartTime = pl.IdleBox.StartTime
	res.EndTime = pl.IdleBox.EndTime

	//任务
	task.Dispatch(ctx, pl, define.TaskGetGuajiAwardTime, 1, 0, true)

	ctx.Send(res)
}

// ReqGetIdleBoxData 获取挂机宝箱数据
func ReqGetIdleBoxData(ctx global.IPlayer, pl *model.Player, req *proto_idlebox.C2SGetIdleBoxData) {
	res := &proto_idlebox.S2CGetIdleBoxData{}
	res.StartTime = pl.IdleBox.StartTime
	res.EndTime = pl.IdleBox.EndTime
	ctx.Send(res)
}

func calcReward(startTime, endTime int64, conf_stage conf2.IdleBox, conf_climb conf2.IdleBox) []conf2.ItemE {
	now := time.Now().Unix()
	t := now
	if t >= endTime {
		t = endTime
	}

	totalTime := t - startTime
	phase := 1 // 阶段序号
	hour := 0  //小时
	totalNums := make(map[int32]int32)

	for timeRemaining := totalTime; timeRemaining > 0; phase++ {
		phaseTime := int64(config.Global.Get().IdleBoxTime) // 每个阶段的时间
		if timeRemaining < phaseTime {
			phaseTime = timeRemaining
		}

		if phase > 360 {
			phase = 1
			hour += 1
		}

		// 计算当前阶段的奖励数量
		for _, v := range conf_stage.StageReward {
			if _, ok := totalNums[v.ItemId]; !ok {
				totalNums[v.ItemId] = 0
			}
			totalNums[v.ItemId] = totalNums[v.ItemId] + v.ItemNum + conf_stage.AddStageRewardNum[v.ItemId]*int32(hour)
		}

		for _, v := range conf_climb.TowerReward {
			if _, ok := totalNums[v.ItemId]; !ok {
				totalNums[v.ItemId] = 0
			}
			totalNums[v.ItemId] = totalNums[v.ItemId] + v.ItemNum + conf_climb.AddTowerRewardNum[v.ItemId]*int32(hour)
		}

		// 减去已计算的时间
		timeRemaining -= phaseTime
	}

	var reward []conf2.ItemE
	for k, v := range totalNums {
		reward = append(reward, conf2.ItemE{ItemId: k, ItemNum: v, ItemType: define.ItemTypeItem})
	}
	return reward
}
