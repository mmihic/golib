// Package ptr contains helpers to create pointers to various types.
package ptr

import "time"

// String returns a string pointer for the given value.
func String(s string) *string {
	return &s
}

// Int returns an int pointer for the given value.
func Int(n int) *int {
	return &n
}

// Float64 returns a float64 pointer for the given value.
func Float64(n float64) *float64 {
	return &n
}

// Time returns a time pointer for the given value.
func Time(t time.Time) *time.Time {
	return &t
}

// Bool returns a bool pointer for the given value.
func Bool(b bool) *bool {
	return &b
}
