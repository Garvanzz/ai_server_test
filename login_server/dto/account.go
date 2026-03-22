package dto

// RegisterRequest 注册请求。
type RegisterRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Platform int    `json:"platform"` // 1 pc 2 ios 3 安卓
	ServerID int    `json:"serverId"`
}

// RegisterResponse 注册响应。
type RegisterResponse struct {
	Account string `json:"account"`
}

// LoginRequest 登录请求。
type LoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Version  string `json:"version"`
	Platform int    `json:"platform"`
	ServerID int    `json:"serverId"`
}

// LoginResponse 登录响应。
type LoginResponse struct {
	ServerID          int    `json:"serverId"`
	EntryServerID     int64  `json:"entryServerId"`
	Token             string `json:"token"`
	UID               string `json:"uid"`
	LastLoginServerID int64  `json:"lastLoginServerId"`
}
