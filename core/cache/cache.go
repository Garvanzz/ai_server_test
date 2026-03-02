package cache

import (
	"container/list"
	"runtime"
	"sync"
	"time"
	"xfx/pkg/log"
)

type SaveFunc[K comparable, V any] func(K, V) bool

type Options[K comparable, V any] struct {
	Capacity      int           // 最大条目数；0 表示不限
	DefaultTTL    time.Duration // Set 的默认 TTL；<= 0 默认 5m
	FlushInterval time.Duration // 后台脏数据刷盘间隔；0 表示不启动后台刷
	SaveFunc      SaveFunc[K, V]
}

type entry[K comparable, V any] struct {
	key       K
	value     V
	dirty     bool
	expiresAt time.Time
	elem      *list.Element
}

type WriteBackCache[K comparable, V any] struct {
	mu        sync.Mutex
	entries   map[K]*entry[K, V]
	lru       *list.List
	opts      Options[K, V]
	stopCh    chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

func New[K comparable, V any](opts Options[K, V]) *WriteBackCache[K, V] {
	if opts.DefaultTTL <= 0 {
		opts.DefaultTTL = 5 * time.Minute
	}
	if opts.Capacity < 0 {
		opts.Capacity = 0
	}

	c := &WriteBackCache[K, V]{
		entries: make(map[K]*entry[K, V]),
		lru:     list.New(),
		opts:    opts,
		stopCh:  make(chan struct{}),
	}

	if opts.FlushInterval > 0 {
		c.wg.Add(1)
		go c.flushLoop()
	}

	return c
}

// Set 写入数据并标记为脏。
func (c *WriteBackCache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.opts.DefaultTTL)
}

// SetWithTTL 写入数据，自定义 TTL，标记为脏。
func (c *WriteBackCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setLocked(key, value, ttl, true)
}

// SetClean 写入数据但不标记脏（用于从 DB 加载的数据放入缓存）。
func (c *WriteBackCache[K, V]) SetClean(key K, value V) {
	c.SetCleanWithTTL(key, value, c.opts.DefaultTTL)
}

// SetCleanWithTTL 写入数据，自定义 TTL，不标记脏。
func (c *WriteBackCache[K, V]) SetCleanWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setLocked(key, value, ttl, false)
}

func (c *WriteBackCache[K, V]) setLocked(key K, value V, ttl time.Duration, markDirty bool) {
	if e, ok := c.entries[key]; ok {
		e.value = value
		if markDirty {
			e.dirty = true
		}
		e.expiresAt = time.Now().Add(ttl)
		c.lru.MoveToFront(e.elem)
		return
	}

	if c.opts.Capacity > 0 && len(c.entries) >= c.opts.Capacity {
		c.evictOldest()
	}

	e := &entry[K, V]{
		key:       key,
		value:     value,
		dirty:     markDirty,
		expiresAt: time.Now().Add(ttl),
	}
	e.elem = c.lru.PushFront(e)
	c.entries[key] = e
}

// Get 获取缓存值。过期条目会先落库（如果脏）再移除。
func (c *WriteBackCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if !ok {
		var zero V
		return zero, false
	}

	if time.Now().After(e.expiresAt) {
		c.removeEntry(e)
		var zero V
		return zero, false
	}

	c.lru.MoveToFront(e.elem)
	return e.value, true
}

// Del 删除条目。脏数据会先落库。
func (c *WriteBackCache[K, V]) Del(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.entries[key]; ok {
		c.removeEntry(e)
	}
}

// removeEntry 落库脏数据并从缓存移除。调用方必须持有 mu。
func (c *WriteBackCache[K, V]) removeEntry(e *entry[K, V]) {
	if e.dirty && c.opts.SaveFunc != nil {
		c.safeSave(e.key, e.value)
	}
	c.lru.Remove(e.elem)
	delete(c.entries, e.key)
}

// evictOldest 驱逐 LRU 尾部条目。调用方必须持有 mu。
func (c *WriteBackCache[K, V]) evictOldest() {
	elem := c.lru.Back()
	if elem == nil {
		return
	}
	c.removeEntry(elem.Value.(*entry[K, V]))
}

func (c *WriteBackCache[K, V]) safeSave(key K, value V) bool {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			log.Error("cache: saveFunc panic, key=%v: %v\n%s", key, r, buf[:n])
		}
	}()
	if ok := c.opts.SaveFunc(key, value); !ok {
		log.Error("cache: save failed, key=%v", key)
		return false
	}
	return true
}

// Flush 将所有脏数据持久化，不移除缓存条目。
// 释放锁后批量写入，写入失败的条目会重新标记为脏。
func (c *WriteBackCache[K, V]) Flush() {
	if c.opts.SaveFunc == nil {
		return
	}

	type snapshot struct {
		key   K
		value V
	}

	c.mu.Lock()
	var dirtyItems []snapshot
	for _, e := range c.entries {
		if e.dirty {
			dirtyItems = append(dirtyItems, snapshot{key: e.key, value: e.value})
			e.dirty = false
		}
	}
	c.mu.Unlock()

	for _, item := range dirtyItems {
		ok := c.safeSave(item.key, item.value)
		if !ok {
			c.mu.Lock()
			if e, exists := c.entries[item.key]; exists {
				e.dirty = true
			}
			c.mu.Unlock()
		}
	}
}

// cleanExpired 移除所有过期条目（脏数据先落库）。
func (c *WriteBackCache[K, V]) cleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var expired []*entry[K, V]
	for _, e := range c.entries {
		if now.After(e.expiresAt) {
			expired = append(expired, e)
		}
	}
	for _, e := range expired {
		c.removeEntry(e)
	}
}

func (c *WriteBackCache[K, V]) flushLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.opts.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.Flush()
			c.cleanExpired()
		case <-c.stopCh:
			return
		}
	}
}

// Close 停止后台刷盘，并将所有脏数据最终落库。
func (c *WriteBackCache[K, V]) Close() {
	c.closeOnce.Do(func() {
		close(c.stopCh)
		c.wg.Wait()
		c.Flush()
	})
}

// Iterate 遍历所有未过期条目。回调返回 false 停止遍历。
func (c *WriteBackCache[K, V]) Iterate(fn func(K, V) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for _, e := range c.entries {
		if now.Before(e.expiresAt) {
			if !fn(e.key, e.value) {
				return
			}
		}
	}
}

// Len 返回缓存条目数。
func (c *WriteBackCache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// DirtyCount 返回脏条目数。
func (c *WriteBackCache[K, V]) DirtyCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, e := range c.entries {
		if e.dirty {
			n++
		}
	}
	return n
}
