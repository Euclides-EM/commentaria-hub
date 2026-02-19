package cache

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/samber/lo"
)

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time // zero means no expiry
}

func (e cacheEntry) expired() bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.expiresAt)
}

type Cache struct {
	mu   sync.RWMutex
	data map[string]cacheEntry
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]cacheEntry),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ent, ok := c.data[key]
	if !ok {
		return nil, false
	}
	if ent.expired() {
		delete(c.data, key)
		return nil, false
	}
	return ent.value, true
}

func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, 0)
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ent := cacheEntry{value: value}
	if ttl > 0 {
		ent.expiresAt = time.Now().Add(ttl)
	}
	c.data[key] = ent
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry)
}

// evictExpiredLocked removes expired entries. Must be called with c.mu held.
func (c *Cache) evictExpiredLocked() {
	for k, ent := range c.data {
		if ent.expired() {
			delete(c.data, k)
		}
	}
}

func (c *Cache) Warmup(loadFunc func() (map[string]interface{}, error)) error {
	data, err := loadFunc()
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry, len(data))
	for k, v := range data {
		c.data[k] = cacheEntry{value: v}
	}
	return nil
}

func (c *Cache) IsWarm() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evictExpiredLocked()
	return len(c.data) > 0
}

func (c *Cache) GetBulk(
	filter func(a any) bool,
	compare func(k1, k2 string, v1, v2 any) int,
	offset int,
	limit int,
) (keys []any, values []any, total int, err error) {
	if offset < 0 {
		return nil, nil, 0, fmt.Errorf("offset must be >= 0, got %d", offset)
	}
	if limit < 0 {
		return nil, nil, 0, fmt.Errorf("limit must be >= 0, got %d", limit)
	}

	c.mu.Lock()
	c.evictExpiredLocked()
	items := make([]lo.Entry[string, any], 0)
	for k, ent := range c.data {
		if filter != nil && !filter(ent.value) {
			continue
		}
		items = append(items, lo.Entry[string, any]{Key: k, Value: ent.value})
	}
	c.mu.Unlock()

	// Default ordering: ascending by key
	if compare == nil {
		sort.SliceStable(items, func(i, j int) bool {
			return items[i].Key < items[j].Key
		})
	} else {
		sort.SliceStable(items, func(i, j int) bool {
			return compare(items[i].Key, items[j].Key, items[i].Value, items[j].Value) < 0
		})
	}

	if offset >= len(items) {
		return []any{}, []any{}, len(items), nil
	}

	start := offset
	end := len(items)

	// Convention: limit == 0 means "no limit" (return everything from offset)
	if limit > 0 {
		if start+limit < end {
			end = start + limit
		}
	}

	selected := items[start:end]
	keys = make([]any, 0, len(selected))
	values = make([]any, 0, len(selected))
	for _, it := range selected {
		keys = append(keys, it.Key)
		values = append(values, it.Value)
	}

	return keys, values, len(items), nil
}
