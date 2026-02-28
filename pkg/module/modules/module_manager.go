package modules

import (
	"fmt"
	"xfx/pkg/agent"
	"xfx/pkg/log"
	"xfx/pkg/module"
)

func NewModuleManager() (m *ModuleManager) {
	m = new(ModuleManager)
	m.named = make(map[string]*DefaultModule)
	return
}

type ModuleManager struct {
	app   module.App
	mods  []*DefaultModule
	named map[string]*DefaultModule
}

func (mm *ModuleManager) Register(mi module.Module) {
	md := new(DefaultModule)
	md.mi = mi
	mm.mods = append(mm.mods, md)
	_, ok := mm.named[mi.GetType()]
	if ok {
		panic(fmt.Sprintf("module typed [%s] is already existed", mi.GetType()))
	}
	mm.named[mi.GetType()] = md
}

func (mm *ModuleManager) Init(app module.App, system *agent.System) {
	mm.app = app
	for i := 0; i < len(mm.mods); i++ {
		m := mm.mods[i]
		m.mi.OnInit(app)
		_, err := system.Create(
			m.mi.GetType(),
			m.mi,
		)

		if err != nil {
			panic(fmt.Sprintf("Init module %v error: %v", m.mi.GetType(), err))
		}
		log.Info("* module [%s] inited", m.mi.GetType())
	}
}

func (mm *ModuleManager) Destroy(system *agent.System) {
	for i := len(mm.mods) - 1; i >= 0; i-- {
		m := mm.mods[i]
		pid := m.mi.Self()
		if pid != nil {
			system.Destroy(pid)
		}
		m.mi.OnDestroy()
		log.Info("* module [%s] destroyed", m.mi.GetType())
	}
}

func (mm *ModuleManager) Get(mod string) module.Module {
	m, ok := mm.named[mod]
	if ok {
		return m.mi
	}
	return nil
}
