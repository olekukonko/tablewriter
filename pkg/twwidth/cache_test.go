package twwidth

import (
	"strconv"
	"sync"
	"testing"
)

func TestLRUCache(t *testing.T) {
	t.Run("Basic Operations", func(t *testing.T) {
		cache := newLRUCache(2)

		// Test Put and Get
		cache.Put("key1", 100)
		if val, ok := cache.Get("key1"); !ok || val != 100 {
			t.Errorf("Get after Put failed: got (%v, %v), want (100, true)", val, ok)
		}

		// Test Update
		cache.Put("key1", 200)
		if val, ok := cache.Get("key1"); !ok || val != 200 {
			t.Errorf("Update failed: got (%v, %v), want (200, true)", val, ok)
		}

		// Test Miss
		if val, ok := cache.Get("nonexistent"); ok || val != 0 {
			t.Errorf("Get nonexistent failed: got (%v, %v), want (0, false)", val, ok)
		}
	})

	t.Run("LRU Eviction", func(t *testing.T) {
		cache := newLRUCache(2)

		// Fill cache
		cache.Put("key1", 1)
		cache.Put("key2", 2)

		// Access key1 to make it MRU
		cache.Get("key1")

		// Add third item, should evict key2 (LRU)
		cache.Put("key3", 3)

		// key2 should be evicted
		if _, ok := cache.Get("key2"); ok {
			t.Error("key2 should have been evicted")
		}

		// key1 and key3 should be present
		if val, ok := cache.Get("key1"); !ok || val != 1 {
			t.Error("key1 should be present")
		}
		if val, ok := cache.Get("key3"); !ok || val != 3 {
			t.Error("key3 should be present")
		}
	})

	t.Run("Capacity Zero", func(t *testing.T) {
		cache := newLRUCache(0)

		// All operations should be no-ops
		cache.Put("key1", 1)
		if val, ok := cache.Get("key1"); ok || val != 0 {
			t.Error("Get should fail on zero capacity cache")
		}
		if cache.Len() != 0 {
			t.Error("Len should be 0 for zero capacity cache")
		}
		if cache.Cap() != 0 {
			t.Error("Cap should be 0 for zero capacity cache")
		}
	})

	t.Run("Nil Cache Safety", func(t *testing.T) {
		var cache *lruCache

		// All methods should handle nil receiver
		cache.Put("key1", 1)
		if val, ok := cache.Get("key1"); ok || val != 0 {
			t.Error("Get on nil cache should return (0, false)")
		}
		cache.Clear()
		if cache.Len() != 0 {
			t.Error("Len on nil cache should be 0")
		}
		if cache.Cap() != 0 {
			t.Error("Cap on nil cache should be 0")
		}
		if cache.HitRate() != 0 {
			t.Error("HitRate on nil cache should be 0")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		cache := newLRUCache(3)

		cache.Put("key1", 1)
		cache.Put("key2", 2)
		cache.Put("key3", 3)

		if cache.Len() != 3 {
			t.Error("Cache should have 3 items before clear")
		}

		cache.Clear()

		if cache.Len() != 0 {
			t.Error("Cache should be empty after clear")
		}

		// Verify all items are gone
		for _, key := range []lruCacheKey{"key1", "key2", "key3"} {
			if _, ok := cache.Get(key); ok {
				t.Errorf("Key %v should not exist after clear", key)
			}
		}
	})

	t.Run("HitRate Tracking", func(t *testing.T) {
		cache := newLRUCache(2)

		// Initial hit rate should be 0
		if rate := cache.HitRate(); rate != 0 {
			t.Errorf("Initial hit rate should be 0, got %v", rate)
		}

		// Add some items
		cache.Put("key1", 1)
		cache.Put("key2", 2)

		// Test hits and misses
		cache.Get("key1") // hit
		cache.Get("key3") // miss
		cache.Get("key2") // hit
		cache.Get("key4") // miss

		// 2 hits, 2 misses = 0.5 hit rate
		expectedRate := 0.5
		if rate := cache.HitRate(); rate != expectedRate {
			t.Errorf("Hit rate should be %v, got %v", expectedRate, rate)
		}

		// Test after clear
		cache.Clear()
		if rate := cache.HitRate(); rate != 0 {
			t.Errorf("Hit rate after clear should be 0, got %v", rate)
		}
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		cache := newLRUCache(100)
		var wg sync.WaitGroup

		// Concurrent writes
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := lruCacheKey("key" + strconv.Itoa(i))
				cache.Put(key, i)
			}(i)
		}

		// Concurrent reads
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := lruCacheKey("key" + strconv.Itoa(i%50))
				cache.Get(key)
			}(i)
		}

		wg.Wait()

		// Verify cache integrity
		if cache.Len() > cache.Cap() {
			t.Errorf("Cache size %d exceeds capacity %d", cache.Len(), cache.Cap())
		}
	})

	t.Run("Boundary Conditions", func(t *testing.T) {
		// Test capacity 1
		cache := newLRUCache(1)
		cache.Put("key1", 1)
		cache.Put("key2", 2) // Should evict key1

		if _, ok := cache.Get("key1"); ok {
			t.Error("key1 should have been evicted")
		}
		if val, ok := cache.Get("key2"); !ok || val != 2 {
			t.Error("key2 should be present")
		}

		// Test large capacity
		cache = newLRUCache(1000)
		for i := 0; i < 1000; i++ {
			cache.Put(lruCacheKey(strconv.Itoa(i)), i)
		}
		if cache.Len() != 1000 {
			t.Errorf("Expected 1000 items, got %d", cache.Len())
		}
	})

	t.Run("MoveToFront Logic", func(t *testing.T) {
		cache := newLRUCache(3)

		cache.Put("key1", 1)
		cache.Put("key2", 2)
		cache.Put("key3", 3)

		// Access key2, should move to front
		cache.Get("key2")

		// Add new item, should evict key1 (oldest accessed)
		cache.Put("key4", 4)

		// key1 should be evicted, key2,3,4 should remain
		if _, ok := cache.Get("key1"); ok {
			t.Error("key1 should have been evicted")
		}
		for _, key := range []lruCacheKey{"key2", "key3", "key4"} {
			if _, ok := cache.Get(key); !ok {
				t.Errorf("key %v should be present", key)
			}
		}
	})
}

