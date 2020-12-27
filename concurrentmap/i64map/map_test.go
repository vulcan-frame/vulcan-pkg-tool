package i64map

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMap_Basic(t *testing.T) {
	m := New(32)
	const goroutines = 10
	const iterations = 100

	// Test Set and Get concurrently
	t.Run("Set and Get", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(base int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					key := int64(base*iterations + j)
					m.Set(key, fmt.Sprintf("value-%d", key))
					val, ok := m.Get(key)
					assert.True(t, ok)
					assert.Equal(t, val, fmt.Sprintf("value-%d", key))
				}
			}(i)
		}
		wg.Wait()
	})

	// Test Has concurrently
	t.Run("Has", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(base int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					key := int64(base*iterations + j)
					assert.True(t, m.Has(key))
					assert.False(t, m.Has(key+1000000))
				}
			}(i)
		}
		wg.Wait()
	})

	// Test Remove concurrently
	t.Run("Remove", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(base int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					key := int64(base*iterations + j)
					m.Remove(key)
					assert.False(t, m.Has(key))
				}
			}(i)
		}
		wg.Wait()
	})
}

func TestConcurrentMap_Concurrent(t *testing.T) {
	m := New(32)
	count := 1000
	var wg sync.WaitGroup

	// Concurrent Set
	t.Run("Concurrent Set", func(t *testing.T) {
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(i int64) {
				defer wg.Done()
				m.Set(i, i)
			}(int64(i))
		}
		wg.Wait()

		assert.Equal(t, m.Count(), count)
	})

	// Concurrent Get
	t.Run("Concurrent Get", func(t *testing.T) {
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func(i int64) {
				defer wg.Done()
				val, ok := m.Get(i)
				assert.True(t, ok)
				assert.Equal(t, val, i)
			}(int64(i))
		}
		wg.Wait()
	})
}

func TestConcurrentMap_BatchOperations(t *testing.T) {
	m := New(32)
	const goroutines = 10
	const batchSize = 100

	// Test MSet concurrently
	t.Run("MSet", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(base int) {
				defer wg.Done()
				data := make(map[int64]interface{})
				for j := 0; j < batchSize; j++ {
					key := int64(base*batchSize + j)
					data[key] = fmt.Sprintf("batch-%d", key)
				}
				m.MSet(data)

				// Verify all values were set correctly
				for k, v := range data {
					val, ok := m.Get(k)
					assert.True(t, ok)
					assert.Equal(t, val, v)
				}
			}(i)
		}
		wg.Wait()
	})

	// Test MGet concurrently
	t.Run("MGet", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for i := 0; i < goroutines; i++ {
			go func(base int) {
				defer wg.Done()
				keys := make([]int64, batchSize)
				for j := 0; j < batchSize; j++ {
					keys[j] = int64(base*batchSize + j)
				}
				results := m.MGet(keys)

				for _, k := range keys {
					val, ok := results[k]
					assert.True(t, ok)
					assert.Equal(t, val, fmt.Sprintf("batch-%d", k))
				}
			}(i)
		}
		wg.Wait()
	})
}

func TestConcurrentMap_SpecialCases(t *testing.T) {
	m := New(32)
	const goroutines = 10
	const iterations = 100

	// Test Clear with concurrent operations
	t.Run("Clear with Concurrent Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(goroutines * 2) // One group setting, one group clearing

		// Goroutines continuously setting values
		for i := 0; i < goroutines; i++ {
			go func(base int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					key := int64(base*iterations + j)
					m.Set(key, fmt.Sprintf("value-%d", key))
				}
			}(i)
		}

		// Goroutines continuously clearing
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations/10; j++ { // Less frequent clears
					m.Clear()
				}
			}()
		}
		wg.Wait()
	})

	// Test Resize with concurrent operations
	t.Run("Resize with Concurrent Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(goroutines * 2) // One group for operations, one for resizing

		// Goroutines performing regular operations
		for i := 0; i < goroutines; i++ {
			go func(base int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					key := int64(base*iterations + j)
					m.Set(key, fmt.Sprintf("value-%d", key))
					m.Get(key)
				}
			}(i)
		}

		// Goroutines performing resize operations
		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				sizes := []int{32, 64, 128, 256}
				for j := 0; j < iterations/10; j++ { // Less frequent resizes
					m.Resize(sizes[j%len(sizes)])
				}
			}()
		}
		wg.Wait()
	})
}

// Benchmarks

func BenchmarkConcurrentMap_Set(b *testing.B) {
	m := New(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Set(rand.Int63(), "value")
		}
	})
}

func BenchmarkConcurrentMap_Get(b *testing.B) {
	m := New(1000)
	for i := 0; i < 1000; i++ {
		m.Set(int64(i), i)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Get(rand.Int63n(1000))
		}
	})
}

func BenchmarkConcurrentMap_MSet(b *testing.B) {
	m := New(1000)
	data := make(map[int64]interface{})
	for i := 0; i < 1000; i++ {
		data[int64(i)] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.MSet(data)
	}
}

func BenchmarkConcurrentMap_GetMultiple(b *testing.B) {
	m := New(1000)
	keys := make([]int64, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = int64(i)
		m.Set(int64(i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.MGet(keys)
		}
	})
}

// Comparison benchmark test
func BenchmarkComparison_SyncMap(b *testing.B) {
	var m sync.Map
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Store(rand.Int63(), "value")
		}
	})
}

// Heavy load test
func BenchmarkConcurrentMap_HeavyLoad(b *testing.B) {
	m := New(32)
	numCPU := runtime.NumCPU()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for i := 0; i < numCPU; i++ {
				key := rand.Int63()
				switch rand.Intn(3) {
				case 0:
					m.Set(key, make([]byte, 4096))
				case 1:
					m.Get(key)
				case 2:
					m.Remove(key)
				}
			}
		}
	})
}

// Test different shard sizes
func BenchmarkConcurrentMap_DifferentShardSizes(b *testing.B) {
	sizes := []int{128, 512, 2048, 8192}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("ShardSize_%d", size), func(b *testing.B) {
			m := New(size)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					key := rand.Int63()
					m.Set(key, make([]byte, 1024))
					m.Get(key)
				}
			})
		})
	}
}
