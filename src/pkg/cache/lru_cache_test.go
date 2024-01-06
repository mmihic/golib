package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// LRUTestSuite is the full suite for LRU tests.
type LRUTestSuite struct {
	TestSuite
}

// TestSyncEviction tests that when using synchronous eviction, the least recently used
// entry is evicted as soon as we reach the max size.
//
// NB(mmihic): We only test this directly against LRU caches because in sharded caches
// the distribution of keys over shards might result in a different eviction order (e.g.
// the LRU entry for the shard will be evicted, but this may not be the LRU entry for the
// cache as a whole)
func (s *LRUTestSuite) TestSyncEviction() {
	entries := map[string]string{
		"foo":       "bar",
		"zed":       "banana",
		"snork":     "mork",
		"gambas":    "camarones",
		"conch":     "snail",
		"ephemeral": "transient",
	}

	const maxSize = 3
	c := s.newCache(maxSize,
		WithLoadFn(func(_ context.Context, key string) (string, time.Time, error) {
			if val, ok := entries[key]; ok {
				return val, time.Time{}, nil
			}

			return "", time.Time{}, ErrNotFound
		}))

	// This is the access pattern
	// snork, zed, foo 		-> 3 loads 		[foo, zed, snork]
	// zed 					-> hit			[zed, foo, snork]
	// gambas				-> evict + load	[gambas, zed, foo]
	// gambas				-> hit			[gambas, zed, foo]
	// foo					-> hit			[foo, gambas, zed]
	// non-existent			-> load + miss	[foo, gambas, zed]
	// conch				-> evict + load	[conch, foo, gambas]
	// gambas				-> hit			[gambas, conch, foo]
	// non-existent			-> load + miss	[foo, gambas, zed]
	// foo					-> hit			[foo, gambas, conch]
	// zed					-> evict + load	[zed, foo, gambas]

	// So in the end we see 8 loads attempts, 3 evictions, 5 hits, 2 misses
	assertEntryFound(s.T(), c, "snork", entries["snork"])
	assertEntryFound(s.T(), c, "zed", entries["zed"])
	assertEntryFound(s.T(), c, "foo", entries["foo"])
	assertEntryFound(s.T(), c, "zed", entries["zed"])
	assertEntryFound(s.T(), c, "gambas", entries["gambas"])
	assertEntryFound(s.T(), c, "gambas", entries["gambas"])
	assertEntryFound(s.T(), c, "foo", entries["foo"])
	assertEntryNotFound(s.T(), c, "non-existent")
	assertEntryFound(s.T(), c, "conch", entries["conch"])
	assertEntryFound(s.T(), c, "gambas", entries["gambas"])
	assertEntryNotFound(s.T(), c, "non-existent")
	assertEntryFound(s.T(), c, "foo", entries["foo"])
	assertEntryFound(s.T(), c, "zed", entries["zed"])

	s.Require().Equal(Statistics{
		LoadAttempts: 8,
		Evictions:    3,
		Hits:         5,
		Misses:       2,
		CurrentSize:  3,
	}, c.Statistics())
}

func TestLRUCache(t *testing.T) {
	suite.Run(t, &LRUTestSuite{
		TestSuite: TestSuite{
			newCache: New[string, string],
		},
	})
}
