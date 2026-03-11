package model

type ServerGroup struct {
	Id        int64  `json:"id" xorm:"pk autoincr"`
	Name      string `json:"name" xorm:"varchar(64) notnull"`
	SortOrder int    `json:"sortOrder" xorm:"sort_order notnull"`
}

type ServerItem struct {
	Id                int64  `json:"id" xorm:"pk autoincr"`
	Channel           int    `json:"channel" xorm:"notnull"`
	GroupId           int    `json:"groupId" xorm:"group_id notnull"`
	Ip                string `json:"ip" xorm:"varchar(64) notnull"`
	Port              int    `json:"port" xorm:"notnull"`
	MainServerHttpUrl string `json:"mainServerHttpUrl" xorm:"main_server_http_url"` // 大厅服 HTTP 地址，GM 转发用（如 http://ip:9505）
	ServerState       int    `json:"serverState" xorm:"server_state"`               // 0：正常 1：拥挤 2：爆满 3：维护 4：未开服 5：停服
	OpenServerTime    int64  `json:"openServerTime" xorm:"open_server_time"`
	StopServerTime    int64  `json:"stopServerTime" xorm:"stop_server_time"`
	ServerName        string `json:"serverName" xorm:"server_name varchar(64)"`
	ExeName           string `json:"exeName" xorm:"exe_name"`
	ExePath           string `json:"exePath" xorm:"exe_path"`
}
