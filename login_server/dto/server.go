package dto

// ServerListRequest 获取服务器列表请求（channel 仅统计用，不参与过滤）。
type ServerListRequest struct {
	Channel int `json:"channel"`
}

// ServerGroupResponse 区组+区服列表响应。
type ServerGroupResponse struct {
	Group   int32            `json:"group"`
	Name    string           `json:"name"`
	Servers []ServerListItem `json:"servers"`
}

// ServerListResponse 区服列表响应，保留旧字段兼容老客户端。
type ServerListResponse struct {
	ServerList []ServerGroupResponse `json:"serverList"`
}

// ServerListItem 区服项（API 返回，camelCase）。
type ServerListItem struct {
	ID             int64  `json:"id"`
	LogicServerID  int64  `json:"logicServerId"`
	MergeState     int    `json:"mergeState"`
	MergeStateText string `json:"mergeStateText"`
	MergeTime      int64  `json:"mergeTime"`
	IP             string `json:"ip"`
	Port           int    `json:"port"`
	Channel        int    `json:"channel"`
	ServerState    int    `json:"serverState"`
	OpenServerTime int64  `json:"openServerTime"`
	StopServerTime int64  `json:"stopServerTime"`
	ServerName     string `json:"serverName"`
	GroupID        int    `json:"groupId"`
	RegionName     string `json:"regionName,omitempty"`
	UDPIP          string `json:"udpIp,omitempty"`
	UDPPort        int    `json:"udpPort,omitempty"`
}
