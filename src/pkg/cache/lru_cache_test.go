package cache

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestLRUCache(t *testing.T) {
	suite.Run(t, &TestSuite{
		newCache: New[string, string],
	})
}
