package main

import "sync"

type cache[K comparable, V any] struct {
	mu    sync.RWMutex
	store map[K]V
}

func newCache[K comparable, V any]() *cache[K, V] {
	return &cache[K, V]{
		store: make(map[K]V),
	}
}

func (c *cache[K, V]) get(k K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.store[k]
	return v, ok
}

func (c *cache[K, V]) put(k K, v V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[k] = v
}

func (c *cache[K, V]) remove(k K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, k)
}
