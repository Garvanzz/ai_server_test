package modules

import (
	"xfx/pkg/module"
)

var _ module.Module = (*BaseModule)(nil)

type BaseModule struct {
	BaseAgent
}

func (m *BaseModule) GetType() string { return "unknown" }
func (m *BaseModule) OnDestroy()      {}
