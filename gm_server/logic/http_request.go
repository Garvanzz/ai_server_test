package logic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"xfx/gm_server/conf"
)

// HttpRequest 向 main_server 发送 GM 请求
// path 示例："/gm/mail"、"/gm/notice"
func HttpRequest(jsonData []byte, path string) (error, string) {
	baseURL := conf.Server.MainServerHttpUrl
	if baseURL == "" {
		// 兼容老配置，默认本机 9505
		baseURL = "http://127.0.0.1:9505"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := baseURL + path

	// 创建自定义客户端
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err), ""
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

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

	// main_server 的 GM 接口统一通过 httpRetGame 返回 {errcode, errmsg, ...}
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
