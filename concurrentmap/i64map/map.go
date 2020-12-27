package i64map

import (
	"encoding/json"
	"sync"
)

const shardCount = 32

// ConcurrentMap is a thread-safe map implementation that uses sharding to reduce lock contention.
// It divides the map into multiple shards, each protected by its own RWMutex.
// This allows for better concurrent access compared to a single map protected by a single lock.
type ConcurrentMap []*mapShared

// mapShared represents a single shard of the concurrent map
type mapShared struct {
	sync.RWMutex // Read Write mutex, guards access to internal map.

	items map[int64]interface{}
}

// New creates a new concurrent map with the specified initial capacity.
// If initCapacity is less than or equal to shardCount, it defaults to 4096.
func New(initCapacity int) ConcurrentMap {
	if initCapacity <= shardCount {
		initCapacity = 4096
	}
	m := make(ConcurrentMap, shardCount)
	for i := 0; i < shardCount; i++ {
		m[i] = &mapShared{items: make(map[int64]interface{}, (initCapacity/shardCount)+1)}
	}
	return m
}

// MGet retrieves multiple items from the map in a single call.
// Returns a map containing only the keys that were found.
func (m ConcurrentMap) MGet(keys []int64) map[int64]interface{} {
	// Optimize by pre-allocating sharded key groups
	shardKeys := make([][]int64, shardCount)
	result := make(map[int64]interface{}, len(keys))

	// Group keys by shard
	for _, key := range keys {
		shard := uint(fnv32(key)) % uint(shardCount)
		shardKeys[shard] = append(shardKeys[shard], key)
	}

	// Process each shard
	for shardIndex, keys := range shardKeys {
		if len(keys) == 0 {
			continue
		}

		shard := m[shardIndex]
		shard.RLock()
		for _, key := range keys {
			if val, ok := shard.items[key]; ok {
				result[key] = val
			}
		}
		shard.RUnlock()
	}

	return result
}

// MSet sets multiple key-value pairs atomically within each shard.
// This provides better performance than setting keys individually.
func (m ConcurrentMap) MSet(data map[int64]interface{}) {
	// Group data by shards
	shardData := make([]map[int64]interface{}, shardCount)
	for i := range shardData {
		shardData[i] = make(map[int64]interface{})
	}

	for key, value := range data {
		shard := uint(fnv32(key)) % uint(shardCount)
		shardData[shard][key] = value
	}

	// Set all values for each shard concurrently
	var wg sync.WaitGroup
	wg.Add(shardCount)
	for i, items := range shardData {
		go func(shard *mapShared, items map[int64]interface{}) {
			defer wg.Done()
			if len(items) == 0 {
				return
			}
			shard.Lock()
			for k, v := range items {
				shard.items[k] = v
			}
			shard.Unlock()
		}(m[i], items)
	}
	wg.Wait()
}

// Set sets the given value under the specified key.
func (m ConcurrentMap) Set(key int64, value interface{}) (old interface{}) {
	// Get map shard.
	shard := m.getShard(key)
	shard.Lock()
	old = shard.items[key]
	shard.items[key] = value
	shard.Unlock()
	return
}

// getShard returns shard under given key
func (m ConcurrentMap) getShard(key int64) *mapShared {
	return m[uint(fnv32(key))%uint(shardCount)]
}

// UpsertCb Callback to return new element to be inserted into the map
// It is called while lock is held, therefore it MUST NOT
// try to access other keys in same map, as it can lead to deadlock since
// Go sync.RWLock is not reentrant
type UpsertCb func(exist bool, valueInMap interface{}, newValue interface{}) interface{}

