package bloom

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBloomFilter(t *testing.T) {
	bf := New(1000, 0.01)

	// test basic functions
	testData := [][]byte{[]byte("hello"), []byte("world")}
	for _, d := range testData {
		bf.Add(d)
		assert.True(t, bf.Contains(d), "Should contain added element")
	}

	// test false positive rate
	falsePositives := 0
	total := 10000
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < total; i++ {
		randomStr := randomString(10)
		if bf.Contains([]byte(randomStr)) {
			falsePositives++
		}
	}

	fpRate := float64(falsePositives) / float64(total)
	assert.True(t, fpRate < 0.02, "False positive rate too high: %f", fpRate)
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty filter", func(t *testing.T) {
		bf := New(100, 0.01)
		assert.False(t, bf.Contains([]byte("test")))
	})

	t.Run("full capacity", func(t *testing.T) {
		bf := New(10, 0.01)
		for i := 0; i < 100; i++ {
			bf.Add([]byte{byte(i)})
		}
		assert.True(t, bf.Contains([]byte{99}))
	})
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func BenchmarkBloomFilter(b *testing.B) {
	bf := New(1000000, 0.01)
	data := make([][]byte, b.N)
	for i := range data {
		data[i] = []byte(randomString(10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bf.Add(data[i])
		bf.Contains(data[i])
	}
}
