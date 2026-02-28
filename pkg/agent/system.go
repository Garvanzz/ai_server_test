package agent

import (
	"fmt"
	"github.com/asynkron/protoactor-go/actor"
	"github.com/asynkron/protoactor-go/remote"
	"time"
	"xfx/pkg/log"
)

type System struct {
	name    string
	system  *actor.ActorSystem
	context *actor.RootContext
	remote  *remote.Remote
	root    PID
	opts    Options
}

func (s *System) Root() PID { return s.root }

func NewSystem(opts ...Option) *System {
	opt := Options{
		CallTTL: DefaultCallTTL,
		Agent:   todoAgent,
	}
	for _, o := range opts {
		o(&opt)
	}

	system := &System{
		opts: opt,
		name: opt.Name,
	}

	system.system = actor.NewActorSystem()
	system.context = system.system.Root

	// todo:system 监听死信
	//system.system.EventStream.Subscribe(func(evt interface{}) {
	//	if e, ok := evt.(*actor.DeadLetterEvent); ok {
	//		msgType := reflect.TypeOf(e.Message)
	//		log.Debug("Dead Letter-Sender:%v,Message:%v,PID:%v", e.Sender, msgType.String(), e.PID)
	//	}
	//})

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
	supervisor := actor.NewOneForOneStrategy(5, time.Second, decider)
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
		panic(fmt.Sprintf("agent: stop system(%s) error(%v)", s.name, err))
	}
	if s.remote != nil {
		s.remote.Shutdown(true)
	}
}

func (s *System) Create(name string, agent Agent, opts ...Option) (PID, error) {
	future := s.system.Root.RequestFuture(s.root, &createMessage{name: name, agent: agent, opts: opts}, time.Second)
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
	future := s.context.PoisonFuture(pid)
	future.Wait()
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
