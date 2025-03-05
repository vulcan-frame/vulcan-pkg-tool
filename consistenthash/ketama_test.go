package consistenthash

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInfo(t *testing.T) {
	ring := NewRing(16)

	nodes := []string{
		"test.server.com#1",
		"test.server.com#2",
		"test.server.com#3",
		"test.server.com#4",
	}

	for _, k := range nodes {
		ring.AddNode(k)
	}

	m := make(map[string]int)
	for i := 0; i < 1e6; i++ {
		node, ok := ring.GetNode("test value" + strconv.FormatUint(uint64(i), 10))
		assert.True(t, ok)
		m[node]++
	}

	for i := 1; i < len(nodes); i++ {
		ring.RemoveNode(nodes[i])
	}

	m = make(map[string]int)
	for i := 0; i < 1e6; i++ {
		node, ok := ring.GetNode("test value" + strconv.FormatUint(uint64(i), 10))
		assert.True(t, ok)
		m[node]++
	}

	ring.RemoveNode(nodes[0])

	for i := 0; i < 1e6; i++ {
		node, ok := ring.GetNode("test value" + strconv.FormatUint(uint64(i), 10))
		assert.False(t, ok)
		assert.Equal(t, node, "")
	}
}

func TestNewRing(t *testing.T) {
	t.Run("default virtual spots", func(t *testing.T) {
		r := NewRing(0)
		assert.Equal(t, r.virtualSpots, DefaultVirtualSpots)
	})

	t.Run("custom virtual spots", func(t *testing.T) {
		customSpots := 200
		r := NewRing(customSpots)
		assert.Equal(t, r.virtualSpots, customSpots)
	})
}

func TestHashRing_AddRemoveNodes(t *testing.T) {
	r := NewRing(100)
	nodes := []string{"node1", "node2", "node3"}

	t.Run("add nodes", func(t *testing.T) {
		for _, n := range nodes {
			r.AddNode(n)
		}

		assert.Equal(t, len(r.nodes), len(nodes)*r.virtualSpots)
	})

	t.Run("remove node", func(t *testing.T) {
		r.RemoveNode("node2")
		expected := (len(nodes) - 1) * r.virtualSpots
		assert.Equal(t, len(r.nodes), expected)

		for _, n := range r.nodes {
			assert.NotEqual(t, n.nodeName, "node2")
		}
	})
}

func TestHashRing_GetNode(t *testing.T) {
	r := NewRing(100)
	nodes := []string{"nodeA", "nodeB", "nodeC"}
	for _, n := range nodes {
		r.AddNode(n)
	}

	testCases := []struct {
		key      string
		expected string
	}{
		{"user123", ""},
		{"session-abc", ""},
		{"data:1", ""},
		{"config:prod", ""},
	}

	// First pass to record distribution
	distribution := make(map[string]string)
	for _, tc := range testCases {
		node, found := r.GetNode(tc.key)
		assert.True(t, found)
		distribution[tc.key] = node
	}

	t.Run("consistent distribution", func(t *testing.T) {
		for _, tc := range testCases {
			node, _ := r.GetNode(tc.key)
			assert.Equal(t, node, distribution[tc.key])
		}
	})

	t.Run("empty ring", func(t *testing.T) {
		emptyRing := NewRing(100)
		node, found := emptyRing.GetNode("anykey")
		assert.False(t, found)
		assert.Equal(t, node, "")
	})

	t.Run("wrap around", func(t *testing.T) {
		// Create predictable ring with known hash values
		r := NewRing(1)
		r.AddNode("nodeX")
		r.AddNode("nodeY")

		// Force wrap around scenario
		highHashKey := "zzzzzzzzzzzzzzzz"
		node, _ := r.GetNode(highHashKey)
		assert.Equal(t, node, r.nodes[1].nodeName)
	})
}

func TestHashRing_Consistency(t *testing.T) {
	r := NewRing(100)
	initialNodes := []string{"node1", "node2", "node3"}
	for _, n := range initialNodes {
		r.AddNode(n)
	}

	keys := make([]string, 1000)
	original := make(map[string]string)
	for i := range keys {
		keys[i] = "key" + strconv.Itoa(i)
		original[keys[i]], _ = r.GetNode(keys[i])
	}

	t.Run("after adding node", func(t *testing.T) {
		r.AddNode("node4")
		changed := 0
		for _, k := range keys {
			node, _ := r.GetNode(k)
			if node != original[k] {
				changed++
			}
		}
		t.Logf("Changed keys after adding node: %.2f%%", float64(changed)/float64(len(keys))*100)
	})

	t.Run("after removing node", func(t *testing.T) {
		r.RemoveNode("node3")
		changed := 0
		for _, k := range keys {
			node, _ := r.GetNode(k)
			if node != original[k] {
				changed++
			}
		}
		t.Logf("Changed keys after removing node: %.2f%%", float64(changed)/float64(len(keys))*100)
	})
}

func TestConcurrentAccess(t *testing.T) {
	r := NewRing(100)
	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			r.AddNode("node" + strconv.Itoa(i))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			r.RemoveNode("node" + strconv.Itoa(i))
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			r.GetNode("key" + strconv.Itoa(i))
		}
	}()

	wg.Wait() // Should not panic
}

func BenchmarkHashRing_GetNode(b *testing.B) {
	r := NewRing(200)
	for i := 0; i < 10; i++ {
		r.AddNode("node" + strconv.Itoa(i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.GetNode("some-key")
		}
	})
}

func BenchmarkAddNode(b *testing.B) {
	r := NewRing(200)
	b.ResetTimer()
	for i := range b.N {
		r.AddNode("node" + strconv.Itoa(i))
	}
}

func BenchmarkRemoveNode(b *testing.B) {
	r := NewRing(200)
	for i := range 100 {
		r.AddNode("node" + strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := range b.N {
		r.RemoveNode("node" + strconv.Itoa(i%100))
	}
}
