package httpx

import (
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
)

type muxEntry struct {
	path    string
	pattern string
	handler http.Handler
}

// A SimpleRouter is a variant of ServeMux that handles method
// based routing in addition to pattern based routing.
type SimpleRouter struct {
	lock     sync.RWMutex
	exact    map[string]muxEntry
	subtrees []muxEntry
}

// NewSimpleRouter creates a new simple router.
func NewSimpleRouter() *SimpleRouter {
	return &SimpleRouter{
		exact:    make(map[string]muxEntry, 10),
		subtrees: make([]muxEntry, 0, 10),
	}
}

// ServeHTTP routes the incoming request to the appropriate handler.
func (r *SimpleRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if h, _ := r.Handler(req); h != nil {
		h.ServeHTTP(w, req)
		return
	}

	http.NotFound(w, req)
}

// Handle registers a handler for a given (optional) method and path.
// Paths can either be fixed rooted paths or subtrees.
func (r *SimpleRouter) Handle(method, path string, h http.Handler) {
	method = strings.ToUpper(method)
	path = cleanPath(path)

	r.lock.Lock()
	defer r.lock.Unlock()

	key := path
	if len(method) != 0 {
		key = method + " " + path
	}

	e := muxEntry{
		path:    path, // used for sorting, so that we match on the longest path regardless of method
		pattern: key,
		handler: h,
	}

	r.exact[key] = e
	if isSubtree(path) {
		r.subtrees = append(r.subtrees, e)
		sort.Slice(r.subtrees, func(i, j int) bool {
			// Longer entries come first
			return r.subtrees[i].path > r.subtrees[j].path
		})
	}
}

// HandleFunc registers a functional handler for a given (optional) method and path.
func (r *SimpleRouter) HandleFunc(method, path string, f func(_ http.ResponseWriter, _ *http.Request)) {
	r.Handle(method, path, http.HandlerFunc(f))
}

// Handler returns the handler and the pattern for the given request.
func (r *SimpleRouter) Handler(req *http.Request) (http.Handler, string) {
	p := cleanPath(req.URL.Path)

	// If the incoming request is /tree and there is a mapping for /tree/ but not /tree,
	// redirect to /tree/
	if u, ok := r.redirectToSubtree(req.Method, p, req.URL); ok {
		return http.RedirectHandler(u.String(), http.StatusMovedPermanently), u.Path
	}

	// Otherwise if the cleaned path doesn't match the incoming path, redirect to the clean path
	if p != req.URL.Path {
		_, pattern := r.handler(req.Method, p)
		u := &url.URL{Path: p, RawQuery: req.URL.RawQuery}
		return http.RedirectHandler(u.String(), http.StatusMovedPermanently), pattern
	}

	return r.handler(req.Method, p)
}

func (r *SimpleRouter) redirectToSubtree(method, path string, u *url.URL) (*url.URL, bool) {
	if isSubtree(path) {
		return nil, false
	}

	r.lock.RLock()
	defer r.lock.RUnlock()

	// If there is an exact match, no need to redirect
	exactKeys := patternsToMatch(method, path)
	for _, k := range exactKeys {
		if _, ok := r.exact[k]; ok {
			return nil, false
		}
	}

	// If we are just doing method matching, or if we are already matching
	// on a full path, no need to redirect
	if len(path) == 0 || path[len(path)-1] == '/' {
		return nil, false
	}

	// Otherwise there is a subtree match but no exact match, redirect to the subtree
	subtreeKeys := patternsToMatch(method, path+"/")
	for _, k := range subtreeKeys {
		if _, ok := r.exact[k]; ok {
			return &url.URL{Path: path + "/", RawQuery: u.RawQuery}, true
		}
	}

	// No matches at all
	return nil, false
}

func (r *SimpleRouter) handler(method, path string) (http.Handler, string) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for _, key := range patternsToMatch(method, path) {
		if h, pattern := r.match(key); h != nil {
			return h, pattern
		}
	}

	return nil, ""
}

func (r *SimpleRouter) match(key string) (http.Handler, string) {
	// Check for an exact match first
	if exact, ok := r.exact[key]; ok {
		return exact.handler, exact.pattern
	}

	// Find the longest subtree with a matching method handler.
	for _, subtree := range r.subtrees {
		if strings.HasPrefix(key, subtree.pattern) {
			return subtree.handler, subtree.pattern
		}
	}

	return nil, ""
}

func isSubtree(path string) bool {
	return path[len(path)-1] == '/'
}

func patternsToMatch(method, path string) []string {
	return []string{
		method + " " + path,
		path,
	}
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	np := path.Clean(p)

	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}
	return np
}
