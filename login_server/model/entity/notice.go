package entity

// NoticeItem 公告表 notice
type NoticeItem struct {
	Id         int64  `json:"id"`
	Channel    int32  `json:"channel"`
	ServerId   int32  `json:"serverId"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	ExpireTime int64  `json:"expireTime"`
	EffectTime int64  `json:"effectTime"`
}

func (NoticeItem) TableName() string { return "notice" }
