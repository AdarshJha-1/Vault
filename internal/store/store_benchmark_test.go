package store

import (
	"strconv"
	"testing"
)

func BenchmarkSetKV(b *testing.B) {
	storage := GetStore()

	for i := 0; i < b.N; i++ {
		storage.Set(strconv.Itoa(i), "uwu")
	}
}
func BenchmarkGetKV(b *testing.B) {
	storage := GetStore()
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = strconv.Itoa(i)
		storage.Set(keys[i], "uwu")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.Get(keys[i])
	}
}
func BenchmarkDeleteKV(b *testing.B) {
	storage := GetStore()
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = strconv.Itoa(i)
		storage.Set(keys[i], "uwu")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		storage.Delete(keys[i])
	}
}
