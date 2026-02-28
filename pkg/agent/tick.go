package agent

import (
	"context"
	"time"
)

func tickgo(ctx Context, tick time.Duration) func() {
	goctx, gocancel := context.WithCancel(context.Background())
	go func() {
		var (
			t1, t2 time.Time
			delta  time.Duration
		)
		ticker := time.NewTicker(tick)
		defer ticker.Stop()
		t1 = time.Now()
	process:
		for {
			select {
			case <-ticker.C:
				t2 = time.Now()
				delta, t1 = t2.Sub(t1), t2
				ctx.Cast(ctx.Self(), tickMessage(delta))
				// fmt.Printf("* tickgo tick %v\n", delta)
			case <-goctx.Done():
				break process
			}
		}
	}()
	return gocancel
}
