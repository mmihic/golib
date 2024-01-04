package cache

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests basic write through
func TestWriteThroughCache(t *testing.T) {
	c := New[string, string](100)

	entries := map[string]string{
		"foo":    "bar",
		"zed":    "banana",
		"snork":  "mork",
		"gambas": "camarones",
	}

	// Add a bunch of entries to the cache asynchronously
	var (
		allWritten   sync.WaitGroup
		allWriting   sync.WaitGroup
		startWriting = make(chan struct{})
	)
	for key, val := range entries {
		key, val := key, val
		allWriting.Add(1)
		allWritten.Add(1)

		go func() {
			allWriting.Done()
			<-startWriting
			c.Put(context.Background(), key, val)
			allWritten.Done()
		}()
	}

	allWriting.Wait()
	close(startWriting)
	allWritten.Wait()

	assert.Equal(t, Statistics{
		CurrentSize: 4,
	}, c.Statistics())

	// Access a bunch of entries, including a non-existent entry. Make
	// sure we get back what we expect (including ErrNotFound for the key
	// that does not match a known entry).
	var (
		allReading   sync.WaitGroup
		allRead      sync.WaitGroup
		startReading = make(chan struct{})

		expectedMisses atomic.Int32
		expectedHits   atomic.Int32
	)

	var keysToRead []string
	for key := range entries {
		keysToRead = append(keysToRead, key)
	}
	keysToRead = append(keysToRead, "non_existent")
	for i := 0; i < 100; i++ {
		allReading.Add(1)
		allRead.Add(1)

		go func() {
			allReading.Done()
			defer allRead.Done()

			key := keysToRead[rand.Int()%len(keysToRead)]
			expectedVal, exists := entries[key]

			<-startReading
			if exists {
				expectedHits.Add(1)
				assertEntryFound(t, c, key, expectedVal)
			} else {
				expectedMisses.Add(1)
				assertEntryNotFound(t, c, key)
			}
		}()
	}

	allReading.Wait()
	close(startReading)
	allRead.Wait()

	// Confirm statistics
	assert.Equal(t, Statistics{
		Hits:        int64(expectedHits.Load()),
		Misses:      int64(expectedMisses.Load()),
		CurrentSize: len(entries),
	}, c.Statistics())
}

// Tests basic reading through the cache
func TestReadThroughCache(t *testing.T) {
	entries := map[string]string{
		"foo":    "bar",
		"zed":    "banana",
		"snork":  "mork",
		"gambas": "camarones",
	}

	const (
		errorKey     = "fail_this_key"
		errorMessage = "failed to load for cache"
	)

	// Create a read-through cache
	c := New[string, string](100,
		WithLoadFn(func(_ context.Context, key string) (string, time.Time, error) {
			if key == errorKey {
				return "", time.Time{}, errors.New(errorMessage)
			}

			if val, ok := entries[key]; ok {
				return val, time.Time{}, nil
			}

			return "", time.Time{}, ErrNotFound
		}))

	// Read from the cache, including reading a key that does not exist and
	// will not be loaded, and a key that will always result in an error
	var keysToRead []string
	for key := range entries {
		keysToRead = append(keysToRead, key)
	}
	keysToRead = append(keysToRead, "non_existent_key")
	keysToRead = append(keysToRead, errorKey)

	var (
		allReading   sync.WaitGroup
		allRead      sync.WaitGroup
		startReading = make(chan struct{})

		expectedLoadAttempts atomic.Int32
		expectedMisses       atomic.Int32
		expectedHits         atomic.Int32
		expectedLoadFailures atomic.Int32
	)

	// We should try to load every key at least once, and will
	// also try to load the error key and the non-existent key
	// each time we access them.
	expectedLoadAttempts.Add(int32(len(entries)))
	expectedHits.Add(-int32(len(entries))) // The initial loads aren't considered a hit

	for i := 0; i < 1000; i++ {
		i := i
		allReading.Add(1)
		allRead.Add(1)

		go func() {
			allReading.Done()
			defer allRead.Done()

			key := keysToRead[i%len(keysToRead)]
			expectedVal, exists := entries[key]

			<-startReading
			if exists {
				expectedHits.Add(1)
				assertEntryFound(t, c, key, expectedVal)
			} else if key == errorKey {
				expectedLoadAttempts.Add(1)
				expectedLoadFailures.Add(1)
				assertEntryError(t, c, key, errorMessage)
			} else {
				expectedLoadAttempts.Add(1)
				expectedMisses.Add(1)
				assertEntryNotFound(t, c, key)
			}
		}()
	}

	allReading.Wait()
	close(startReading)
	allRead.Wait()

	// Confirm statistics
	assert.Equal(t, Statistics{
		LoadFailures: int64(expectedLoadFailures.Load()),
		LoadAttempts: int64(expectedLoadAttempts.Load()),
		Hits:         int64(expectedHits.Load()),
		Misses:       int64(expectedMisses.Load()),
		CurrentSize:  len(entries),
	}, c.Statistics())
}

