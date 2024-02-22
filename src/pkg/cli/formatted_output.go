package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/mmihic/golib/src/pkg/csvmarshal"
)

// A Formatter formats an output.
type Formatter interface {
	WriteFormatted(w io.Writer, pretty bool, val any) error
}

// A FormatterFn is a function that act as a formatter.
type FormatterFn func(w io.Writer, pretty bool, val any) error

// WriteFormatted writes the given value in the appropriate format.
func (fn FormatterFn) WriteFormatted(w io.Writer, pretty bool, val any) error {
	return fn(w, pretty, val)
}

// Format controls the output.
type Format string

// Known formats
const (
	FormatUnknown Format = ""
	FormatJSON    Format = "json"
	FormatCSV     Format = "csv"
)

// FormattedOutput renders output using formatting
type FormattedOutput struct {
	Output
	Format Format `help:"the format to use for output" default:"json"`
}

// WriteFormatted writes the given output according to the format.
func (cmd *FormattedOutput) WriteFormatted(val any) error {
	formatter, ok := LookupFormatter(cmd.Format)
	if !ok {
		return fmt.Errorf("unknown format '%s'", cmd.Format)
	}

	return cmd.WriteOutput(func(w io.Writer) error {
		return formatter.WriteFormatted(w, cmd.Compact, val)
	})
}

// RegisterFormatter registers a new Formatter for a given Format.
func RegisterFormatter(format Format, formatter Formatter) {
	formatters.Store(format, formatter)
}

// LookupFormatter returns the Formatter for the given Format.
func LookupFormatter(format Format) (Formatter, bool) {
	formatter, ok := formatters.Load(format)
	if !ok {
		return nil, ok
	}

	return formatter.(Formatter), ok
}

var (
	formatters = sync.Map{}
)

func init() {
	RegisterFormatter(FormatJSON, FormatterFn(func(w io.Writer, compact bool, val any) error {
		enc := json.NewEncoder(w)
		if !compact {
			enc.SetIndent("", "  ")
		}
		return enc.Encode(val)
	}))

	RegisterFormatter(FormatCSV, FormatterFn(func(w io.Writer, compact bool, val any) error {
		m, err := csvmarshal.NewMarshaller(reflect.TypeOf(val))
		if err != nil {
			return err
		}

		csvw := csv.NewWriter(w)
		defer csvw.Flush()

		if !compact {
			if err := csvw.Write(m.Headers()); err != nil {
				return fmt.Errorf("unable to write header: %w", err)
			}
		}

		return m.Marshal(csvw, val, "")
	}))
}
