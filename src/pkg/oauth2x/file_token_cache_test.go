package oauth2x

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFileTokenCache(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	defer func() { _ = os.RemoveAll(dir) }()

	log, err := zap.NewDevelopment()
	require.NoError(t, err)

	clock := clockwork.NewFakeClockAt(time.Now())
	tokenSource := newTestTokenSource("my-token", clock)

	c := NewFileTokenCache(filepath.Join(dir, "cache.json"))

	// Create a new caching token source
	ts1 := CachingTokenSource(c, tokenSource, log, clock)
	require.NoError(t, err)

	// Retrieve the initial token, should go to the original token source,
	// since there is nothing in the cache
	tok, err := ts1.Token()
	require.NoError(t, err)
	require.Equal(t, "my-token-1-refresh", tok.RefreshToken)

	// Retrieving the token again should just fetch it from the in-memory cache.
	tok, err = ts1.Token()
	require.NoError(t, err)
	require.Equal(t, "my-token-1-refresh", tok.RefreshToken)

	// Create a second token source using the cache, retrieving the token
	// should fetch it from the cache.
	ts2 := CachingTokenSource(c, tokenSource, log, clock)
	tok, err = ts2.Token()
	require.NoError(t, err)
	require.Equal(t, "my-token-1-refresh", tok.RefreshToken)

	// Expire the token, should result in a call to the original token source
	// and a re-save of the token in the cache.
	clock.Advance(time.Hour * 2)

	tok, err = ts1.Token()
	require.NoError(t, err)
	require.Equal(t, "my-token-2-refresh", tok.RefreshToken) // and we obtained a new token from the source

	// Obtain the token again, should return the newly cached token.
	tok, err = ts1.Token()
	require.NoError(t, err)
	require.Equal(t, "my-token-2-refresh", tok.RefreshToken) // we used the same token
}
