package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RespondWithError responds with an error.
func RespondWithError(w http.ResponseWriter, err error) {
	switch terr := err.(type) {
	case httpError:
		w.Header().Set("Content-Type", terr.contentType)
		w.WriteHeader(terr.statusCode)
		_, _ = w.Write(terr.body)
	default:
		// NB(mmihic): Assumes this error is logged at a layer above
		w.Header().Set("Content-Type", ContentTypeTextPlain)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}
}

// Error returns an error with a status code, content type, and body.
func Error(statusCode int, contentType string, body []byte) error {
	return httpError{
		statusCode:  statusCode,
		contentType: contentType,
		body:        body,
	}
}

// Errorf returns an error with a status and message.
func Errorf(statusCode int, msg string, args ...interface{}) error {
	return Error(statusCode, ContentTypeTextPlain, []byte(fmt.Sprintf(msg, args...)))
}

type httpError struct {
	statusCode  int
	contentType string
	body        []byte
}

func (e httpError) Error() string {
	return string(e.body)
}

// JSONError is an error type that is returned as a JSON response.
func JSONError(statusCode int, body interface{}) error {
	msg, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unable to marshal error body: %w", err)
	}

	return Error(statusCode, ContentTypeApplicationJSON, msg)
}

// JSONErrorf returns a JSON error with just a message.
func JSONErrorf(statusCode int, msg string, args ...interface{}) error {
	return JSONError(statusCode, struct {
		Message string `json:"message"`
	}{fmt.Sprintf(msg, args...)})
}

// JSONInternalServerError returns a generic "internal server error" response,
// for errors which would not leak.
var JSONInternalServerError = JSONErrorf(http.StatusInternalServerError, "internal server error")
