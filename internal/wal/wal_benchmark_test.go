package wal

import (
	"fmt"
	"testing"

	"github.com/AdarshJha-1/Vault/internal/store"
)

func BenchmarkWriteEntry(b *testing.B) {
	walDir := b.TempDir()
	wal, err := OpenWAL(walDir, 16*1024*1024, 10)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		wal.Close()
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := wal.WriteEntry("SET me you")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadToVault(b *testing.B) {
	sizes := []int{1000, 10000}

	wals := make(map[int]WAL)

	for _, size := range sizes {
		walDir := b.TempDir()
		wal, err := OpenWAL(walDir, 16*1024*1024, 10)
		if err != nil {
			b.Fatal(err)
		}
		for i := 0; i < size; i++ {
			err := wal.WriteEntry(fmt.Sprintf("SET me%d you", i))
			if err != nil {
				b.Fatal(err)
			}
		}
		wals[size] = wal
		b.Cleanup(func() {
			wal.Close()
		})

	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			wal := wals[size]
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store := store.GetStore(1000)
				err := wal.LoadToVault(store)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

}
