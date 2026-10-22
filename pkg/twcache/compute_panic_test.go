package twcache

import (
	"testing"
	"time"
)

func TestGetOrComputePanicReleasesLock(t *testing.T) {
	cache := NewLRU[string, int](2)
	func() {
		defer func() {
			if got := recover(); got != "compute failed" {
				t.Fatalf("got panic %v, want compute failed", got)
			}
		}()
		cache.GetOrCompute("key", func() int { panic("compute failed") })
	}()
	done := make(chan int, 1)
	go func() {
		done <- cache.GetOrCompute("key", func() int { return 42 })
	}()
	select {
	case got := <-done:
		if got != 42 {
			t.Fatalf("got %d, want 42", got)
		}
	case <-time.After(time.Second):
		t.Fatal("cache remained locked after the compute panic")
	}
}
