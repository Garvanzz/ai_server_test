// clock_test.go: 游戏逻辑时间源的功能测试与性能对比（Atomic vs RWMutex）。
//
// 运行方式：
//   go test -v ./pkg/utils -run "TestSetTimeOffset|TestNow"   # 功能测试
//   go test -bench=BenchmarkNow -benchmem ./pkg/utils        # 单 goroutine 性能
//   go test -bench=BenchmarkNow -benchmem -count=3 ./pkg/utils # 并行 + 多次取平均
package utils

import (
	"sync"
	"testing"
	"time"
)

// ========== 功能测试 ==========

func TestSetTimeOffset_Zero(t *testing.T) {
	SetTimeOffsetEnabled(true)
	SetTimeOffset(0)
	got := Now()
	real := time.Now()
	diff := got.Sub(real)
	if diff < -time.Millisecond || diff > time.Millisecond {
		t.Errorf("offset=0: Now() 应与真实时间接近, got diff=%v", diff)
	}
}

func TestSetTimeOffset_NonZero(t *testing.T) {
	SetTimeOffsetEnabled(true)
	offset := 7 * 24 * time.Hour
	SetTimeOffset(offset)
	got := Now()
	real := time.Now()
	expected := real.Add(offset)
	diff := got.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("offset=7d: Now() 应比真实时间快约 7 天, got diff from expected=%v", diff)
	}
	// 恢复，避免影响其他测试
	SetTimeOffset(0)
}

func TestSetTimeOffset_Overwrite(t *testing.T) {
	SetTimeOffsetEnabled(true)
	SetTimeOffset(24 * time.Hour)
	t1 := Now()
	SetTimeOffset(2 * 24 * time.Hour)
	t2 := Now()
	// t2 应比 t1 多约 1 天（第二次设置后）
	diff := t2.Sub(t1)
	if diff < time.Hour || diff > 3*24*time.Hour {
		t.Errorf("覆盖偏移后 Now() 应反映新偏移, t2-t1=%v", diff)
	}
	SetTimeOffset(0)
}

func TestNow_MonotonicIncrease(t *testing.T) {
	SetTimeOffsetEnabled(true)
	SetTimeOffset(0)
	prev := Now()
	for i := 0; i < 5; i++ {
		time.Sleep(10 * time.Millisecond)
		cur := Now()
		if cur.Before(prev) {
			t.Errorf("Now() 应单调递增: prev=%v cur=%v", prev, cur)
		}
		prev = cur
	}
}

func TestNow_ConcurrentReads(t *testing.T) {
	SetTimeOffsetEnabled(true)
	SetTimeOffset(5 * time.Hour)
	defer SetTimeOffset(0)
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				_ = Now()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

// ========== RWMutex 实现（仅用于性能对比） ==========

var (
	timeOffsetMutex time.Duration
	offsetMu        sync.RWMutex
)

func setTimeOffsetMutex(offset time.Duration) {
	offsetMu.Lock()
	timeOffsetMutex = offset
	offsetMu.Unlock()
}

func nowMutex() time.Time {
	offsetMu.RLock()
	d := timeOffsetMutex
	offsetMu.RUnlock()
	return time.Now().Add(d)
}

// ========== 性能测试：Atomic vs RWMutex ==========

func BenchmarkNow_Atomic(b *testing.B) {
	SetTimeOffsetEnabled(true)
	SetTimeOffset(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Now()
	}
}

func BenchmarkNow_RWMutex(b *testing.B) {
	SetTimeOffsetEnabled(true)
	setTimeOffsetMutex(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nowMutex()
	}
}

func BenchmarkNow_Atomic_Parallel(b *testing.B) {
	SetTimeOffsetEnabled(true)
	SetTimeOffset(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = Now()
		}
	})
}

func BenchmarkNow_RWMutex_Parallel(b *testing.B) {
	SetTimeOffsetEnabled(true)
	setTimeOffsetMutex(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = nowMutex()
		}
	})
}

// SetTimeOffset 写路径对比（通常只调一次，可选跑）
func BenchmarkSetTimeOffset_Atomic(b *testing.B) {
	SetTimeOffsetEnabled(true)
	offset := 7 * 24 * time.Hour
	for i := 0; i < b.N; i++ {
		SetTimeOffset(offset)
	}
}

func BenchmarkSetTimeOffset_RWMutex(b *testing.B) {
	SetTimeOffsetEnabled(true)
	offset := 7 * 24 * time.Hour
	for i := 0; i < b.N; i++ {
		setTimeOffsetMutex(offset)
	}
}
