package entity

import "xfx/login_server/define"

type ServerGroupMeta struct {
	Id        int64  `json:"id" xorm:"pk autoincr"`
	Name      string `json:"name" xorm:"varchar(64) notnull"`
	SortOrder int    `json:"sortOrder" xorm:"sort_order notnull"`
}

func (*ServerGroupMeta) TableName() string { return define.TableServerGroup }

type ServerItem struct {
	Id             int64  `json:"id" xorm:"pk autoincr"`
	Channel        int    `json:"channel" xorm:"notnull"`
	GroupId        int    `json:"groupId" xorm:"group_id notnull"`
	Ip             string `json:"ip" xorm:"varchar(64) notnull"`
	Port           int    `json:"port" xorm:"notnull"`
	ServerState    int    `json:"serverState" xorm:"server_state"`
	OpenServerTime int64  `json:"openServerTime" xorm:"open_server_time"`
	StopServerTime int64  `json:"stopServerTime" xorm:"stop_server_time"`
	ServerName     string `json:"serverName" xorm:"server_name varchar(64)"`
	ExeName        string `json:"exeName"`
	ExePath        string `json:"exePath"`
}

func (*ServerItem) TableName() string { return define.TableGameServer }
