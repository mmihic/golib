package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONHandler(t *testing.T) {
	type SimpleStruct struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}

	for _, tt := range []struct {
		name             string
		f                interface{}
		request          *http.Request
		expectedResponse response
	}{
		{"no inputs, error only output",
			func(_ context.Context) error {
				return JSONErrorf(http.StatusBadRequest, "this failed")
			},
			newTestJSONRequest(t, "POST", "/whatever", nil),
			response{
				code:        http.StatusBadRequest,
				contentType: ContentTypeApplicationJSON,
				body:        `{"message":"this failed"}`,
			},
		},
		{"no inputs, struct and error output, returns error",
			func(_ context.Context) (*SimpleStruct, error) {
				return nil, JSONErrorf(http.StatusBadRequest, "this failed")
			},
			newTestJSONRequest(t, "POST", "/whatever", nil),
			response{
				code:        http.StatusBadRequest,
				contentType: ContentTypeApplicationJSON,
				body:        `{"message":"this failed"}`,
			},
		},
		{"no inputs, struct and error output, returns struct",
			func(_ context.Context) (*SimpleStruct, error) {
				return &SimpleStruct{
					Name:  "Home",
					Value: "Over There",
				}, nil
			},
			newTestJSONRequest(t, "POST", "/whatever", nil),
			response{
				code:        http.StatusOK,
				contentType: ContentTypeApplicationJSON,
				body: `{"name":"Home","value":"Over There"}
`,
			},
		},
		{"request only input",
			func(_ context.Context, r *http.Request) (*SimpleStruct, error) {
				return &SimpleStruct{
					Name:  r.Method,
					Value: r.RequestURI,
				}, nil
			},
			newTestJSONRequest(t, "GET", "/whatever", nil),
			response{
				code:        http.StatusOK,
				contentType: ContentTypeApplicationJSON,
				body: `{"name":"GET","value":"/whatever"}
`,
			},
		},
		{"struct only input",
			func(_ context.Context, in *SimpleStruct) (*SimpleStruct, error) {
				return &SimpleStruct{
					Name:  in.Name + " Response",
					Value: in.Value + " Response",
				}, nil
			},
			newTestJSONRequest(t, "GET", "/whatever", &SimpleStruct{
				Name:  "Home",
				Value: "Over There",
			}),
			response{
				code:        http.StatusOK,
				contentType: ContentTypeApplicationJSON,
				body: `{"name":"Home Response","value":"Over There Response"}
`,
			},
		},
		{"request and struct input",
			func(_ context.Context, r *http.Request, in *SimpleStruct) (*SimpleStruct, error) {
				return &SimpleStruct{
					Name:  r.Method + " " + in.Name,
					Value: r.RequestURI + " " + in.Value,
				}, nil
			},
			newTestJSONRequest(t, "GET", "/whatever", &SimpleStruct{
				Name:  "Home",
				Value: "Over There",
			}),
			response{
				code:        http.StatusOK,
				contentType: ContentTypeApplicationJSON,
				body: `{"name":"GET Home","value":"/whatever Over There"}
`,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h, err := JSONHandler(tt.f)
			if !assert.NoError(t, err) {
				return
			}

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, tt.request)
			tt.expectedResponse.AssertMatches(t, rr)
		})
	}
}

func TestInvalidJSONHandlers(t *testing.T) {
	for _, tt := range []struct {
		name   string
		method interface{}
		msg    string
	}{
		{"no inputs at all",
			func() error { return nil },
			"invalid signature for func() error: must take 1-3 inputs",
		},
		{"one input, no context",
			func(_ *http.Request) error { return nil },
			"invalid signature for func(*http.Request) error: first input must be a context.Context",
		},
		{"multiple inputs, context is not the first input",
			func(_ *http.Request, _ context.Context) error { return nil },
			"invalid signature for func(*http.Request, context.Context) error: first input must be a context.Context",
		},
		{"one input, not struct or request",
			func(_ context.Context, _ int) error { return nil },
			"invalid signature func(context.Context, int) error: input must be a *http.Request or a struct ptr",
		},
		{"too many inputs",
			func(_ context.Context, _ *http.Request, _ *struct{}, _ *struct{}) error { return nil },
			"invalid signature for func(context.Context, *http.Request, *struct {}, *struct {}) error: must take 1-3 inputs",
		},
		{"two inputs, first is not request",
			func(_ context.Context, _ *struct{}, _ *http.Request) error { return nil },
			"invalid signature func(context.Context, *struct {}, *http.Request) error: first input must be a *http.Request",
		},
		{"two inputs, second is not struct",
			func(_ context.Context, _ *http.Request, _ int) error { return nil },
			"invalid signature func(context.Context, *http.Request, int) error: second input must be a struct ptr",
		},
		{"invalid signature func(context.Context, *http.Request, struct {}) error: second input must be a struct ptr",
			func(_ context.Context, _ *http.Request, _ struct{}) error { return nil },
			"invalid signature func(context.Context, *http.Request, struct {}) error: second input must be a struct ptr",
		},
		{
			"no outputs",
			func(_ context.Context) {},
			"invalid signature for func(context.Context); must return at least return an error",
		},
		{
			"one output, not an error",
			func(_ context.Context) *struct{} { return nil },
			"invalid signature for func(context.Context) *struct {}; last output must be an error",
		},
		{
			"two outputs, last is not an error",
			func(_ context.Context) (*struct{}, *struct{}) { return nil, nil },
			"invalid signature for func(context.Context) (*struct {}, *struct {}); last output must be an error",
		},
		{
			"two outputs, first is not a struct",
			func(_ context.Context) (int, error) { return 0, nil },
			"invalid signature for func(context.Context) (int, error); can only return struct ptr for response body",
		},
		{
			"two outputs, first is not a struct ptr",
			func(_ context.Context) (struct{}, error) { return struct{}{}, nil },
			"invalid signature for func(context.Context) (struct {}, error); can only return struct ptr for response body",
		},
		{
			"too many outputs",
			func(_ context.Context) (struct{}, struct{}, error) { return struct{}{}, struct{}{}, nil },
			"invalid signature for func(context.Context) (struct {}, struct {}, error); can only return an output structure and an error",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := JSONHandler(tt.method)
			if !assert.Error(t, err) {
				return
			}

			assert.Equal(t, tt.msg, err.Error())
		})
	}
}

func newTestJSONRequest(t *testing.T, method, target string, v interface{}) *http.Request {
	var b bytes.Buffer
	if v != nil {
		if err := json.NewEncoder(&b).Encode(v); err != nil {
			require.NoError(t, err)
		}
	}

	return httptest.NewRequest(method, target, &b)
}
