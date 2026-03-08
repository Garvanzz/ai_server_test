package huaguoshan

import (
	"xfx/pkg/utils"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/player/bag"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/proto/proto_huaguoshan"
	"xfx/proto/proto_public"
)

// ReqStartMakeWine 开始酿酒
func ReqStartMakeWine(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SStartMakeWine) {
	resp := &proto_huaguoshan.S2CStartMakeWine{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Huaguoshan == nil || pl.Huaguoshan.Wine == nil {
		initHuaguoshanData(pl)
	}

	// 检查是否正在酿造
	if pl.Huaguoshan.Wine.CurMakingWineId > 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(resp)
		return
	}

	// 获取酒架配置
	wineRackConfs := config.WineRack.All()
	wineRackConf := wineRackConfs[int64(pl.Huaguoshan.Wine.CurWineRack)]
	if !wineRackConf.CanMakeWineType(req.Type) {
		log.Error("CanMakeWineType ： %v, %v", req.Type, wineRackConf.Type)
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 获取消耗材料
	costItems := wineRackConf.GetWineCostByType(req.Type)
	if len(costItems) == 0 {
		log.Error("costItems ： %v", costItems)
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 检查材料是否足够
	costs := make(map[int32]int32)
	for _, item := range costItems {
		costs[item.ItemId] = item.ItemNum
	}
	if !internal.CheckItemsEnough(pl, costs) {
		log.Error("CheckItemsEnough ： %v", costs)
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(resp)
		return
	}

	// 扣除材料
	internal.SubItems(ctx, pl, costs)

	// 开始酿造
	now := utils.Now().Unix()
	pl.Huaguoshan.Wine.CurMakingWineId = req.Id
	pl.Huaguoshan.Wine.CurMakingWineStarTime = int32(now)
	pl.Huaguoshan.Wine.CurMakingWineEndTime = int32(now + int64(wineRackConf.MakeTime))

	// 返回最新状态
	resp.Opt = pl.Huaguoshan.Wine.ToMakeWineOption()

	log.Debug("Player %d start make wine type=%d, endTime=%d", pl.Id, req.Type, pl.Huaguoshan.Wine.CurMakingWineEndTime)
	ctx.Send(resp)
}

// 切换酒架
func ReqCutMakeWine(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SCutWineRack) {
	resp := &proto_huaguoshan.S2CCutWineRack{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Huaguoshan == nil || pl.Huaguoshan.Wine == nil {
		initHuaguoshanData(pl)
	}

	// 检查是否正在酿造
	if pl.Huaguoshan.Wine.CurMakingWineId > 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(resp)
		return
	}

	if !utils.ContainsInt32(pl.Huaguoshan.Wine.OwerWineRack, req.RackId) {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	pl.Huaguoshan.Wine.CurWineRack = req.RackId

	// 返回最新状态
	resp.Opt = pl.Huaguoshan.Wine.ToMakeWineOption()
	resp.Code = proto_public.CommonErrorCode_ERR_OK

	log.Debug("Player %d cut, endTime=%d", pl.Id, pl.Huaguoshan.Wine.CurMakingWineEndTime)
	ctx.Send(resp)
}

// ReqCollectMakeWine 收集酿酒
func ReqCollectMakeWine(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SCollectWine) {
	resp := &proto_huaguoshan.S2CCollectWine{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Huaguoshan == nil || pl.Huaguoshan.Wine == nil {
		initHuaguoshanData(pl)
	}

	// 检查是否正在酿造
	if pl.Huaguoshan.Wine.CurMakingWineId <= 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	//检查时间
	if pl.Huaguoshan.Wine.CurMakingWineId > 0 {
		now := utils.Now().Unix()
		// 如果已完成，自动收取
		if now >= int64(pl.Huaguoshan.Wine.CurMakingWineEndTime) {
			autoCollectWine(ctx, pl)
		} else {
			// 正在酿造中
			resp.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
			ctx.Send(resp)
			return
		}
	}

	// 返回最新状态
	resp.Opt = pl.Huaguoshan.Wine.ToMakeWineOption()

	log.Debug("Player %d start , endTime=%d", pl.Id, pl.Huaguoshan.Wine.CurMakingWineEndTime)
	ctx.Send(resp)
}

// autoCollectWine 自动收取酿造完成的酒
func autoCollectWine(ctx global.IPlayer, pl *model.Player) {
	if pl.Huaguoshan.Wine.CurMakingWineId == 0 {
		return
	}

	wineItemId := pl.Huaguoshan.Wine.CurMakingWineId
	awards := []conf.ItemE{
		conf.ItemE{
			ItemType: define.ItemTypeItem,
			ItemNum:  1,
			ItemId:   pl.Huaguoshan.Wine.CurMakingWineId,
		},
	}

	bag.AddAward(ctx, pl, awards, true)

	// 清空酿造状态
	pl.Huaguoshan.Wine.CurMakingWineId = 0
	pl.Huaguoshan.Wine.CurWineRack = 0
	pl.Huaguoshan.Wine.CurMakingWineStarTime = 0
	pl.Huaguoshan.Wine.CurMakingWineEndTime = 0

	log.Debug("Player %d auto collect wine, wineId=%d", pl.Id, wineItemId)
}
