package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMean(t *testing.T) {
	assert.Equal(t, Mean([]int{1098, 100, 125, 375, 290}), 397)
}

func TestMin(t *testing.T) {
	assert.Equal(t, Min([]int{1098, 100, 125, 375, 290}), 100)
}

func TestMax(t *testing.T) {
	assert.Equal(t, Max([]int{1098, 100, 125, 375, 290}), 1098)
}

func TestMedian(t *testing.T) {
	assert.Equal(t, Median([]int{100, 125, 375, 290}), 250)
	assert.Equal(t, Median([]int{1098, 100, 125, 375, 290}), 125)
}
