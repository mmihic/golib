// Package oauth2x contains utilities for building OAuth 2 clients and servers.
package oauth2x

import "golang.org/x/oauth2"

// TokenSourceFunc is a function that returns a token and can
// act as an oauth2.TokenSource.
type TokenSourceFunc func() (*oauth2.Token, error)

// Token returns the token by calling the function.
func (fn TokenSourceFunc) Token() (*oauth2.Token, error) {
	return fn()
}
