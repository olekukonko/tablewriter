package twcache

import (
	"strconv"
	"testing"
)

func BenchmarkLRU_GetHit(b *testing.B) {
	cache := NewLRU[string, int](100)
	for i := 0; i < 100; i++ {
		cache.Add(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get(strconv.Itoa(i % 100))
			i++
		}
	})
}

func BenchmarkLRU_GetMiss(b *testing.B) {
	cache := NewLRU[string, int](100)
	for i := 0; i < 50; i++ {
		cache.Add(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get("miss_" + strconv.Itoa(i))
			i++
		}
	})
}

func BenchmarkLRU_AddNew(b *testing.B) {
	cache := NewLRU[string, int](1000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Add(strconv.Itoa(i), i)
			i++
		}
	})
}

func BenchmarkLRU_AddUpdate(b *testing.B) {
	cache := NewLRU[string, int](100)
	for i := 0; i < 100; i++ {
		cache.Add(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Add(strconv.Itoa(i%100), i)
			i++
		}
	})
}

func BenchmarkLRU_GetOrCompute_Hit(b *testing.B) {
	cache := NewLRU[string, int](100)
	for i := 0; i < 100; i++ {
		cache.Add(strconv.Itoa(i), i)
	}
	compute := func() int { return 42 }
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.GetOrCompute(strconv.Itoa(i%100), compute)
			i++
		}
	})
}

func BenchmarkLRU_GetOrCompute_Miss(b *testing.B) {
	cache := NewLRU[string, int](1000)
	compute := func() int { return 42 }
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.GetOrCompute("key_"+strconv.Itoa(i), compute)
			i++
		}
	})
}
