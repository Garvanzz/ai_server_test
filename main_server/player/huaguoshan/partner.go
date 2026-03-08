package huaguoshan

import (
	"fmt"
	"xfx/core/config"
	"xfx/core/config/conf"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/main_server/invoke"
	"xfx/main_server/player/internal"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_huaguoshan"
	"xfx/proto/proto_public"
)

// ReqInitPartner 获取伴侣详情
func ReqInitPartner(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SInitPartner) {
	resp := &proto_huaguoshan.S2CInitPartner{}

	// 初始化数据
	if pl.Huaguoshan == nil || pl.Huaguoshan.Partner == nil {
		initHuaguoshanData(pl)
	}

	// 检查每日重置
	checkDailyReset(pl.Huaguoshan.Partner)

	// 获取伴侣信息
	var partnerInfo *proto_public.CommonPlayerInfo
	if pl.Huaguoshan.Partner.PartnerId > 0 {
		partnerInfo = getPlayerInfo(ctx, pl.Huaguoshan.Partner.PartnerId)
	}
	log.Debug("pl.Huaguoshan.Partner: %v", pl.Huaguoshan.Partner)
	resp.Option = pl.Huaguoshan.Partner.ToPartnerOption(partnerInfo)
	ctx.Send(resp)
}

// ReqGetPartnerInviteList 获取邀请列表
func ReqGetPartnerInviteList(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SGetPartnerInviteList) {
	resp := &proto_huaguoshan.S2CGetPartnerInviteList{}

	// 从logic层获取邀请列表
	invites := invoke.HuaguoshanClient(ctx).GetReceiverInvites(pl.Id)

	resp.List = make([]*proto_huaguoshan.PartnerInviteOption, 0)
	for _, invite := range invites {
		// 获取发送者信息
		sendInfo := getPlayerInfo(ctx, invite.SenderId)
		resp.List = append(resp.List, invite.ToPartnerInviteOption(sendInfo))
	}

	ctx.Send(resp)
}

