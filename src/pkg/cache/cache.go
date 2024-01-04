// Package cache contains a simple fixed-sized Cache, supporting
// both read-through and write-ahead caching, with evictions using
// an LRU eviction policy.
package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/mmihic/golib/src/pkg/genericlist"
	"github.com/mmihic/golib/src/pkg/lifecycle"
)

var (
	ErrNotFound = errors.New("cache entry not foune")
)

// Statistics are statistics about the cache.
type Statistics struct {
	Hits         int64
	Misses       int64
	LoadAttempts int64
	LoadFailures int64
	Expirations  int64
	Evictions    int64
	CurrentSize  int
}

// A LoadFn is a function that loads a cache entry on demand.
type LoadFn[K comparable, V any] func(context.Context, K) (V, time.Time, error)

// Cache is a fixed sized cache that evicts entries using a least-recently-used
// eviction policy. Supports both read-through and write-ahead caching, allows
// for synchronous and asynchronous eviction, and supports TTLs on cache entries.
type Cache[K comparable, V any] struct {
	mut      sync.Mutex
	byKey    map[K]*entry[K, V]
	byAccess *genericlist.List[*entry[K, V]]
	maxSize  int
	stats    Statistics

	evictionJob       *lifecycle.OnDemandJob
	evictInBackground bool
	reloadSchedule    time.Duration
	defaultTTL        time.Duration
	clock             clockwork.Clock
	loadFn            LoadFn[K, V]
}

// An Option is an option to an LRU cache.
type Option[K comparable, V any] func(c *Cache[K, V])

// WithDefaultTTL sets the default expiry for newly loaded cache entries. Can
// be overridden on a per-load basis.
func WithDefaultTTL[K comparable, V any](defaultTTL time.Duration) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.defaultTTL = defaultTTL
	}
}

// WithClock configures the clock to use with the cache.
func WithClock[K comparable, V any](clock clockwork.Clock) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.clock = clock
	}
}

// WithLoadFn sets the read-through load function to use when a cache entry
// does not exist.
func WithLoadFn[K comparable, V any](loader LoadFn[K, V]) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.loadFn = loader
	}
}

// New creates a new Cache with a max size and a set of options.
func New[K comparable, V any](maxSize int, opts ...Option[K, V]) *Cache[K, V] {
	c := &Cache[K, V]{}
	for _, opt := range opts {
		opt(c)
	}

	c.maxSize = maxSize
	c.byKey = make(map[K]*entry[K, V], maxSize)
	c.byAccess = genericlist.New[*entry[K, V]]()
	return c
}

// MaxSize is the maximum size of the cache.
func (c *Cache[K, V]) MaxSize() int {
	return c.maxSize
}

// Get retrieves an entry from the cache, loading it on demand
// if the entry is not present and a load function has been
// specified.
func (c *Cache[K, V]) Get(ctx context.Context, k K) (V, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	if entry, ok := c.byKey[k]; ok {
		return c.accessValueLocked(ctx, k, entry)
	}

	return c.loadLocked(ctx, k)
}

// Put writes an entry to the cache, replacing the existing entry.
func (c *Cache[K, V]) Put(_ context.Context, k K, v V) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.putLocked(k, v, time.Time{})
}

// PutWithTTL writes an entry to the cache with an explicit TTL, replacing
// the existing entry and TTL.
func (c *Cache[K, V]) PutWithTTL(_ context.Context, k K, v V, expiry time.Time) {
	c.mut.Lock()
	defer c.mut.Unlock()

	c.putLocked(k, v, expiry)
}

// Statistics returns a copy of the cache statistics.
func (c *Cache[K, V]) Statistics() Statistics {
	c.mut.Lock()
	defer c.mut.Unlock()

	stats := c.stats
	stats.CurrentSize = c.byAccess.Len()
	return stats
}

func (c *Cache[K, V]) accessValueLocked(
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

func (c *Cache[K, V]) loadLocked(ctx context.Context, k K) (V, error) {
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

func (c *Cache[K, V]) putLocked(k K, v V, expiry time.Time) {
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

func (c *Cache[K, V]) moveToFrontLocked(entry *entry[K, V]) {
	if entry.element == nil {
		entry.element = c.byAccess.PushFront(entry)
	} else {
		c.byAccess.MoveToFront(entry.element)
	}
}

func (c *Cache[K, V]) evictToSize() {
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
