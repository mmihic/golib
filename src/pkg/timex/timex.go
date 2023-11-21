// Package timex contained time-related extensions, notably
// representations of a Date without a time, and a MonthYear
// without a day.
package timex

import "time"

// MustParseTime parses the given string according to the provided layout,
// panicking if the time cannot be parsed. Useful for tests.
func MustParseTime(layout, s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		panic(err)
	}
	return t
}
