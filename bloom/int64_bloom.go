package bloom

import (
	"github.com/vulcan-frame/vulcan-pkg-tools/bitmap"
)

// Int64BloomFilter optimized Bloom filter for int64
type Int64BloomFilter struct {
	bitmap   *bitmap.Bitmap
	hashFunc []func(int64) uint32
	size     uint32
}

// NewInt64Bloom create int64 optimized Bloom filter
// n: expected element count
// p: expected false positive rate (0 < p < 1)
func NewInt64Bloom(n uint32, p float64) *Int64BloomFilter {
	m, k := estimateParameters(n, p)
	// limit max hash function count to 8
	if k > 8 {
		k = 8
	}
	return &Int64BloomFilter{
		bitmap:   bitmap.NewBitmap(int(m)),
		hashFunc: createInt64HashFunctions(k),
		size:     m,
	}
}

// Add add int64 element
func (bf *Int64BloomFilter) Add(data int64) {
	for _, fn := range bf.hashFunc {
		h := fn(data) % bf.size
		bf.bitmap.Set(int(h))
	}
}

// AddMany add multiple int64 elements
func (bf *Int64BloomFilter) AddMany(data []int64) {
	indexes := make([]int, len(data))
	for i, d := range data {
		for _, fn := range bf.hashFunc {
			h := fn(d) % bf.size
			indexes[i] = int(h)
		}
	}
	bf.bitmap.MSet(indexes)
}

// Contains check if the element may exist
func (bf *Int64BloomFilter) Contains(data int64) bool {
	for _, fn := range bf.hashFunc {
		h := fn(data) % bf.size
		if !bf.bitmap.IsSet(int(h)) {
			return false
		}
	}
	return true
}

// create int64 optimized hash functions
func createInt64HashFunctions(k uint32) []func(int64) uint32 {
	base := []func(int64) uint32{
		func(data int64) uint32 { // use high bits
			return uint32(uint64(data) >> 32)
		},
		func(data int64) uint32 { // use low bits
			return uint32(data)
		},
		func(data int64) uint32 { // mix bits
			return uint32(uint64(data)>>16) ^ uint32(data)
		},
	}

	// generate more hash functions using double hash technique
	fns := make([]func(int64) uint32, 0, k)
	for i := uint32(0); i < k; i++ {
		idx := i % uint32(len(base))
		factor := i/uint32(len(base)) + 1
		fns = append(fns, func(data int64) uint32 {
			return base[idx](data) * uint32(factor)
		})
	}
	return fns
}