// Multiple goroutines accessing different keys at the same time, only one
// call to the load function is performed per key, but multiple keys can be
// loaded simultaneously.
func TestAllowConcurrentReadThroughOnDifferentKeysButNotOnSameKey(t *testing.T) {
	var totalLoading atomic.Int32
	keyLoadingCounters := map[string]*atomic.Int32{
		"foo":  {},
		"bar":  {},
		"zed":  {},
		"klue": {},
	}

	// Kick off multiple concurrent loads for each key.
	var (
		finishLoading  = make(chan struct{})
		accessFinished sync.WaitGroup
	)

	c := New[string, string](100,
		WithLoadFn(func(_ context.Context, key string) (string, time.Time, error) {
			keyLoadingCounters[key].Add(1)
			totalLoading.Add(1)
			<-finishLoading
			return key, time.Time{}, nil
		}))

	const goroutinesPerKey = 5

	for key := range keyLoadingCounters {
		key := key
		for i := 0; i < goroutinesPerKey; i++ {
			accessFinished.Add(1)
			go func() {
				val, err := c.Get(context.Background(), key)
				require.NoError(t, err)
				require.Equal(t, key, val)
				accessFinished.Done()
			}()
		}
	}

	// Check the counters - eventually we should reach a state where there
	// is a loader running for every distinct key, but never have more
	// than one loader running for any given key.
	allLoading := false
	timeout := time.Now().Add(time.Second * 30)
	for !allLoading && time.Now().Before(timeout) {
		// Should never have more loaders than there are keys - at most one loader
		// should be loading per key
		loadingNow := int(totalLoading.Load())
		if loadingNow == len(keyLoadingCounters) {
			// If the number of loaders running matches the number of keys, confirm
			// that there is only one loader per key (meaning there we never have
			// concurrent loading going on for any specific key)
			for key, keyCounter := range keyLoadingCounters {
				require.Equalf(t, int(keyCounter.Load()), 1, "multiple loaders for key %s", key)
			}

			allLoading = true
		} else if loadingNow > len(keyLoadingCounters) {
			// There are more loaders running then keys - this should not happen,
			// at most there should be a single loader running per key
			t.Fatalf("there are %d loaders running, more than %d keys",
				loadingNow, len(keyLoadingCounters))
		}
	}

	require.True(t, allLoading, "never reached state where all loaders were running")

	// Release all loaders, allowing them to complete and the
	// accessors to continue running
	close(finishLoading)

	// Wait for all goroutines to finish, confirm that we never called
	// the loader for a given key more than once.
	accessFinished.Wait()
	for key, keyCounter := range keyLoadingCounters {
		require.Equalf(t, int(keyCounter.Load()), 1, "multiple loaders for key %s", key)
	}
}

// Multiple goroutines access the same key at the same time, blocked goroutines
// honor context cancellation and deadlines.
func TestReadThroughHonorsContextCancellation(t *testing.T) {
	var (
		startedLoading = make(chan struct{})
		finishLoading  = make(chan struct{})
	)
	c := New[string, string](100,
		WithLoadFn(func(ctx context.Context, key string) (string, time.Time, error) {
			close(startedLoading)
			<-finishLoading
			return key, time.Time{}, nil
		}))

	// Spin up a goroutine that accesses (and therefore loads) a key
	var allGoRoutinesComplete sync.WaitGroup
	allGoRoutinesComplete.Add(1)
	go func() {
		defer allGoRoutinesComplete.Done()
		key, err := c.Get(context.Background(), "my_key")
		require.NoError(t, err)
		require.Equal(t, key, "my_key")
	}()

	// Wait until we enter the loader
	<-startedLoading

	// Spin up goroutines that access the key again, using a context which we'll
	// cancel. These should all block waiting for the current loader to complete.
	cancelCtx, cancel := context.WithCancel(context.Background())

	var secondGoRoutinesComplete sync.WaitGroup
	for i := 0; i < 5; i++ {
		secondGoRoutinesComplete.Add(1)
		go func() {
			defer secondGoRoutinesComplete.Done()
			_, err := c.Get(cancelCtx, "my_key")
			require.Error(t, err, context.Canceled)
		}()
	}

	// Wait a few seconds so they goroutines all block.
	time.Sleep(time.Second * 5)

	// Cancel the context. All secondary goroutines should error out with
	// a cancellation
	cancel()

	// Wait for all secondary goroutine to complete before releasing
	// the initial goroutine
	secondGoRoutinesComplete.Wait()

	// Release the initial loader, this should complete successfully
	close(finishLoading)

	// Check stats
	allGoRoutinesComplete.Wait()
	assert.Equal(t, Statistics{
		Hits:         0,
		Misses:       0,
		LoadAttempts: 1,
		LoadFailures: 0,
		Expirations:  0,
		Evictions:    0,
		CurrentSize:  1,
	}, c.Statistics())
}

