package twwidth

import (
	"sync"
	"sync/atomic"
)

const (
	cacheMinCapacity = 1024
	cacheMaxCapacity = 100000
)

type lruCacheKey string

func (k lruCacheKey) String() string {
	return string(k)
}

// cacheEntry represents a node in the doubly-linked LRU list.
type cacheEntry struct {
	key   lruCacheKey
	value int
	prev  *cacheEntry
	next  *cacheEntry
}

// lruCache is a simple, bounded LRU cache for lruCacheKey -> int.
//
// It is fully thread-safe via a single mutex. For this use case
// (string width caching), the mutex overhead is negligible compared
// to Unicode/regex work.
type lruCache struct {
	mu       sync.Mutex
	items    map[lruCacheKey]*cacheEntry
	head     *cacheEntry // Most recently used
	tail     *cacheEntry // Least recently used
	capacity int
	hits     int64 // atomic access for hit tracking
	misses   int64 // atomic access for miss tracking
}

// newLRUCache creates a new LRU cache with the given capacity.
// If capacity <= 0, the cache effectively disables itself.
func newLRUCache(capacity int) *lruCache {
	if capacity <= 0 {
		// Disabled cache
		return &lruCache{
			items:    nil,
			capacity: 0,
		}
	}

	// Optional upper bound for very large capacities
	if capacity > cacheMaxCapacity {
		capacity = cacheMaxCapacity
	}

	return &lruCache{
		items:    make(map[lruCacheKey]*cacheEntry, capacity),
		capacity: capacity,
	}
}

// Get looks up a key in the cache.
// It returns (value, true) on hit, or (0, false) on miss.
func (c *lruCache) Get(key lruCacheKey) (int, bool) {
	if c == nil || c.capacity <= 0 {
		return 0, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.items[key]
	if !ok {
		atomic.AddInt64(&c.misses, 1)
		return 0, false
	}

	atomic.AddInt64(&c.hits, 1)
	// Move the accessed entry to the front (MRU).
	c.moveToFront(entry)

	return entry.value, true
}

// Put inserts or updates a key in the cache.
func (c *lruCache) Put(key lruCacheKey, value int) {
	if c == nil || c.capacity <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry.
	if entry, ok := c.items[key]; ok {
		entry.value = value
		c.moveToFront(entry)
		return
	}

	// Evict LRU if at capacity.
	if len(c.items) >= c.capacity && c.tail != nil {
		delete(c.items, c.tail.key)
		c.removeNode(c.tail)
	}

	// Insert new entry at front.
	entry := &cacheEntry{
		key:   key,
		value: value,
	}
	c.addToFront(entry)
	c.items[key] = entry
}

// Clear removes all entries from the cache.
func (c *lruCache) Clear() {
	if c == nil || c.capacity <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear the map by reallocating with the same capacity
	// This is more efficient than iterating for large maps
	if len(c.items) > 0 {
		c.items = make(map[lruCacheKey]*cacheEntry, c.capacity)
	}
	c.head = nil
	c.tail = nil

	// Reset metrics
	atomic.StoreInt64(&c.hits, 0)
	atomic.StoreInt64(&c.misses, 0)
}

// Len returns the current number of entries in the cache.
func (c *lruCache) Len() int {
	if c == nil || c.capacity <= 0 {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// Cap returns the maximum capacity of the cache.
func (c *lruCache) Cap() int {
	if c == nil {
		return 0
	}
	return c.capacity
}

// HitRate returns the cache hit rate (for debugging/monitoring).
// Returns 0 if the cache is disabled or has no accesses.
func (c *lruCache) HitRate() float64 {
	if c == nil || c.capacity <= 0 {
		return 0
	}
	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total)
}

// moveToFront moves an existing entry to the head (MRU position).
// Caller must hold c.mu.
func (c *lruCache) moveToFront(e *cacheEntry) {
	if e == c.head {
		return
	}
	c.removeNode(e)
	c.addToFront(e)
}

// addToFront inserts an entry at the head of the list.
// Caller must hold c.mu.
func (c *lruCache) addToFront(e *cacheEntry) {
	e.prev = nil
	e.next = c.head

	if c.head != nil {
		c.head.prev = e
	}
	c.head = e

	if c.tail == nil {
		// First element in the list
		c.tail = e
	}
}

// removeNode removes an entry from the list (but not from the map).
// Caller must hold c.mu.
func (c *lruCache) removeNode(e *cacheEntry) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		// e is head
		c.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		// e is tail
		c.tail = e.prev
	}
	// Explicitly break references to help garbage collection
	e.prev = nil
	e.next = nil
}
