package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespondWithError(t *testing.T) {
	for _, tt := range []struct {
		name     string
		err      error
		expected response
	}{
		{"Errorf",
			Errorf(http.StatusBadRequest, "this is a bad request: %d", 100),
			response{
				code:        http.StatusBadRequest,
				contentType: ContentTypeTextPlain,
				body:        "this is a bad request: 100",
			},
		},
		{"JSONError",
			JSONError(http.StatusBadRequest, struct {
				InternalCode    string `json:"internal_code"`
				FriendlyMessage string `json:"friendly_message"`
			}{
				InternalCode:    "A67CD",
				FriendlyMessage: "something bad happened",
			}),
			response{
				code:        http.StatusBadRequest,
				contentType: ContentTypeApplicationJSON,
				body:        `{"internal_code":"A67CD","friendly_message":"something bad happened"}`,
			},
		},
		{"JSONErrorf",
			JSONErrorf(http.StatusBadRequest, "this is a bad request: %d", 100),
			response{
				code:        http.StatusBadRequest,
				contentType: ContentTypeApplicationJSON,
				body:        `{"message":"this is a bad request: 100"}`,
			},
		},
		{"JSONInternalServerError",
			JSONInternalServerError,
			response{
				code:        http.StatusInternalServerError,
				contentType: ContentTypeApplicationJSON,
				body:        `{"message":"internal server error"}`,
			},
		},
		{"unknown error",
			errors.New("this gets eaten"),
			response{
				code:        http.StatusInternalServerError,
				contentType: ContentTypeTextPlain,
				body:        "internal server error",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			RespondWithError(rr, tt.err)
		})
	}
}

type response struct {
	code        int
	contentType string
	body        string
}

func (r *response) AssertMatches(t *testing.T, rr *httptest.ResponseRecorder) bool {
	if !assert.Equal(t, r.code, rr.Code) {
		return false
	}

	if !assert.Equal(t, r.contentType, rr.Header().Get("Content-Type")) {
		return false
	}

	if !assert.Equal(t, r.body, rr.Body.String()) {
		return false
	}

	return true
}