// When using synchronous eviction, the least recently used entry is evicted as
// soon as we reach the max size.
func TestSyncEviction(t *testing.T) {
	entries := map[string]string{
		"foo":       "bar",
		"zed":       "banana",
		"snork":     "mork",
		"gambas":    "camarones",
		"conch":     "snail",
		"ephemeral": "transient",
	}

	const maxSize = 3
	c := New[string, string](maxSize,
		WithLoadFn(func(_ context.Context, key string) (string, time.Time, error) {
			if val, ok := entries[key]; ok {
				return val, time.Time{}, nil
			}

			return "", time.Time{}, ErrNotFound
		}))

	// This is the access pattern
	// snork, zed, foo 		-> 3 loads 		[foo, zed, snork]
	// zed 					-> hit			[zed, foo, snork]
	// gambas				-> evict + load	[gambas, zed, foo]
	// gambas				-> hit			[gambas, zed, foo]
	// foo					-> hit			[foo, gambas, zed]
	// non-existent			-> load + miss	[foo, gambas, zed]
	// conch				-> evict + load	[conch, foo, gambas]
	// gambas				-> hit			[gambas, conch, foo]
	// non-existent			-> load + miss	[foo, gambas, zed]
	// foo					-> hit			[foo, gambas, conch]
	// zed					-> evict + load	[zed, foo, gambas]

	// So in the end we see 8 loads attempts, 3 evictions, 5 hits, 2 misses
	assertEntryFound(t, c, "snork", entries["snork"])
	assertEntryFound(t, c, "zed", entries["zed"])
	assertEntryFound(t, c, "foo", entries["foo"])
	assertEntryFound(t, c, "zed", entries["zed"])
	assertEntryFound(t, c, "gambas", entries["gambas"])
	assertEntryFound(t, c, "gambas", entries["gambas"])
	assertEntryFound(t, c, "foo", entries["foo"])
	assertEntryNotFound(t, c, "non-existent")
	assertEntryFound(t, c, "conch", entries["conch"])
	assertEntryFound(t, c, "gambas", entries["gambas"])
	assertEntryNotFound(t, c, "non-existent")
	assertEntryFound(t, c, "foo", entries["foo"])
	assertEntryFound(t, c, "zed", entries["zed"])

	require.Equal(t, Statistics{
		LoadAttempts: 8,
		Evictions:    3,
		Hits:         5,
		Misses:       2,
		CurrentSize:  3,
	}, c.Statistics())
}

// Honors expirations if set, uses default expiration if explicit expirations
// are not set.
func TestHonorsExpirations(t *testing.T) {
	ctx := context.Background()

	clock := clockwork.NewFakeClock()
	c := New[string, string](100,
		WithClock[string, string](clock),
		WithDefaultTTL[string, string](time.Minute))

	// Put a bunch of entries with explicit expirations
	c.PutWithTTL(ctx, "expires_first", "bar", clock.Now().Add(time.Second*10))
	c.PutWithTTL(ctx, "expires_last", "zed", clock.Now().Add(time.Minute*5))
	c.PutWithTTL(ctx, "expires_second", "banana", clock.Now().Add(time.Minute+time.Second))

	// Put a bunch of entries with no explicit expiration
	c.Put(ctx, "first_default_expiry", "nock")
	c.Put(ctx, "second_default_expiry", "mork")

	// Advance time to only expire the first element and then access
	clock.Advance(time.Second * 15)
	assertEntryNotFound(t, c, "expires_first")
	assertEntryFound(t, c, "expires_second", "banana")
	assertEntryFound(t, c, "expires_last", "zed")
	assertEntryFound(t, c, "first_default_expiry", "nock")
	assertEntryFound(t, c, "second_default_expiry", "mork")
	assert.Equal(t, Statistics{
		Hits:         4,
		Misses:       1,
		LoadAttempts: 0,
		LoadFailures: 0,
		Expirations:  1,
		Evictions:    0,
		CurrentSize:  4,
	}, c.Statistics())

	// Move time forward past the default expiration and some explicit expirations
	clock.Advance(time.Minute)
	assertEntryNotFound(t, c, "expires_first")
	assertEntryNotFound(t, c, "expires_second")
	assertEntryFound(t, c, "expires_last", "zed")
	assertEntryNotFound(t, c, "first_default_expiry")
	assertEntryNotFound(t, c, "second_default_expiry")
	assert.Equal(t, Statistics{
		Hits:         5,
		Misses:       5,
		LoadAttempts: 0,
		LoadFailures: 0,
		Expirations:  4,
		Evictions:    0,
		CurrentSize:  1,
	}, c.Statistics())
}

func assertEntryFound[K comparable, V any](t *testing.T, c *Cache[K, V], key K, expected V) {
	actual, err := c.Get(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func assertEntryNotFound[K comparable, V any](t *testing.T, c *Cache[K, V], key K) {
	assertEntryError(t, c, key, ErrNotFound.Error())
}

func assertEntryError[K comparable, V any](t *testing.T, c *Cache[K, V], key K, expected string) {
	_, err := c.Get(context.Background(), key)
	require.Error(t, err)
	require.Contains(t, expected, err.Error())
}