// ReqPartnerInvite 邀请伴侣
func ReqPartnerInvite(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SPartnerInvite) {
	resp := &proto_huaguoshan.S2CPartnerInvite{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Huaguoshan == nil || pl.Huaguoshan.Partner == nil {
		initHuaguoshanData(pl)
	}

	// 校验: 自己没有伴侣
	if pl.Huaguoshan.Partner.PartnerId > 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ALHASPATERNER
		ctx.Send(resp)
		return
	}

	// 校验: 自己不在冷却期
	if isInCooldown(pl.Huaguoshan.Partner) {
		resp.Code = proto_public.CommonErrorCode_ERR_InCooldown
		ctx.Send(resp)
		return
	}

	// 校验: 不能邀请自己
	if req.Id == pl.Id {
		resp.Code = proto_public.CommonErrorCode_ERR_CannotInviteSelf
		ctx.Send(resp)
		return
	}

	costs := make(map[int32]int32, 0)
	cost := config.Global.Get().PaternerInviteCost
	costs[cost[0].ItemId] = cost[0].ItemNum
	//判断消耗
	if !internal.CheckItemsEnough(pl, costs) {
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(resp)
		return
	}

	invite := invoke.HuaguoshanClient(ctx).CreateInvite(pl.Id, pl.Base.Name, req.Id)
	if invite == nil {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	//扣除
	internal.SubItems(ctx, pl, costs)

	log.Debug("Player %d invite partner %d, inviteId: %d", pl.Id, req.Id, invite.Id)
	ctx.Send(resp)
}

// ReqLogicPartnerInvite 处理邀请
func ReqLogicPartnerInvite(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SLogicPartnerInvite) {
	resp := &proto_huaguoshan.S2CLogicPartnerInvite{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Huaguoshan == nil || pl.Huaguoshan.Partner == nil {
		initHuaguoshanData(pl)
	}

	// 校验: 自己没有伴侣
	if pl.Huaguoshan.Partner.PartnerId > 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ALHASPATERNER
		ctx.Send(resp)
		return
	}

	// 校验: 自己不在冷却期
	if isInCooldown(pl.Huaguoshan.Partner) {
		resp.Code = proto_public.CommonErrorCode_ERR_InCooldown
		ctx.Send(resp)
		return
	}

	// 处理邀请
	invite, err := invoke.HuaguoshanClient(ctx).ProcessInvite(req.Id, req.Status)
	if err != nil {
		log.Error("ProcessInvite error: %v", err)
		resp.Code = proto_public.CommonErrorCode_ERR_InviteInvalid
		ctx.Send(resp)
		return
	}

	// 如果拒绝，直接返回
	if !req.Status {
		ctx.Send(resp)
		return
	}

	// 同意邀请: 建立伴侣关系
	// 设置自己的伴侣信息
	pl.Huaguoshan.Partner.PartnerId = invite.SenderId
	pl.Huaguoshan.Partner.Intimacy = 0
	pl.Huaguoshan.Partner.IntimacyLevel = 0
	pl.Huaguoshan.Partner.PartnerType = 0
	pl.Huaguoshan.Partner.LastRelieveTime = 0
	pl.Huaguoshan.Partner.GiveCount = 0
	pl.Huaguoshan.Partner.LastGiveResetTime = utils.Now().Unix()
	pl.Huaguoshan.Partner.UnlockedSkills = []int32{}
	pl.Huaguoshan.Partner.UnlockedBraces = []int32{}
	pl.Huaguoshan.Partner.UnlockedMounts = []int32{}
	pl.Huaguoshan.Partner.UnlockedHeadWears = []int32{}
	pl.Huaguoshan.Partner.UnlockedBuffs = []int32{}
	pl.Huaguoshan.Partner.UnlockedHeadFrames = []int32{}

	// 返回伴侣信息
	partnerInfo := getPlayerInfo(ctx, pl.Huaguoshan.Partner.PartnerId)
	resp.Option = pl.Huaguoshan.Partner.ToPartnerOption(partnerInfo)

	log.Debug("Player %d accept partner invite from %d", pl.Id, invite.SenderId)
	ctx.Send(resp)
}

// ReqRelievePartner 解除伴侣
func ReqRelievePartner(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SRelieveParterner) {
	resp := &proto_huaguoshan.S2CRelieveParterner{
		Code:       proto_public.CommonErrorCode_ERR_OK,
		HasPartner: false,
	}

	// 初始化数据
	if pl.Huaguoshan == nil || pl.Huaguoshan.Partner == nil {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 校验: 有伴侣
	if pl.Huaguoshan.Partner.PartnerId == 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	partnerId := pl.Huaguoshan.Partner.PartnerId

	// 清空伴侣数据
	pl.Huaguoshan.Partner.PartnerId = 0
	pl.Huaguoshan.Partner.Intimacy = 0
	pl.Huaguoshan.Partner.IntimacyLevel = 0
	pl.Huaguoshan.Partner.PartnerType = 0
	pl.Huaguoshan.Partner.LastRelieveTime = utils.Now().Unix()
	pl.Huaguoshan.Partner.GiveCount = 0
	pl.Huaguoshan.Partner.UnlockedSkills = []int32{}
	pl.Huaguoshan.Partner.UnlockedBraces = []int32{}
	pl.Huaguoshan.Partner.UnlockedMounts = []int32{}
	pl.Huaguoshan.Partner.UnlockedHeadWears = []int32{}
	pl.Huaguoshan.Partner.UnlockedBuffs = []int32{}
	pl.Huaguoshan.Partner.UnlockedHeadFrames = []int32{}

	log.Debug("Player %d relieve partner %d", pl.Id, partnerId)
	ctx.Send(resp)
}

// ReqPartnerGive 赠送礼物
func ReqPartnerGive(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SParternerGive) {
	resp := &proto_huaguoshan.S2CParternerGive{
		Code: proto_public.CommonErrorCode_ERR_OK,
	}

	// 初始化数据
	if pl.Huaguoshan == nil || pl.Huaguoshan.Partner == nil {
		resp.Code = proto_public.CommonErrorCode_ERR_ParamTypeError
		ctx.Send(resp)
		return
	}

	// 校验: 有伴侣
	if pl.Huaguoshan.Partner.PartnerId == 0 {
		resp.Code = proto_public.CommonErrorCode_ERR_NoPlayer
		ctx.Send(resp)
		return
	}

	// 检查每日重置
	checkDailyReset(pl.Huaguoshan.Partner)

	// 校验: 今日赠送次数
	day := config.Global.Get().PaternerGiveCount
	if pl.Huaguoshan.Partner.GiveCount >= day {
		resp.Code = proto_public.CommonErrorCode_ERR_OutPutLimit
		ctx.Send(resp)
		return
	}

	cost := make(map[int32]int32)
	cost[req.Id] = req.Count

	if !internal.CheckItemsEnough(pl, cost) {
		resp.Code = proto_public.CommonErrorCode_ERR_NumNotEnough
		ctx.Send(resp)
		return
	}

	addIntimacys := config.Global.Get().GiftRaceAddIntimacy
	// 校验并扣除消耗
	itemConf, _ := config.Item.Find(int64(req.Id))
	addIntimacy := addIntimacys[itemConf.Rare-4] * req.Count

	// 增加自己的亲密度
	pl.Huaguoshan.Partner.Intimacy += addIntimacy

	// 检查是否升级
	checkIntimacyLevelUp(pl)

	// 增加赠送次数
	pl.Huaguoshan.Partner.GiveCount++

	// 给对方发送邮件奖励
	if len(cost) > 0 {
		costItem := []conf.ItemE{
			conf.ItemE{
				ItemId:   req.Id,
				ItemNum:  1,
				ItemType: define.ItemTypeItem,
			},
		}
		invoke.MailClient(ctx).SendMail(
			define.PlayerMail,
			"伴侣赠送",
			fmt.Sprintf("%s赠送了您礼物，请查收", pl.Base.Name),
			"", "", "系统",
			costItem,
			[]int64{pl.Huaguoshan.Partner.PartnerId},
			0, 0, false, []string{},
		)
	}

	resp.GiveCount = pl.Huaguoshan.Partner.GiveCount
	resp.Intimacy = pl.Huaguoshan.Partner.Intimacy
	resp.IntimacyLevel = pl.Huaguoshan.Partner.IntimacyLevel

	log.Debug("Player %d give gift %d to partner %d, intimacy: %d", pl.Id, req.Id, pl.Huaguoshan.Partner.PartnerId, pl.Huaguoshan.Partner.Intimacy)
	ctx.Send(resp)
}

// initHuaguoshanData 初始化花果山数据
func initHuaguoshanData(pl *model.Player) {
	if pl.Huaguoshan == nil {
		pl.Huaguoshan = &model.Huaguoshan{}
	}
	if pl.Huaguoshan.Partner == nil {
		pl.Huaguoshan.Partner = &model.HuaguoshanPartner{
			PartnerId:          0,
			Intimacy:           0,
			IntimacyLevel:      0,
			PartnerType:        0,
			LastRelieveTime:    0,
			GiveCount:          0,
			LastGiveResetTime:  0,
			CurStageId:         0,
			UnlockedSkills:     []int32{},
			UnlockedBraces:     []int32{},
			UnlockedMounts:     []int32{},
			UnlockedHeadWears:  []int32{},
			UnlockedBuffs:      []int32{},
			UnlockedHeadFrames: []int32{},
		}
	}
	if pl.Huaguoshan.Wine == nil {
		pl.Huaguoshan.Wine = &model.HuaguoshanWine{
			CurMakingWineId:       0,
			CurMakingWineStarTime: 0,
			CurMakingWineEndTime:  0,
			CurWineRack:           101,
			OwerWineRack:          []int32{101},
		}
	}
	if pl.Huaguoshan.Peach == nil {
		pl.Huaguoshan.Peach = &model.HuaguoshanPeach{
			CurTreeId:              0,
			CurPlantPeachStage:     0,
			CurPlantPeachStartTime: 0,
			CurPlantPeachEndTime:   0,
			OwerTreeId:             []int32{201},
			Awards:                 make([]conf.ItemE, 0),
		}
	}
}

// isInCooldown 是否在冷却期
func isInCooldown(partner *model.HuaguoshanPartner) bool {
	if partner.LastRelieveTime == 0 {
		return false
	}
	now := utils.Now().Unix()
	day := config.Global.Get().PaternerCoolDown
	cooldownSeconds := int64(day * 24 * 3600)
	return now < partner.LastRelieveTime+cooldownSeconds
}

// checkIntimacyLevelUp 检查亲密度升级
func checkIntimacyLevelUp(pl *model.Player) {
	// 获取所有亲密度配置
	intimacyConfigs := config.ParternerIntimacy.All()
	if len(intimacyConfigs) == 0 {
		return
	}

	// 找到当前应该达到的等级
	newLevel := int32(0)
	for _, cfg := range intimacyConfigs {
		if pl.Huaguoshan.Partner.Intimacy >= cfg.Exp {
			if cfg.Stage > newLevel {
				newLevel = cfg.Stage
			}
		}
	}

	// 如果等级有变化，解锁对应奖励
	if newLevel > pl.Huaguoshan.Partner.IntimacyLevel {
		oldLevel := pl.Huaguoshan.Partner.IntimacyLevel
		pl.Huaguoshan.Partner.IntimacyLevel = newLevel

		// 解锁从旧等级+1到新等级之间的所有奖励
		for level := oldLevel + 1; level <= newLevel; level++ {
			unlockRewards(pl, level, intimacyConfigs)
		}

		log.Debug("Player %d intimacy level up: %d -> %d", pl.Id, oldLevel, newLevel)
	}
}

// unlockRewards 解锁指定等级的奖励
func unlockRewards(pl *model.Player, level int32, configs map[int64]conf.ParternerIntimacy) {
	var cfg *conf.ParternerIntimacy
	for _, c := range configs {
		if c.Stage == level {
			cfg = &c
			break
		}
	}

	if cfg == nil {
		return
	}

	// 解锁BUFF
	if cfg.UnLockBuffValue > 0 {
		pl.Huaguoshan.Partner.UnlockedBuffs = append(pl.Huaguoshan.Partner.UnlockedBuffs, cfg.UnLockBuffValue)
	}

	// 解锁头像框
	if cfg.UnLockHeadFrameValue > 0 {
		pl.Huaguoshan.Partner.UnlockedHeadFrames = append(pl.Huaguoshan.Partner.UnlockedHeadFrames, cfg.UnLockHeadFrameValue)
	}

	// 解锁头饰
	if cfg.UnLockHeadWearValue > 0 {
		pl.Huaguoshan.Partner.UnlockedHeadWears = append(pl.Huaguoshan.Partner.UnlockedHeadWears, cfg.UnLockHeadWearValue)
	}

	// 解锁背饰
	if cfg.UnLockBraceValue > 0 {
		pl.Huaguoshan.Partner.UnlockedBraces = append(pl.Huaguoshan.Partner.UnlockedBraces, cfg.UnLockBraceValue)
	}

	// 解锁坐骑
	if cfg.UnLockMountValue > 0 {
		pl.Huaguoshan.Partner.UnlockedMounts = append(pl.Huaguoshan.Partner.UnlockedMounts, cfg.UnLockMountValue)
	}

	// 解锁技能
	if cfg.UnLockSkillValue > 0 {
		pl.Huaguoshan.Partner.UnlockedSkills = append(pl.Huaguoshan.Partner.UnlockedSkills, cfg.UnLockSkillValue)
	}
}
