package cache

import (
	"container/list"
	"reflect"
	"sync"
	"time"
)

const maxEntries = 1000

type Entry struct {
	Value     interface{}
	ExpiresAt time.Time
}

type Cache struct {
	mu    sync.Mutex
	items map[string]*list.Element
	lru   *list.List
	ttl   time.Duration
}

type cacheItem struct {
	key   string
	entry Entry
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]*list.Element),
		lru:   list.New(),
		ttl:   ttl,
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := Entry{Value: value, ExpiresAt: time.Now().Add(c.ttl)}
	if element, ok := c.items[key]; ok {
		element.Value.(*cacheItem).entry = entry
		c.lru.MoveToFront(element)
		return
	}
	if len(c.items) >= maxEntries {
		c.remove(c.lru.Back())
	}
	element := c.lru.PushFront(&cacheItem{key: key, entry: entry})
	c.items[key] = element
}

// Get copies the stored snapshot into out. Stored values must be immutable after
// Set, and callers must not mutate reference fields returned from the snapshot.
func (c *Cache) Get(key string, out interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	target := reflect.ValueOf(out)
	if target.Kind() != reflect.Ptr || target.IsNil() {
		return false
	}
	element, ok := c.items[key]
	if !ok {
		return false
	}
	item := element.Value.(*cacheItem)
	if time.Now().After(item.entry.ExpiresAt) {
		c.remove(element)
		return false
	}
	c.lru.MoveToFront(element)

	value := reflect.ValueOf(item.entry.Value)
	if !value.IsValid() {
		target.Elem().Set(reflect.Zero(target.Elem().Type()))
		return true
	}
	if !value.Type().AssignableTo(target.Elem().Type()) {
		return false
	}
	target.Elem().Set(value)
	return true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, ok := c.items[key]; ok {
		c.remove(element)
	}
}

func (c *Cache) remove(element *list.Element) {
	if element == nil {
		return
	}
	item := element.Value.(*cacheItem)
	delete(c.items, item.key)
	c.lru.Remove(element)
}
