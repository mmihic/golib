package oauth2x

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// A TokenCache is a cache for local tokens.
type TokenCache interface {
	Load() ([]byte, error)
	Store([]byte) error
	Name() string
}

// A CachingTokenSource retrieves tokens from a cache, and if that
// fails triggers consults another token source.
func CachingTokenSource(
	cache TokenCache, fallback oauth2.TokenSource,
	log *zap.Logger, clock clockwork.Clock,
) oauth2.TokenSource {
	if clock == nil {
		clock = clockwork.NewRealClock()
	}

	// Attempt to load from the cache. Cache failures
	// won't prevent creating the token source, it'll just
	// trigger a new fetch on first access.
	tok, err := readTokenFromCache(cache)
	if err != nil {
		log.Warn("unable to load token from cache",
			zap.String("cacheName", cache.Name()),
			zap.Error(err))
		tok = nil
	}

	return &cachingTokenSource{
		cache:    cache,
		fallback: fallback,
		log:      log,
		clock:    clock,
		token:    tok,
	}
}

func (src *cachingTokenSource) Token() (*oauth2.Token, error) {
	if token := src.checkTokenValid(); token != nil {
		return token, nil
	}

	src.log.Info("token invalid, fetching new token")

	// Fetch a new token and save it off
	token, err := src.fallback.Token()
	if err != nil {
		return nil, err
	}

	src.mut.Lock()
	src.token = token
	src.mut.Unlock()

	// Write to the cache, but don't treat it like an error
	if err := writeTokenToCache(src.cache, token); err != nil {
		src.log.Error("unable to write token to cache",
			zap.String("cache_name", src.cache.Name()),
			zap.Error(err))
	}

	return token, nil
}

func (src *cachingTokenSource) checkTokenValid() *oauth2.Token {
	src.mut.RLock()
	defer src.mut.RUnlock()

	if src.token == nil ||
		src.token.AccessToken == "" ||
		src.clock.Now().After(src.token.Expiry) {
		return nil
	}

	return src.token
}

func readTokenFromCache(cache TokenCache) (*oauth2.Token, error) {
	var tok oauth2.Token
	tokenBytes, err := cache.Load()
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	if err := json.Unmarshal(tokenBytes, &tok); err != nil {
		return nil, err
	}

	return &tok, nil
}

func writeTokenToCache(cache TokenCache, tok *oauth2.Token) error {
	tokenBytes, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}

	return cache.Store(tokenBytes)
}

type cachingTokenSource struct {
	mut      sync.RWMutex
	cache    TokenCache
	token    *oauth2.Token
	log      *zap.Logger
	fallback oauth2.TokenSource
	clock    clockwork.Clock
}
