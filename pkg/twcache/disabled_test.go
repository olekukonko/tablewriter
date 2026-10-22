package twcache

import "testing"

func TestDisabledCacheHitRate(t *testing.T) {
	for _, size := range []int{0, -1} {
		cache := NewLRU[string, int](size)
		if got := cache.HitRate(); got != 0 {
			t.Fatalf("size %d: got %v, want 0", size, got)
		}
	}
}
