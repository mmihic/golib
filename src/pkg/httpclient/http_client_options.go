package httpclient

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

// CallOption is an option to modify a request.
type CallOption func(r *http.Request) error

// WithQueryParams adds query parameters to the request.
func WithQueryParams(k string, v ...string) CallOption {
	all := append([]string{k}, v...)
	return func(r *http.Request) error {
		if (len(all) % 2) != 0 {
			return errors.Errorf("mismatched query params; %d is not an even number", len(all))
		}

		qv := r.URL.Query()

		for i := 0; i < len(all); i += 2 {
			k, v := all[i], all[i+1]
			qv.Add(k, v)
		}

		r.URL.RawQuery = qv.Encode()

		return nil
	}
}

// SetHeader adds a header to a request.
func SetHeader(key, val string) CallOption {
	return func(r *http.Request) error {
		r.Header.Set(key, val)
		return nil
	}
}

// SetBearerToken sets a bearer token on a request.
func SetBearerToken(token string) CallOption {
	return SetHeader("Authorization", fmt.Sprintf("Bearer %s", token))
}
