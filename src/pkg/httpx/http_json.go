package httpx

import (
	"encoding/json"
	"net/http"
)

// Common content types.
const (
	ContentTypeApplicationJSON = "application/json"
	ContentTypeTextPlain       = "text/plain"
)

// RespondWithJSONStatus writes the given value as a JSON response with the specific response code.
func RespondWithJSONStatus(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// TODO(mmihic): Should log this somewhere
	}
}

// RespondWithJSON writes the given value as a JSON response.
func RespondWithJSON(w http.ResponseWriter, v interface{}) {
	RespondWithJSONStatus(w, 200, v)
}

// ReadJSONBody reads the JSON body from a request message.
func ReadJSONBody(r *http.Request, v interface{}) error {
	defer func() { _ = r.Body.Close() }()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return Errorf(http.StatusBadRequest, "unable to parse request body: %s", err)
	}

	return nil
}
