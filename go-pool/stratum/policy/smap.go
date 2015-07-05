// Generated from https://github.com/streamrail/concurrent-map
package policy

import (
	"hash/fnv"
	"sync"
)

var SHARD_COUNT = 32

// TODO: Add Keys function which returns an array of keys for the map.

// A "thread" safe map of type string:*Stats.
// To avoid lock bottlenecks this map is dived to several (SHARD_COUNT) map shards.
type StatsMap []*StatsMapShared
type StatsMapShared struct {
	items        map[string]*Stats
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

// Creates a new concurrent map.
func NewStatsMap() StatsMap {
	m := make(StatsMap, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &StatsMapShared{items: make(map[string]*Stats)}
	}
	return m
}

// Returns shard under given key
func (m StatsMap) GetShard(key string) *StatsMapShared {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	return m[int(hasher.Sum32())%SHARD_COUNT]
}

// Sets the given value under the specified key.
func (m *StatsMap) Set(key string, value *Stats) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.items[key] = value
}

// Retrieves an element from map under given key.
func (m StatsMap) Get(key string) (*Stats, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// Get item from shard.
	val, ok := shard.items[key]
	return val, ok
}

// Returns the number of elements within the map.
func (m StatsMap) Count() int {
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
func (m *StatsMap) Has(key string) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	defer shard.RUnlock()

	// See if element is within shard.
	_, ok := shard.items[key]
	return ok
}

// Removes an element from the map.
func (m *StatsMap) Remove(key string) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.items, key)
}

// Checks if map is empty.
func (m *StatsMap) IsEmpty() bool {
	return m.Count() == 0
}

// Used by the Iter & IterBuffered functions to wrap two variables together over a channel,
type Tuple struct {
	Key string
	Val *Stats
}

// Returns an iterator which could be used in a for range loop.
func (m StatsMap) Iter() <-chan Tuple {
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
func (m StatsMap) IterBuffered() <-chan Tuple {
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
