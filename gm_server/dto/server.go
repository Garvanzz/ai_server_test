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

// GameServerItem 游戏服表行（与 DB 映射）
type GameServerItem struct {
	Id         int64  `json:"id" xorm:"pk autoincr"`
	Ip         string `json:"ip" xorm:"ip"`
	Port       int    `json:"port" xorm:"port"`
	ServerName string `json:"serverName" xorm:"server_name"`
	ExeName    string `json:"exeName" xorm:"exe_name"`
	ExePath    string `json:"exePath" xorm:"exe_path"`
	RedisPort  int    `json:"redisPort" xorm:"redis_port"`
	MysqlAddr  string `json:"mysqlAddr" xorm:"mysql_addr"`
}

// GmSetServerTime 设置服务器时间请求
type GmSetServerTime struct {
	Time    string `json:"time"`
	SetTime string `json:"settime"`
	Server  int32  `json:"server"`
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
