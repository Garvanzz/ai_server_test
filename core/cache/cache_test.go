package cache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"xfx/pkg/log"
)

func TestMain(m *testing.M) {
	log.DefaultInit()
	m.Run()
}

// ─── helpers ───

func newTestCache(cap int, ttl time.Duration, save SaveFunc[string, int]) *WriteBackCache[string, int] {
	return New[string, int](Options[string, int]{
		Capacity:   cap,
		DefaultTTL: ttl,
		SaveFunc:   save,
	})
}

// ─── Basic Set / Get / Del ───

func TestSetGet(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Fatalf("expected (1, true), got (%d, %v)", v, ok)
	}
}

func TestGetMiss(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected miss")
	}
}

func TestDel(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	c.Del("a")
	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected miss after Del")
	}
}

func TestDelNonExistent(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()
	c.Del("nope") // should not panic
}

func TestOverwrite(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	c.Set("a", 2)
	v, ok := c.Get("a")
	if !ok || v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
	if c.Len() != 1 {
		t.Fatalf("expected len=1, got %d", c.Len())
	}
}

// ─── TTL ───

func TestTTLExpiration(t *testing.T) {
	c := newTestCache(100, 50*time.Millisecond, nil)
	defer c.Close()

	c.Set("a", 1)
	time.Sleep(80 * time.Millisecond)

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected miss after TTL")
	}
}

func TestTTLNotExpired(t *testing.T) {
	c := newTestCache(100, 200*time.Millisecond, nil)
	defer c.Close()

	c.Set("a", 1)
	time.Sleep(50 * time.Millisecond)

	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Fatal("should not expire yet")
	}
}

func TestSetWithTTL(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	c.SetWithTTL("short", 1, 50*time.Millisecond)
	c.SetWithTTL("long", 2, time.Minute)

	time.Sleep(80 * time.Millisecond)

	_, ok := c.Get("short")
	if ok {
		t.Fatal("short should have expired")
	}
	v, ok := c.Get("long")
	if !ok || v != 2 {
		t.Fatal("long should still be alive")
	}
}

// ─── TTL 过期触发脏数据落库 ───

func TestTTLExpirySavesDirty(t *testing.T) {
	var saved sync.Map
	save := func(k string, v int) bool {
		saved.Store(k, v)
		return true
	}
	c := newTestCache(100, 50*time.Millisecond, save)
	defer c.Close()

	c.Set("dirty", 42)
	time.Sleep(80 * time.Millisecond)

	// Get 触发过期检测 → removeEntry → safeSave
	_, ok := c.Get("dirty")
	if ok {
		t.Fatal("should be expired")
	}

	v, loaded := saved.Load("dirty")
	if !loaded || v.(int) != 42 {
		t.Fatalf("dirty entry should have been saved, got loaded=%v v=%v", loaded, v)
	}
}

func TestTTLExpirySkipsClean(t *testing.T) {
	saveCalled := int32(0)
	save := func(k string, v int) bool {
		atomic.AddInt32(&saveCalled, 1)
		return true
	}
	c := newTestCache(100, 50*time.Millisecond, save)
	defer c.Close()

	c.SetClean("clean", 99)
	time.Sleep(80 * time.Millisecond)

	c.Get("clean") // expired clean entry
	if atomic.LoadInt32(&saveCalled) != 0 {
		t.Fatal("clean entry should not trigger save on expiry")
	}
}

// ─── Dirty tracking ───

func TestDirtyTracking(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	if c.DirtyCount() != 1 {
		t.Fatalf("expected 1 dirty, got %d", c.DirtyCount())
	}

	c.SetClean("b", 2)
	if c.DirtyCount() != 1 {
		t.Fatalf("SetClean should not increase dirty count, got %d", c.DirtyCount())
	}
}

func TestSetCleanDoesNotClearExistingDirty(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1) // dirty
	c.SetClean("a", 2)
	if c.DirtyCount() != 1 {
		t.Fatal("SetClean on existing dirty entry should preserve dirty flag")
	}
}

// ─── LRU capacity eviction ───

