package gm_model

// 大厅服务器列表信息
type ServerItem struct {
	Id             int64  `json:"id"`
	Ip             string `json:"ip"`
	Port           int    `json:"port"`
	Channel        int    `json:"channel"`        //渠道
	RedisPort      int    `json:"redisPort"`      //redis账户端口
	MysqlAddr      string `json:"mysqlAddr"`      // mysql地址
	ServerState    int    `json:"serverState"`    //0：正常 1：拥挤 2：爆满 3：维护 4：未开服 5：停服
	OpenServerTime int64  `json:"openServerTime"` //开服时间
	StopServerTime int64  `json:"stopServerTime"` //停服时间
	ServerName     string `json:"serverName"`     //服务器名字
	LoginServerUrl string `json:"loginServerUrl"` //登录服务器
	ServerGroup    int    `json:"servergroup"`    //服务器组
	ExeName        string `json:"exeName"`        //可执行文件
	ExePath        string `json:"exePath"`        //可执行路径
	GameServer     int    `json:"gameServer"`     //游戏服
}

// 游戏服务器列表信息
type GameServerItem struct {
	Id         int64  `json:"id"`
	Ip         string `json:"ip"`
	Port       int    `json:"port"`
	RedisPort  int    `json:"redisPort"`  //redis账户端口
	MysqlAddr  string `json:"mysqlAddr"`  // mysql地址
	ServerName string `json:"serverName"` //服务器名字
	ExeName    string `json:"exeName"`    //可执行文件
	ExePath    string `json:"exePath"`    //可执行路径
}

// 热更列表信息
type HotUpdateItem struct {
	Id          int64  `json:"id"`
	Channel     string `json:"channel"`
	Version     string `json:"version"`
	ChannelName string `json:"channelName"`
}

type GMRespHotUpdateItem struct {
	Id          int64
	Channel     string
	Version     string
	ChannelName string
}

// 设置服务器时间
type GmSetServerTime struct {
	Time    string `json:"time"`
	SetTime string `json:"settime"`
	Server  int32  `json:"server"`
}
