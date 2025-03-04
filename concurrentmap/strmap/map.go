// Package strmap provides a high performance, thread-safe concurrent map implementation
// optimized for string keys. It uses sharding to minimize lock contention.
package strmap

import (
	"encoding/json"
	"sync"
)

const defaultShardCount = 32

// ConcurrentMap is a "thread" safe map of type string:interface{}.
// To avoid lock bottlenecks this map is divided into several (defaultShardCount) map shards.
type ConcurrentMap []*mapShard

// mapShard is a "thread" safe string to anything map segment.
type mapShard struct {
	sync.RWMutex // Read Write mutex, guards access to internal map.
	items        map[string]interface{}
}

type Options struct {
	ShardCount int
}

// NewWithOptions creates a new concurrent map with custom options.
func NewWithOptions(opts Options) ConcurrentMap {
	shardCount := defaultShardCount
	if opts.ShardCount > 0 {
		shardCount = opts.ShardCount
	}

	m := make(ConcurrentMap, shardCount)
	for i := 0; i < shardCount; i++ {
		m[i] = &mapShard{items: make(map[string]interface{})}
	}
	return m
}

func (m ConcurrentMap) MSet(data map[string]interface{}) {
	var wg sync.WaitGroup
	for key, value := range data {
		wg.Add(1)
		go func(key string, value interface{}) {
			defer wg.Done()
			m.Set(key, value)
		}(key, value)
	}
	wg.Wait()
}

func (m ConcurrentMap) MGet(keys []string) map[string]interface{} {
	result := make(map[string]interface{}, len(keys))
	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			if val, ok := m.Get(key); ok {
				mutex.Lock()
				result[key] = val
				mutex.Unlock()
			}
		}(key)
	}
	wg.Wait()
	return result
}

func (m ConcurrentMap) Clear() {
	var wg sync.WaitGroup
	wg.Add(len(m))
	for _, shard := range m {
		go func(shard *mapShard) {
			defer wg.Done()
			shard.Lock()
			shard.items = make(map[string]interface{})
			shard.Unlock()
		}(shard)
	}
	wg.Wait()
}

func (m ConcurrentMap) GetOrSet(key string, value interface{}) interface{} {
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	if val, ok := shard.items[key]; ok {
		return val
	}
	shard.items[key] = value
	return value
}

func fnv32(key string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime32
	}
	return hash
}

func New() ConcurrentMap {
	m := make(ConcurrentMap, defaultShardCount)
	for i := 0; i < defaultShardCount; i++ {
		m[i] = &mapShard{items: make(map[string]interface{})}
	}
	return m
}

func (m ConcurrentMap) Set(key string, value interface{}) {
	shard := m.getShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// getShard returns shard under given key
func (m ConcurrentMap) getShard(key string) *mapShard {
	return m[uint(fnv32(key))%uint(defaultShardCount)]
}

// UpsertCb Callback to return new element to be inserted into the map
// It is called while lock is held, therefore it MUST NOT
// try to access other keys in same map, as it can lead to deadlock since
// Go sync.RWLock is not reentrant
type UpsertCb func(exist bool, valueInMap interface{}, newValue interface{}) interface{}

// Upsert Insert or Update - updates existing element or inserts a new one using UpsertCb
func (m ConcurrentMap) Upsert(key string, value interface{}, cb UpsertCb) (res interface{}) {
	shard := m.getShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// SetIfAbsent sets the given value under the specified key if no value was associated with it.
func (m ConcurrentMap) SetIfAbsent(key string, value interface{}) bool {
	// Get map shard.
	shard := m.getShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

// Get retrieves an element from map under given key.
func (m ConcurrentMap) Get(key string) (interface{}, bool) {
	// Get shard
	shard := m.getShard(key)
	shard.RLock()
	// Get item from shard.
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Count returns the number of elements within the map.
func (m ConcurrentMap) Count() int {
	count := 0
	for i := 0; i < defaultShardCount; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has looks up an item under specified key
func (m ConcurrentMap) Has(key string) bool {
	// Get shard
	shard := m.getShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Remove removes an element from the map.
func (m ConcurrentMap) Remove(key string) {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// RemoveCb is a callback executed in a map.RemoveCb() call, while Lock is held
// If returns true, the element will be removed from the map
type RemoveCb func(key string, v interface{}, exists bool) bool

// RemoveCb locks the shard containing the key, retrieves its current value and calls the callback with those params
// If callback returns true and element exists, it will remove it from the map
// Returns the value returned by the callback (even if element was not present in the map)
func (m ConcurrentMap) RemoveCb(key string, cb RemoveCb) bool {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

// Pop removes an element from the map and returns it
func (m ConcurrentMap) Pop(key string) (v interface{}, exists bool) {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

// IsEmpty checks if map is empty.
func (m ConcurrentMap) IsEmpty() bool {
	return m.Count() == 0
}

// Tuple used by the Iter functions to wrap two variables together over a channel,
type Tuple struct {
	Key string
	Val interface{}
}

// Iter returns a buffered iterator which could be used in a for range loop.
func (m ConcurrentMap) Iter() <-chan Tuple {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple, total)
	go fanIn(chans, ch)
	return ch
}

// fanIn reads elements from channels `chans` into channel `out`
func fanIn(chans []chan Tuple, out chan Tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

// snapshot returns an array of channels that contains elements in each shard,
// which likely takes a snapshot of `m`.
// It returns once the size of each buffered channel is determined,
// before all the channels are populated using goroutines.
func snapshot(m ConcurrentMap) (chans []chan Tuple) {
	chans = make([]chan Tuple, defaultShardCount)
	wg := sync.WaitGroup{}
	wg.Add(defaultShardCount)
	// Foreach shard.
	for index, shard := range m {
		go func(index int, shard *mapShard) {
			// Foreach key, value pair.
			shard.RLock()
			chans[index] = make(chan Tuple, len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- Tuple{key, val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

// Items returns all items as map[string]interface{}
func (m ConcurrentMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

// Keys returns all keys as []string
func (m ConcurrentMap) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(defaultShardCount)
		for _, shard := range m {
			go func(shard *mapShard) {
				// Foreach key, value pair.
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

// MarshalJSON reviles ConcurrentMap "private" variables to json marshal.
func (m ConcurrentMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}
