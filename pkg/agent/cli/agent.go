package main

import (
	"fmt"
	"time"
	"xfx/pkg/agent"
)

type Gigi struct{}

func (a *Gigi) OnStart(ctx agent.Context) {
	fmt.Printf("* Gigi start\n")

	ctx.Create("lala", &Lala{})
	ctx.Create("yoyo", &Yoyo{})
}

func (a *Gigi) OnStop() {
	fmt.Printf("* Gigi stop\n")
}
func (a *Gigi) OnTerminated(pid agent.PID, reason int) {}
func (a *Gigi) OnMessage(msg interface{}) interface{}  { return nil }
func (a *Gigi) OnTick(delta time.Duration)             {}

type Lala struct {
	context agent.Context
}

func (a *Lala) OnStart(ctx agent.Context) {
	a.context = ctx
	fmt.Printf("* Lala start\n")
}
func (a *Lala) OnStop() {
	fmt.Printf("* Lala stop\n")
}
func (a *Lala) OnTerminated(pid agent.PID, reason int) {}
func (a *Lala) OnTick(delta time.Duration)             {}
func (a *Lala) OnMessage(msg interface{}) interface{} {
	switch msg.(type) {
	case *Hi:
		fmt.Printf("* Lala receive message Hi from %s\n", agent.Address(a.context.Sender()))
		return &Hi{Say: fmt.Sprintf("hello %v ", agent.Address(a.context.Sender()))}
	}
	return nil
}

type Yoyo struct {
	countdown time.Duration
	context   agent.Context
}

func (a *Yoyo) OnStart(ctx agent.Context) {
	a.context = ctx
	fmt.Printf("* Yoyo start\n")
	a.countdown = time.Second * 3
}
func (a *Yoyo) OnStop() {
	fmt.Printf("* Yoyo stop\n")
}
func (a *Yoyo) OnTerminated(pid agent.PID, reason int) {}
func (a *Yoyo) OnMessage(msg interface{}) interface{}  { return nil }
func (a *Yoyo) OnTick(delta time.Duration) {
	if a.countdown > 0 {
		a.countdown -= delta
		if a.countdown <= 0 {
			lalaPid, ok := agent.Lookup("127.0.0.1:10001/test/gigi/lala")
			if ok {
				fmt.Printf("* Lala pid is %s\n", agent.Address(lalaPid))
				fmt.Printf("* Yoyo call Lala message\n")
				msg, _ := a.context.Call(lalaPid, &Hi{Say: "hello lala"})
				fmt.Printf("* result %v\n", msg)
				// panic("stop")
				a.context.Stop()
			}
		}
	}
}

type Hi struct {
	Say string
}