func TestCapacityEviction(t *testing.T) {
	c := newTestCache(3, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	if c.Len() != 3 {
		t.Fatalf("expected 3, got %d", c.Len())
	}

	c.Set("d", 4) // should evict "a" (LRU)
	if c.Len() != 3 {
		t.Fatalf("expected 3 after eviction, got %d", c.Len())
	}

	_, ok := c.Get("a")
	if ok {
		t.Fatal("'a' should have been evicted (LRU)")
	}
	v, ok := c.Get("d")
	if !ok || v != 4 {
		t.Fatal("'d' should exist")
	}
}

func TestCapacityEvictionSavesDirty(t *testing.T) {
	var saved sync.Map
	save := func(k string, v int) bool {
		saved.Store(k, v)
		return true
	}
	c := newTestCache(2, time.Minute, save)
	defer c.Close()

	c.Set("a", 1) // dirty
	c.Set("b", 2) // dirty
	c.Set("c", 3) // evicts "a"

	v, ok := saved.Load("a")
	if !ok || v.(int) != 1 {
		t.Fatalf("evicted dirty 'a' should be saved, got ok=%v v=%v", ok, v)
	}
}

func TestCapacityEvictionSkipsClean(t *testing.T) {
	saveCalled := int32(0)
	save := func(k string, v int) bool {
		atomic.AddInt32(&saveCalled, 1)
		return true
	}
	c := newTestCache(2, time.Minute, save)
	defer c.Close()

	c.SetClean("a", 1) // clean
	c.SetClean("b", 2) // clean
	c.Set("c", 3)      // evicts "a" (clean)

	if atomic.LoadInt32(&saveCalled) != 0 {
		t.Fatal("evicting clean entry should not call saveFunc")
	}
}

func TestLRUOrder(t *testing.T) {
	c := newTestCache(3, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// access "a" → moves to front
	c.Get("a")

	c.Set("d", 4) // should evict "b" (now LRU)

	_, ok := c.Get("b")
	if ok {
		t.Fatal("'b' should have been evicted")
	}
	_, ok = c.Get("a")
	if !ok {
		t.Fatal("'a' should still be present (recently accessed)")
	}
}

// ─── Flush ───

func TestFlush(t *testing.T) {
	var saved sync.Map
	save := func(k string, v int) bool {
		saved.Store(k, v)
		return true
	}
	c := newTestCache(100, time.Minute, save)
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.SetClean("c", 3)

	c.Flush()

	if _, ok := saved.Load("a"); !ok {
		t.Fatal("'a' should be flushed")
	}
	if _, ok := saved.Load("b"); !ok {
		t.Fatal("'b' should be flushed")
	}
	if _, ok := saved.Load("c"); ok {
		t.Fatal("'c' (clean) should not be flushed")
	}

	if c.DirtyCount() != 0 {
		t.Fatal("dirty count should be 0 after flush")
	}

	// entries still in cache
	if c.Len() != 3 {
		t.Fatal("flush should not remove entries")
	}
}

func TestFlushRetryOnFailure(t *testing.T) {
	failCount := int32(0)
	save := func(k string, v int) bool {
		if k == "fail" && atomic.AddInt32(&failCount, 1) == 1 {
			return false // first attempt fails
		}
		return true
	}
	c := newTestCache(100, time.Minute, save)
	defer c.Close()

	c.Set("fail", 1)
	c.Flush() // first flush: save fails → re-marked dirty

	if c.DirtyCount() != 1 {
		t.Fatal("failed save should re-mark entry as dirty")
	}

	c.Flush() // second flush: succeeds
	if c.DirtyCount() != 0 {
		t.Fatal("dirty count should be 0 after successful retry")
	}
}

// ─── Close ───

func TestCloseFlushesAll(t *testing.T) {
	var saved sync.Map
	save := func(k string, v int) bool {
		saved.Store(k, v)
		return true
	}
	c := New[string, int](Options[string, int]{
		Capacity:      100,
		DefaultTTL:    time.Minute,
		FlushInterval: time.Hour, // won't trigger during test
		SaveFunc:      save,
	})

	c.Set("a", 1)
	c.Set("b", 2)

	c.Close()

	if _, ok := saved.Load("a"); !ok {
		t.Fatal("Close should flush 'a'")
	}
	if _, ok := saved.Load("b"); !ok {
		t.Fatal("Close should flush 'b'")
	}
}

func TestDoubleClose(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	c.Close()
	c.Close() // should not panic
}

// ─── Del saves dirty ───

func TestDelSavesDirty(t *testing.T) {
	var saved sync.Map
	save := func(k string, v int) bool {
		saved.Store(k, v)
		return true
	}
	c := newTestCache(100, time.Minute, save)
	defer c.Close()

	c.Set("a", 42)
	c.Del("a")

	v, ok := saved.Load("a")
	if !ok || v.(int) != 42 {
		t.Fatalf("Del on dirty entry should save, got ok=%v v=%v", ok, v)
	}
}

func TestDelCleanNoSave(t *testing.T) {
	saveCalled := int32(0)
	save := func(k string, v int) bool {
		atomic.AddInt32(&saveCalled, 1)
		return true
	}
	c := newTestCache(100, time.Minute, save)
	defer c.Close()

	c.SetClean("a", 1)
	c.Del("a")

	if atomic.LoadInt32(&saveCalled) != 0 {
		t.Fatal("Del on clean entry should not call saveFunc")
	}
}

// ─── Iterate ───

func TestIterate(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.SetClean("c", 3)

	sum := 0
	count := 0
	c.Iterate(func(k string, v int) bool {
		sum += v
		count++
		return true
	})
	if count != 3 || sum != 6 {
		t.Fatalf("expected 3 entries sum=6, got count=%d sum=%d", count, sum)
	}
}

func TestIterateSkipsExpired(t *testing.T) {
	c := newTestCache(100, 50*time.Millisecond, nil)
	defer c.Close()

	c.Set("short", 1)
	c.SetWithTTL("long", 2, time.Minute)

	time.Sleep(80 * time.Millisecond)

	count := 0
	c.Iterate(func(k string, v int) bool {
		count++
		return true
	})
	if count != 1 {
		t.Fatalf("expected 1 (long only), got %d", count)
	}
}

func TestIterateEarlyStop(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	for i := 0; i < 10; i++ {
		c.Set(fmt.Sprintf("k%d", i), i)
	}

	count := 0
	c.Iterate(func(k string, v int) bool {
		count++
		return count < 3
	})
	if count != 3 {
		t.Fatalf("expected 3 iterations, got %d", count)
	}
}

// ─── Unlimited capacity ───

func TestUnlimitedCapacity(t *testing.T) {
	c := newTestCache(0, time.Minute, nil)
	defer c.Close()

	for i := 0; i < 1000; i++ {
		c.Set(fmt.Sprintf("k%d", i), i)
	}
	if c.Len() != 1000 {
		t.Fatalf("expected 1000, got %d", c.Len())
	}
}

// ─── Background flush loop ───

func TestAutoFlush(t *testing.T) {
	var saved sync.Map
	save := func(k string, v int) bool {
		saved.Store(k, v)
		return true
	}
	c := New[string, int](Options[string, int]{
		Capacity:      100,
		DefaultTTL:    time.Minute,
		FlushInterval: 50 * time.Millisecond,
		SaveFunc:      save,
	})
	defer c.Close()

	c.Set("auto", 99)

	time.Sleep(150 * time.Millisecond)

	if _, ok := saved.Load("auto"); !ok {
		t.Fatal("auto-flush should have saved 'auto'")
	}
	if c.DirtyCount() != 0 {
		t.Fatal("dirty count should be 0 after auto-flush")
	}
	if c.Len() != 1 {
		t.Fatal("entry should still be in cache after flush")
	}
}

func TestAutoCleanExpired(t *testing.T) {
	saveCalled := int32(0)
	save := func(k string, v int) bool {
		atomic.AddInt32(&saveCalled, 1)
		return true
	}
	c := New[string, int](Options[string, int]{
		Capacity:      100,
		DefaultTTL:    80 * time.Millisecond,
		FlushInterval: 50 * time.Millisecond,
		SaveFunc:      save,
	})
	defer c.Close()

	c.Set("will-expire", 1) // dirty, short TTL

	time.Sleep(200 * time.Millisecond)

	if c.Len() != 0 {
		t.Fatalf("expired entry should be cleaned, len=%d", c.Len())
	}
	if atomic.LoadInt32(&saveCalled) < 1 {
		t.Fatal("dirty expired entry should be saved before removal")
	}
}

// ─── saveFunc panic recovery ───

func TestSaveFuncPanicRecovery(t *testing.T) {
	c := newTestCache(2, time.Minute, func(k string, v int) bool {
		panic("boom")
	})
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)

	// eviction triggers saveFunc which panics → should not crash
	c.Set("c", 3)

	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}

func TestFlushPanicRecovery(t *testing.T) {
	c := newTestCache(100, time.Minute, func(k string, v int) bool {
		panic("boom in flush")
	})
	defer c.Close()

	c.Set("a", 1)
	c.Flush() // should not crash
}

// ─── Concurrent access ───

func TestConcurrentReadWrite(t *testing.T) {
	c := newTestCache(1000, time.Minute, func(k string, v int) bool {
		return true
	})
	defer c.Close()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("g%d-k%d", id, j)
				c.Set(key, j)
				c.Get(key)
			}
		}(i)
	}
	wg.Wait()

	if c.Len() > 1000 {
		t.Fatalf("should not exceed capacity, got %d", c.Len())
	}
}

