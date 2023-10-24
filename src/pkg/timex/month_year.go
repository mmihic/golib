package timex

import (
	"encoding"
	"fmt"
	"time"
)

// MonthYear is a month and year combination.
type MonthYear struct {
	Month time.Month
	Year  int
}

// ParseMonthYear parses a month-year.
func ParseMonthYear(s string) (MonthYear, error) {
	tm, err := time.Parse("2006-01", s)
	if err != nil {
		return MonthYear{}, err
	}

	return MonthYear{
		Month: tm.Month(),
		Year:  tm.Year(),
	}, nil
}

// MustParseMonthYear parses a month-year, panicking
// if the month-year cannot be parsed. Useful for tests.
func MustParseMonthYear(s string) MonthYear {
	my, err := ParseMonthYear(s)
	if err != nil {
		panic(err)
	}

	return my
}

// UnmarshalText unmarshalls the date from a text value.
// Implements the TextUnmarshaler interface.
func (my *MonthYear) UnmarshalText(text []byte) error {
	u, err := ParseMonthYear(string(text))
	if err != nil {
		return err
	}

	*my = u
	return nil
}

// String returns a string format of the MonthYear
func (my MonthYear) String() string {
	return fmt.Sprintf("%04d-%02d", my.Year, my.Month)
}

// MonthStart returns the date that is the start of the month.
func (my MonthYear) MonthStart() Date {
	return Date{
		Day:   1,
		Month: my.Month,
		Year:  my.Year,
	}
}

// MonthEnd returns the date that is the end of the month.
func (my MonthYear) MonthEnd() Date {
	// Go to the start of the next month and backoff
	endOfMonth := my.NextMonth().MonthStart().DayStart().Add(-time.Nanosecond)
	return Date{
		Day:   endOfMonth.Day(),
		Month: endOfMonth.Month(),
		Year:  endOfMonth.Year(),
	}
}

// NextMonth returns the next month.
func (my MonthYear) NextMonth() MonthYear {
	month, year := my.Month+1, my.Year
	if month > 12 {
		year++
		month = time.January
	}

	return MonthYear{
		Month: month,
		Year:  year,
	}
}

// CompareTo compares two (month, year). Returns:
//
//	-1 if this MonthYear is earlier than the provided MonthYear
//	1 if this MonthYear is later than the provided MonthYear
//	0 if this MonthYear is the same as the provided MonthYear
func (my MonthYear) CompareTo(other MonthYear) int {
	if my.Year > other.Year {
		return 1
	}

	if my.Year < other.Year {
		return -1
	}

	if my.Month > other.Month {
		return 1
	}

	if my.Month < other.Month {
		return -1
	}

	return 0
}

// Less returns true if the MonthYear is earlier than the provided
// MonthYear.
func (my MonthYear) Less(other MonthYear) bool {
	return my.CompareTo(other) < 0
}

// MonthsBetween returns the number of months between two (month, year)
func MonthsBetween(from, to MonthYear) int {
	if to.Less(from) {
		from, to = to, from
	}

	return (from.Year + int(from.Month)) - (to.Year + int(to.Month))
}

var (
	_ encoding.TextUnmarshaler = &MonthYear{}
)
