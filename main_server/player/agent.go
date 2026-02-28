package player

import (
	"time"
	"xfx/core/model"
	"xfx/pkg/agent"
	"xfx/pkg/module"
	"xfx/pkg/module/modules"
)

const SaveTick = 10 * 10

type PlayerAgent struct {
	modules.BaseAgent
	model *model.Player
}

func New(model *model.Player, session agent.PID, app module.App) *PlayerAgent {
	a := &PlayerAgent{
		model: model,
	}
	a.model.Cache.Session = session
	a.model.Cache.App = app
	return a
}

func (pl *PlayerAgent) OnStart(ctx agent.Context) {
	pl.BaseAgent.OnStart(ctx)
	pl.model.Cache.Self = pl.Context.Self()
}

func (pl *PlayerAgent) OnStop() {
}

func (pl *PlayerAgent) OnTerminated(pid agent.PID, reason int) {
	pidStr := agent.Address(pid)
	if pidStr == agent.Address(pl.model.Session) {
		//game.OnDisconnected(pl, pl.model)
	} else if pidStr == agent.Address(pl.model.GameRun.PID) {
		//game.OnGameClosed(pl, pl.model)
	}
}

func (pl *PlayerAgent) Send(msg any) {
	if pl.model.Cache.Session != nil {
		pl.Context.Cast(pl.model.Cache.Session, msg)
	}
}

func (pl *PlayerAgent) Call(pid agent.PID, msg any) (any, error) {
	return pl.Context.Call(pid, msg)
}

func (pl *PlayerAgent) Cast(pid agent.PID, msg any) {
	pl.Context.Cast(pid, msg)
}

func (pl *PlayerAgent) SaveNow() {
	pl.model.SaveTick = 1
}

func (pl *PlayerAgent) GetPlayer() *model.Player {
	return pl.model
}

func (pl *PlayerAgent) OnTick(delta time.Duration) {
	if pl.model.SaveTick > 0 {
		pl.model.SaveTick--
		if pl.model.SaveTick <= 0 {
			pl.OnSave(false)
			pl.model.SaveTick = SaveTick
		}
	}
}

func (pl *PlayerAgent) Stop() {
	pl.Context.Stop()
}

func (pl *PlayerAgent) OnMessage(msg any) any {
	return dispatch(pl, pl.model, msg)
}

func (pl *PlayerAgent) Watch(pid agent.PID) {
	pl.Context.Watch(pid)
}

func (pl *PlayerAgent) Unwatch(pid agent.PID) {
	pl.Context.Unwatch(pid)
}
