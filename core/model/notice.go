package model

// NoticeItem 公告表 notice
type NoticeItem struct {
	Id         int64  `json:"id"`
	Channel    int32  `json:"channel"`
	ServerId   int32  `json:"serverId"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	ExpireTime int64  `json:"expireTime"`
	EffectTime int64  `json:"effectTime"`
	//IsImmediately bool
}

type NoticeOpt struct {
	Id         int64
	Channel    int32
	ServerId   int32
	Content    string
	ExpireTime int64
	EffectTime int64
	//Title      string
}

// HorseOpt 跑马灯参数（与 main_server GM 接口一致）
type HorseOpt struct {
	Channel   int32
	ServerId  int32
	Content   string
	VaildTime int32
	Scene     int32
	Priority  int32
}

// HorseItem 跑马灯请求体（GM 后台用，字段与 HorseOpt 一致）
type HorseItem struct {
	Channel   int32  `json:"channel"`
	ServerId  int32  `json:"serverId"`
	Content   string `json:"content"`
	VaildTime int32  `json:"vaildTime"`
	Scene     int32  `json:"scene"`
	Priority  int32  `json:"priority"`
}
