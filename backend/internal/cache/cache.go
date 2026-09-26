package cache

import (
	"encoding/json"
	"sync"
	"time"
)

type Entry struct {
	Value     interface{}
	ExpiresAt time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]Entry
	ttl   time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{items: make(map[string]Entry), ttl: ttl}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Entry{Value: value, ExpiresAt: time.Now().Add(c.ttl)}
}

func (c *Cache) Get(key string, out interface{}) bool {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return false
	}
	if time.Now().After(item.ExpiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return false
	}
	bytes, err := json.Marshal(item.Value)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(bytes, out); err != nil {
		return false
	}
	return true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}
