package httpx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSimpleRouter_ExactMatchesPrefersMethod(t *testing.T) {
	// Given handlers for
	//		GET /foo.gif
	//		POST /foo.gif
	//		/foo.gif without a method
	r := NewSimpleRouter()
	registerTestSimpleRouterHandler(r, "GET", "/foo.gif")
	registerTestSimpleRouterHandler(r, "POST", "/foo.gif")
	registerTestSimpleRouterHandler(r, "", "/foo.gif")

	// GET 	/foo.gif => handler for GET /foo
	// POST /foo.gif => handler for POST /foo
	// PUT 	/foo.gif => handler for /foo without a method
	assertRouting(t, r, "GET", "/foo.gif", "GET", "/foo.gif")
	assertRouting(t, r, "POST", "/foo.gif", "POST", "/foo.gif")
	assertRouting(t, r, "PUT", "/foo.gif", "", "/foo.gif")
}

func TestSimpleRouter_SubtreeMatchesPrefersMethod(t *testing.T) {
	// Given handlers for
	//		GET /foo/
	//		POST /foo/
	//		/foo/ without a method
	r := NewSimpleRouter()
	registerTestSimpleRouterHandler(r, "GET", "/foo/")
	registerTestSimpleRouterHandler(r, "POST", "/foo/")
	registerTestSimpleRouterHandler(r, "", "/foo/")

	// GET 	/foo/a/b/c/zed.gif => handler for GET /foo
	// POST /foo/a/b/c/zed.gif => handler for POST /foo
	// PUT 	/foo/a/b/c/zed.gif => handler for /foo without a method
	assertRouting(t, r, "GET", "/foo/a/b/c/zed.gif", "GET", "/foo/")
	assertRouting(t, r, "GET", "/foo/zed.gif", "GET", "/foo/")
	assertRouting(t, r, "POST", "/foo/a/b/c/zed.gif", "POST", "/foo/")
	assertRouting(t, r, "POST", "/foo/zed.gif", "POST", "/foo/")
	assertRouting(t, r, "PUT", "/foo/a/b/c/zed.gif", "", "/foo/")
	assertRouting(t, r, "PUT", "/foo/zed.gif", "", "/foo/")

	// Should not match to elements with the same name as the path
	assertNoRoute(t, r, "GET", "/foo.gif")
	assertNoRoute(t, r, "PUT", "/foo.gif")
}

func TestSimpleRouter_SubtreeMatchesPrefersLongerPath(t *testing.T) {
	// Given handlers for
	//    GET /foo/a/b/
	//    GET /foo/a/
	//    POST /foo/a
	//    /foo/a/b/ without a method
	r := NewSimpleRouter()
	registerTestSimpleRouterHandler(r, "GET", "/foo/a/b/")
	registerTestSimpleRouterHandler(r, "GET", "/foo/a/")
	registerTestSimpleRouterHandler(r, "POST", "/foo/a/")
	registerTestSimpleRouterHandler(r, "", "/foo/a/b/")

	// GET 	/foo/a/b/c/zed.gif 	=> handler for GET /foo/a/b/
	// GET  /foo/a/zed.gif 		=> handler for GET /foo/a/
	// POST /foo/a/b/c/zed.gif 	=> handler for POST /foo/a/
	// PUT 	/foo/a/b/c/zed.gif 	=> handler for /foo/a/b/ without a method
	assertRouting(t, r, "GET", "/foo/a/b/c/zed.gif", "GET", "/foo/a/b/")
	assertRouting(t, r, "GET", "/foo/a/zed.gif", "GET", "/foo/a/")
	assertRouting(t, r, "POST", "/foo/a/b/c/zed.gif", "POST", "/foo/a/")
	assertRouting(t, r, "PUT", "/foo/a/b/c/zed.gif", "", "/foo/a/b/")
}

func TestSimpleRouter_PrefersExactOverSubtree(t *testing.T) {
	// Given handlers for
	// 		GET /foo/bar/zed.gif
	//		/foo/bar/zed.gif without a method
	//		GET /foo/bar/
	//		POST /foo/bar/
	//		/foo/bar/ without a method
	r := NewSimpleRouter()
	registerTestSimpleRouterHandler(r, "GET", "/foo/bar/zed.gif")
	registerTestSimpleRouterHandler(r, "", "/foo/bar/zed.gif")
	registerTestSimpleRouterHandler(r, "GET", "/foo/bar/")
	registerTestSimpleRouterHandler(r, "POST", "/foo/bar/")
	registerTestSimpleRouterHandler(r, "", "/foo/bar/")

	// GET 	/foo/bar/zed.gif	=> handler for GET /foo/bar/zed.gif
	// POST /foo/bar/zed.gif	=> handler for POST /foo/bar/
	// POST /foo/bar/what.gif	=> handler for POST /foo/bar
	// PUT 	/foo/bar/zed.gif	=> handler for /foo/bar/zed.gif without a method
	// PUT 	/foo/bar/what.gif	=> handler for /foo/bar/ without a method
	assertRouting(t, r, "GET", "/foo/bar/zed.gif", "GET", "/foo/bar/zed.gif")
	assertRouting(t, r, "POST", "/foo/bar/zed.gif", "POST", "/foo/bar/")
	assertRouting(t, r, "POST", "/foo/bar/what.gif", "POST", "/foo/bar/")
	assertRouting(t, r, "PUT", "/foo/bar/zed.gif", "", "/foo/bar/zed.gif")
	assertRouting(t, r, "PUT", "/foo/bar/what.gif", "", "/foo/bar/")
}

