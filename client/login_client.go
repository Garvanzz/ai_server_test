package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LoginResult 登录成功后的数据
type LoginResult struct {
	Token    string
	UID      string
	ServerID int64
}

// LoginClient 登录服 HTTP 客户端
type LoginClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewLoginClient(baseURL string) *LoginClient {
	// URL 必须带 scheme，否则 url.Parse 会报错：first path segment in URL cannot contain colon
	if baseURL != "" && !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "http://" + baseURL
	}
	return &LoginClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// post 统一 POST JSON，返回解析后的 map
func (lc *LoginClient) post(path string, body any) (map[string]interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}
	u, err := url.Parse(lc.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := lc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return m, nil
}

const (
	loginRouteAuthRegister = "/auth/register"
	loginRouteAuthLogin    = "/auth/login"
	loginRouteServersList  = "/servers/list"
)

// Register 注册账号
func (lc *LoginClient) Register(account, password string, platform, serverID int) error {
	body := map[string]interface{}{
		"account":  account,
		"password": password,
		"platform": platform,
		"serverId": serverID,
	}
	m, err := lc.post(loginRouteAuthRegister, body)
	if err != nil {
		return err
	}
	code, _ := m["errcode"].(float64)
	if code != 0 {
		msg, _ := m["errmsg"].(string)
		return fmt.Errorf("register failed: errcode=%v errmsg=%s", code, msg)
	}
	return nil
}

// Login 登录，返回 token、uid、serverId
func (lc *LoginClient) Login(account, password string, serverID int, platform int, version string) (*LoginResult, error) {
	if version == "" {
		version = "0.1"
	}
	if platform <= 0 {
		platform = 1
	}
	body := map[string]interface{}{
		"account":  account,
		"password": password,
		"version":  version,
		"platform": platform,
		"serverId": serverID,
	}
	m, err := lc.post(loginRouteAuthLogin, body)
	if err != nil {
		return nil, err
	}
	code, _ := m["errcode"].(float64)
	if code != 0 {
		msg, _ := m["errmsg"].(string)
		return nil, fmt.Errorf("login failed: errcode=%v errmsg=%s", code, msg)
	}
	payload, _ := m["data"].(map[string]interface{})
	if payload == nil {
		payload = m
	}
	token, _ := payload["token"].(string)
	uid, _ := payload["uid"].(string)
	if token == "" || uid == "" {
		return nil, fmt.Errorf("login resp missing token or uid: %v", m)
	}
	var serverId int64 = int64(serverID)
	if s, ok := payload["serverId"]; ok {
		switch v := s.(type) {
		case float64:
			serverId = int64(v)
		case int:
			serverId = int64(v)
		}
	}
	return &LoginResult{Token: token, UID: uid, ServerID: serverId}, nil
}

// GetServerList 获取区服列表，返回 JSON 字符串（可解析出各服 Ip/Port）
func (lc *LoginClient) GetServerList(channel int) (serverListJSON string, err error) {
	body := map[string]interface{}{"channel": channel}
	m, err := lc.post(loginRouteServersList, body)
	if err != nil {
		return "", err
	}
	code, _ := m["errcode"].(float64)
	if code != 0 {
		msg, _ := m["errmsg"].(string)
		return "", fmt.Errorf("server list failed: errcode=%v errmsg=%s", code, msg)
	}
	payload, _ := m["data"].(map[string]interface{})
	if payload == nil {
		payload = m
	}
	if serverList, ok := payload["serverList"]; ok {
		b, marshalErr := json.Marshal(serverList)
		if marshalErr != nil {
			return "", fmt.Errorf("marshal server list: %w", marshalErr)
		}
		return string(b), nil
	}
	s, _ := m["ServerList"].(string)
	if s == "" {
		s, _ = m["serverList"].(string)
	}
	return s, nil
}
