package huaguoshan

import (
	"encoding/json"
	"fmt"
	"xfx/core/config/conf"
	"xfx/core/db"
	"xfx/core/define"
	"xfx/core/model"
	"xfx/main_server/global"
	"xfx/pkg/log"
	"xfx/pkg/utils"
	"xfx/proto/proto_huaguoshan"
	"xfx/proto/proto_public"
)

func Init(pl *model.Player) {
	pl.Huaguoshan = &model.Huaguoshan{
		Partner: &model.HuaguoshanPartner{
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
		},
		Wine: &model.HuaguoshanWine{
			CurMakingWineId:       0,
			CurMakingWineStarTime: 0,
			CurMakingWineEndTime:  0,
			CurWineRack:           101,
			OwerWineRack:          []int32{101},
		},
		Peach: &model.HuaguoshanPeach{
			CurTreeId:              0,
			CurPlantPeachStage:     0,
			CurPlantPeachStartTime: 0,
			CurPlantPeachEndTime:   0,
			OwerTreeId:             []int32{201},
			Awards:                 make([]conf.ItemE, 0),
		},
	}
}

func Save(pl *model.Player, isSync bool) {
	j, err := json.Marshal(pl.Huaguoshan)
	if err != nil {
		log.Error("player[%v],save huaguoshan marshal error:%v", pl.Id, err)
		return
	}

	if isSync {
		rdb, err := db.GetEngineByPlayerId(pl.Id)
		if err != nil {
			log.Error("save huaguoshan error, no this server:%v", err)
			return
		}
		rdb.RedisExec("SET", fmt.Sprintf("%s:%d", define.PlayerHuaguoshan, pl.Id), j)
	} else {
		// TODO: 异步存储
		//global.ServerG.GetDBEngine().Request(p, EVENTYPE_DB_RET_SET_SHOP, int64(0), "SET", fmt.Sprintf("shop:%d", p.dbId), j)
	}
}

func Load(pl *model.Player) {
	rdb, err := db.GetEngineByPlayerId(pl.Id)
	if err != nil {
		log.Error("save huaguoshan error, no this server:%v", err)
		return
	}
	reply, err := rdb.RedisExec("GET", fmt.Sprintf("%s:%d", define.PlayerHuaguoshan, pl.Id))
	if err != nil {
		log.Error("player[%v],load huaguoshan error:%v", pl.Id, err)
		return
	}

	if reply == nil {
		Init(pl)
		return
	}

	m := new(model.Huaguoshan)
	err = json.Unmarshal(reply.([]byte), &m)
	if err != nil {
		log.Error("player[%v],load huaguoshan unmarshal error:%v", pl.Id, err)
	}
	pl.Huaguoshan = m

	// TODO:load new tasks
}

// ReqInitHuaguoshan 初始化花果山
func ReqInitHuaguoshan(ctx global.IPlayer, pl *model.Player, req *proto_huaguoshan.C2SInitHuaguoshan) {
	resp := &proto_huaguoshan.S2CInitHuaguoshan{}

	// 初始化花果山数据(如果不存在)
	if pl.Huaguoshan == nil {
		pl.Huaguoshan = &model.Huaguoshan{
			Partner: &model.HuaguoshanPartner{
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
			},
			Wine: &model.HuaguoshanWine{
				CurMakingWineId:       0,
				CurMakingWineStarTime: 0,
				CurMakingWineEndTime:  0,
				CurWineRack:           101,
				OwerWineRack:          []int32{101},
			},
			Peach: &model.HuaguoshanPeach{
				CurTreeId:              0,
				CurPlantPeachStage:     0,
				CurPlantPeachStartTime: 0,
				CurPlantPeachEndTime:   0,
				OwerTreeId:             []int32{201},
				Awards:                 make([]conf.ItemE, 0),
			},
		}
	}

	// 返回伴侣信息
	resp.HasPartner = pl.Huaguoshan.Partner.PartnerId > 0
	if resp.HasPartner {
		// 查询伴侣玩家信息
		partnerInfo := getPlayerInfo(ctx, pl.Huaguoshan.Partner.PartnerId)
		resp.PartnerInfo = partnerInfo
	}

	// 返回酿酒数据
	resp.Wine = pl.Huaguoshan.Wine.ToMakeWineOption()

	// 检查是否完成当前阶段，如果完成则自动进入下一阶段
	checkAndAdvanceStage(pl)

	// 返回种桃数据
	resp.Peach = pl.Huaguoshan.Peach.ToPlantPeachOption()
	log.Debug("初始花果山数据:%v", resp)
	ctx.Send(resp)
}

// getPlayerInfo 获取玩家信息
func getPlayerInfo(ctx global.IPlayer, playerId int64) *proto_public.CommonPlayerInfo {
	if playerId == 0 {
		return nil
	}

	mod := global.GetPlayerInfo(playerId)
	return mod.ToCommonPlayer()
}

// checkDailyReset 检查每日重置
func checkDailyReset(partner *model.HuaguoshanPartner) {
	now := utils.Now()
	if partner.LastRelieveTime > 0 {
		if !utils.CheckIsSameDayBySec(partner.LastGiveResetTime, now.Unix(), 0) {
			partner.GiveCount = 0
			partner.LastGiveResetTime = now.Unix()
		}
	}
}
