package model

// 请求公告
type ReqNoticeList struct {
	Channel  int
	ServerId int
}

type NoticeItem struct {
	Id         int64
	Channel    int32
	ServerId   int32
	Title      string
	Content    string
	ExpireTime int64
	EffectTime int64
}
