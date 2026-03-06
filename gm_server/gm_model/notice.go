package gm_model

type NoticeItem struct {
	Channel       int32
	ServerId      int32
	Title         string
	Content       string
	ExpireTime    int64
	EffectTime    int64
	IsImmediately bool
}

type NoticeOpt struct {
	Id         int64 `xorm:"'id' pk autoincr"`
	Channel    int32
	ServerId   int32
	Title      string
	Content    string
	ExpireTime int64
	EffectTime int64
}

type HorseItem struct {
	Channel   int32
	ServerId  int32
	Content   string
	VaildTime int32
	Scene     int32
	Priority  int32
}
