package model

import "encoding/json"

// GMKickReq 踢人请求
type GMKickReq struct {
	PlayerId int64  `json:"player_id" binding:"required"`
	Reason   string `json:"reason"`
}

// GMGrantItemReq 发放道具请求（Items 格式同邮件附件）
// GrantAll 为 true 时表示一键发放全部配置道具，此时 Items 可为空，由游戏服从 config 构建
type GMGrantItemReq struct {
	PlayerId int64      `json:"player_id" binding:"required"`
	Items    []MailItem `json:"items"`   // 单条/多条发放时必填；GrantAll 时可为空
	GrantAll bool       `json:"grant_all"` // 一键发放全部道具时由 gm_server 设为 true
}

// GMPlayerIdReq 仅需 player_id 的 GM 请求（查背包/装备/关卡/英雄/玩家游戏信息等）
type GMPlayerIdReq struct {
	PlayerId int64 `json:"player_id" binding:"required"`
}

// GMPlayerIdsReq 多玩家 ID 请求（如批量查玩家游戏信息）
type GMPlayerIdsReq struct {
	PlayerIds []int64 `json:"player_ids" binding:"required"`
}

// GMItemDeleteReq 删除背包道具请求
type GMItemDeleteReq struct {
	PlayerId int64   `json:"player_id" binding:"required"`
	ItemIds  []int32 `json:"item_ids" binding:"required"`
}

// GMEquipSetReq 设置装备数据请求（Data 为 Equip 的 JSON）
type GMEquipSetReq struct {
	PlayerId int64           `json:"player_id" binding:"required"`
	Data     json.RawMessage `json:"data" binding:"required"`
}

// GMEquipDeleteReq 删除装备请求（按装备唯一 ID 列表）
type GMEquipDeleteReq struct {
	PlayerId int64   `json:"player_id" binding:"required"`
	Ids      []int32 `json:"ids" binding:"required"`
}

// GMStageSetReq 设置关卡数据请求（Data 为 Stage 的 JSON）
type GMStageSetReq struct {
	PlayerId int64           `json:"player_id" binding:"required"`
	Data     json.RawMessage `json:"data" binding:"required"`
}

// GMHeroSetReq 设置英雄数据请求（Data 为 Hero 的 JSON）
type GMHeroSetReq struct {
	PlayerId int64           `json:"player_id" binding:"required"`
	Data     json.RawMessage `json:"data" binding:"required"`
}
