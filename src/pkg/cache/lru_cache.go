package cache

import (
	"context"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/mmihic/golib/src/pkg/genericlist"
)

type lruCache[K comparable, V any] struct {
	mut      sync.Mutex
	byKey    map[K]*entry[K, V]
	byAccess *genericlist.List[*entry[K, V]]
	maxSize  int
	stats    Statistics

	defaultTTL time.Duration
	clock      clockwork.Clock
	loadFn     LoadFn[K, V]
}

func newLRUCache[K comparable, V any](props *cacheProperties[K, V]) Cache[K, V] {
	c := &lruCache[K, V]{}

	c.maxSize = props.maxSize
	c.byKey = make(map[K]*entry[K, V], props.maxSize)
	c.byAccess = genericlist.New[*entry[K, V]]()
	c.clock = props.clock
	c.defaultTTL = props.defaultTTL
	c.loadFn = props.loadFn
	return c
}

// MaxSize is the maximum size of the cache.
func (c *lruCache[K, V]) MaxSize() int {
	return c.maxSize
}

// Get retrieves an entry from the cache, loading it on demand
// if the entry is not present and a load function has been
// specified.
func (c *lruCache[K, V]) Get(ctx context.Context, k K) (V, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	if entry, ok := c.byKey[k]; ok {
		return c.accessValueLocked(ctx, k, entry)
	}

	return c.loadLocked(ctx, k)
}

// Put writes an entry to the cache, replacing the existing entry.
func (c *lruCache[K, V]) Put(_ context.Context, k K, v V) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.putLocked(k, v, time.Time{})
}

// PutWithTTL writes an entry to the cache with an explicit TTL, replacing
// the existing entry and TTL.
func (c *lruCache[K, V]) PutWithTTL(_ context.Context, k K, v V, expiry time.Time) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.putLocked(k, v, expiry)
}

// Statistics returns a copy of the cache statistics.
func (c *lruCache[K, V]) Statistics() Statistics {
	c.mut.Lock()
	defer c.mut.Unlock()

	stats := c.stats
	stats.CurrentSize = c.byAccess.Len()
	return stats
}

// ShardStatistics returns the statistics for each shard.
// LRUCaches are just a single shard (themselves)
func (c *lruCache[K, V]) ShardStatistics() []Statistics {
	return []Statistics{c.Statistics()}
}

func (c *lruCache[K, V]) accessValueLocked(
	ctx context.Context, k K, entry *entry[K, V],
) (V, error) {
	// If another goroutine is loading, wait until that goroutine is done
	for entry.loading != nil {
		loading := entry.loading
		c.mut.Unlock()

		select {
		case <-ctx.Done():
			c.mut.Lock()
			var noop V
			return noop, ctx.Err()
		case <-loading:
			c.mut.Lock()

			// We can't rely on the entry still being valid, since the load
			// we were waiting on might have failed and another load proceeded,
			// or the loaded entry might expire or be evicted
			entry = c.byKey[k]
			if entry == nil {
				return c.loadLocked(ctx, k)
			}
		}
	}

	// If the entry is expired, evict or reload it
	if !entry.expiry.IsZero() && c.clock.Now().After(entry.expiry) {
		c.stats.Expirations++
		delete(c.byKey, entry.k)
		c.byAccess.Remove(entry.element)
		return c.loadLocked(ctx, k)
	}

	// The entry is valid and contains a value, so mark it as the most recently
	// used and return the new value.
	c.moveToFrontLocked(entry)
	c.stats.Hits++
	val := entry.v
	return val, nil
}

func (c *lruCache[K, V]) loadLocked(ctx context.Context, k K) (V, error) {
	if c.loadFn == nil {
		c.stats.Misses++
		var noop V
		return noop, ErrNotFound
	}

	// Create an entry marked as loading, to prevent a thundering herd
	// on this particular value.
	entry := &entry[K, V]{
		loading: make(chan struct{}),
	}
	c.byKey[k] = entry
	c.stats.LoadAttempts++

	// Call the loader without the mutex locked
	c.mut.Unlock()
	v, expiry, err := c.loadFn(ctx, k)
	c.mut.Lock()

	// Regardless of the outcome, tell other goroutines waiting on this
	// load that we're done. Because we have the mutex, those goroutines
	// will block until we've fully recorded the result of the outcome
	// in the state of the cache i.e. have either updated the entry with the
	// value (if the load succeeded) or remove the entry from the map
	// (if the load failed)
	close(entry.loading)
	entry.loading = nil // Tells subsequent accesses that there is no load in progress
	if err != nil {
		// Remove the entry,so that subsequent goroutines (including those
		// that are waiting) will not find it and will attempt a load
		// themselves
		delete(c.byKey, k)

		if err == ErrNotFound {
			c.stats.Misses++
		} else {
			c.stats.LoadFailures++
		}

		var noop V
		return noop, err
	}

	// Save the loaded value in the cache
	c.putLocked(k, v, expiry)
	return v, nil
}

func (c *lruCache[K, V]) putLocked(k K, v V, expiry time.Time) {
	e, ok := c.byKey[k]
	if !ok {
		e = &entry[K, V]{}
		c.byKey[k] = e
	}

	if !expiry.IsZero() {
		e.expiry = expiry
	} else if c.defaultTTL != 0 {
		e.expiry = c.clock.Now().Add(c.defaultTTL)
	}

	e.k = k
	e.v = v
	c.moveToFrontLocked(e)

	if c.byAccess.Len() > c.maxSize {
		c.evictToSize()
	}
}

func (c *lruCache[K, V]) moveToFrontLocked(entry *entry[K, V]) {
	if entry.element == nil {
		entry.element = c.byAccess.PushFront(entry)
	} else {
		c.byAccess.MoveToFront(entry.element)
	}
}

func (c *lruCache[K, V]) evictToSize() {
	for c.byAccess.Len() > c.maxSize {
		c.stats.Evictions++
		toEvict := c.byAccess.Back()
		delete(c.byKey, toEvict.Value.k)
		c.byAccess.Remove(toEvict)

		// TODO(mmihic): Add to the list of evicted and call the eviction callback
	}
}

type entry[K comparable, V any] struct {
	k       K
	v       V
	expiry  time.Time
	element *genericlist.Element[*entry[K, V]]
	loading chan struct{}
}
