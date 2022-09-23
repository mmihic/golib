// Package assertx contains additional assertions for testing.
package assertx

import (
	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"testing"
)

// JSONEq asserts to JSON documents are exact matches, producing a friendlier
// output if not.
func JSONEq(t *testing.T, expected, actual string) bool {
	opts := jsondiff.DefaultJSONOptions()
	diff, msg := jsondiff.Compare([]byte(expected), []byte(actual), &opts)
	if diff == jsondiff.FullMatch {
		return true
	}

	return assert.Fail(t, msg)
}
