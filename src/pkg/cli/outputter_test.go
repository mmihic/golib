package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	writeGoldenFile = false
)

func TestOutputter_WriteOutputJSON(t *testing.T) {
	out := &Outputter{
		Output: OutputToTemp,
	}

	err := out.WriteOutput(struct {
		Title   string            `json:"title"`
		Authors []string          `json:"authors"`
		Tags    map[string]string `json:"tags"`
	}{
		Title: "This is my title",
		Authors: []string{
			"joe@banana.com",
			"jeff@banana.com",
			"jane@banana.com",
		},
		Tags: map[string]string{
			"bar":  "zed",
			"foo":  "zork",
			"mork": "ork",
		},
	})

	require.NoError(t, err)
	assertFileMatches(t, out, "testdata/raw.json")
}

func TestOutputter_WriteOutputPrettyJSON(t *testing.T) {
	out := &Outputter{
		Output:      OutputToTemp,
		PrettyPrint: true,
	}

	err := out.WriteOutput(struct {
		Title   string            `json:"title"`
		Authors []string          `json:"authors"`
		Tags    map[string]string `json:"tags"`
	}{
		Title: "This is my title",
		Authors: []string{
			"joe@banana.com",
			"jeff@banana.com",
			"jane@banana.com",
		},
		Tags: map[string]string{
			"bar":  "zed",
			"foo":  "zork",
			"mork": "ork",
		},
	})

	require.NoError(t, err)
	assertFileMatches(t, out, "testdata/pretty.json")

}

func TestOutputter_WriteOutputWriter(t *testing.T) {
	out := &Outputter{
		Output: OutputToTemp,
	}

	err := out.WriteOutput(func(w io.Writer) error {
		_, err := w.Write([]byte(strings.Join(
			[]string{
				"this is the first line",
				"this is the second line",
				"this is the third line",
			}, "\n")))
		return err
	})

	require.NoError(t, err)
	assertFileMatches(t, out, "testdata/plain.txt")
}

func assertFileMatches(t *testing.T, out *Outputter, testdata string) {
	outputf, err := os.Open(out.Output)
	require.NoError(t, err)

	if writeGoldenFile {
		goldenf, err := os.Create(testdata)
		require.NoError(t, err)

		_, err = io.Copy(goldenf, outputf)
		require.NoError(t, err)
		return
	}

	actual, err := io.ReadAll(outputf)
	require.NoError(t, err)

	goldenf, err := os.Open(testdata)
	require.NoError(t, err)
	expected, err := io.ReadAll(goldenf)

	assert.Equal(t, string(actual), string(expected))
}
