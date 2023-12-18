package oauth2x

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func newTestTokenSource(tokName string, clock clockwork.Clock) oauth2.TokenSource {
	// Create a token source that tracks how often it's been called
	// and returns a test token with the count.
	numCalls := 0
	return TokenSourceFunc(func() (*oauth2.Token, error) {
		// Record the number of times we've been called, and use this as
		// the prefix for the token
		numCalls++
		tokPrefix := fmt.Sprintf("%s-%d", tokName, numCalls)
		return newTestToken(tokPrefix, "bearer", clock), nil
	})
}

func TestCachingTokenSource_LoadErrors(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	clock := clockwork.NewFakeClockAt(time.Now())
	tokenSource := newTestTokenSource("my-token", clock)

	// Create a cache that fails on load
	c := &fakeTokenCache{name: "my-cache", loadErr: errors.New("this failed")}
	ts1 := CachingTokenSource(c, tokenSource, log, clock)

	// Retrieve the initial token, will fetch a new token since
	// the load from the cache will have failed.
	tok, err := ts1.Token()
	require.NoError(t, err)
	require.Equal(t, 1, c.numLoadAttempts) // we tried the cache
	require.Equal(t, 0, c.numLoadSuccess)  // was not in the cache
	require.Equal(t, 1, c.numStores)       // we stored it back into the cache
	require.Equal(t, "my-token-1-refresh", tok.RefreshToken)
}

func TestCachingTokenSource_StoreErrors(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	clock := clockwork.NewFakeClockAt(time.Now())
	tokenSource := newTestTokenSource("my-token", clock)

	// Create a cache that fails on store
	c := &fakeTokenCache{name: "my-cache", storeErr: errors.New("this failed")}
	ts1 := CachingTokenSource(c, tokenSource, log, clock)

	// Retrieve the initial token, the storage should fail but
	// there should be no issue using the returned token (it
	// just won't be stored in the cached)
	tok, err := ts1.Token()
	require.NoError(t, err)
	require.Equal(t, 1, c.numLoadAttempts) // we tried the cache
	require.Equal(t, 0, c.numLoadSuccess)  // was not in the cache
	require.Equal(t, 0, c.numStores)       // storage to the cache will have failed
	require.Equal(t, "my-token-1-refresh", tok.RefreshToken)
}

func TestCachingTokenSource(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	clock := clockwork.NewFakeClockAt(time.Now())
	tokenSource := newTestTokenSource("my-token", clock)

	c := &fakeTokenCache{name: "my-cache"}

	// Create a new caching token source
	ts1 := CachingTokenSource(c, tokenSource, log, clock)
	require.NoError(t, err)

	// Retrieve the initial token, should go to the original token source,
	// since there is nothing in the cache
	tok, err := ts1.Token()
	require.NoError(t, err)
	require.Equal(t, 1, c.numLoadAttempts) // we tried the cache
	require.Equal(t, 0, c.numLoadSuccess)  // was not in the cache
	require.Equal(t, 1, c.numStores)       // we stored it back into the cache
	require.Equal(t, "my-token-1-refresh", tok.RefreshToken)

	// Retrieving the token again should just fetch it from the in-memory cache.
	tok, err = ts1.Token()
	require.NoError(t, err)
	require.Equal(t, 1, c.numLoadAttempts) // we did not try the cache again
	require.Equal(t, 1, c.numStores)       // we did not write back to the cache
	require.Equal(t, "my-token-1-refresh", tok.RefreshToken)

	// Create a second token source using the cache, retrieving the token
	// should fetch it from the cache.
	ts2 := CachingTokenSource(c, tokenSource, log, clock)
	tok, err = ts2.Token()
	require.NoError(t, err)
	require.Equal(t, 2, c.numLoadAttempts) // we tried the cache again
	require.Equal(t, 1, c.numLoadSuccess)  // in this case it was in the cache
	require.Equal(t, 1, c.numStores)       // we did not write back to the cache
	require.Equal(t, "my-token-1-refresh", tok.RefreshToken)

	// Expire the token, should result in a call to the original token source
	// and a re-save of the token in the cache.
	clock.Advance(time.Hour * 2)

	tok, err = ts1.Token()
	require.NoError(t, err)
	require.Equal(t, 2, c.numLoadAttempts)                   // we did not try the cache again
	require.Equal(t, 2, c.numStores)                         // but we did write back to the cache
	require.Equal(t, "my-token-2-refresh", tok.RefreshToken) // and we obtained a new token from the source

	// Obtain the token again, should return the newly cached token.
	tok, err = ts1.Token()
	require.NoError(t, err)
	require.Equal(t, 2, c.numLoadAttempts)                   // we did not try the cache again
	require.Equal(t, 2, c.numStores)                         // we did not write back to the cache
	require.Equal(t, "my-token-2-refresh", tok.RefreshToken) // we used the same token
}

type fakeTokenCache struct {
	contents        []byte
	name            string
	numStores       int
	numLoadSuccess  int
	numLoadAttempts int
	loadErr         error
	storeErr        error
}

func (c *fakeTokenCache) Load() ([]byte, error) {
	c.numLoadAttempts++
	if c.loadErr != nil {
		return nil, c.loadErr
	}

	if len(c.contents) == 0 {
		return nil, os.ErrNotExist
	}

	c.numLoadSuccess++
	return c.contents, nil
}

func (c *fakeTokenCache) Store(contents []byte) error {
	if c.storeErr != nil {
		return c.storeErr
	}

	c.numStores++
	c.contents = contents
	return nil
}

func (c *fakeTokenCache) Name() string {
	return c.name
}

func newTestToken(tok, tokenType string, clock clockwork.Clock) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  tok + "-access",
		RefreshToken: tok + "-refresh",
		TokenType:    tokenType,
		Expiry:       clock.Now().Add(time.Minute * 30),
	}
}
