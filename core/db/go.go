package db

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"runtime"
	"sync/atomic"
	"xfx/pkg/agent"
	"xfx/pkg/log"
)

type Go struct {
	pendingGo int64          // 活跃协程计数
	isRunning bool           // 运行状态
	taskQueue chan *RedisJob // 任务队列
	system    *agent.System
}

type RedisJob struct {
	Pool    *redis.Pool
	Command string
	Args    []any
	Result  chan any
	Cb      func(any, error)
}

type RedisRet struct {
	OpType int
	Params []int64
	Reply  any
	Err    error
}

func NewGo(system *agent.System) *Go {
	return &Go{
		taskQueue: make(chan *RedisJob, 2000),
		system:    system,
	}
}

func (g *Go) start() {
	if g.isRunning {
		panic("already running")
	}

	g.isRunning = true
	go g.worker()
}

func (g *Go) stop() {
	if !g.isRunning {
		return
	}

	g.isRunning = false

	close(g.taskQueue)
}

func (g *Go) worker() {
	for job := range g.taskQueue {
		go g.executeJob(job)
	}
}

func (g *Go) executeJob(job *RedisJob) {
	defer func() {
		if r := recover(); r != nil {
			atomic.AddInt64(&g.pendingGo, -1)

			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			log.Error("Error: %v\nStack: %s", r, buf[:n])
			return
		}
	}()

	atomic.AddInt64(&g.pendingGo, 1)

	conn := job.Pool.Get()
	defer conn.Close()

	reply, err := conn.Do(job.Command, job.Args...)

	atomic.AddInt64(&g.pendingGo, -1)

	job.Cb(reply, err)
}

func (g *Go) submitJob(pool *redis.Pool, command string, args []any, cb func(any, error)) error {
	if !g.isRunning {
		return fmt.Errorf("service not running")
	}

	job := &RedisJob{
		Pool:    pool,
		Command: command,
		Args:    args,
		Result:  make(chan any, 1),
		Cb:      cb,
	}

	select {
	case g.taskQueue <- job:
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}
