package timex

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMonthYear(t *testing.T) {
	m, err := ParseMonthYear("2023-10")
	require.NoError(t, err)
	require.Equal(t, MonthYear{
		Month: time.October,
		Year:  2023,
	}, m)
	require.Equal(t, "2023-10", m.String())
}

func TestMonthYear_CompareTo(t *testing.T) {
	baseline := MustParseMonthYear("2023-10")
	for _, tt := range []struct {
		other string
		want  int
	}{
		{"2021-12", 1},
		{"2024-01", -1},
		{"2023-09", 1},
		{"2023-11", -1},
		{"2023-10", 0},
	} {
		t.Run(fmt.Sprintf("%s vs %s", baseline, tt.other), func(t *testing.T) {
			other := MustParseMonthYear(tt.other)
			result := baseline.CompareTo(other)
			if !assert.Equal(t, tt.want, result) {
				return
			}

			assert.Equal(t, tt.want < 0, baseline.Less(other))
		})
	}
}

func TestMonthYear_StartOfMonth(t *testing.T) {
	for _, tt := range []struct {
		baseline string
		want     string
	}{
		{"2004-01", "2004-01-01"},
		{"2004-02", "2004-02-01"},
		{"2004-03", "2004-03-01"},
		{"2004-04", "2004-04-01"},
		{"2004-05", "2004-05-01"},
		{"2004-06", "2004-06-01"},
		{"2004-07", "2004-07-01"},
		{"2004-08", "2004-08-01"},
		{"2004-09", "2004-09-01"},
		{"2004-10", "2004-10-01"},
		{"2004-11", "2004-11-01"},
		{"2004-12", "2004-12-01"},
	} {
		t.Run(tt.baseline, func(t *testing.T) {
			var (
				baseline = MustParseMonthYear(tt.baseline)
				want     = MustParseDate(tt.want)
				actual   = baseline.MonthStart()
			)

			assert.Equal(t, want, actual)
		})
	}
}

func TestMonthYear_EndOfMonth(t *testing.T) {
	for _, tt := range []struct {
		baseline string
		want     string
	}{
		{"2004-01", "2004-01-31"},
		{"2004-02", "2004-02-29"}, // leap year
		{"2000-02", "2000-02-29"}, // leap year
		{"2001-02", "2001-02-28"}, // not a leap year
		{"1700-02", "1700-02-28"}, // not a leap year
		{"2004-03", "2004-03-31"},
		{"2004-04", "2004-04-30"},
		{"2004-05", "2004-05-31"},
		{"2004-06", "2004-06-30"},
		{"2004-07", "2004-07-31"},
		{"2004-08", "2004-08-31"},
		{"2004-09", "2004-09-30"},
		{"2004-10", "2004-10-31"},
		{"2004-11", "2004-11-30"},
		{"2004-12", "2004-12-31"},
	} {
		t.Run(tt.baseline, func(t *testing.T) {
			var (
				baseline = MustParseMonthYear(tt.baseline)
				want     = MustParseDate(tt.want)
				actual   = baseline.MonthEnd()
			)

			assert.Equal(t, want, actual)
		})
	}
}

func TestMonthYear_NextMonth(t *testing.T) {
	actual := MustParseMonthYear("2021-10")
	for _, want := range []MonthYear{
		MustParseMonthYear("2021-11"),
		MustParseMonthYear("2021-12"),
		MustParseMonthYear("2022-01"),
		MustParseMonthYear("2022-02"),
		MustParseMonthYear("2022-03"),
	} {
		actual = actual.NextMonth()
		require.Equal(t, actual, want)
	}
}

func TestMonthsBetween(t *testing.T) {
	first, second := MustParseMonthYear("2021-10"), MustParseMonthYear("2022-03")
	assert.Equal(t, 6, MonthsBetween(first, second))
	assert.Equal(t, 6, MonthsBetween(second, first))
}
