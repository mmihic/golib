package timex

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDate(t *testing.T) {
	d, err := ParseDate("2023-10-14")
	require.NoError(t, err)
	require.Equal(t, Date{
		Day:   14,
		Month: time.October,
		Year:  2023,
	}, d)
	require.Equal(t, "2023-10-14", d.String())
	require.Equal(t,
		time.Date(2023, time.October, 14, 0, 0, 0, 0, time.UTC),
		d.DayStart())
	require.Equal(t,
		time.Date(2023, time.October, 14, 23, 59, 59, 999999999, time.UTC),
		d.DayEnd())
}

func TestDate_CompareTo(t *testing.T) {
	baseline := MustParseDate("2023-10-14")
	for _, tt := range []struct {
		against string
		want    int
	}{
		{"2024-10-14", -1},
		{"2022-10-14", 1},
		{"2023-11-14", -1},
		{"2023-04-14", 1},
		{"2023-10-15", -1},
		{"2023-10-13", 1},
		{"2023-10-14", 0},
	} {
		t.Run(fmt.Sprintf("%s vs %s", baseline, tt.against), func(t *testing.T) {
			against := MustParseDate(tt.against)
			assert.Equal(t, tt.want, baseline.CompareTo(against))
			assert.Equal(t, 0-tt.want, against.CompareTo(baseline))
			assert.Equal(t, tt.want < 0, baseline.Less(against))
			assert.Equal(t, tt.want < 0, baseline.Before(against))
			assert.Equal(t, tt.want > 0, baseline.After(against))
			assert.Equal(t, tt.want == 0, baseline.Equal(against))
		})
	}
}

func TestDate_NextDay(t *testing.T) {
	for _, tt := range []struct {
		start string
		want  string
	}{
		{"2004-01-16", "2004-01-17"},
		{"2004-01-30", "2004-01-31"},
		{"2004-01-31", "2004-02-01"},
		{"2004-02-28", "2004-02-29"}, // leap year
		{"2000-02-28", "2000-02-29"}, // leap year
		{"2004-02-29", "2004-03-01"}, // leap year
		{"2000-02-29", "2000-03-01"}, // leap year
		{"2001-02-28", "2001-03-01"}, // not a leap year
		{"1700-02-28", "1700-03-01"}, // not a leap year
		{"2004-03-30", "2004-03-31"},
		{"2004-04-29", "2004-04-30"},
		{"2004-04-30", "2004-05-01"},
		{"2004-05-30", "2004-05-31"},
		{"2004-05-31", "2004-06-01"},
		{"2004-06-29", "2004-06-30"},
		{"2004-06-30", "2004-07-01"},
		{"2004-07-30", "2004-07-31"},
		{"2004-07-31", "2004-08-01"},
		{"2004-08-30", "2004-08-31"},
		{"2004-08-31", "2004-09-01"},
		{"2004-09-29", "2004-09-30"},
		{"2004-09-30", "2004-10-01"},
		{"2004-10-30", "2004-10-31"},
		{"2004-10-31", "2004-11-01"},
		{"2004-11-29", "2004-11-30"},
		{"2004-11-30", "2004-12-01"},
		{"2004-12-30", "2004-12-31"},
		{"2004-12-31", "2005-01-01"},
		{"2004-05-01", "2004-05-02"},
		{"2004-05-02", "2004-05-03"},
	} {
		t.Run(tt.start, func(t *testing.T) {
			var (
				start  = MustParseDate(tt.start)
				want   = MustParseDate(tt.want)
				actual = start.NextDay()
			)
			assert.Equal(t, want, actual)
		})
	}
}

func TestDateJSON(t *testing.T) {
	type Embedded struct {
		When Date `json:"when"`
	}

	var em Embedded
	err := json.Unmarshal([]byte(`{"when":"2023-11-14"}`), &em)
	require.NoError(t, err)
	assert.Equal(t, Embedded{
		When: Date{Day: 14, Month: time.November, Year: 2023},
	}, em)
}
