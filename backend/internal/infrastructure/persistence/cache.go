package persistence

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrCacheMiss = errors.New("cache entry not found or expired")

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]cacheItem),
	}
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, ErrCacheMiss
	}

	if time.Now().After(item.expiresAt) {
		return nil, ErrCacheMiss
	}

	return item.value, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}
