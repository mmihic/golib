package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mmihic/golib/src/pkg/httptestx"
	"github.com/mmihic/golib/src/pkg/httpx"
)

func TestWithQueryParams(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := r.URL.Query().Get("my-param")
		if !httptestx.ServerCheck(w, assert.Equal(t, "my-value", s)) {
			return
		}

		s2 := r.URL.Query().Get("my-other-param")
		if !httptestx.ServerCheck(w, assert.Equal(t, "my-other-value", s2)) {
			return
		}

		httpx.RespondWithJSON(w, struct{}{})
	})

	svr := httptest.NewServer(h)
	defer svr.Close()

	c, err := NewClient(svr.URL)
	require.NoError(t, err)

	var nothing struct{}
	err = c.Get(context.Background(), "", JSON(&nothing),
		WithQueryParams(
			"my-param", "my-value",
			"my-other-param", "my-other-value"))
	require.NoError(t, err)
}

func TestSetBearerToken(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !httptestx.ServerCheck(w, assert.Equal(t, "Bearer A_BEARER_TOKEN", authHeader)) {
			return
		}

		httpx.RespondWithJSON(w, struct{}{})
	})

	svr := httptest.NewServer(h)
	defer svr.Close()

	c, err := NewClient(svr.URL)
	require.NoError(t, err)

	var nothing struct{}
	err = c.Get(context.Background(), "", JSON(&nothing),
		SetBearerToken("A_BEARER_TOKEN"))
	require.NoError(t, err)
}

func TestSetHeader(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := r.Header.Get("my-header")
		if !httptestx.ServerCheck(w, assert.Equal(t, "my-value", s)) {
			return
		}

		s2 := r.Header.Get("my-other-header")
		if !httptestx.ServerCheck(w, assert.Equal(t, "my-other-value", s2)) {
			return
		}

		httpx.RespondWithJSON(w, struct{}{})
	})

	svr := httptest.NewServer(h)
	defer svr.Close()

	c, err := NewClient(svr.URL)
	require.NoError(t, err)

	var nothing struct{}
	err = c.Get(context.Background(), "", JSON(&nothing),
		SetHeader("my-header", "my-value"),
		SetHeader("my-other-header", "my-other-value"))
	require.NoError(t, err)
}
