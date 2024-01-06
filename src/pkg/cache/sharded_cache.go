package cache

import (
	"context"
	"hash/maphash"
	"time"
)

// A HashFn hashes a key into a shard bucket.
type HashFn[K comparable] func(k K) int

// StringHash is a standard string hash function.
func StringHash(k string) int {
	return int(maphash.String(stdSeed, k))
}

// BytesHash is a standard byte slice has function.
func BytesHash(k []byte) int {
	return int(maphash.Bytes(stdSeed, k))
}

// IntHash is a standard int hash function.
func IntHash(k int) int {
	return int(k)
}

var (
	stdSeed = maphash.MakeSeed()
)

// Sharded configures a cache to be sharded, allowing for concurrent
// access across shards.
func Sharded[K comparable, V any](shardCount uint, hashFn HashFn[K]) Option[K, V] {
	return func(c *cacheProperties[K, V]) {
		c.shardCount = shardCount
		c.hashFn = hashFn
	}
}

type shardedCache[K comparable, V any] struct {
	shards []Cache[K, V]
	hashFn HashFn[K]
}

func (c *shardedCache[K, V]) Get(ctx context.Context, k K) (V, error) {
	return c.shard(k).Get(ctx, k)
}

func (c *shardedCache[K, V]) Put(ctx context.Context, k K, v V) {
	c.shard(k).Put(ctx, k, v)
}

func (c *shardedCache[K, V]) PutWithTTL(ctx context.Context, k K, v V, expiry time.Time) {
	c.shard(k).PutWithTTL(ctx, k, v, expiry)
}

func (c *shardedCache[K, V]) Statistics() Statistics {
	var stats Statistics
	for _, shard := range c.shards {
		stats = stats.Add(shard.Statistics())
	}

	return stats
}

func (c *shardedCache[K, V]) shard(k K) Cache[K, V] {
	hash := c.hashFn(k)
	if hash < 0 {
		hash = -hash
	}

	return c.shards[hash%len(c.shards)]
}

func newShardedCache[K comparable, V any](props *cacheProperties[K, V]) Cache[K, V] {
	c := &shardedCache[K, V]{
		shards: make([]Cache[K, V], props.shardCount),
		hashFn: props.hashFn,
	}

	lruProps := *props
	lruProps.shardCount = 0
	lruProps.hashFn = nil
	for i := range c.shards {
		c.shards[i] = newLRUCache(&lruProps)
	}

	return c
}

var (
	_ Cache[string, string] = &shardedCache[string, string]{}
)
