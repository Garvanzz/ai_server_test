package model

// 热更列表信息
type HotUpdateItem struct {
	Id          int64  `json:"id"`
	Channel     string `json:"channel"`
	Version     string `json:"version"`
	ChannelName string `json:"channelName"`
}