func BenchmarkLRUCache(b *testing.B) {
	b.Run("Get Hit", func(b *testing.B) {
		cache := newLRUCache(100)
		for i := 0; i < 100; i++ {
			cache.Put(lruCacheKey(strconv.Itoa(i)), i)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				cache.Get(lruCacheKey(strconv.Itoa(i % 100)))
				i++
			}
		})
	})

	b.Run("Get Miss", func(b *testing.B) {
		cache := newLRUCache(100)
		// Only fill half the cache
		for i := 0; i < 50; i++ {
			cache.Put(lruCacheKey(strconv.Itoa(i)), i)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				// Always query keys that don't exist
				cache.Get(lruCacheKey("miss_" + strconv.Itoa(i)))
				i++
			}
		})
	})

	b.Run("Put New", func(b *testing.B) {
		cache := newLRUCache(1000)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				cache.Put(lruCacheKey(strconv.Itoa(i)), i)
				i++
			}
		})
	})

	b.Run("Put Update", func(b *testing.B) {
		cache := newLRUCache(100)
		for i := 0; i < 100; i++ {
			cache.Put(lruCacheKey(strconv.Itoa(i)), i)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				cache.Put(lruCacheKey(strconv.Itoa(i%100)), i)
				i++
			}
		})
	})

	b.Run("Mixed Workload", func(b *testing.B) {
		cache := newLRUCache(100)
		for i := 0; i < 50; i++ {
			cache.Put(lruCacheKey(strconv.Itoa(i)), i)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				op := i % 10
				switch op {
				case 0, 1, 2, 3: // 40% gets on existing keys
					cache.Get(lruCacheKey(strconv.Itoa(i % 50)))
				case 4, 5: // 20% gets on missing keys
					cache.Get(lruCacheKey("miss_" + strconv.Itoa(i)))
				case 6, 7: // 20% updates
					cache.Put(lruCacheKey(strconv.Itoa(i%50)), i)
				case 8, 9: // 20% new inserts
					cache.Put(lruCacheKey("new_"+strconv.Itoa(i)), i)
				}
				i++
			}
		})
	})
}

func BenchmarkNativeMap(b *testing.B) {
	b.Run("Get Hit", func(b *testing.B) {
		m := make(map[string]any)
		for i := 0; i < 100; i++ {
			m[strconv.Itoa(i)] = i
		}
		var mu sync.RWMutex

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				mu.RLock()
				_ = m[strconv.Itoa(i%100)]
				mu.RUnlock()
				i++
			}
		})
	})

	b.Run("Get Miss", func(b *testing.B) {
		m := make(map[string]any)
		for i := 0; i < 50; i++ {
			m[strconv.Itoa(i)] = i
		}
		var mu sync.RWMutex

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				mu.RLock()
				_ = m["miss_"+strconv.Itoa(i)]
				mu.RUnlock()
				i++
			}
		})
	})

	b.Run("Put New", func(b *testing.B) {
		m := make(map[string]any)
		var mu sync.Mutex

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				mu.Lock()
				m[strconv.Itoa(i)] = i
				mu.Unlock()
				i++
			}
		})
	})

	b.Run("Put Update", func(b *testing.B) {
		m := make(map[string]any)
		for i := 0; i < 100; i++ {
			m[strconv.Itoa(i)] = i
		}
		var mu sync.Mutex

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				mu.Lock()
				m[strconv.Itoa(i%100)] = i
				mu.Unlock()
				i++
			}
		})
	})

	b.Run("Mixed Workload", func(b *testing.B) {
		m := make(map[string]any)
		for i := 0; i < 50; i++ {
			m[strconv.Itoa(i)] = i
		}
		var mu sync.RWMutex

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				op := i % 10
				switch op {
				case 0, 1, 2, 3: // 40% reads on existing keys
					mu.RLock()
					_ = m[strconv.Itoa(i%50)]
					mu.RUnlock()
				case 4, 5: // 20% reads on missing keys
					mu.RLock()
					_ = m["miss_"+strconv.Itoa(i)]
					mu.RUnlock()
				case 6, 7: // 20% updates
					mu.Lock()
					m[strconv.Itoa(i%50)] = i
					mu.Unlock()
				case 8, 9: // 20% new inserts
					mu.Lock()
					m["new_"+strconv.Itoa(i)] = i
					mu.Unlock()
				}
				i++
			}
		})
	})
}

// Test coverage for internal methods
func TestInternalMethods(t *testing.T) {
	cache := newLRUCache(3)

	// Test addToFront and removeNode indirectly through public methods
	cache.Put("key1", 1)
	cache.Put("key2", 2)
	cache.Put("key3", 3)

	// Test moveToFront by accessing middle element
	cache.Get("key2") // This should move key2 to front

	// Verify key2 is now at front by checking eviction order
	cache.Put("key4", 4) // Should evict key1 (least recently used)

	if _, ok := cache.Get("key1"); ok {
		t.Error("key1 should have been evicted")
	}

	// All other keys should exist
	for _, key := range []lruCacheKey{"key2", "key3", "key4"} {
		if _, ok := cache.Get(key); !ok {
			t.Errorf("key %v should exist", key)
		}
	}
}
