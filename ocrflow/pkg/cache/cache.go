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
	sizeBytes int64
	lastUsed  uint64
}

func (e cacheEntry) expired() bool {
	if e.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.expiresAt)
}

type Cache struct {
	mu           sync.RWMutex
	data         map[string]cacheEntry
	warmed       bool
	maxEntries   int
	maxSizeBytes int64
	sizeBytes    int64
	useCounter   uint64
}

type Option func(*Cache)

// WithLimits bounds a cache by both entry count and caller-estimated value
// bytes. A non-positive limit disables that particular bound.
func WithLimits(maxEntries int, maxSizeBytes int64) Option {
	return func(c *Cache) {
		c.maxEntries = maxEntries
		c.maxSizeBytes = maxSizeBytes
	}
}

func NewCache(options ...Option) *Cache {
	c := &Cache{
		data: make(map[string]cacheEntry),
	}
	for _, option := range options {
		option(c)
	}
	return c
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ent, ok := c.data[key]
	if !ok {
		return nil, false
	}
	if ent.expired() {
		c.deleteLocked(key)
		return nil, false
	}
	c.useCounter++
	ent.lastUsed = c.useCounter
	c.data[key] = ent
	return ent.value, true
}

func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, 0)
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.SetWithTTLAndSize(key, value, ttl, 0)
}

// SetWithTTLAndSize stores a value with an estimated retained size. Bounded
// caches evict least-recently-used entries until both configured limits are
// satisfied. Values larger than the byte limit are deliberately not cached.
func (c *Cache) SetWithTTLAndSize(key string, value interface{}, ttl time.Duration, sizeBytes int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if sizeBytes < 0 {
		sizeBytes = 0
	}
	// Include the retained key and a conservative per-entry bookkeeping
	// allowance in the limit, not only the caller-estimated value size.
	sizeBytes += int64(len(key)) + 64
	if c.maxSizeBytes > 0 && sizeBytes > c.maxSizeBytes {
		c.deleteLocked(key)
		return
	}
	c.deleteLocked(key)
	c.useCounter++
	ent := cacheEntry{value: value, sizeBytes: sizeBytes, lastUsed: c.useCounter}
	if ttl > 0 {
		ent.expiresAt = time.Now().Add(ttl)
	}
	c.data[key] = ent
	c.sizeBytes += sizeBytes
	c.evictToLimitsLocked()
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deleteLocked(key)
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]cacheEntry)
	c.sizeBytes = 0
	c.warmed = false
}

func (c *Cache) deleteLocked(key string) {
	if ent, ok := c.data[key]; ok {
		c.sizeBytes -= ent.sizeBytes
		delete(c.data, key)
	}
}

// evictExpiredLocked removes expired entries. Must be called with c.mu held.
func (c *Cache) evictExpiredLocked() {
	for k, ent := range c.data {
		if ent.expired() {
			c.deleteLocked(k)
		}
	}
}

func (c *Cache) evictToLimitsLocked() {
	c.evictExpiredLocked()
	for (c.maxEntries > 0 && len(c.data) > c.maxEntries) ||
		(c.maxSizeBytes > 0 && c.sizeBytes > c.maxSizeBytes) {
		var oldestKey string
		var oldestUse uint64
		first := true
		for key, ent := range c.data {
			if first || ent.lastUsed < oldestUse {
				oldestKey = key
				oldestUse = ent.lastUsed
				first = false
			}
		}
		if first {
			return
		}
		c.deleteLocked(oldestKey)
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
	c.sizeBytes = 0
	for k, v := range data {
		c.useCounter++
		c.data[k] = cacheEntry{value: v, lastUsed: c.useCounter}
	}
	c.evictToLimitsLocked()
	c.warmed = true
	return nil
}

func (c *Cache) IsWarm() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evictExpiredLocked()
	return c.warmed
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
