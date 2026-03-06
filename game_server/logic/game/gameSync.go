package game

import (
	"xfx/game_server/logic/base"
)

type GameSync struct {
	base.ISync
	lister *lockstep
}

// NewGame 构造游戏
func NewGameSync(lister *lockstep) *GameSync {
	g := &GameSync{}
	g.lister = lister

	return g
}

// ----------------------------------------------------------------------
