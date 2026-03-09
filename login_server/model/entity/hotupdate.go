package entity

// HotUpdateItem 热更表 hot_update
type HotUpdateItem struct {
	Id          int64  `json:"id"`
	Channel     string `json:"channel"`
	Version     string `json:"version"`
	ChannelName string `json:"channelName"`
}

func (HotUpdateItem) TableName() string { return "hot_update" }
