package bitmap

import (
	"math/bits"
	"sync"
)

// Bitmap represents a thread-safe bitmap using a byte array
type Bitmap struct {
	mutex sync.Mutex
	bits  []byte
	size  int // Track original size for bounds checking
}

// NewBitmap creates a new Bitmap with the given size (in bits)
func NewBitmap(size int) *Bitmap {
	if size < 0 {
		panic("bitmap size must be non-negative")
	}
	return &Bitmap{
		bits: make([]byte, (size+7)/8),
		size: size,
	}
}

// Set sets the bit at the given index to 1
func (b *Bitmap) Set(index int) {
	b.validateIndex(index)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.bits[index/8] |= 1 << (index % 8)
}

func (b *Bitmap) MSet(indexes []int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	for _, index := range indexes {
		b.Set(index)
	}
}

// Clear clears the bit at the given index to 0
func (b *Bitmap) Clear(index int) {
	b.validateIndex(index)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.bits[index/8] &^= 1 << (index % 8)
}

// IsSet checks if the bit at the given index is set to 1
func (b *Bitmap) IsSet(index int) bool {
	b.validateIndex(index)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.bits[index/8]&(1<<(index%8)) != 0
}

// Count returns the number of bits set to 1 using efficient bit counting
func (b *Bitmap) Count() int {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	count := 0
	for _, byteVal := range b.bits {
		count += bits.OnesCount8(byteVal)
	}
	return count
}

// Size returns the capacity of the bitmap in bits
func (b *Bitmap) Size() int {
	return b.size
}

// validateIndex checks if index is within valid range
func (b *Bitmap) validateIndex(index int) {
	if index < 0 || index >= b.size {
		panic("bitmap index out of range")
	}
}
