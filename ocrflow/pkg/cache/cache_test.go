package cache

import (
	"testing"
	"time"
)

func TestBoundedCacheEvictsLeastRecentlyUsed(t *testing.T) {
	c := NewCache(WithLimits(2, 1000))
	c.SetWithTTLAndSize("a", "a", time.Minute, 10)
	c.SetWithTTLAndSize("b", "b", time.Minute, 10)
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected a in cache")
	}
	c.SetWithTTLAndSize("c", "c", time.Minute, 10)

	if _, ok := c.Get("b"); ok {
		t.Fatal("expected least-recently-used entry b to be evicted")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected recently-used entry a to remain")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("expected new entry c to remain")
	}
}

func TestBoundedCacheEnforcesByteLimit(t *testing.T) {
	c := NewCache(WithLimits(10, 100))
	c.SetWithTTLAndSize("a", "a", time.Minute, 30)
	c.SetWithTTLAndSize("b", "b", time.Minute, 30)

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected a to be evicted to satisfy byte limit")
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatal("expected b to remain")
	}

	c.SetWithTTLAndSize("oversized", "value", time.Minute, 101)
	if _, ok := c.Get("oversized"); ok {
		t.Fatal("value larger than total byte limit must not be cached")
	}
}
