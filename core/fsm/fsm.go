package fsm

import (
	"errors"
	"fmt"
)

var ErrTransitionNotFound = errors.New("fsm transition not found")

type Transition struct {
	From   string
	Event  string
	To     string
	Action string
}

type Delegate interface {
	HandleEvent(action string, fromState string, toState string, args []interface{}) error
}

type StateMachine struct {
	delegate    Delegate
	transitions []Transition
	index       map[string]map[string]Transition
}

func NewStateMachine(delegate Delegate, transitions ...Transition) *StateMachine {
	index := make(map[string]map[string]Transition, len(transitions))
	for _, trans := range transitions {
		events, ok := index[trans.From]
		if !ok {
			events = make(map[string]Transition)
			index[trans.From] = events
		}
		events[trans.Event] = trans
	}
	return &StateMachine{delegate: delegate, transitions: transitions, index: index}
}

func (m *StateMachine) Trigger(currentState string, event string, args ...interface{}) error {
	trans := m.findTransMatching(currentState, event)
	if trans == nil {
		return fmt.Errorf("%w: currentState=%s, event=%s", ErrTransitionNotFound, currentState, event)

	}

	var err error
	if trans.Action != "" {
		err = m.delegate.HandleEvent(trans.Action, currentState, trans.To, args)
	}
	return err
}

func (m *StateMachine) findTransMatching(fromState string, event string) *Transition {
	if events, ok := m.index[fromState]; ok {
		if trans, ok := events[event]; ok {
			return &trans
		}
	}
	return nil
}
