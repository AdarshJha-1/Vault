package store

import (
	"fmt"
	"testing"
)

func BenchmarkSet(b *testing.B) {
	store := GetStore(10000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
	}
}

func BenchmarkSetEviction(b *testing.B) {
	store := GetStore(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
	}
}

func BenchmarkGet(b *testing.B) {
	store := GetStore(10000)
	for i := 0; i < 10000; i++ {
		store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get("key3000")
	}
}

// This is not a good way...
func BenchmarkDelete(b *testing.B) {
	store := GetStore(10000)
	for i := 0; i < 10000; i++ {
		store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Delete(fmt.Sprintf("key%d", i%10000))
	}
}

func BenchmarkConcurrentGet(b *testing.B) {
	store := GetStore(10000)
	for i := 0; i < 10000; i++ {
		store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
	}
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			store.Get("key3000")
		}
	})
}
