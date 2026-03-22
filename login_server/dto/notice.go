package dto

import "xfx/core/model"

// NoticeListRequest 获取公告列表请求。
type NoticeListRequest struct {
	Channel  int `json:"channel"`
	ServerID int `json:"serverId"`
}

// NoticeListResponse 获取公告列表响应。
type NoticeListResponse struct {
	Notices []model.NoticeItem `json:"notices"`
}
