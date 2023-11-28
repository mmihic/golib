// Package cli contains command line tools and base commands.
package cli

import (
	"encoding/json"
	"io"
	"os"
)

const (
	// OutputToTemp is a marker Output for writing to a temp file. Used to
	// test output.
	OutputToTemp = "~~temp~~"
)

// Outputter is a mix-in for commands that generate output.
type Outputter struct {
	PrettyPrint bool   `name:"pretty-print"`
	Output      string `name:"output" short:"o" description:"output location"`
}

// WriteOutput writes the given value to the requested output.
func (cmd *Outputter) WriteOutput(v any) error {
	var out io.Writer
	if cmd.Output == "" || cmd.Output == "--" {
		out = os.Stdout
	} else if cmd.Output == OutputToTemp {
		// Write to a temp file. Mostly helpful for tests
		f, err := os.CreateTemp("", "base-cmp-test")
		if err != nil {
			return err
		}

		// Replace the name with the temp file
		cmd.Output = f.Name()

		out = f
	} else {
		f, err := os.Create(cmd.Output)
		if err != nil {
			return err
		}

		out = f
	}

	// If the output is a function, call the function
	if fn, ok := v.(func(_ io.Writer) error); ok {
		return fn(out)
	}

	// Otherwise marshal as JSON
	enc := json.NewEncoder(out)
	if cmd.PrettyPrint {
		enc.SetIndent("", "  ")
	}

	return enc.Encode(v)
}