func TestConcurrentFlush(t *testing.T) {
	var saveCount int64
	c := newTestCache(1000, time.Minute, func(k string, v int) bool {
		atomic.AddInt64(&saveCount, 1)
		return true
	})
	defer c.Close()

	for i := 0; i < 100; i++ {
		c.Set(fmt.Sprintf("k%d", i), i)
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Flush()
		}()
	}
	wg.Wait()

	// each entry should be saved at least once across all flushes
	if atomic.LoadInt64(&saveCount) < 100 {
		t.Fatalf("expected at least 100 saves, got %d", atomic.LoadInt64(&saveCount))
	}
	if c.DirtyCount() != 0 {
		t.Fatal("all entries should be clean after flush")
	}
}

// ─── Len / DirtyCount ───

func TestLenAndDirtyCount(t *testing.T) {
	c := newTestCache(100, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	c.SetClean("b", 2)
	c.Set("c", 3)

	if c.Len() != 3 {
		t.Fatalf("expected len=3, got %d", c.Len())
	}
	if c.DirtyCount() != 2 {
		t.Fatalf("expected dirty=2, got %d", c.DirtyCount())
	}
}

// ─── Edge: nil saveFunc ───

func TestNilSaveFunc(t *testing.T) {
	c := newTestCache(2, time.Minute, nil)
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // evicts "a", no saveFunc → should not panic
	c.Del("b")    // should not panic
	c.Flush()     // should not panic
}

// ─── Integration: full lifecycle ───

func TestFullLifecycle(t *testing.T) {
	var saved sync.Map
	save := func(k string, v int) bool {
		saved.Store(k, v)
		return true
	}

	c := New[string, int](Options[string, int]{
		Capacity:      5,
		DefaultTTL:    100 * time.Millisecond,
		FlushInterval: 60 * time.Millisecond,
		SaveFunc:      save,
	})

	// phase 1: load data from "DB" (clean)
	c.SetClean("player:1", 100)
	c.SetClean("player:2", 200)

	// phase 2: modify some data (dirty)
	c.Set("player:1", 150)
	c.Set("player:3", 300)

	// phase 3: wait for auto-flush
	time.Sleep(100 * time.Millisecond)

	v, ok := saved.Load("player:1")
	if !ok || v.(int) != 150 {
		t.Fatalf("player:1 should be flushed with value 150, got ok=%v v=%v", ok, v)
	}
	v, ok = saved.Load("player:3")
	if !ok || v.(int) != 300 {
		t.Fatalf("player:3 should be flushed")
	}
	if _, ok := saved.Load("player:2"); ok {
		t.Fatal("player:2 (clean) should not be flushed")
	}

	// phase 4: modify after flush
	c.Set("player:2", 250)

	// phase 5: close → final flush
	c.Close()

	v, ok = saved.Load("player:2")
	if !ok || v.(int) != 250 {
		t.Fatalf("player:2 should be saved on Close, got ok=%v v=%v", ok, v)
	}
}
