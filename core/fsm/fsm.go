package fsm

import (
	"errors"
	"fmt"
)

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
}

func NewStateMachine(delegate Delegate, transitions ...Transition) *StateMachine {
	return &StateMachine{delegate: delegate, transitions: transitions}
}

func (m *StateMachine) Trigger(currentState string, event string, args ...interface{}) error {
	trans := m.findTransMatching(currentState, event)
	if trans == nil {
		return errors.New(fmt.Sprintf("stateMechine trigger error,currentState:%s,event:%s", currentState, event))

	}

	var err error
	if trans.Action != "" {
		err = m.delegate.HandleEvent(trans.Action, currentState, trans.To, args)
	}
	return err
}

func (m *StateMachine) findTransMatching(fromState string, event string) *Transition {
	for _, v := range m.transitions {
		if v.From == fromState && v.Event == event {
			return &v
		}
	}
	return nil
}
