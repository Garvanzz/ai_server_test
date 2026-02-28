package activity

import "xfx/core/fsm"

const (
	StateWaiting = "waiting"
	StateRunning = "running"
	StateStopped = "stopped"
	StateClosed  = "closed"

	EventNone    = ""
	EventStart   = "event_start"
	EventStop    = "event_stop"
	EventClose   = "event_close"
	EventRecover = "event_recover"
	EventRestart = "event_restart"

	ActionStart   = "action_start"
	ActionClose   = "action_close"
	ActionStop    = "action_stop"
	ActionRecover = "action_recover"
	ActionRestart = "action_restart"
)

var (
	transitions = []fsm.Transition{
		{StateWaiting, EventStart, StateRunning, ActionStart},
		{StateWaiting, EventClose, StateClosed, ActionClose},

		{StateRunning, EventStop, StateStopped, ActionStop},
		{StateRunning, EventClose, StateClosed, ActionClose},

		{StateStopped, EventRecover, StateRunning, ActionRecover},
		{StateStopped, EventClose, StateClosed, ActionClose},
		{StateStopped, EventRestart, StateWaiting, ActionRestart},

		{StateClosed, EventRestart, StateWaiting, ActionRestart},
	}
)
