package agent

import (
	"errors"
	"fmt"
	"github.com/asynkron/protoactor-go/actor"
	"strings"
	"time"
	"xfx/pkg/log"
)

var (
	ErrNilPID      = errors.New("agent: pid is nil")
	ErrEmptyName   = errors.New("agent: name is empty")
	ErrInvalidName = errors.New("agent: invalid name(maybe contains ':' or '/')")
)

type MessageFilter func(msg interface{}) (bool, interface{})

type Context interface {
	Self() PID
	Cast(pid PID, msg interface{})
	Call(pid PID, msg interface{}) (interface{}, error)
	CallWithTimeout(pid PID, msg interface{}, timeout time.Duration) (interface{}, error)
	CallNR(pid PID, msg interface{}) error
	Watch(pid PID)
	Unwatch(pid PID)
	Stop()
	Create(name string, agent Agent, opts ...Option) (PID, error)
	SetMessageFilter(MessageFilter)
	Message() interface{}
	Sender() PID
	Metrics() *Metrics
	// Stash saves the current message for later replay.
	// Only effective within OnMessage; Call-type messages should not be stashed.
	Stash()
	// UnstashAll re-sends all stashed messages back to this actor's mailbox.
	UnstashAll()
}

type agentContext struct {
	context     actor.Context
	system      *System
	opts        *Options
	states      map[string]interface{}
	metrics     *Metrics
	filter      MessageFilter
	message     interface{}
	sender      PID
	senderA     string
	rawMsg      interface{}
	stashedMsgs []interface{}
}

func (c *agentContext) Self() PID {
	return c.context.Self()
}

func (c *agentContext) Cast(pid PID, msg interface{}) {
	if pid == nil {
		return
	}
	m, e := wrapMessage(c.Self(), pid, msg, false)
	if e != nil {
		log.Error("Cast:%v", e)
		return
	}

	c.context.Request(pid, m)
}

func (c *agentContext) Call(pid PID, msg interface{}) (interface{}, error) {
	if pid == nil {
		return nil, ErrNilPID
	}
	m, err := wrapMessage(c.Self(), pid, msg, true)
	if err != nil {
		return nil, err
	}

	f := c.context.RequestFuture(pid, m, c.opts.CallTTL)
	if result, err := f.Result(); err != nil {
		return nil, fmt.Errorf("agent context: call error(%v)", err)
	} else {
		return result, nil
	}
}

func (c *agentContext) CallNR(pid PID, msg interface{}) error {
	if pid == nil {
		return ErrNilPID
	}
	m, err := wrapMessage(c.Self(), pid, msg, true)
	if err != nil {
		return err
	}

	f := c.context.RequestFuture(pid, m, c.opts.CallTTL)
	if err := f.Wait(); err != nil {
		return fmt.Errorf("agent context: call error(%v)", err)
	} else {
		return nil
	}
}

func (c *agentContext) CallWithTimeout(pid PID, msg interface{}, timeout time.Duration) (interface{}, error) {
	if pid == nil {
		return nil, ErrNilPID
	}
	m, err := wrapMessage(c.Self(), pid, msg, true)
	if err != nil {
		return nil, err
	}

	f := c.context.RequestFuture(pid, m, timeout)
	if result, err := f.Result(); err != nil {
		return nil, fmt.Errorf("agent context: call error(%v)", err)
	} else {
		return result, nil
	}
}

func (c *agentContext) Metrics() *Metrics { return c.metrics }

func (c *agentContext) Watch(pid PID)   { c.context.Watch(pid) }
func (c *agentContext) Unwatch(pid PID) { c.context.Unwatch(pid) }
func (c *agentContext) Stop()           { c.context.Stop(c.Self()) }

func (c *agentContext) Set(key string, value interface{}) {
	c.states[key] = value
}

func (c *agentContext) Get(key string) (interface{}, bool) {
	v, ok := c.states[key]
	return v, ok
}

func (c *agentContext) Create(name string, agent Agent, opts ...Option) (PID, error) {
	opt := Options{
		CallTTL:          DefaultCallTTL,
		Name:             name,
		Agent:            agent,
		MaxRetries:       DefaultMaxRetries,
		SupervisorWindow: DefaultSupervisorWindow,
	}
	for _, o := range opts {
		o(&opt)
	}

	if opt.Name == "" {
		return nil, ErrEmptyName
	}

	if strings.ContainsAny(opt.Name, ":/") {
		return nil, ErrInvalidName
	}

	if opt.Agent == nil {
		opt.Agent = todoAgent
	}

	decider := func(reason interface{}) actor.Directive {
		if opt.Restart {
			return actor.RestartDirective
		} else {
			return actor.StopDirective
		}
	}
	supervisor := actor.NewOneForOneStrategy(opt.MaxRetries, opt.SupervisorWindow, decider)
	props := actor.PropsFromProducer(func() actor.Actor {
		return &defaultActor{opts: opt, system: c.system}
	}, actor.WithSupervisor(supervisor))

	var (
		pid PID
		err error
	)

	if name != "" {
		pid, err = c.context.SpawnNamed(props, opt.Name)
	} else {
		pid = c.context.Spawn(props)
	}
	return pid, err
}

func (c *agentContext) SetMessageFilter(filter MessageFilter) { c.filter = filter }
func (c *agentContext) Message() interface{}                  { return c.message }
func (c *agentContext) Sender() PID {
	if c.sender != nil {
		return c.sender
	}
	if c.senderA != "" {
		c.sender, _ = Parse(c.senderA)
		return c.sender
	}
	return nil
}

func (c *agentContext) Stash() {
	if c.rawMsg != nil {
		c.stashedMsgs = append(c.stashedMsgs, c.rawMsg)
	}
}

func (c *agentContext) UnstashAll() {
	for _, m := range c.stashedMsgs {
		c.context.Send(c.Self(), m)
	}
	c.stashedMsgs = c.stashedMsgs[:0]
}
