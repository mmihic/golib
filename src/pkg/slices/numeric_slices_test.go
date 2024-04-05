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
	assert.Equal(t, 207.5, Median([]float64{100, 125, 375, 290}))
	assert.Equal(t, 290.0, Median([]float64{1098, 100, 125, 375, 290}))
}
