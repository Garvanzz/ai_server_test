package main

import (
	"flag"
	"fmt"
	"time"
)

// Config 客户端运行配置，支持命令行与默认值
type Config struct {
	// 登录服
	LoginServerURL string
	// 游戏服 TCP 地址（若为空则从 GetServerList 取 serverId 对应服）
	MainServerAddr string
	// 区服 ID（与登录请求、GetServerList 一致）
	ServerID int
	// 账号密码（可选，不填则用 AccountPrefix+序号）
	Account  string
	Password string
	// 多客户端时账号前缀与数量
	AccountPrefix string
	ClientCount   int
	// 是否先注册再登录
	DoRegister bool
	// 随机打接口间隔
	TestInterval time.Duration
	// 运行时长，0 表示一直跑
	RunDuration time.Duration
}

func LoadConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.LoginServerURL, "login", "http://127.0.0.1:9033", "登录服 HTTP 地址")
	flag.StringVar(&cfg.MainServerAddr, "main", "127.0.0.1:8082", "游戏服 TCP 地址 ip:port")
	flag.IntVar(&cfg.ServerID, "server", 1, "区服 ID")
	flag.StringVar(&cfg.Account, "account", "", "账号（单客户端时使用）")
	flag.StringVar(&cfg.Password, "password", "", "密码（单客户端时使用）")
	flag.StringVar(&cfg.AccountPrefix, "prefix", "test_user", "多客户端账号前缀")
	flag.IntVar(&cfg.ClientCount, "n", 1, "并发客户端数量")
	flag.BoolVar(&cfg.DoRegister, "register", true, "是否先注册再登录")
	flag.DurationVar(&cfg.TestInterval, "interval", 2*time.Second, "随机请求间隔")
	flag.DurationVar(&cfg.RunDuration, "duration", 0, "运行时长，0 为一直运行")
	flag.Parse()
	return cfg
}

// AccountForIndex 返回第 i 个客户端的账号（从 1 开始）
func (c *Config) AccountForIndex(i int) string {
	if c.Account != "" && i == 1 {
		return c.Account
	}
	return fmt.Sprintf("%s_%d", c.AccountPrefix, i)
}

// PasswordForIndex 返回第 i 个客户端的密码
func (c *Config) PasswordForIndex(i int) string {
	if c.Password != "" && i == 1 {
		return c.Password
	}
	return c.AccountForIndex(i)
}
