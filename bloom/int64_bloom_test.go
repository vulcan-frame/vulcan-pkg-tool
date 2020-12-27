package bloom

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInt64BloomFilter(t *testing.T) {
	bf := NewInt64Bloom(1000, 0.01)

	// test basic functions
	testData := []int64{0, -1, 123456789, 1<<63 - 1}
	for _, d := range testData {
		bf.Add(d)
		assert.True(t, bf.Contains(d), "Should contain added element")
	}

	// test false positive rate
	falsePositives := 0
	total := 10000
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < total; i++ {
		randomNum := rand.Int63()
		if bf.Contains(randomNum) && !contains(testData, randomNum) {
			falsePositives++
		}
	}

	fpRate := float64(falsePositives) / float64(total)
	assert.True(t, fpRate < 0.02, "False positive rate too high: %f", fpRate)
}

func TestInt64EdgeCases(t *testing.T) {
	t.Run("min and max", func(t *testing.T) {
		bf := NewInt64Bloom(100, 0.01)
		bf.Add(-1 << 63)
		bf.Add(1<<63 - 1)
		assert.True(t, bf.Contains(-1<<63))
		assert.True(t, bf.Contains(1<<63-1))
	})

	t.Run("zero value", func(t *testing.T) {
		bf := NewInt64Bloom(10, 0.01)
		bf.Add(0)
		assert.True(t, bf.Contains(0))
	})
}

func BenchmarkInt64Bloom(b *testing.B) {
	bf := NewInt64Bloom(1000000, 0.01)
	data := make([]int64, b.N)
	for i := range data {
		data[i] = rand.Int63()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(data[i])
		bf.Contains(data[i])
	}
}

func contains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
