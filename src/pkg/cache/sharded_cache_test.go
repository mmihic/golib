package cache

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestShardedCache(t *testing.T) {
	suite.Run(t, &TestSuite{
		newCache: func(maxSize int, opts ...Option[string, string]) Cache[string, string] {
			allOpts := make([]Option[string, string], 0, len(opts)+1)
			allOpts = append(allOpts, opts...)
			allOpts = append(allOpts, Sharded[string, string](10, StringHash))
			return New(maxSize, allOpts...)
		},
	})
}
