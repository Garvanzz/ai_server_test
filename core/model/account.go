package model

import "time"

type Account struct {
	Id                     int64     `json:"id"`
	Uid                    string    `json:"uid"`      // 用户 uid
	Account                string    `json:"account"`  // 账号
	Password               string    `json:"password"` // 密码
	Type                   int       `json:"type"`
	CreateTime             time.Time `json:"createTime"`
	OnlineTime             time.Time `json:"onlineTime"`
	OfflineTime            time.Time `json:"offlineTime"`
	DeviceId               string    `json:"deviceId"`
	IsWhiteAcc             int       `json:"isWhiteAcc"`
	LoginBan               int64     `json:"loginBan"`
	LoginBanReason         string    `json:"loginBanReason"`
	Platform               int       `json:"platform"`
	ChatBan                int64     `json:"chatBan"`
	ChatBanReason          string    `json:"chatBanReason"`
	LastLoginEntryServerId int       `json:"lastLoginEntryServerId"`
	LastLoginLogicServerId int       `json:"lastLoginLogicServerId"`
}

type AccountRole struct {
	Id             int64     `json:"id"`
	AccountId      int64     `json:"accountId"`
	Uid            string    `json:"uid"`
	EntryServerId  int       `json:"entryServerId"`
	LogicServerId  int       `json:"logicServerId"`
	OriginServerId int       `json:"originServerId"`
	NickName       string    `json:"nickName"`
	RedisId        int64     `json:"redisId"`
	SystemMailId   int64     `json:"systemMailId"`
	LastToken      string    `json:"lastToken"`
	CreateTime     time.Time `json:"createTime"`
	OnlineTime     time.Time `json:"onlineTime"`
	OfflineTime    time.Time `json:"offlineTime"`
	LastLoginTime  time.Time `json:"lastLoginTime"`
}

type AccountRoleProfile struct {
	AccountId              int64     `json:"accountId" xorm:"account_id"`
	RoleId                 int64     `json:"roleId" xorm:"role_id"`
	Uid                    string    `json:"uid" xorm:"uid"`
	Account                string    `json:"account" xorm:"account"`
	NickName               string    `json:"nickName" xorm:"nick_name"`
	Type                   int       `json:"type" xorm:"type"`
	DeviceId               string    `json:"deviceId" xorm:"device_id"`
	Platform               int       `json:"platform" xorm:"platform"`
	IsWhiteAcc             int       `json:"isWhiteAcc" xorm:"is_white_acc"`
	LoginBan               int64     `json:"loginBan" xorm:"login_ban"`
	LoginBanReason         string    `json:"loginBanReason" xorm:"login_ban_reason"`
	ChatBan                int64     `json:"chatBan" xorm:"chat_ban"`
	ChatBanReason          string    `json:"chatBanReason" xorm:"chat_ban_reason"`
	EntryServerId          int       `json:"entryServerId" xorm:"entry_server_id"`
	LogicServerId          int       `json:"logicServerId" xorm:"logic_server_id"`
	OriginServerId         int       `json:"originServerId" xorm:"origin_server_id"`
	RedisId                int64     `json:"redisId" xorm:"redis_id"`
	SystemMailId           int64     `json:"systemMailId" xorm:"system_mail_id"`
	RoleCreateTime         time.Time `json:"roleCreateTime" xorm:"role_create_time"`
	RoleOnlineTime         time.Time `json:"roleOnlineTime" xorm:"role_online_time"`
	RoleOfflineTime        time.Time `json:"roleOfflineTime" xorm:"role_offline_time"`
	RoleLastLoginTime      time.Time `json:"roleLastLoginTime" xorm:"role_last_login_time"`
	AccountCreateTime      time.Time `json:"accountCreateTime" xorm:"account_create_time"`
	LastLoginEntryServerId int       `json:"lastLoginEntryServerId" xorm:"last_login_entry_server_id"`
	LastLoginLogicServerId int       `json:"lastLoginLogicServerId" xorm:"last_login_logic_server_id"`
}
