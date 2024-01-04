package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverse(t *testing.T) {
	elements := []string{"foo", "bar", "zed", "mork", "ork"}
	Reverse(elements)
	assert.Equal(t, []string{"ork", "mork", "zed", "bar", "foo"}, elements)
}

func TestReversed(t *testing.T) {
	elements := []string{"foo", "bar", "zed", "mork", "ork"}

	reversed := Reversed(elements)
	assert.Equal(t, []string{"foo", "bar", "zed", "mork", "ork"}, elements)
	assert.Equal(t, []string{"ork", "mork", "zed", "bar", "foo"}, reversed)
}

func TestClone(t *testing.T) {
	elements := []string{"foo", "bar", "zed", "mork", "ork"}
	cloned := Clone(elements)

	assert.Equal(t, []string{"foo", "bar", "zed", "mork", "ork"}, elements)
	assert.Equal(t, []string{"foo", "bar", "zed", "mork", "ork"}, cloned)

	// Change the original, the copy should not change
	elements[3] = "quork"
	assert.Equal(t, []string{"foo", "bar", "zed", "quork", "ork"}, elements)
	assert.Equal(t, []string{"foo", "bar", "zed", "mork", "ork"}, cloned)
}
