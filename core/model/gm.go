package model

// GMKickReq 踢人请求
type GMKickReq struct {
	PlayerId int64  `json:"player_id" binding:"required"`
	Reason   string `json:"reason"`
}

// GMGrantItemReq 发放道具请求（Items 格式同邮件附件）
type GMGrantItemReq struct {
	PlayerId int64      `json:"player_id" binding:"required"`
	Items    []MailItem `json:"items" binding:"required"`
}
