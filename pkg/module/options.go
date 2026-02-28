package module

import (
	"time"
)

type Option func(*Options)

type Options struct {
	Version     string
	KillWaitTTL time.Duration
	Fps         int
}

func WithVersion(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

func WithKillWaitTTL(v time.Duration) Option {
	return func(o *Options) {
		o.KillWaitTTL = v
	}
}

func WithFps(fps int) Option {
	return func(o *Options) {
		o.Fps = fps
	}
}
