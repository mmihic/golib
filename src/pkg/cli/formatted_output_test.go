package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormattedOutput_CSV(t *testing.T) {
	out := &FormattedOutput{
		Format: FormatCSV,
		Output: Output{
			Output: OutputToTemp,
		},
	}

	err := out.WriteFormatted([]struct {
		Title  string `json:"title"`
		Author string `json:"author"`
	}{
		{
			Title:  "This is my title",
			Author: "joe@banana.com",
		},
		{
			Title:  "This is my other title",
			Author: "jane@banana.com",
		},
		{
			Title:  "This is my third title",
			Author: "jackie@banana.com",
		},
	})

	require.NoError(t, err)
	assertFileMatches(t, &out.Output, "testdata/formatted.csv")
}

func TestFormattedOutput_JSON(t *testing.T) {
	out := &FormattedOutput{
		Format: FormatJSON,
		Output: Output{
			Output: OutputToTemp,
		},
	}

	err := out.WriteFormatted([]struct {
		Title  string `json:"title"`
		Author string `json:"author"`
	}{
		{
			Title:  "This is my title",
			Author: "joe@banana.com",
		},
		{
			Title:  "This is my other title",
			Author: "jane@banana.com",
		},
		{
			Title:  "This is my third title",
			Author: "jackie@banana.com",
		},
	})

	require.NoError(t, err)
	assertFileMatches(t, &out.Output, "testdata/formatted.json")
}
