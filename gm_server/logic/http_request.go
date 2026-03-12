package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HttpRequestToServer 向指定区服的 main_server 发送 GM 请求；serverId 为区服 id，<=0 时用配置默认 URL
// path 示例："/gm/mail"、"/gm/player/game-info"
func HttpRequestToServer(serverId int, jsonData []byte, path string) (error, string) {
	baseURL := getMainServerURL(serverId)
	if baseURL == "" {
		return fmt.Errorf("main_server URL not configured"), ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := baseURL + path

	// 创建自定义客户端
	client := &http.Client{
		Timeout: time.Second * 30, // 增加超时时间到30秒
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err), ""
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err), ""
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err), ""
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %d: %s", resp.StatusCode, string(body)), string(body)
	}

	// main_server 的 GM 接口统一通过 HTTPRetGame 返回 {errcode, errmsg, ...}
	var wrapper struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &wrapper); err == nil {
		if wrapper.ErrCode != 0 {
			return fmt.Errorf("gm api error %d: %s", wrapper.ErrCode, wrapper.ErrMsg), string(body)
		}
	}

	return nil, string(body)
}