func TestSimpleRouter_RedirectsToFolderIfExactMatchForFolderExists(t *testing.T) {
	// Given a handler for
	//		GET /foo/bar/
	//		GET /foo/bar/zed/
	//		GET /foo/chort/something
	//		/foo/chort/something/
	r := NewSimpleRouter()
	registerTestSimpleRouterHandler(r, "GET", "/foo/bar/")
	registerTestSimpleRouterHandler(r, "GET", "/foo/bar/zed/")
	registerTestSimpleRouterHandler(r, "GET", "/foo/chort/something")
	registerTestSimpleRouterHandler(r, "", "/foo/chort/something/")

	//	GET /foo/bar 					=> redirect to /foo/bar/				(exact match for folder)
	//  GET /foo/chort/something		=> handler for /foo/chort/something 	(exact match)
	//  PUT /foo/chort/something		=> redirect to /foo/chort/something/	(exact match for folder but not item)
	//  GET /foo/bar/zed				=> redirect to /foo/bar/zed/			(exact match for folder)
	//  GET /foo/bar/zed/whatever.gif	=> handler for /foo/bar/zed/			(no exact subtree match)
	assertRedirect(t, r, "GET", "/foo/bar", "/foo/bar/")
	assertRouting(t, r, "GET", "/foo/chort/something", "GET", "/foo/chort/something")
	assertRedirect(t, r, "PUT", "/foo/chort/something", "/foo/chort/something/")
	assertRedirect(t, r, "GET", "/foo/bar/zed", "/foo/bar/zed/")
	assertRouting(t, r, "GET", "/foo/bar/zed/whatever.gif", "GET", "/foo/bar/zed/")
}

func TestSimpleRouter_RedirectsOnUncleanPaths(t *testing.T) {
	r := NewSimpleRouter()
	registerTestSimpleRouterHandler(r, "GET", "/foo/bar/")

	assertRedirect(t, r, "GET", "/foo/bar/zed/./../blerp..gif", "/foo/bar/blerp..gif")
}

func TestSimpleRouter_HandlesRootPaths(t *testing.T) {
	// GET foo.gif 		becomes GET /foo.gif
	// GET foo/bar/zed/ becomes GET /foo/bar/zed
	// GET 				becomes GET /
	r := NewSimpleRouter()
	registerTestSimpleRouterHandler(r, "GET", "foo.gif")
	registerTestSimpleRouterHandler(r, "GET", "foo/bar/zed/")
	registerTestSimpleRouterHandler(r, "GET", "")

	assertRouting(t, r, "GET", "/foo.gif", "GET", "foo.gif")
	assertRouting(t, r, "GET", "/foo/bar/zed/mork.gif", "GET", "foo/bar/zed/")
	assertRouting(t, r, "GET", "/foo/mork/zed.gif", "GET", "")
	assertRouting(t, r, "GET", "/zed.gif", "GET", "")
	assertRouting(t, r, "GET", "/", "GET", "")
}

func registerTestSimpleRouterHandler(r *SimpleRouter, method, path string) {
	r.HandleFunc(method, path, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "%s%s", method, path)
	})
}

func assertRouting(t *testing.T, r http.Handler, method, path string, expectedMethod, expectedPath string) bool {
	req := httptest.NewRequest(method, path, nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)
	if !assert.Equal(t, http.StatusOK, resp.Code) {
		return false
	}

	return assert.Equal(t, fmt.Sprintf("%s%s", expectedMethod, expectedPath), resp.Body.String())
}

func assertNoRoute(t *testing.T, r http.Handler, method, path string) bool {
	req := httptest.NewRequest(method, path, nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)
	return assert.Equal(t, http.StatusNotFound, resp.Code)
}

func assertRedirect(t *testing.T, r http.Handler, method, path, expectedPath string) bool {
	req := httptest.NewRequest(method, path+"?q=something", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)
	if !assert.Equal(t, http.StatusMovedPermanently, resp.Code) {
		return false
	}

	return assert.Equal(t, expectedPath+"?q=something", resp.Header().Get("Location"))
}
