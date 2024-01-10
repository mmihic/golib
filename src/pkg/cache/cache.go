// Package cache contains a simple fixed-sized Cache, supporting
// both read-through and write-ahead caching, with evictions using
// an LRU eviction policy.
package cache

import (
	"context"
	"errors"
	"time"

	"github.com/jonboulle/clockwork"
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

// Add adds two Statistics, returning the aggregate result.
func (s Statistics) Add(other Statistics) Statistics {
	return Statistics{
		Hits:         s.Hits + other.Hits,
		Misses:       s.Misses + other.Misses,
		LoadAttempts: s.LoadAttempts + other.LoadAttempts,
		LoadFailures: s.LoadFailures + other.LoadFailures,
		Expirations:  s.Expirations + other.Expirations,
		Evictions:    s.Evictions + other.Evictions,
		CurrentSize:  s.CurrentSize + other.CurrentSize,
	}
}

// A LoadFn is a function that loads a cache entry on demand.
type LoadFn[K comparable, V any] func(context.Context, K) (V, time.Time, error)

// Cache is a fixed sized cache that evicts entries using a least-recently-used
// eviction policy. Supports both read-through and write-ahead caching, allows
// for synchronous and asynchronous eviction, and supports TTLs on cache entries.
type Cache[K comparable, V any] interface {
	// Get retrieves an entry from the cache, loading it on demand
	// if the entry is not present and a load function has been
	// specified.
	Get(ctx context.Context, k K) (V, error)

	// Put writes an entry to the cache, replacing the existing entry.
	Put(_ context.Context, k K, v V)

	// PutWithTTL writes an entry to the cache with an explicit TTL, replacing
	// the existing entry and TTL.
	PutWithTTL(_ context.Context, k K, v V, expiry time.Time)

	// Statistics returns a copy of the cache statistics.
	Statistics() Statistics

	// ShardStatistics returns the statistics for each shard.
	ShardStatistics() []Statistics
}

// cacheProperties are the properties to the cache
type cacheProperties[K comparable, V any] struct {
	maxSize    int
	defaultTTL time.Duration
	clock      clockwork.Clock
	loadFn     LoadFn[K, V]
	hashFn     HashFn[K]
	shardCount uint
}

// An Option is an option to a sharded cache.
type Option[K comparable, V any] func(c *cacheProperties[K, V])

// WithDefaultTTL sets the default expiry for newly loaded cache entries. Can
// be overridden on a per-load basis.
func WithDefaultTTL[K comparable, V any](defaultTTL time.Duration) Option[K, V] {
	return func(c *cacheProperties[K, V]) {
		c.defaultTTL = defaultTTL
	}
}

// WithClock configures the clock to use with the cache.
func WithClock[K comparable, V any](clock clockwork.Clock) Option[K, V] {
	return func(c *cacheProperties[K, V]) {
		c.clock = clock
	}
}

// WithLoadFn sets the read-through load function to use when a cache entry
// does not exist.
func WithLoadFn[K comparable, V any](loader LoadFn[K, V]) Option[K, V] {
	return func(c *cacheProperties[K, V]) {
		c.loadFn = loader
	}
}

// New creates a new Cache with a max size and a set of options.
func New[K comparable, V any](maxSize int, opts ...Option[K, V]) Cache[K, V] {
	props := &cacheProperties[K, V]{
		maxSize: maxSize,
	}

	for _, opt := range opts {
		opt(props)
	}

	if props.clock == nil {
		props.clock = clockwork.NewRealClock()
	}

	if props.shardCount > 1 {
		return newShardedCache(props)
	}

	return newLRUCache(props)
}