// Upsert atomically updates or inserts a value for the given key.
// The callback function is called under lock to ensure atomic operation.
// WARNING: The callback must not access the map to avoid deadlocks.
func (m ConcurrentMap) Upsert(key int64, value interface{}, cb UpsertCb) (res interface{}) {
	shard := m.getShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// SetIfAbsent sets the given value under the specified key if no value was associated with it.
func (m ConcurrentMap) SetIfAbsent(key int64, value interface{}) bool {
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
func (m ConcurrentMap) Get(key int64) (interface{}, bool) {
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
	for i := 0; i < shardCount; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Has looks up an item under specified key
func (m ConcurrentMap) Has(key int64) bool {
	// Get shard
	shard := m.getShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Remove removes an element from the map.
func (m ConcurrentMap) Remove(key int64) {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// RemoveCb is a callback executed in a map.RemoveCb() call, while Lock is held
// If returns true, the element will be removed from the map
type RemoveCb func(key int64, v interface{}, exists bool) bool

// RemoveCb locks the shard containing the key, retrieves its current value and calls the callback with those params
// If callback returns true and element exists, it will remove it from the map
// Returns the value returned by the callback (even if element was not present in the map)
func (m ConcurrentMap) RemoveCb(key int64, cb RemoveCb) bool {
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
func (m ConcurrentMap) Pop(key int64) (v interface{}, exists bool) {
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

// Clear removes all items from the map efficiently.
// It creates new internal maps rather than deleting items one by one.
func (m ConcurrentMap) Clear() {
	for i := 0; i < shardCount; i++ {
		shard := m[i]
		shard.Lock()
		// Preserve the original capacity when clearing
		capacity := len(shard.items)
		shard.items = make(map[int64]interface{}, capacity)
		shard.Unlock()
	}
}

// Tuple used by the Iter functions to wrap two variables together over a channel,
type Tuple struct {
	Key int64
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
	chans = make([]chan Tuple, shardCount)
	for index, shard := range m {
		shard.RLock()
		chans[index] = make(chan Tuple, len(shard.items))
		for key, val := range shard.items {
			chans[index] <- Tuple{key, val}
		}
		shard.RUnlock()
		close(chans[index])
	}
	return chans
}

// Items returns all items as map[string]interface{}
func (m ConcurrentMap) Items() map[int64]interface{} {
	tmp := make(map[int64]interface{})

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

// Keys returns all keys as []string
func (m ConcurrentMap) Keys() []int64 {
	count := m.Count()
	ch := make(chan int64, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(shardCount)
		for _, shard := range m {
			go func(shard *mapShared) {
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
	keys := make([]int64, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

// MarshalJSON reviles ConcurrentMap "private" variables to json marshal.
func (m ConcurrentMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[int64]interface{})

	// Insert items to temporary map.
	for item := range m.Iter() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

// fnv32 generates a 32-bit hash for the given 64-bit key using FNV-1a algorithm.
func fnv32(key int64) uint32 {
	// Use FNV-1a algorithm for better distribution
	const (
		offset32 = uint32(2166136261)
		prime32  = uint32(16777619)
	)

	hash := offset32
	var bytes [8]byte
	for i := uint(0); i < 8; i++ {
		bytes[7-i] = byte(key >> (i * 8))
	}

	for _, b := range bytes {
		hash ^= uint32(b)
		hash *= prime32
	}

	return hash
}

// Resize adjusts the capacity of all shards in the map.
// This can be useful for optimizing memory usage after many items have been removed.
func (m ConcurrentMap) Resize(newCapacity int) {
	if newCapacity < 0 {
		return
	}

	shardCapacity := (newCapacity / shardCount) + 1
	for i := 0; i < shardCount; i++ {
		shard := m[i]
		shard.Lock()
		newItems := make(map[int64]interface{}, shardCapacity)
		for k, v := range shard.items {
			newItems[k] = v
		}
		shard.items = newItems
		shard.Unlock()
	}
}

// ForEach executes the given function for each key-value pair in the map.
// The iteration is done in a concurrent manner across shards.
func (m ConcurrentMap) ForEach(fn func(key int64, value interface{})) {
	var wg sync.WaitGroup
	wg.Add(shardCount)

	for i := 0; i < shardCount; i++ {
		go func(shard *mapShared) {
			defer wg.Done()
			shard.RLock()
			for k, v := range shard.items {
				fn(k, v)
			}
			shard.RUnlock()
		}(m[i])
	}

	wg.Wait()
}
