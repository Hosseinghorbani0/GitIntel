package cache

import (
	"fmt"
	"testing"
	"time"
)

func TestCacheEvictsOldestEntriesAtCapacity(t *testing.T) {
	cache := NewCache(time.Minute)
	for index := range 1500 {
		cache.Set(fmt.Sprintf("profile:%04d", index), index)
	}

	if got := len(cache.items); got != maxEntries {
		t.Fatalf("cache contains %d entries, want exactly %d", got, maxEntries)
	}
	for index := range 1500 {
		var value int
		found := cache.Get(fmt.Sprintf("profile:%04d", index), &value)
		if want := index >= 500; found != want {
			t.Errorf("entry %d found = %t, want %t", index, found, want)
		}
		if found && value != index {
			t.Errorf("entry %d contains value %d", index, value)
		}
	}
}

func TestCacheGetRefreshesLRURecency(t *testing.T) {
	cache := NewCache(time.Minute)
	cache.Set("oldest", "oldest")
	cache.Set("middle", "middle")
	cache.Set("newest", "newest")

	var value string
	if !cache.Get("oldest", &value) {
		t.Fatal("expected to read oldest entry before eviction")
	}
	for index := 0; index < maxEntries-3; index++ {
		cache.Set(fmt.Sprintf("filler:%d", index), index)
	}
	cache.Set("last", "last")

	if cache.Get("middle", &value) {
		t.Fatal("least recently used entry should have been evicted")
	}
	for _, key := range []string{"oldest", "newest", "last"} {
		if !cache.Get(key, &value) {
			t.Errorf("recent entry %q was unexpectedly evicted", key)
		}
	}
	if got := len(cache.items); got != maxEntries {
		t.Fatalf("cache contains %d entries, want exactly %d", got, maxEntries)
	}
}

func TestCacheUpdateDoesNotIncreaseCapacityUsage(t *testing.T) {
	cache := NewCache(time.Minute)
	for index := 0; index < maxEntries; index++ {
		cache.Set(fmt.Sprintf("profile:%d", index), index)
	}
	cache.Set("profile:0", "updated")

	if got := len(cache.items); got != maxEntries {
		t.Fatalf("updating an entry changed cache size to %d, want %d", got, maxEntries)
	}
	var value string
	if !cache.Get("profile:0", &value) || value != "updated" {
		t.Fatalf("updated cache value = %q, want updated", value)
	}
}

func TestCacheDeleteRemovesEntryAndLRUNode(t *testing.T) {
	cache := NewCache(time.Minute)
	cache.Set("profile", "octocat")
	cache.Delete("profile")
	cache.Delete("missing")

	var value string
	if cache.Get("profile", &value) {
		t.Fatal("deleted entry is still present")
	}
	if len(cache.items) != 0 || cache.lru.Len() != 0 {
		t.Fatalf("delete left stale cache state: map=%d list=%d", len(cache.items), cache.lru.Len())
	}
}

func TestCacheGetHasZeroAllocationsForScalarSnapshots(t *testing.T) {
	cache := NewCache(time.Minute)
	cache.Set("value", 42)
	var value int
	if !cache.Get("value", &value) {
		t.Fatal("expected cache hit")
	}

	allocations := testing.AllocsPerRun(1000, func() {
		if !cache.Get("value", &value) {
			t.Fatal("expected cache hit")
		}
	})
	if allocations != 0 {
		t.Fatalf("cache hit allocated %.2f objects, want zero", allocations)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	type profileSnapshot struct {
		Username     string
		Repositories []struct {
			Name     string
			Topics   []string
			Language map[string]int
		}
	}
	cache := NewCache(time.Minute)
	cache.Set("profile", profileSnapshot{
		Username: "octocat",
		Repositories: []struct {
			Name     string
			Topics   []string
			Language map[string]int
		}{{Name: "project", Topics: []string{"go", "api"}, Language: map[string]int{"Go": 1200}}},
	})
	var value profileSnapshot
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if !cache.Get("profile", &value) {
			b.Fatal("expected cache hit")
		}
	}
}

func TestCacheGetRejectsInvalidDestination(t *testing.T) {
	cache := NewCache(time.Minute)
	cache.Set("value", 42)

	if cache.Get("value", 42) {
		t.Fatal("expected non-pointer output destination to be rejected")
	}
	var value string
	if cache.Get("value", &value) {
		t.Fatal("expected incompatible output type to be rejected")
	}
}

func TestCacheGetRemovesExpiredEntries(t *testing.T) {
	cache := NewCache(time.Minute)
	cache.Set("expired", "value")
	cache.mu.Lock()
	cache.items["expired"].Value.(*cacheItem).entry.ExpiresAt = time.Now().Add(-time.Second)
	cache.mu.Unlock()

	var value string
	if cache.Get("expired", &value) {
		t.Fatal("expired cache entry was returned")
	}
	if len(cache.items) != 0 || cache.lru.Len() != 0 {
		t.Fatal("expired entry remains in cache storage")
	}
}
