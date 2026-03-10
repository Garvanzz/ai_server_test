package dto

// ForceUpdate 强制更新检查请求
type ForceUpdate struct {
	Version string `json:"version"`
	Channel int    `json:"channel"`
}

// ForceUpdateRes 强制更新响应（Status/Url 等由业务填充）
type ForceUpdateRes struct {
	Status  int    `json:"status"`
	Url     string `json:"url"`
	Version string `json:"version,omitempty"`
}
