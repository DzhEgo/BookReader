package cache

import (
	"sync"
	"time"
)

type MemoryCacheService interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Delete(key string)
	Clean()
}

type item struct {
	value   any
	expires time.Time
}

type memoryCache struct {
	mu   sync.RWMutex
	data map[string]item
	ttl  time.Duration
}

func NewService() MemoryCacheService {
	c := &memoryCache{
		data: make(map[string]item),
		ttl:  time.Minute * 30,
	}

	go func() {
		ticker := time.NewTicker(time.Minute * 2)
		defer ticker.Stop()
		for _ = range ticker.C {
			c.cleanup()
		}
	}()

	return c
}

func (mc *memoryCache) Get(key string) (any, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	v, ok := mc.data[key]
	if !ok || time.Now().After(v.expires) {
		return nil, false
	}

	return v.value, ok
}

func (mc *memoryCache) Set(key string, value any) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.data[key] = item{
		value:   value,
		expires: time.Now().Add(mc.ttl),
	}
}

func (mc *memoryCache) Delete(key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	delete(mc.data, key)
}

func (mc *memoryCache) Clean() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.data = make(map[string]item)
}

func (mc *memoryCache) cleanup() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	for k, v := range mc.data {
		if now.After(v.expires) {
			delete(mc.data, k)
		}
	}
}
