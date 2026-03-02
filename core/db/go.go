package db

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"runtime"
	"sync"
	"sync/atomic"
	"xfx/pkg/agent"
	"xfx/pkg/log"
)

const defaultWorkerCount = 8

type Go struct {
	pendingGo int64
	running   int32 // 原子操作，替代无保护的 bool
	taskQueue chan *RedisJob
	system    *agent.System
	wg        sync.WaitGroup // 等待所有 worker 退出
}

type RedisJob struct {
	Pool    *redis.Pool
	Command string
	Args    []any
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
	if !atomic.CompareAndSwapInt32(&g.running, 0, 1) {
		panic("already running")
	}

	for i := 0; i < defaultWorkerCount; i++ {
		g.wg.Add(1)
		go g.worker()
	}
}

func (g *Go) stop() {
	if !atomic.CompareAndSwapInt32(&g.running, 1, 0) {
		return
	}
	close(g.taskQueue)
	g.wg.Wait()
}

// worker 从 taskQueue 消费并直接执行，固定数量的 worker 控制并发度
func (g *Go) worker() {
	defer g.wg.Done()
	for job := range g.taskQueue {
		g.executeJob(job)
	}
}

func (g *Go) executeJob(job *RedisJob) {
	atomic.AddInt64(&g.pendingGo, 1)
	defer atomic.AddInt64(&g.pendingGo, -1)

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			log.Error("redis async job panic: %v\nStack: %s", r, buf[:n])
		}
	}()

	conn := job.Pool.Get()
	defer conn.Close()

	reply, err := conn.Do(job.Command, job.Args...)
	job.Cb(reply, err)
}

func (g *Go) submitJob(pool *redis.Pool, command string, args []any, cb func(any, error)) error {
	if atomic.LoadInt32(&g.running) == 0 {
		return fmt.Errorf("service not running")
	}

	job := &RedisJob{
		Pool:    pool,
		Command: command,
		Args:    args,
		Cb:      cb,
	}

	select {
	case g.taskQueue <- job:
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}
