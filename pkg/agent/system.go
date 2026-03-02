package agent

import (
	"fmt"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
	"time"
	"xfx/pkg/log"
)

type System struct {
	name     string
	system   *actor.ActorSystem
	context  *actor.RootContext
	remote   *remote.Remote
	root     PID
	opts     Options
	eventBus *eventBus
}

func (s *System) Root() PID { return s.root }

func NewSystem(opts ...Option) *System {
	opt := Options{
		CallTTL:          DefaultCallTTL,
		Agent:            todoAgent,
		MaxRetries:       DefaultMaxRetries,
		SupervisorWindow: DefaultSupervisorWindow,
	}
	for _, o := range opts {
		o(&opt)
	}

	system := &System{
		opts:     opt,
		name:     opt.Name,
		eventBus: newEventBus(),
	}

	system.system = actor.NewActorSystem()
	system.context = system.system.Root

	dlHandler := opt.DeadLetterHandler
	system.system.EventStream.Subscribe(func(evt interface{}) {
		if e, ok := evt.(*actor.DeadLetterEvent); ok {
			actualMsg := e.Message
			if lm, ok := actualMsg.(*LocalMessage); ok {
				actualMsg = lm.msg
			}
			if dlHandler != nil {
				dlHandler(e.PID, actualMsg, e.Sender)
			} else {
				log.Warn("dead letter: target=%v, sender=%v, message=%T", e.PID, e.Sender, actualMsg)
			}
		}
	})

	if opt.Host != "" {
		configure := remote.Configure(opt.Host, opt.Port)
		system.remote = remote.NewRemote(system.system, configure)
	}
	return system
}

func (s *System) Start() {
	if s.remote != nil {
		s.remote.Start()
	}

	decider := func(reason interface{}) actor.Directive {
		if s.opts.Restart {
			return actor.RestartDirective
		} else {
			return actor.StopDirective
		}
	}
	supervisor := actor.NewOneForOneStrategy(s.opts.MaxRetries, s.opts.SupervisorWindow, decider)
	props := actor.PropsFromProducer(func() actor.Actor {
		return &defaultActor{opts: s.opts, system: s}
	}, actor.WithSupervisor(supervisor))

	if s.name != "" {
		root, err := s.system.Root.SpawnNamed(props, s.name)
		if err != nil {
			panic(fmt.Sprintf("agent: start system error(%v)", err))
		}
		s.root = root
	} else {
		s.root = s.system.Root.Spawn(props)
	}
}

func (s *System) Stop() {
	future := s.context.PoisonFuture(s.root)
	if err := future.Wait(); err != nil {
		log.Error("agent: stop system(%s) error: %v", s.name, err)
	}
	if s.remote != nil {
		s.remote.Shutdown(true)
	}
}

func (s *System) StopGraceful(timeout time.Duration) error {
	ch := make(chan error, 1)
	go func() {
		future := s.context.PoisonFuture(s.root)
		ch <- future.Wait()
	}()

	var err error
	select {
	case err = <-ch:
	case <-time.After(timeout):
		err = fmt.Errorf("agent: stop system(%s) timeout after %v", s.name, timeout)
	}

	if s.remote != nil {
		s.remote.Shutdown(true)
	}
	return err
}

func (s *System) Create(name string, agent Agent, opts ...Option) (PID, error) {
	future := s.system.Root.RequestFuture(s.root, &createMessage{name: name, agent: agent, opts: opts}, s.opts.CallTTL)
	if result, err := future.Result(); err != nil {
		return nil, fmt.Errorf("agent: system(%s) create %s error(%v)", s.name, name, err)
	} else {
		if pid, ok := result.(PID); ok {
			return pid, nil
		} else {
			if err, ok := result.(error); ok {
				return nil, err
			} else {
				return nil, fmt.Errorf("agent: system(%s) create error(%v)", s.name, err)
			}
		}
	}
}

func (s *System) Destroy(pid PID) {
	if pid == nil {
		return
	}
	future := s.context.PoisonFuture(pid)
	if err := future.Wait(); err != nil {
		log.Error("agent: destroy %s error: %v", Address(pid), err)
	}
}

func (s *System) Cast(pid PID, message any) {
	if pid == nil {
		return
	}
	m, e := wrapMessage(s.root, pid, message, false)
	if e != nil {
		log.Error("system cast error:%v", e)
		return
	}

	s.context.Request(pid, m)
}
