package mapx

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	keys := Keys(map[string]string{
		"foo":   "bar",
		"zed":   "clark",
		"quork": "ork",
		"blerk": "berk",
	})

	sort.Strings(keys)
	assert.Equal(t, []string{"blerk", "foo", "quork", "zed"}, keys)
}