package dto

// RegisterUser 注册请求
type RegisterUser struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Platform int    `json:"platform"` // 1 pc 2 ios 3 安卓
	ServerId int    `json:"serverId"`
}

// LoginUser 登录请求
type LoginUser struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Version  string `json:"version"`
	Platform int    `json:"platform"`
	ServerId int    `json:"serverId"`
}
