package bloom

import (
	"hash/fnv"
	"math"

	"github.com/vulcan-frame/vulcan-pkg-tool/bitmap"
)

// BloomFilter represents a thread-safe Bloom filter
type BloomFilter struct {
	bitmap   *bitmap.Bitmap
	hashFunc []func([]byte) uint32
	size     uint32
}

// New create bloom filter
// n: expected element count
// p: expected false positive rate (0 < p < 1)
func New(n uint32, p float64) *BloomFilter {
	m, k := estimateParameters(n, p)
	if k > 8 {
		k = 8
	}
	return &BloomFilter{
		bitmap:   bitmap.NewBitmap(int(m)),
		hashFunc: createHashFunctions(k),
		size:     m,
	}
}

// Add add element to bloom filter
func (bf *BloomFilter) Add(data []byte) {
	for _, fn := range bf.hashFunc {
		h := fn(data) % bf.size
		bf.bitmap.Set(int(h))
	}
}

// Contains check if the element may exist
func (bf *BloomFilter) Contains(data []byte) bool {
	for _, fn := range bf.hashFunc {
		h := fn(data) % bf.size
		if !bf.bitmap.IsSet(int(h)) {
			return false
		}
	}
	return true
}

// estimateParameters calculate optimal parameters (m: array size, k: hash function count)
func estimateParameters(n uint32, p float64) (uint32, uint32) {
	m := uint32(math.Ceil(-float64(n) * math.Log(p) / (math.Pow(math.Log(2), 2))))
	k := uint32(math.Ceil(math.Log(2) * float64(m) / float64(n)))
	return m, k
}

// createHashFunctions create k hash functions (using double hash technique)
func createHashFunctions(k uint32) []func([]byte) uint32 {
	h1 := fnv.New32a()
	h2 := fnv.New32()

	base := []func([]byte) uint32{
		func(data []byte) uint32 { h1.Reset(); h1.Write(data); return h1.Sum32() },
		func(data []byte) uint32 { h2.Reset(); h2.Write(data); return h2.Sum32() },
	}

	fns := make([]func([]byte) uint32, 0, k)
	for i := uint32(0); i < k; i++ {
		idx := i % uint32(len(base))
		factor := i/uint32(len(base)) + 1
		fns = append(fns, func(data []byte) uint32 {
			return base[idx](data) * uint32(factor)
		})
	}
	return fns
}
