package workers

import (
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestErrorCh_HasErrors(t *testing.T) {
	errorCh := NewErrorCh(2) // smaller than the number of workers

	require.NoError(t, errorCh.Start())

	// Spin up goroutines that generate errors
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()
			errorCh.AddError(errors.Errorf("error %d", i))
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close and confirm we have errors
	require.NoError(t, errorCh.Close())

	err := errorCh.Error()
	require.Error(t, err)
}

func TestErrorCh_HasNoError(t *testing.T) {
	errorCh := NewErrorCh(2) // smaller than the number of workers

	require.NoError(t, errorCh.Start())
	require.NoError(t, errorCh.Close())
	require.NoError(t, errorCh.Error())
}

func TestErrorCh_StateChanges(t *testing.T) {
	errorCh := NewErrorCh(1)

	// Not yet started, so can't close or get errors
	require.ErrorContains(t, errorCh.Close(), ErrChannelNotOpen.Error())
	require.ErrorContains(t, errorCh.Error(), ErrChannelStillOpen.Error())

	// Start
	require.NoError(t, errorCh.Start())

	// Don't allow double starts
	require.ErrorContains(t, errorCh.Start(), ErrChannelAlreadyStarted.Error())

	// Still running so can't get error
	require.ErrorContains(t, errorCh.Error(), ErrChannelStillOpen.Error())

	// Close
	require.NoError(t, errorCh.Close())

	// Don't allow start after close
	require.ErrorContains(t, errorCh.Start(), ErrChannelAlreadyStarted.Error())

	// Don't allow double close
	require.ErrorContains(t, errorCh.Close(), ErrChannelNotOpen.Error())

	// Can now get errors
	require.NoError(t, errorCh.Error())
}
