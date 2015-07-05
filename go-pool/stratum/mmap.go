// Generated from https://github.com/streamrail/concurrent-map
package stratum

import (
	"hash/fnv"
	"sync"
)

var SHARD_COUNT = 32

// TODO: Add Keys function which returns an array of keys for the map.

// A "thread" safe map of type string:*Miner.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type MinersMap []*MinersMapShared
type MinersMapShared struct {
	items        map[string]*Miner
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

// Creates a new concurrent map.
func NewMinersMap() MinersMap {
	m := make(MinersMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &MinersMapShared{items: make(map[string]*Miner)}
	}
	return m
}

// Returns shard under given key
func (m MinersMap) GetShard(key string) *MinersMapShared {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return m[int(hasher.Sum32())%SHARD_COUNT]
}

// Sets the given value under the specified key.
func (m *MinersMap) Set(key string, value *Miner) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.items[key] = value
}

// Retrieves an element from map under given key.
func (m MinersMap) Get(key string) (*Miner, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// Get item from shard.
	val, ok := shard.items[key]
	return val, ok
}

// Returns the number of elements within the map.
func (m MinersMap) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Looks up an item under specified key
func (m *MinersMap) Has(key string) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// See if element is within shard.
	_, ok := shard.items[key]
	return ok
}

// Removes an element from the map.
func (m *MinersMap) Remove(key string) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.items, key)
}

// Checks if map is empty.
func (m *MinersMap) IsEmpty() bool {
	return m.Count() == 0
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type Tuple struct {
	Key string
	Val *Miner
}

// Returns an iterator which could be used in a for range loop.
func (m MinersMap) Iter() <-chan Tuple {
	ch := make(chan Tuple)
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- Tuple{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}

// Returns a buffered iterator which could be used in a for range loop.
func (m MinersMap) IterBuffered() <-chan Tuple {
	ch := make(chan Tuple, m.Count())
	go func() {
		// Foreach shard.
		for _, shard := range m {
			// Foreach key, value pair.
			shard.RLock()
			for key, val := range shard.items {
				ch <- Tuple{key, val}
			}
			shard.RUnlock()
		}
		close(ch)
	}()
	return ch
}
