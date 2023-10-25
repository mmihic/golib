package ptr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	assert.True(t, *Bool(true))
}

func TestFloat64(t *testing.T) {
	assert.Equal(t, 209309340.32934, *Float64(209309340.32934))
}

func TestInt(t *testing.T) {
	assert.Equal(t, 3454, *Int(3454))
}

func TestString(t *testing.T) {
	assert.Equal(t, "hello there!", *String("hello there!"))
}

func TestTime(t *testing.T) {
	n := time.Date(2024, time.October, 14, 22, 56, 47, 0, time.UTC)
	assert.Equal(t, n, *Time(n))
}
