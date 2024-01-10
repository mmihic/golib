package cache

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ShardedCacheTestSuite is the full suite for LRU tests.
type ShardedCacheTestSuite struct {
	TestSuite
}

func TestShardedCache(t *testing.T) {
	// Use a bad but stable hashing function
	hashFn := func(k string) int {
		return int(k[0])
	}

	suite.Run(t, &ShardedCacheTestSuite{
		TestSuite: TestSuite{
			newCache: func(maxSize int, opts ...Option[string, string]) Cache[string, string] {
				allOpts := make([]Option[string, string], 0, len(opts)+1)
				allOpts = append(allOpts, opts...)
				allOpts = append(allOpts, Sharded[string, string](2, hashFn))
				return New(maxSize, allOpts...)
			},
		},
	})
}
