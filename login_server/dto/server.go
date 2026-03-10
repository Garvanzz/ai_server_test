package dto

// ReqServerList 获取服务器列表请求（channel 仅统计用，不参与过滤）
type ReqServerList struct {
	Channel int `json:"channel"`
}

// ServerGroupResp 区组+区服列表响应
type ServerGroupResp struct {
	Group   int32             `json:"group"`
	Name    string            `json:"name"`
	Servers []LoginServerItem `json:"servers"`
}

// LoginServerItem 区服项（API 返回，camelCase）
type LoginServerItem struct {
	Id             int64  `json:"id"`
	Ip             string `json:"ip"`
	Port           int    `json:"port"`
	Channel        int    `json:"channel"`
	ServerState    int    `json:"serverState"`
	OpenServerTime int64  `json:"openServerTime"`
	StopServerTime int64  `json:"stopServerTime"`
	ServerName     string `json:"serverName"`
	GroupId        int    `json:"groupId"`
	RegionName     string `json:"regionName,omitempty"`
	UdpIp          string `json:"udpIp,omitempty"`
	UdpPort        int    `json:"udpPort,omitempty"`
}
