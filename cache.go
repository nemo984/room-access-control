package main

import (
	"sync"
	"time"
)

type cache[K comparable, V any] struct {
	mu    sync.RWMutex
	store map[K]struct {
		V         V
		updatedAt time.Time
		expireDur time.Duration
	}
}

func newCache[K comparable, V any]() *cache[K, V] {
	return &cache[K, V]{
		store: make(map[K]struct {
			V         V
			updatedAt time.Time
			expireDur time.Duration
		}),
	}
}

func (c *cache[K, V]) get(k K) (V, bool) {
	c.mu.RLock()
	v, ok := c.store[k]
	isExpired := ok && time.Since(v.updatedAt) > v.expireDur
	c.mu.RUnlock()
	if !ok {
		return v.V, false
	}

	if isExpired {
		c.remove(k)
		return v.V, false
	}
	return v.V, ok
}

func (c *cache[K, V]) put(k K, v V, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[k] = struct {
		V         V
		updatedAt time.Time
		expireDur time.Duration
	}{
		V:         v,
		updatedAt: time.Now(),
		expireDur: d,
	}
}

func (c *cache[K, V]) remove(k K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, k)
}

func (c *cache[K, V]) removeKeys(keys []K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range keys {
		delete(c.store, k)
	}
}

func (c *cache[K, V]) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[K]struct {
		V         V
		updatedAt time.Time
		expireDur time.Duration
	})
}
