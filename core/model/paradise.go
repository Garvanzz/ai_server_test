package model

import (
	"xfx/core/config/conf"
	"xfx/proto/proto_huaguoshan"
	"xfx/proto/proto_public"
)

// Paradise 乐园数据(玩家维度)
type Paradise struct {
	Partner *ParadisePartner `redis:"partner" json:"partner"` // 伴侣数据
	Wine    *ParadiseWine    `redis:"wine" json:"wine"`       // 酿酒数据(后续实现)
	Peach   *ParadisePeach   `redis:"peach" json:"peach"`     // 种桃数据(后续实现)
}

// ParadisePartner 伴侣数据
type ParadisePartner struct {
	PartnerId          int64   `redis:"partner_id" json:"partner_id"`                     // 伴侣ID (0=无伴侣)
	Intimacy           int32   `redis:"intimacy" json:"intimacy"`                         // 当前亲密度
	IntimacyLevel      int32   `redis:"intimacy_level" json:"intimacy_level"`             // 亲密度等级
	PartnerType        int32   `redis:"partner_type" json:"partner_type"`                 // 伴侣类型
	LastRelieveTime    int64   `redis:"last_relieve_time" json:"last_relieve_time"`       // 上次解除时间(Unix时间戳)
	GiveCount          int32   `redis:"give_count" json:"give_count"`                     // 今日赠送次数
	LastGiveResetTime  int64   `redis:"last_give_reset_time" json:"last_give_reset_time"` // 上次重置时间
	CurStageId         int32   `redis:"cur_stage_id" json:"cur_stage_id"`                 // 当前副本ID(后续实现)
	UnlockedBuffs      []int32 `redis:"unlocked_buffs" json:"unlocked_buffs"`             // 已解锁BUFFID
	UnlockedSkills     []int32 `redis:"unlocked_skills" json:"unlocked_skills"`           // 已解锁SKILLID
	UnlockedHeadFrames []int32 `redis:"unlocked_headFrames" json:"unlocked_headFrames"`   // 已解锁头像框
	UnlockedHeadWears  []int32 `redis:"unlocked_headWears" json:"unlocked_headWears"`     // 已解锁
	UnlockedBraces     []int32 `redis:"unlocked_braces" json:"unlocked_braces"`           // 已解锁
	UnlockedMounts     []int32 `redis:"unlocked_mounts" json:"unlocked_mounts"`           // 已解锁
}

// ParadiseWine 酿酒数据(后续实现)
type ParadiseWine struct {
	CurMakingWineId       int32   `redis:"cur_making_wine_id" json:"cur_making_wine_id"`
	CurMakingWineStarTime int32   `redis:"cur_making_wine_star_time" json:"cur_making_wine_star_time"`
	CurMakingWineEndTime  int32   `redis:"cur_making_wine_end_time" json:"cur_making_wine_end_time"`
	CurWineRack           int32   `redis:"cur_wine_rack" json:"cur_wine_rack"`
	OwerWineRack          []int32 `redis:"ower_wine_rack" json:"ower_wine_rack"`
}

// ParadisePeach 种桃数据(后续实现)
type ParadisePeach struct {
	CurTreeId              int32        `redis:"cur_tree_id" json:"cur_tree_id"`
	CurPlantPeachStage     int32        `redis:"cur_plant_peach_stage" json:"cur_plant_peach_stage"`
	CurPlantPeachStartTime int64        `redis:"cur_plant_peach_start_time" json:"cur_plant_peach_start_time"`
	CurPlantPeachEndTime   int64        `redis:"cur_plant_peach_end_time" json:"cur_plant_peach_end_time"`
	OwerTreeId             []int32      `redis:"ower_tree_id" json:"ower_tree_id"`
	Awards                 []conf.ItemE `redis:"awards" json:"awards"`
}

// PartnerInvite 伴侣邀请数据(Redis全局)
type PartnerInvite struct {
	Id         int64  `json:"id"`          // 邀请ID(自增)
	SenderId   int64  `json:"sender_id"`   // 发送者ID
	SenderName string `json:"sender_name"` // 发送者名字
	ReceiverId int64  `json:"receiver_id"` // 接收者ID
	Status     int32  `json:"status"`      // 状态: 1=待处理, 2=已同意, 3=已拒绝
	CreateTime int64  `json:"create_time"` // 创建时间
	ExpireTime int64  `json:"expire_time"` // 过期时间(7天后)
}

// ToPartnerOption 转换为协议结构
func (p *ParadisePartner) ToPartnerOption(partnerInfo *proto_public.CommonPlayerInfo) *proto_huaguoshan.PartnerOption {
	opt := &proto_huaguoshan.PartnerOption{
		HasPartner:      p.PartnerId > 0,
		PartnerInfo:     partnerInfo,
		Intimacy:        p.Intimacy,
		PartnerType:     p.PartnerType,
		LastRelieveTime: p.LastRelieveTime,
		IntimacyLevel:   p.IntimacyLevel,
		GiveCount:       p.GiveCount,
		CurStageId:      p.CurStageId,
		GetBrace:        p.UnlockedBraces,
		GetBuffs:        p.UnlockedBuffs,
		GetHeadFrame:    p.UnlockedHeadFrames,
		GetHeadWear:     p.UnlockedHeadWears,
		GetMount:        p.UnlockedMounts,
		GetSkills:       p.UnlockedSkills,
	}
	return opt
}

// ToMakeWineOption 转换为酿酒协议(后续实现)
func (w *ParadiseWine) ToMakeWineOption() *proto_huaguoshan.MakeWineOption {
	if w == nil {
		return &proto_huaguoshan.MakeWineOption{}
	}
	return &proto_huaguoshan.MakeWineOption{
		CurMakingWineId:       w.CurMakingWineId,
		CurMakingWineStarTime: w.CurMakingWineStarTime,
		CurMakingWineEndTime:  w.CurMakingWineEndTime,
		CurWineRack:           w.CurWineRack,
		OwerWineRack:          w.OwerWineRack,
	}
}

// ToPlantPeachOption 转换为种桃协议(后续实现)
func (p *ParadisePeach) ToPlantPeachOption() *proto_huaguoshan.PlantPeachOption {
	if p == nil {
		return &proto_huaguoshan.PlantPeachOption{}
	}
	return &proto_huaguoshan.PlantPeachOption{
		CurTreeId:              p.CurTreeId,
		CurPlantPeachStage:     p.CurPlantPeachStage,
		CurPlantPeachStartTime: p.CurPlantPeachStartTime,
		CurPlantPeachEndTime:   p.CurPlantPeachEndTime,
		OwerTreeId:             p.OwerTreeId,
	}
}

// ToPartnerInviteOption 转换邀请为协议结构
func (inv *PartnerInvite) ToPartnerInviteOption(sendInfo *proto_public.CommonPlayerInfo) *proto_huaguoshan.PartnerInviteOption {
	return &proto_huaguoshan.PartnerInviteOption{
		Id:       inv.Id,
		SendInfo: sendInfo,
		Status:   inv.Status,
	}
}
