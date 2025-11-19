package twcache

import (
	"strconv"
	"sync"
	"testing"
)

func TestLRU(t *testing.T) {
	t.Run("Basic Operations", func(t *testing.T) {
		cache := NewLRU[string, int](128)

		// Add
		cache.Add("key1", 100)
		val, ok := cache.Get("key1")
		if !ok || val != 100 {
			t.Errorf("expected 100, true; got %v, %v", val, ok)
		}

		// Update
		cache.Add("key1", 200)
		val, ok = cache.Get("key1")
		if !ok || val != 200 {
			t.Errorf("update failed: got %v, %v", val, ok)
		}

		// Miss
		_, ok = cache.Get("nonexistent")
		if ok {
			t.Error("miss should return ok=false")
		}
	})

	t.Run("GetOrCompute", func(t *testing.T) {
		cache := NewLRU[string, int](10)

		counter := 0
		compute := func() int {
			counter++
			return 42
		}

		v1 := cache.GetOrCompute("a", compute)
		if v1 != 42 || counter != 1 {
			t.Errorf("first call should compute: v=%d, counter=%d", v1, counter)
		}

		v2 := cache.GetOrCompute("a", compute)
		if v2 != 42 || counter != 1 {
			t.Errorf("second call should hit cache: v=%d, counter=%d", v2, counter)
		}

		// Concurrent safety
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cache.GetOrCompute("b", compute)
			}()
		}
		wg.Wait()

		// Only one computation should have happened
		if counter != 2 { // "a" and "b"
			t.Errorf("expected exactly 2 computations under concurrency, got %d", counter)
		}
	})

	t.Run("LRU Eviction", func(t *testing.T) {
		cache := NewLRU[string, int](2)

		cache.Add("key1", 1)
		cache.Add("key2", 2)

		// Make key1 most recent
		cache.Get("key1")

		// This should evict key2
		cache.Add("key3", 3)

		if _, ok := cache.Get("key2"); ok {
			t.Error("key2 should have been evicted")
		}
		if v, ok := cache.Get("key1"); !ok || v != 1 {
			t.Error("key1 should still be present")
		}
		if v, ok := cache.Get("key3"); !ok || v != 3 {
			t.Error("key3 should be present")
		}
	})

	t.Run("Eviction Callback", func(t *testing.T) {
		evicted := make([]string, 0)
		callback := func(key string, value int) {
			evicted = append(evicted, key)
		}

		cache := NewLRUEvict[string, int](2, callback)

		cache.Add("a", 1)
		cache.Add("b", 2)
		cache.Add("c", 3) // evicts "a"

		if len(evicted) != 1 || evicted[0] != "a" {
			t.Errorf("expected eviction of 'a', got %v", evicted)
		}

		cache.Purge()
		if len(evicted) != 3 {
			t.Errorf("Purge should call callback for all items, got %v", evicted)
		}
	})

	t.Run("Disabled Cache (size <= 0)", func(t *testing.T) {
		cache := NewLRU[string, int](0)

		if cache != nil {
			t.Error("NewLRU(0) should return nil")
		}

		var nilCache *LRU[string, int]

		// All operations should be safe no-ops
		nilCache.Add("x", 1)
		nilCache.GetOrCompute("x", func() int { return 99 })
		if v, ok := nilCache.Get("x"); ok || v != 0 {
			t.Errorf("nil cache Get should return (0, false), got %v, %v", v, ok)
		}
		nilCache.Purge()
		if nilCache.Len() != 0 || nilCache.Cap() != 0 {
			t.Error("nil cache should report zero size")
		}
	})

	t.Run("Purge", func(t *testing.T) {
		cache := NewLRU[string, int](10)
		cache.Add("a", 1)
		cache.Add("b", 2)

		cache.Purge()
		if cache.Len() != 0 {
			t.Errorf("Len after Purge should be 0, got %d", cache.Len())
		}
		if cache.HitRate() != 0 {
			t.Error("HitRate should reset after Purge")
		}
	})

	t.Run("HitRate Tracking", func(t *testing.T) {
		cache := NewLRU[string, int](10)

		if r := cache.HitRate(); r != 0 {
			t.Errorf("initial hit rate should be 0, got %f", r)
		}

		cache.Get("miss1") // Miss (Total: 1)
		cache.Add("k", 1)  // Add does NOT affect Hit/Miss count
		cache.Get("k")     // Hit  (Total: 2)
		cache.Get("miss2") // Miss (Total: 3)
		cache.Get("miss3") // Miss (Total: 4) <-- ADDED THIS LINE TO FIX TEST

		expected := 1.0 / 4.0 // 1 hit, 4 total accesses
		if r := cache.HitRate(); r != expected {
			t.Errorf("expected hit rate %.2f, got %.2f", expected, r)
		}
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		cache := NewLRU[string, int](1000)
		var wg sync.WaitGroup

		for i := 0; i < 200; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := "key" + strconv.Itoa(i)
				cache.Add(key, i)
			}(i)
		}

		for i := 0; i < 200; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := "key" + strconv.Itoa(i%100)
				cache.Get(key)
			}(i)
		}

		wg.Wait()

		if n := cache.Len(); n > 1000 {
			t.Errorf("cache size exceeded capacity: %d > 1000", n)
		}
	})

	t.Run("MoveToFront Logic", func(t *testing.T) {
		cache := NewLRU[string, int](3)

		cache.Add("a", 1)
		cache.Add("b", 2)
		cache.Add("c", 3)

		cache.Get("b") // make b most recent

		cache.Add("d", 4) // should evict "a"

		if _, ok := cache.Get("a"); ok {
			t.Error("'a' should have been evicted")
		}
		for _, k := range []string{"b", "c", "d"} {
			if _, ok := cache.Get(k); !ok {
				t.Errorf("%s should be present", k)
			}
		}
	})
}

// Add this to cache_test.go

func TestCoverageGaps(t *testing.T) {
	t.Run("LRU Manual Removal", func(t *testing.T) {
		l := NewLRU[string, int](10)
		l.Add("a", 1)
		l.Add("b", 2)

		if !l.Remove("a") {
			t.Error("Remove('a') should return true")
		}
		if l.Len() != 1 {
			t.Errorf("Expected len 1, got %d", l.Len())
		}

		if l.Remove("z") {
			t.Error("Remove('z') should return false")
		}

		key, val, ok := l.RemoveOldest()
		if !ok || key != "b" || val != 2 {
			t.Error("RemoveOldest should return 'b', 2")
		}

		_, _, ok = l.RemoveOldest()
		if ok {
			t.Error("RemoveOldest on empty should return false")
		}
	})

	t.Run("LRU Safety Cap", func(t *testing.T) {
		l := NewLRU[string, int](1_000_000)
		if l.Cap() != 100_000 {
			t.Errorf("Expected capacity capped at 100,000, got %d", l.Cap())
		}
	})

}
