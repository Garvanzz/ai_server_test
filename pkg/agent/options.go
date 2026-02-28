package agent

import (
	"time"
)

const (
	DefaultCallTTL = time.Millisecond * 1000
)

type Option func(*Options)

type Options struct {
	Name    string        // 节点名称
	Host    string        // 节点地址
	Port    int           // 节点端口
	CallTTL time.Duration // 发起Call的超时时间
	Restart bool
	Agent   Agent
	Tick    time.Duration
}

func WithTick(tick time.Duration) Option {
	return func(o *Options) {
		o.Tick = tick
	}
}

func WithCallTTL(ttl time.Duration) Option {
	return func(o *Options) {
		o.CallTTL = ttl
	}
}

func WithAgent(agent Agent) Option {
	return func(o *Options) {
		o.Agent = agent
	}
}

func WithName(v string) Option {
	return func(o *Options) {
		o.Name = v
	}
}

func WithRestart() Option {
	return func(o *Options) {
		o.Restart = true
	}
}

func WithHost(v string) Option {
	return func(o *Options) {
		o.Host = v
	}
}

func WithPort(v int) Option {
	return func(o *Options) {
		o.Port = v
	}
}
