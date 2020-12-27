package strmap

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// basic operations
func TestConcurrentMap_BasicOperations(t *testing.T) {
	m := New()

	// Test Set and Get
	m.Set("key1", "value1")
	val, ok := m.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, val, "value1")

	// Test Has
	assert.True(t, m.Has("key1"))

	// Test Remove
	m.Remove("key1")
	assert.False(t, m.Has("key1"))

	// Test Count
	m.Set("key2", "value2")
	m.Set("key3", "value3")
	assert.Equal(t, m.Count(), 2)
}

// concurrent access
func TestConcurrentMap_ConcurrentAccess(t *testing.T) {
	m := New()
	const goroutines = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // Writers and readers

	// Writers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := string(rune(id))
				m.Set(key, j)
			}
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := string(rune(id))
				m.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

// 
func TestConcurrentMap_BatchOperations(t *testing.T) {
	m := New()

	// Test MSet
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	m.MSet(testData)

	// Test MGet
	keys := []string{"key1", "key2", "key3"}
	results := m.MGet(keys)

	for k, v := range testData {
		assert.Equal(t, results[k], v)
	}
}

// benchmark
func BenchmarkConcurrentMap_Set(b *testing.B) {
	m := New()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(string(rune(i)), i)
			i++
		}
	})
}

func BenchmarkConcurrentMap_Get(b *testing.B) {
	m := New()
	for i := 0; i < b.N; i++ {
		m.Set(string(rune(i)), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Get(string(rune(i)))
			i++
		}
	})
}

func BenchmarkConcurrentMap_MSet(b *testing.B) {
	m := New()
	data := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		data[string(rune(i))] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.MSet(data)
	}
}

func BenchmarkConcurrentMap_MGet(b *testing.B) {
	m := New()
	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		key := string(rune(i))
		keys[i] = key
		m.Set(key, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.MGet(keys)
	}
}

// edge cases
func TestConcurrentMap_EdgeCases(t *testing.T) {
	m := New()

	// Test empty map
	assert.True(t, m.IsEmpty())

	// Test Pop on empty map
	_, exists := m.Pop("nonexistent")
	assert.False(t, exists)

	// Test GetOrSet
	val := m.GetOrSet("key", "value")
	assert.Equal(t, val, "value")

	val = m.GetOrSet("key", "newvalue")
	assert.Equal(t, val, "value")

	// Test Clear
	m.Set("key1", "value1")
	m.Set("key2", "value2")
	m.Clear()
	assert.True(t, m.IsEmpty())
}
