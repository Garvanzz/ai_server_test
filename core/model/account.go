package model

import "time"

type Account struct {
	Id             int64     `json:"id"`
	Uid            string    `json:"uid"`      //用户uid
	Account        string    `json:"account"`  //账号
	Password       string    `json:"password"` //密码
	Type           int       `json:"type"`
	NickName       string    `json:"nickName"`
	CreateTime     time.Time `json:"createTime"`
	OnlineTime     time.Time `json:"onlineTime"`
	OfflineTime    time.Time `json:"offlineTime"`
	DeviceId       string    `json:"deviceId"`
	IsWhiteAcc     int       `json:"isWhiteAcc"` //白名单账号 0不是 1是
	LoginBan       int64     `json:"loginBan"`   //登录封禁 1/封禁中
	LoginBanReason string    `json:"loginBanReason"`
	Platform       int       `json:"platform"`  //平台 1pc 2ios 3安卓
	RedisId        int64     `json:"redisId"`   //dbId
	LastToken      string    `json:"lastToken"` //上次使用token
	SystemMailId   int64     `json:"systemMailId"`
	ChatBan        int64     `json:"chatBan"`       //聊天封禁 是否被ban 0没有 其他是具体时间戳
	ChatBanReason  string    `json:"chatBanReason"` //聊天封禁原因
	ServerId       int       `json:"serverId"`      //服务器ID
}

// ServerItem 服务器列表信息
type ServerItem struct {
	Id             int64  `json:"id"`
	Channel        int    `json:"channel"` //渠道
	Ip             string `json:"ip"`
	Port           int    `json:"port"`
	RedisPort      int    `json:"redisPort"`      //redis账户端口
	MysqlAddr      string `json:"mysqlAddr"`      // mysql地址
	ServerState    int    `json:"serverState"`    //0：正常 1：拥挤 2：爆满 3：维护 4：未开服 5：停服
	OpenServerTime int64  `json:"openServerTime"` //开服时间
	StopServerTime int64  `json:"stopServerTime"` //停服时间
	ServerName     string `json:"serverName"`     //服务器名字
	LoginServerUrl string `json:"loginServerUrl"` //登录服务器
	ServerGroup    int    `json:"server_group"`   //服务器组
}
