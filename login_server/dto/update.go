package dto

// ForceUpdateRequest 强制更新检查请求。
type ForceUpdateRequest struct {
	Version string `json:"version"`
	Channel int    `json:"channel"`
}

// ForceUpdateResponse 强制更新响应，保留旧字段兼容老客户端。
type ForceUpdateResponse struct {
	Status  int    `json:"status"`
	URL     string `json:"url"`
	Version string `json:"version,omitempty"`

	LegacyStatus  int    `json:"Status,omitempty"`
	LegacyURL     string `json:"Url,omitempty"`
	LegacyVersion string `json:"Version,omitempty"`
}
