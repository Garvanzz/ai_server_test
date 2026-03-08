package define

const (
	EventTypePlayerOnline = iota + 1
	EventTypePlayerOffline
	EventTypeActivity
	EventTypeConfigReload // 配置热更成功，各模块可据此重新拉取配置
)
