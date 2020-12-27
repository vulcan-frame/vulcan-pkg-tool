package bitmap

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitmap(t *testing.T) {
	size := 10_000
	bitmap := NewBitmap(size)

	// Test setting bits
	for i := 0; i < size; i++ {
		bitmap.Set(i)
	}

	// Verify if all bits are set to 1
	for i := 0; i < size; i++ {
		assert.Truef(t, bitmap.IsSet(i), "Bit %d is not set", i)
	}

	// Test clearing bits
	for i := 0; i < size; i++ {
		bitmap.Clear(i)
	}

	// Verify if all bits are cleared
	for i := 0; i < size; i++ {
		assert.Falsef(t, bitmap.IsSet(i), "Bit %d is set", i)
	}

	// Test counting bits
	for i := 0; i < size; i += 2 {
		bitmap.Set(i)
	}
	count := bitmap.Count()
	assert.Equalf(t, size/2, count, "Expected %d bits to be set, but got %d", size/2, count)
}

func TestNewBitmap(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		wantSize int
	}{
		{"zero size", 0, 0},
		{"single byte", 7, 7},
		{"exact byte", 8, 8},
		{"multiple bytes", 15, 15},
		{"large size", 1024, 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBitmap(tt.size)
			assert.Equal(t, bm.Size(), tt.wantSize)
		})
	}

	t.Run("negative size", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for negative size")
			}
		}()
		_ = NewBitmap(-1)
	})
}

func TestSetAndIsSet(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		setBits  []int
		checkBit int
		want     bool
	}{
		{"set first bit", 8, []int{0}, 0, true},
		{"set last bit of byte", 8, []int{7}, 7, true},
		{"set cross-byte bit", 16, []int{8}, 8, true},
		{"unset bit", 8, []int{1}, 0, false},
		{"out of bounds (untested due to panic)", 8, []int{}, 8, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBitmap(tt.size)
			for _, bit := range tt.setBits {
				bm.Set(bit)
			}

			if tt.checkBit >= tt.size {
				assert.Panics(t, func() { bm.IsSet(tt.checkBit) }, "index out of range")
				return
			}

			if got := bm.IsSet(tt.checkBit); got != tt.want {
				t.Errorf("IsSet(%d) = %v, want %v", tt.checkBit, got, tt.want)
			}
		})
	}
}

func TestClear(t *testing.T) {
	bm := NewBitmap(16)
	bm.Set(8)
	bm.Clear(8)

	t.Run("clear set bit", func(t *testing.T) {
		if bm.IsSet(8) {
			t.Error("Bit should be cleared")
		}
	})

	t.Run("clear unset bit", func(t *testing.T) {
		bm.Clear(0) // Shouldn't panic
		if bm.IsSet(0) {
			t.Error("Bit should remain unset")
		}
	})
}

func TestCount(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		setBits   []int
		wantCount int
	}{
		{"empty", 8, []int{}, 0},
		{"single bit", 8, []int{0}, 1},
		{"multiple bits", 8, []int{0, 2, 4, 6}, 4},
		{"full byte", 8, []int{0, 1, 2, 3, 4, 5, 6, 7}, 8},
		{"cross bytes", 16, []int{7, 8, 15}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bm := NewBitmap(tt.size)
			for _, bit := range tt.setBits {
				bm.Set(bit)
			}
			if count := bm.Count(); count != tt.wantCount {
				t.Errorf("Count() = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestConcurrency(t *testing.T) {
	size := 1000
	bm := NewBitmap(size)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bm.Set(idx)
			bm.IsSet(idx)
			bm.Clear(idx)
		}(i)
	}
	wg.Wait()

	// Final check
	if count := bm.Count(); count != 0 {
		t.Errorf("Expected empty bitmap after concurrency test, got %d", count)
	}
}

// Benchmark tests
func BenchmarkSet(b *testing.B) {
	bm := NewBitmap(b.N * 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.Set(i % bm.Size())
	}
}

func BenchmarkIsSet(b *testing.B) {
	bm := NewBitmap(b.N * 8)
	for i := 0; i < b.N; i++ {
		bm.Set(i % bm.Size())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.IsSet(i % bm.Size())
	}
}

func BenchmarkCount(b *testing.B) {
	bm := NewBitmap(1024 * 1024)
	for i := 0; i < bm.Size(); i += 2 {
		bm.Set(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.Count()
	}
}

func BenchmarkConcurrentAccess(b *testing.B) {
	bm := NewBitmap(1024)
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			bm.Set(i % bm.Size())
			bm.IsSet(i % bm.Size())
			i++
		}
	})
}
