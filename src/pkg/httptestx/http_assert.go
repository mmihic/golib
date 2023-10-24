// Package httptestx contains utilities for writing server-side assertions,
// returning an error if the assertion failed.
package httptestx

import (
	"fmt"
	"github.com/mmihic/golib/src/pkg/httpx"
	"net/http"
)

// ServerCheck returns a client error if a server-side test assertion failed.
func ServerCheck(w http.ResponseWriter, result bool) bool {
	if !result {
		httpx.RespondWithError(w, fmt.Errorf("assertion failed on server"))
		return false
	}

	return true
}
