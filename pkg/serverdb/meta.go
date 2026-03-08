package serverdb

// ServerMeta 服列表元数据（仅业务/展示，不含 Redis/MySQL 连接信息）。
// 存 MySQL 表或配置中心，供客户端选服、GM 展示；连接配置仅来自 Config。
// 若从现有 servergroup 表加载，可用列：id, serverName, server_group, serverState, openServerTime, stopServerTime, channel, ip, port, loginServerUrl（无 redis/mysql）
type ServerMeta struct {
	Id             int64  `json:"id" xorm:"pk"`
	ServerName     string `json:"serverName" xorm:"serverName"`
	ServerGroup    int    `json:"serverGroup" xorm:"server_group"`
	State          int    `json:"state" xorm:"serverState"`
	OpenServerTime int64  `json:"openServerTime" xorm:"openServerTime"`
	StopServerTime int64  `json:"stopServerTime" xorm:"stopServerTime"`
	Channel        int    `json:"channel" xorm:"channel"`
	Ip             string `json:"ip" xorm:"ip"`
	Port           int    `json:"port" xorm:"port"`
	ClientGateAddr string `json:"clientGateAddr" xorm:"-"` // 客户端连接地址，可填或由 Ip:Port 得到
	LoginServerUrl string `json:"loginServerUrl" xorm:"loginServerUrl"`
}

// TableName 默认表名；LoadServerList 可传入 "servergroup" 兼容现有表。
func (ServerMeta) TableName() string {
	return "server_list"
}
