package model

type ServerGroup struct {
	Id        int64  `json:"id" xorm:"pk autoincr"`
	Name      string `json:"name" xorm:"varchar(64) notnull"`
	SortOrder int    `json:"sortOrder" xorm:"sort_order notnull"`
}

type ServerItem struct {
	Id             int64  `json:"id" xorm:"pk autoincr"`
	Channel        int    `json:"channel" xorm:"notnull"`
	GroupId        int    `json:"groupId" xorm:"group_id notnull"`
	Ip             string `json:"ip" xorm:"varchar(64) notnull"`
	Port           int    `json:"port" xorm:"notnull"`
	RedisPort      int    `json:"redisPort" xorm:"redis_port"`           // Redis 端口，GM/多服用
	MysqlAddr      string `json:"mysqlAddr" xorm:"mysql_addr"`          // MySQL 地址（可选）
	LoginServerUrl string `json:"loginServerUrl" xorm:"login_server_url"` // 登录服地址（可选）
	ServerState    int    `json:"serverState" xorm:"server_state"`        // 0：正常 1：拥挤 2：爆满 3：维护 4：未开服 5：停服
	OpenServerTime int64  `json:"openServerTime" xorm:"open_server_time"`
	StopServerTime int64  `json:"stopServerTime" xorm:"stop_server_time"`
	ServerName     string `json:"serverName" xorm:"server_name varchar(64)"`
	ExeName        string `json:"exeName" xorm:"exe_name"`
	ExePath        string `json:"exePath" xorm:"exe_path"`
	ServerGroup    int    `json:"serverGroup" xorm:"server_group"` // 区服组展示用
	GameServer     int64  `json:"gameServer" xorm:"game_server"`  // 关联游戏服 ID（可选）
}
