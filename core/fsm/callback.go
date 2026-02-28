package fsm

type EventProcessor interface {
	OnExit(fromState string, args []interface{})
	Action(action string, fromState string, toState string, args []interface{}) error
	OnActionFailure(action string, fromState string, toState string, args []interface{}, err error)
	OnEnter(toState string, args []interface{})
}

type DefaultDelegate struct {
	P EventProcessor
}

func (dd *DefaultDelegate) HandleEvent(action string, fromState string, toState string, args []interface{}) error {
	if fromState != toState {
		dd.P.OnExit(fromState, args)
	}

	err := dd.P.Action(action, fromState, toState, args)
	if err != nil {
		dd.P.OnActionFailure(action, fromState, toState, args, err)
		return err
	}

	if fromState != toState {
		dd.P.OnEnter(toState, args)
	}

	return nil
}
