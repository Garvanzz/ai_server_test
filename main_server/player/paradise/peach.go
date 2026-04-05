package paradise

import (
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/main_server/player/task"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_huaguoshan"
	"xfx/proto/proto_public"
)

// ReqStartPlantPeach 种树相关操作
func ReqStartPlantPeach(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SStartPlantPeach) {
	resp := &proto_huaguoshan.S2CStartPlantPeach{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Paradise == nil || pl.Paradise.Peach == nil {
		initParadiseData(pl)
	}

	// 获取树配置
	peachTreeConfs := config.PeachTree.All()
	treeConf, exists := peachTreeConfs[int64(req.TreeId)]
	if !exists {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 检查是否拥有该树
	hasTree := false
	for _, id := range pl.Paradise.Peach.OwerTreeId {
		if id == req.TreeId {
			hasTree = true
			break
		}
	}
	if !hasTree {
		resp.Code = proto_public.CommonErrorCode_ERR_NoConfig
		ctx.Send(resp)
		return
	}

	// 检查是否已经在种植中
	if pl.Paradise.Peach.CurTreeId > 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(resp)
		return
	}

	// 开始第一阶段
	now := utils.Now().Unix()
	pl.Paradise.Peach.CurTreeId = req.TreeId
	pl.Paradise.Peach.CurPlantPeachStage = 1
	pl.Paradise.Peach.CurPlantPeachStartTime = now
	pl.Paradise.Peach.Awards = treeConf.Award
	pl.Paradise.Peach.CurPlantPeachEndTime = now + int64(treeConf.GetStageTime(1))

	// 返回最新状态
	resp.Option = pl.Paradise.Peach.ToPlantPeachOption()

	log.Debug("Player %d start plant peach treeId=%d, stage=1, endTime=%d", pl.Id, req.TreeId, pl.Paradise.Peach.CurPlantPeachEndTime)
	ctx.Send(resp)
}

// ReqLogicPlantPeach 种树相关操作
func ReqLogicPlantPeach(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SLogicPlantPeach) {
	if req.Type == 1 {
		// 浇水
		waterPeach(ctx, pl)
	} else if req.Type == 2 {
		// 施肥
		fertilizePeach(ctx, pl)
	} else if req.Type == 3 {
		// 收获
		collectPeach(ctx, pl)
	} else {
		resp := &proto_huaguoshan.S2CLogicPlantPeach{
			Code: proto_public.CommonErrorCode_ERR_ParamTypeError,
		}
		ctx.Send(resp)
	}
}

// waterPeach 浇水
func waterPeach(ctx global.IPlayer, pl *model.Player) {
	resp := &proto_huaguoshan.S2CLogicPlantPeach{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Paradise == nil || pl.Paradise.Peach == nil {
		initParadiseData(pl)
	}

	// 检查是否在种植中
	if pl.Paradise.Peach.CurTreeId == 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 获取树配置
	peachTreeConfs := config.PeachTree.All()
	treeConf, exists := peachTreeConfs[int64(pl.Paradise.Peach.CurTreeId)]
	if !exists {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 检查并扣除浇水消耗
	costs := make(map[int32]int32)
	for _, item := range treeConf.CoolDownneedCost {
		costs[item.ItemId] = item.ItemNum
	}
	if !internal.CheckItemsEnough(pl, costs) {
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(resp)
		return
	}

	internal.SubItems(ctx, pl, costs)

	// 缩短当前阶段时间
	now := utils.Now().Unix()
	newEndTime := pl.Paradise.Peach.CurPlantPeachEndTime - int64(treeConf.CoolDownTime)

	// 确保不会缩短到当前时间之前
	if newEndTime < now {
		newEndTime = now
	}

	pl.Paradise.Peach.CurPlantPeachEndTime = newEndTime

	// 检查是否完成当前阶段，如果完成则自动进入下一阶段
	checkAndAdvanceStageByConf(pl, &treeConf)

	// 返回最新状态
	resp.Opt = pl.Paradise.Peach.ToPlantPeachOption()
	resp.Code = proto_public.CommonErrorCode_ERR_OK

	//任务
	task.Dispatch(ctx, pl, define.TaskParadiseTreeWaterTime, 1, 0, true)

	log.Debug("Player %d water peach, new endTime=%d, stage=%d", pl.Id, pl.Paradise.Peach.CurPlantPeachEndTime, pl.Paradise.Peach.CurPlantPeachStage)
	ctx.Send(resp)
}

// fertilizePeach 施肥
func fertilizePeach(ctx global.IPlayer, pl *model.Player) {
	resp := &proto_huaguoshan.S2CLogicPlantPeach{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Paradise == nil || pl.Paradise.Peach == nil {
		initParadiseData(pl)
	}

	// 检查是否在种植中
	if pl.Paradise.Peach.CurTreeId == 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 获取树配置
	peachTreeConfs := config.PeachTree.All()
	treeConf, exists := peachTreeConfs[int64(pl.Paradise.Peach.CurTreeId)]
	if !exists {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 检查并扣除施肥消耗
	costs := make(map[int32]int32)
	for _, item := range treeConf.AddNumneedCost {
		costs[item.ItemId] = item.ItemNum
	}
	if !internal.CheckItemsEnough(pl, costs) {
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(resp)
		return
	}

	internal.SubItems(ctx, pl, costs)

	// 检查是否在成熟期(阶段3产量翻倍)
	if pl.Paradise.Peach.CurPlantPeachStage == 3 {
		for i := range pl.Paradise.Peach.Awards {
			pl.Paradise.Peach.Awards[i].ItemNum += treeConf.AddNum * 2
		}
	} else {
		for i := range pl.Paradise.Peach.Awards {
			pl.Paradise.Peach.Awards[i].ItemNum += treeConf.AddNum
		}
	}

	// 返回最新状态
	resp.Opt = pl.Paradise.Peach.ToPlantPeachOption()
	resp.Code = proto_public.CommonErrorCode_ERR_OK

	log.Debug("Player %d fertilize peach, treeId=%d", pl.Id, pl.Paradise.Peach.CurTreeId)
	ctx.Send(resp)
}

// collectPeach 收获
func collectPeach(ctx global.IPlayer, pl *model.Player) {
	resp := &proto_huaguoshan.S2CStartPlantPeach{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Paradise == nil || pl.Paradise.Peach == nil {
		initParadiseData(pl)
	}

	// 检查是否在种植中
	if pl.Paradise.Peach.CurTreeId == 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 检查时间是否已到
	now := utils.Now().Unix()
	if now < pl.Paradise.Peach.CurPlantPeachEndTime {
		resp.Code = proto_public.CommonErrorCode_ERR_LIMITTIME
		ctx.Send(resp)
		return
	}

	// 获取树配置
	peachTreeConfs := config.PeachTree.All()
	treeConf, exists := peachTreeConfs[int64(pl.Paradise.Peach.CurTreeId)]
	if !exists {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 发放奖励
	bag.AddAward(ctx, pl, pl.Paradise.Peach.Awards, true)

	// 清空种植状态
	pl.Paradise.Peach.CurTreeId = 0
	pl.Paradise.Peach.CurPlantPeachStage = 0
	pl.Paradise.Peach.CurPlantPeachStartTime = 0
	pl.Paradise.Peach.CurPlantPeachEndTime = 0
	pl.Paradise.Peach.Awards = make([]conf.ItemE, 0)

	// 返回最新状态
	resp.Option = pl.Paradise.Peach.ToPlantPeachOption()
	resp.Code = proto_public.CommonErrorCode_ERR_OK

	log.Debug("Player %d collect peach, treeId=%d", pl.Id, treeConf.Id)
	ctx.Send(resp)
}

// checkAndAdvanceStage 检查并推进阶段
func checkAndAdvanceStage(pl *model.Player) {
	// 检查是否在种植中
	if pl.Paradise.Peach.CurTreeId <= 0 {
		return
	}

	// 检查时间是否已到
	now := utils.Now().Unix()
	if now >= pl.Paradise.Peach.CurPlantPeachEndTime {
		return
	}

	// 获取树配置
	peachTreeConfs := config.PeachTree.All()
	treeConf, exists := peachTreeConfs[int64(pl.Paradise.Peach.CurTreeId)]
	if !exists {
		return
	}
	checkAndAdvanceStageByConf(pl, &treeConf)
}

// checkAndAdvanceStage 检查并推进阶段
func checkAndAdvanceStageByConf(pl *model.Player, treeConf *conf.PeachTree) {
	now := utils.Now().Unix()

	// 循环检查是否可以进入下一阶段
	for pl.Paradise.Peach.CurPlantPeachStage < 5 && now >= pl.Paradise.Peach.CurPlantPeachEndTime {
		// 进入下一阶段
		pl.Paradise.Peach.CurPlantPeachStage++
		pl.Paradise.Peach.CurPlantPeachStartTime = now

		nextStageTime := treeConf.GetStageTime(pl.Paradise.Peach.CurPlantPeachStage)
		pl.Paradise.Peach.CurPlantPeachEndTime = now + int64(nextStageTime)

		log.Debug("Player %d advance to stage %d, endTime=%d", pl.Id, pl.Paradise.Peach.CurPlantPeachStage, pl.Paradise.Peach.CurPlantPeachEndTime)
	}
}
