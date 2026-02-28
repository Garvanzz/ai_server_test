package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"xfx/pkg/agent"
)

type testAgent struct{}

func (a *testAgent) OnStart(ctx agent.Context) {
	fmt.Printf("* testAgent start\n")
	//ctx.Create("gigi", &Gigi{})
}

func (a *testAgent) OnStop() {
	fmt.Printf("* testAgent stop\n")
}

func (a *testAgent) OnTerminated(pid agent.PID, reason int) {}
func (a *testAgent) OnMessage(msg interface{}) interface{}  { return nil }
func (a *testAgent) OnTick(delta time.Duration)             {}

func main() {
	system := agent.NewSystem(
		agent.WithName("test"),
		agent.WithHost("192.168.3.34"),
		agent.WithPort(10001),
		agent.WithAgent(&testAgent{}),
		agent.WithTick(time.Second),
		agent.WithRestart(),
	)
	system.Start()
	fmt.Print("agent system start\n")

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c

	// 一定时间内关不了则强制关闭
	timeout := time.NewTimer(time.Second)
	wait := make(chan struct{})
	go func() {
		system.Stop()
		wait <- struct{}{}
	}()
	select {
	case <-timeout.C:
		panic(fmt.Sprintf("app close timeout (signal: %v)", sig))
	case <-wait:
		fmt.Printf("app closing down (signal: %v)\n", sig)
	}
}
