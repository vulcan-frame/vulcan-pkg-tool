package consistenthash

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInt64Ring(t *testing.T) {
	t.Run("default virtual spots", func(t *testing.T) {
		r := NewInt64Ring(0)
		if r.virtualSpots != DefaultVirtualSpots {
			t.Errorf("Expected %d virtual spots, got %d", DefaultVirtualSpots, r.virtualSpots)
		}
	})

	t.Run("custom virtual spots", func(t *testing.T) {
		const customSpots = 200
		r := NewInt64Ring(customSpots)
		if r.virtualSpots != customSpots {
			t.Errorf("Expected %d virtual spots, got %d", customSpots, r.virtualSpots)
		}
	})
}

func TestInt64HashRing_AddNode(t *testing.T) {
	r := NewInt64Ring(100)
	nodes := []string{"node1", "node2", "node3"}

	t.Run("add single node", func(t *testing.T) {
		r.AddNode(nodes[0])
		if len(r.nodes) != 100 {
			t.Errorf("Expected 100 virtual nodes, got %d", len(r.nodes))
		}
	})

	t.Run("add multiple nodes", func(t *testing.T) {
		r.AddNode(nodes[1])
		r.AddNode(nodes[2])
		if len(r.nodes) != 300 {
			t.Errorf("Expected 300 virtual nodes, got %d", len(r.nodes))
		}
	})

	t.Run("duplicate node addition", func(t *testing.T) {
		originalCount := len(r.nodes)
		r.AddNode(nodes[0])
		if len(r.nodes) != originalCount+100 {
			t.Errorf("Expected %d virtual nodes after duplicate add, got %d", originalCount+100, len(r.nodes))
		}
	})
}

func TestInt64HashRing_RemoveNode(t *testing.T) {
	r := NewInt64Ring(50)
	nodes := []string{"nodeA", "nodeB", "nodeC"}
	for _, n := range nodes {
		r.AddNode(n)
	}

	t.Run("remove existing node", func(t *testing.T) {
		r.RemoveNode(nodes[1])
		for _, n := range r.nodes {
			assert.NotEqual(t, n.nodeName, nodes[1])
		}
	})

	t.Run("remove non-existent node", func(t *testing.T) {
		originalCount := len(r.nodes)
		r.RemoveNode("ghost_node")
		assert.Equal(t, len(r.nodes), originalCount)
	})

	t.Run("remove all nodes", func(t *testing.T) {
		for _, n := range nodes {
			r.RemoveNode(n)
		}
		assert.Equal(t, len(r.nodes), 0)
	})
}

func TestInt64HashRing_GetNode(t *testing.T) {
	r := NewInt64Ring(100)
	nodes := []string{"server1", "server2", "server3"}
	for _, n := range nodes {
		r.AddNode(n)
	}

	// Test key distribution
	distribution := make(map[string]int)
	const testKeys = 10_000
	for i := range testKeys {
		key := int64(i)
		node, _ := r.GetNode(key)
		distribution[node]++
	}
	for node, count := range distribution {
		t.Logf("Node %s received %d keys (%.1f%%)", node, count, float64(count)/testKeys*100)
	}

	t.Run("empty ring", func(t *testing.T) {
		emptyRing := NewInt64Ring(100)
		_, ok := emptyRing.GetNode(123)
		assert.False(t, ok)
	})

	t.Run("consistent hashing", func(t *testing.T) {
		key := int64(42)
		node1, _ := r.GetNode(key)
		node2, _ := r.GetNode(key)
		assert.Equal(t, node1, node2)
	})

	t.Run("ring wrap-around", func(t *testing.T) {
		// Find the highest hash value
		maxHash := r.nodes[len(r.nodes)-1].hash
		testKey := maxHash + 1 // Force wrap-around
		node, _ := r.GetNode(int64(testKey))
		assert.Equal(t, node, r.nodes[0].nodeName)
	})
}

func TestInt64HashRing_ConcurrentAccess(t *testing.T) {
	r := NewInt64Ring(160)
	var wg sync.WaitGroup

	// Concurrent writers
	for i := range 4 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				r.AddNode("node" + strconv.Itoa(id*100+j))
				r.RemoveNode("node" + strconv.Itoa((id*100+j)-1))
			}
		}(i)
	}

	// Concurrent readers
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range 1000 {
				r.GetNode(int64(j))
			}
		}()
	}

	wg.Wait()
}

func BenchmarkInt64HashRing_GetNode(b *testing.B) {
	r := NewInt64Ring(160)
	for i := range 10 {
		r.AddNode("node" + strconv.Itoa(i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.GetNode(int64(b.N))
		}
	})
}

func BenchmarkHashRing_AddNode(b *testing.B) {
	r := NewInt64Ring(160)
	nodeNames := make([]string, b.N)
	for i := range nodeNames {
		nodeNames[i] = "node" + strconv.Itoa(i)
	}

	b.ResetTimer()
	for i := range b.N {
		r.AddNode(nodeNames[i])
	}
}

func BenchmarkInt64HashRing_RemoveNode(b *testing.B) {
	r := NewInt64Ring(160)
	for i := range 10 {
		r.AddNode("node" + strconv.Itoa(i))
	}

	b.ResetTimer()
	for i := range b.N {
		r.RemoveNode("node" + strconv.Itoa(i))
	}
}
