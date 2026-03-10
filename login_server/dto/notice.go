package dto

// ReqNoticeList 获取公告列表请求
type ReqNoticeList struct {
	Channel  int `json:"channel"`
	ServerId int `json:"serverId"`
}
