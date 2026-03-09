package entity

import "time"

// Account 账号表 account
type Account struct {
	Id             int64     `json:"id"`
	Uid            string    `json:"uid"`
	Account        string    `json:"account"`
	Password       string    `json:"password"`
	Type           int       `json:"type"`
	NickName       string    `json:"nickName"`
	CreateTime     time.Time `json:"createTime"`
	OnlineTime     time.Time `json:"onlineTime"`
	OfflineTime    time.Time `json:"offlineTime"`
	DeviceId       string    `json:"deviceId"`
	IsWhiteAcc     int       `json:"isWhiteAcc"`
	LoginBan       int64     `json:"loginBan"`
	LoginBanReason string    `json:"loginBanReason"`
	Platform       int       `json:"platform"`
	RedisId        int64     `json:"redisId"`
	LastToken      string    `json:"lastToken"`
	SystemMailId   int64     `json:"systemMailId"`
	ChatBan        int64     `json:"chatBan"`
	ChatBanReason  string    `json:"chatBanReason"`
	ServerId       int       `json:"serverId"`
}
