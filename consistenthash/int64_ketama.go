package consistenthash

import (
	"hash"
	"sort"
	"strconv"
	"sync"

	"github.com/spaolacci/murmur3"
)

type int64RingNode struct {
	nodeName string
	key      int64
	hash     uint64
}

type int64RingNodes []int64RingNode

func (r int64RingNodes) Len() int           { return len(r) }
func (r int64RingNodes) Less(i, j int) bool { return r[i].hash < r[j].hash }
func (r int64RingNodes) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

type Int64HashRing struct {
	sync.RWMutex
	virtualSpots int
	nodes        int64RingNodes
	hashCache    sync.Pool
}

func NewInt64Ring(virtualSpots int) *Int64HashRing {
	if virtualSpots <= 0 {
		virtualSpots = DefaultVirtualSpots
	}

	return &Int64HashRing{
		virtualSpots: virtualSpots,
		hashCache: sync.Pool{
			New: func() any {
				return murmur3.New64()
			},
		},
	}
}

func (h *Int64HashRing) AddNode(nodeName string) {
	h.Lock()
	defer h.Unlock()

	hasher := h.hashCache.Get().(hash.Hash64)
	defer h.hashCache.Put(hasher)

	nodes := make(int64RingNodes, 0, h.virtualSpots)
	for i := range h.virtualSpots {
		keyStr := nodeName + ":" + strconv.Itoa(i)

		hasher.Reset()
		hasher.Write([]byte(keyStr))
		hash64 := hasher.Sum64()

		nodes = append(nodes, int64RingNode{
			nodeName: nodeName,
			key:      int64(hash64),
			hash:     hash64,
		})
	}

	h.nodes = append(h.nodes, nodes...)
	sort.Sort(h.nodes)
}

func (h *Int64HashRing) RemoveNode(nodeName string) {
	h.Lock()
	defer h.Unlock()

	filtered := h.nodes[:0]
	for _, n := range h.nodes {
		if n.nodeName != nodeName {
			filtered = append(filtered, n)
		}
	}
	h.nodes = filtered
}

func (h *Int64HashRing) GetNode(key int64) (string, bool) {
	h.RLock()
	defer h.RUnlock()

	if len(h.nodes) == 0 {
		return "", false
	}

	targetHash := uint64(key)
	idx := sort.Search(len(h.nodes), func(i int) bool {
		return h.nodes[i].hash >= targetHash
	})

	if idx == len(h.nodes) {
		idx = 0
	}

	return h.nodes[idx].nodeName, true
}
