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

func BenchmarkConcurrentSet(b *testing.B) {
	store := GetStore(10000)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		i := 0
		for p.Next() {
			store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
			i++
		}
	})
}

func BenchmarkMixedWorkload(b *testing.B) {
	store := GetStore(10000)

	for i := 0; i < 10000; i++ {
		store.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("val%d", i))
	}

	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		i := 0
		for p.Next() {
			switch i % 10 {
			case 0:
				store.Delete(fmt.Sprintf("key%d", i%10000))
			case 1, 2:
				store.Set(fmt.Sprintf("key%d", i%10000), "value")
			default:
				store.Get("key3000")
			}
			i++
		}
	})

}
