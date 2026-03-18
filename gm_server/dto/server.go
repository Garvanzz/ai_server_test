package dto

// GMRespHotUpdateItem 热更列表项响应
type GMRespHotUpdateItem struct {
	Id          int64
	Channel     string
	Version     string
	ChannelName string
}

// HotUpdateItem 热更表行（与 DB 映射）
type HotUpdateItem struct {
	Id          int64  `json:"id" xorm:"pk autoincr"`
	Channel     string `json:"channel" xorm:"channel"`
	ChannelName string `json:"channelName" xorm:"channel_name"`
	Version     string `json:"version" xorm:"version"`
}

// GmSetServerTime 设置服务器时间请求
type GmSetServerTime struct {
	Time         string `json:"time"`
	SetTime      string `json:"settime"`
	SetTimeCamel string `json:"setTime"`
	Server       int32  `json:"server"`
	ServerID     int32  `json:"serverId"`
}

type GMServerTimePayload struct {
	ServerID      int64  `json:"serverId"`
	Time          string `json:"time"`
	ServerTime    string `json:"serverTime"`
	GameTime      int64  `json:"gameTime"`
	GameIso       string `json:"gameIso"`
	RealTime      int64  `json:"realTime"`
	RealIso       string `json:"realIso"`
	OffsetDays    int64  `json:"offsetDays"`
	OffsetEnabled bool   `json:"offsetEnabled"`
}

// GmHotUpdateVersionReq 编辑/创建热更版本请求
type GmHotUpdateVersionReq struct {
	Channel     string `json:"channel"`
	ChannelName string `json:"channelName"`
	Version     string `json:"version"`
}

// GmHotUpdateDeleteReq 删除热更版本请求（按 channel 列表删除）
type GmHotUpdateDeleteReq struct {
	Channels []string `json:"channel"` // 或 "channels"，前端传数组
}

// GmHotUpdatePathReq 创建热更路径请求（仅创建目录，不写表）
type GmHotUpdatePathReq struct {
	Channel string `json:"channel"`
	Version string `json:"version"`
}

type GmServerManageUpsertReq struct {
	Id                int64  `json:"id"`
	Channel           int    `json:"channel"`
	GroupId           int    `json:"groupId"`
	LogicServerId     int64  `json:"logicServerId"`
	Ip                string `json:"ip"`
	Port              int    `json:"port"`
	MainServerHttpUrl string `json:"mainServerHttpUrl"`
	ServerState       int    `json:"serverState"`
	OpenServerTime    string `json:"openServerTime"`
	StopServerTime    string `json:"stopServerTime"`
	ServerName        string `json:"serverName"`
	ManageMode        string `json:"manageMode"`
	ProcessName       string `json:"processName"`
	StartCommand      string `json:"startCommand"`
	WorkDir           string `json:"workDir"`
	ExeName           string `json:"exeName"`
	ExePath           string `json:"exePath"`
}

type GmServerManageDeleteReq struct {
	Ids []int64 `json:"ids"`
}

type GmServerBatchUpdateReq struct {
	Ids           []int64 `json:"ids"`
	GroupId       *int    `json:"groupId"`
	LogicServerId *int64  `json:"logicServerId"`
	ServerState   *int    `json:"serverState"`
}

type GmServerGroupUpsertReq struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
	GroupType int    `json:"groupType"`
	IsVisible int    `json:"isVisible"`
}

type GmServerGroupDeleteReq struct {
	Ids []int64 `json:"ids"`
}

type GmServerGroupManageItem struct {
	Id            int64  `json:"id"`
	Name          string `json:"name"`
	SortOrder     int    `json:"sortOrder"`
	GroupType     int    `json:"groupType"`
	GroupTypeText string `json:"groupTypeText"`
	IsVisible     int    `json:"isVisible"`
	ServerCount   int64  `json:"serverCount"`
}
