package cache

import (
	"github.com/dgraph-io/ristretto/v2"
	"sync"
	"time"
	"xfx/pkg/log"
)

type AutoTLCache[K ristretto.Key, V any] struct {
	cache    *ristretto.Cache[K, *cacheItem[K, V]]
	keys     map[any]struct{}
	keysMu   sync.RWMutex
	saveFunc func(K, V) bool
}

type cacheItem[K, V any] struct {
	key       K
	value     V
	expiresAt time.Time
}

func New[K ristretto.Key, V any](
	capacity int64,
	keyToHash func(K) (uint64, uint64),
	saveFunc func(K, V) bool,
) (*AutoTLCache[K, V], error) {

	config := &ristretto.Config[K, *cacheItem[K, V]]{
		NumCounters: capacity * 10,
		MaxCost:     capacity,
		BufferItems: 64,
		KeyToHash:   keyToHash,
		OnEvict: func(item *ristretto.Item[*cacheItem[K, V]]) {
			if saveFunc == nil {
				return
			}

			ok := saveFunc(item.Value.key, item.Value.value)
			if !ok {
				log.Error("cache on evict error")
			}
		},
		TtlTickerDurationInSec: 5,
	}

	cache, err := ristretto.NewCache(config)
	if err != nil {
		return nil, err
	}

	return &AutoTLCache[K, V]{
		cache:    cache,
		saveFunc: saveFunc,
		keys:     make(map[any]struct{}),
	}, nil
}

// Set 默认五分钟过期
func (c *AutoTLCache[K, V]) Set(key K, value V) bool {
	return c.SetWithTTL(key, value, 5*time.Minute)
}

func (c *AutoTLCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) bool {
	item := &cacheItem[K, V]{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	c.keysMu.Lock()
	c.keys[key] = struct{}{}
	c.keysMu.Unlock()

	return c.cache.SetWithTTL(key, item, 1, ttl)
}

func (c *AutoTLCache[K, V]) Get(key K) (V, bool) {
	if item, found := c.cache.Get(key); found {
		if time.Now().Before(item.expiresAt) {
			return item.value, true
		}

		c.cache.Del(key)
	}

	var Z V
	return Z, false
}

func (c *AutoTLCache[K, V]) Close() {
	c.cache.Wait()
	c.cache.Close()
}

// Iterate 迭代器方法
func (c *AutoTLCache[K, V]) Iterate(fn func(key K, value V) bool) {
	c.keysMu.Lock()
	defer c.keysMu.Unlock()

	for _key := range c.keys {
		key := _key.(K)
		item, ok := c.cache.Get(key)
		if !ok {
			delete(c.keys, key)
			continue
		}

		if time.Now().Before(item.expiresAt) {
			if !fn(item.key, item.value) {
				return
			}
		} else {
			c.cache.Del(key)
		}
	}
}

func (c *AutoTLCache[K, V]) Del(key K) {
	c.cache.Del(key)
}
