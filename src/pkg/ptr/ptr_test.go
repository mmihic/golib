package ptr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToDeref(t *testing.T) {
	assert.Equal(t, Deref(To("foo"), "my-default"), "foo")
	assert.Equal(t, Deref(nil, "my-default"), "my-default")
}
