package timex

import (
	"encoding/json"
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
			assert.Equal(t, tt.want < 0, baseline.Before(other))
			assert.Equal(t, tt.want > 0, baseline.After(other))
			assert.Equal(t, tt.want == 0, baseline.Equal(other))
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

func TestMonthYear_AddMonths(t *testing.T) {
	start := MustParseMonthYear("2021-10")
	for _, tt := range []struct {
		numMonths int
		expected  MonthYear
	}{
		{0, start},
		{1, MustParseMonthYear("2021-11")},
		{2, MustParseMonthYear("2021-12")},
		{3, MustParseMonthYear("2022-01")},
		{4, MustParseMonthYear("2022-02")},
		{12, MustParseMonthYear("2022-10")},
		{24, MustParseMonthYear("2023-10")},
		{26, MustParseMonthYear("2023-12")},
		{-1, MustParseMonthYear("2021-09")},
		{-2, MustParseMonthYear("2021-08")},
		{-3, MustParseMonthYear("2021-07")},
		{-4, MustParseMonthYear("2021-06")},
		{-5, MustParseMonthYear("2021-05")},
		{-10, MustParseMonthYear("2020-12")},
		{-12, MustParseMonthYear("2020-10")},
		{-24, MustParseMonthYear("2019-10")},
		{-26, MustParseMonthYear("2019-08")},
	} {
		actual := start.AddMonths(tt.numMonths)
		assert.Equalf(t, tt.expected, actual,
			"adding %d months did not produce correct results", tt.numMonths)
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

func TestMonthYear_PriorMonth(t *testing.T) {
	actual := MustParseMonthYear("2021-02")
	for _, want := range []MonthYear{
		MustParseMonthYear("2021-01"),
		MustParseMonthYear("2020-12"),
		MustParseMonthYear("2020-11"),
		MustParseMonthYear("2020-10"),
		MustParseMonthYear("2020-09"),
	} {
		actual = actual.PriorMonth()
		require.Equal(t, actual, want)
	}
}

func TestMonthsBetween(t *testing.T) {
	first, second := MustParseMonthYear("2021-10"), MustParseMonthYear("2022-03")
	assert.Equal(t, 5, MonthsBetween(first, second))
	assert.Equal(t, 5, MonthsBetween(second, first))

	first, second = MustParseMonthYear("2021-10"), MustParseMonthYear("2024-04")
	assert.Equal(t, 30, MonthsBetween(first, second))
	assert.Equal(t, 30, MonthsBetween(second, first))

	first, second = MustParseMonthYear("2021-10"), MustParseMonthYear("2023-03")
	assert.Equal(t, 17, MonthsBetween(first, second))
	assert.Equal(t, 17, MonthsBetween(second, first))
}

func TestMonthYearJSON(t *testing.T) {
	type Embedded struct {
		When MonthYear `json:"when"`
	}

	var em Embedded
	err := json.Unmarshal([]byte(`{"when": "2022-09"}`), &em)
	require.NoError(t, err)

	assert.Equal(t, Embedded{
		When: MonthYear{Year: 2022, Month: time.September},
	}, em)
}
