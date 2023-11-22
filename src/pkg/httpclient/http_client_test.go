package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mmihic/golib/src/pkg/httpx"
)

var (
	rePath = regexp.MustCompile("/api/json/(?P<ID>.*)")
)

func TestClient_GetPutDelete(t *testing.T) {
	for _, tt := range []struct {
		name     string
		basePath string
		path     string
	}{
		{
			"baseURL is just the hostPort",
			"",
			"/api/json/2",
		},
		{
			"baseURL includes partial path but lacks trailing sep, request path lacks leading sep",
			"/api/json",
			"2",
		},
		{
			"baseURL includes partial path but lacks trailing sep, request path has leading sep",
			"/api/json",
			"/2",
		},
		{
			"baseURL includes partial path and has trailing sep, request path has leading sep",
			"/api/json/",
			"/2",
		},
		{
			"baseURL includes partial path and has trailing sep, request path lacks leading sep",
			"/api/json/",
			"2",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tst := newTester(t)
			defer tst.Close()

			c, err := NewClient(tst.svr.URL + tt.basePath)
			if !assert.NoError(t, err) {
				return
			}

			var stored map[string]string
			err = c.Get(ctx, tt.path, JSON(&stored))
			if !assertHTTPError(t, err, http.StatusNotFound) {
				return
			}

			data := map[string]string{
				"key1": "val1",
				"key2": "val2",
				"joe":  "banana",
			}

			err = c.Put(ctx, tt.path, JSONPretty(data), Empty())
			if !assert.NoError(t, err) {
				return
			}

			err = c.Get(ctx, tt.path, JSON(&stored))
			if !assert.NoError(t, err) {
				return
			}

			if !assert.Equal(t, data, stored) {
				return
			}

			err = c.Delete(ctx, tt.path, Empty(), Empty())
			if !assert.NoError(t, err) {
				return
			}

			err = c.Get(ctx, tt.path, JSON(&stored))
			if !assertHTTPError(t, err, http.StatusNotFound) {
				return
			}
		})
	}
}

type tester struct {
	svr *httptest.Server
}

func (tt *tester) Close() {
	tt.svr.Close()
}

func newTester(_ *testing.T) *tester {
	data := map[int][]byte{}

	h := httpx.NewSimpleRouter()
	h.HandleFunc("GET", "/api/json/", func(w http.ResponseWriter, r *http.Request) {
		id, err := getID(r.URL.Path)
		if err != nil {
			httpx.RespondWithError(w, httpx.Errorf(http.StatusInternalServerError, "invalid id: %s", err))
			return
		}

		value := data[id]
		if value == nil {
			httpx.RespondWithError(w, httpx.Errorf(http.StatusNotFound, "id %d not found", id))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(value)
	})

	h.HandleFunc("PUT", "/api/json/", func(w http.ResponseWriter, r *http.Request) {
		id, err := getID(r.URL.Path)
		if err != nil {
			httpx.RespondWithError(w, httpx.Errorf(http.StatusInternalServerError, "invalid id: %s", err))
			return
		}

		defer func() { _ = r.Body.Close() }()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.RespondWithError(w, httpx.Errorf(http.StatusInternalServerError, "unable to read body:%s", err))
			return
		}

		data[id] = body
		httpx.RespondWithJSON(w, map[string]string{})
	})

	h.HandleFunc("DELETE", "/api/json/", func(w http.ResponseWriter, r *http.Request) {
		id, err := getID(r.URL.Path)
		if err != nil {
			httpx.RespondWithError(w, httpx.Errorf(http.StatusInternalServerError, "invalid id: %s", err))
			return
		}

		_, exists := data[id]
		if !exists {
			httpx.RespondWithError(w, httpx.Errorf(http.StatusNotFound, "id %d not found", id))
			return
		}

		delete(data, id)

		httpx.RespondWithJSON(w, map[string]string{})
	})

	return &tester{
		svr: httptest.NewServer(h),
	}
}

func getID(path string) (int, error) {
	groups := rePath.FindStringSubmatch(path)
	idIdx := rePath.SubexpIndex("ID")
	if len(groups) < idIdx {
		return 0, fmt.Errorf("could not find ID in %s", path)
	}

	id, err := strconv.ParseInt(groups[idIdx], 10, 32)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func assertHTTPError(t *testing.T, err error, statusCode int) bool {
	if !assert.Error(t, err) {
		return false
	}

	httpErr, ok := UnwrapError(err)
	if !assert.True(t, ok, "error is a %s not an HTTP error", reflect.TypeOf(err).Name()) {
		return false
	}

	if !assert.Equal(t, httpErr.StatusCode, statusCode) {
		return false
	}

	return true
}
