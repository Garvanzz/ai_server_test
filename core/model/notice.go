package model

type NoticeOpt struct {
	Id         int64
	Channel    int32
	ServerId   int32
	Content    string
	ExpireTime int64
	EffectTime int64
}

type HorseOpt struct {
	Channel   int32
	ServerId  int32
	Content   string
	VaildTime int32
	Scene     int32
	Priority  int32
}
