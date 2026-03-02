package agent

import (
	"fmt"
	"github.com/asynkron/protoactor-go/actor"
	"reflect"
	"runtime"
	"time"
	"xfx/pkg/log"
)

type defaultActor struct {
	name    string
	opts    Options
	system  *System
	agent   Agent
	context *agentContext
	cancel  func()
}

func (da *defaultActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		da.onStart(context)
	case *actor.Stopping:
	case *actor.Stopped:
		da.onStop()
	case *actor.Restarting:
	case *actor.Restart:
	case *actor.Terminated:
		da.OnTerminated(msg.Who, msg.Why)
	case *createMessage:
		da.onCreate(context, msg.name, msg.agent, msg.opts...)
	case tickMessage:
		da.onTick(context, time.Duration(msg))
	case *LocalMessage, *RemoteMessage:
		da.onMessage(context, msg)
	default:
		msgType := reflect.TypeOf(msg)
		msgValue := reflect.ValueOf(msg)
		log.Error("unknown message: type=%s, value=%+v", msgType, msgValue.Interface())
	}
}

func (da *defaultActor) onStart(context actor.Context) {
	da.context = &agentContext{
		context: context,
		system:  da.system,
		opts:    &da.opts,
		states:  make(map[string]interface{}),
		metrics: &Metrics{},
	}
	da.name = da.opts.Name
	da.agent = da.opts.Agent
	_Store(da.name, da.context.Self())
	da.agent.OnStart(da.context)
	if da.opts.Tick > 0 {
		da.cancel = tickgo(da.context, da.opts.Tick)
	}
}

func (da *defaultActor) onStop() {
	_Delete(da.name, da.context.Self())
	da.agent.OnStop()
	if da.cancel != nil {
		da.cancel()
		da.cancel = nil
	}
}

func (da *defaultActor) OnTerminated(who PID, reason actor.TerminatedReason) {
	da.agent.OnTerminated(who, int(reason))
}

func (da *defaultActor) onCreate(context actor.Context, name string, agent Agent, opts ...Option) {
	if pid, err := da.context.Create(name, agent, opts...); err != nil {
		context.Respond(err)
	} else {
		context.Respond(pid)
	}
}

func (da *defaultActor) onTick(context actor.Context, delta time.Duration) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			da.context.metrics.recordPanic()
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			log.Error("agent OnTick panic: %v\nStack: %s", r, buf[:n])
		}
		da.context.metrics.recordTick(time.Since(start))
	}()
	da.agent.OnTick(delta)
	for _, child := range context.Children() {
		context.Send(child, tickMessage(delta))
	}
}

func (da *defaultActor) onMessage(context actor.Context, msg interface{}) {
	start := time.Now()
	m, sender, senderA, response, err := unwrapMessage(msg)
	if err != nil {
		log.Error("unwrap message: %v", err)
		if response {
			context.Respond(err)
		}
		return
	}

	da.context.message = m
	da.context.sender = sender
	da.context.senderA = senderA
	da.context.rawMsg = msg
	defer func() {
		da.context.message = nil
		da.context.sender = nil
		da.context.senderA = ""
		da.context.rawMsg = nil
		da.context.metrics.recordMsg(time.Since(start))
	}()

	defer func() {
		if r := recover(); r != nil {
			da.context.metrics.recordPanic()
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			log.Error("agent OnMessage panic: %v\nStack: %s", r, buf[:n])
			if response {
				context.Respond(fmt.Errorf("agent panic: %v", r))
			}
		}
	}()

	var result interface{}
	if da.context.filter != nil {
		var discard bool
		discard, result = da.context.filter(m)
		if discard {
			if response {
				context.Respond(result)
			}
			return
		}
	}

	result = da.agent.OnMessage(m)
	if response {
		context.Respond(result)
	}
}
