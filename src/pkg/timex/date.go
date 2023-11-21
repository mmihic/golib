package timex

import (
	"encoding"
	"fmt"
	"time"
)

// A Date is a month, day, year.
type Date struct {
	Day   int
	Month time.Month
	Year  int
}

// ParseDate parses a date.
func ParseDate(s string) (Date, error) {
	tm, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, err
	}

	return Date{
		Day:   tm.Day(),
		Month: tm.Month(),
		Year:  tm.Year(),
	}, nil
}

// MustParseDate parses a date, panicking if the date can't be parsed.
// Useful for tests.
func MustParseDate(s string) Date {
	d, err := ParseDate(s)
	if err != nil {
		panic(err)
	}

	return d
}

// UnmarshalText unmarshalls the date from a text value.
// Implements the TextUnmarshaler interface.
func (d *Date) UnmarshalText(text []byte) error {
	u, err := ParseDate(string(text))
	if err != nil {
		return err
	}

	*d = u
	return nil
}

// String returns a string format of the date.
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

// DayStart returns the time at the start of the day, in UTC.
func (d Date) DayStart() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

// DayEnd returns the time at the end of the day, in UTC.
func (d Date) DayEnd() time.Time {
	return d.DayStart().Add(time.Hour * 24).Add(-time.Nanosecond)
}

// NextDay returns the next day.
func (d Date) NextDay() Date {
	nextDay := d.DayStart().Add(time.Hour * 24)
	return Date{
		Day:   nextDay.Day(),
		Month: nextDay.Month(),
		Year:  nextDay.Year(),
	}
}

// After returns true if this date is after another.
func (d Date) After(other Date) bool {
	return d.CompareTo(other) > 0
}

// Before returns true if this date is before another date.
func (d Date) Before(other Date) bool {
	return d.CompareTo(other) < 0
}

// Equal returns true if this date is equal to another date.
func (d Date) Equal(other Date) bool {
	return d.CompareTo(other) == 0
}

// CompareTo compares this date to another;
func (d Date) CompareTo(other Date) int {
	if d.Year > other.Year {
		return 1
	}

	if d.Year < other.Year {
		return -1
	}

	if d.Month > other.Month {
		return 1
	}

	if d.Month < other.Month {
		return -1
	}

	if d.Day > other.Day {
		return 1
	}

	if d.Day < other.Day {
		return -1
	}

	return 0
}

// Less returns true if this date is earlier than the other date.
func (d Date) Less(other Date) bool {
	return d.CompareTo(other) < 0
}

var (
	_ encoding.TextUnmarshaler = &Date{}
)
