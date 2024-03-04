package lifecycle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStateTransitions(t *testing.T) {
	var (
		startGoRoutineDone = make(chan struct{})
		stopGoRoutineDone  = make(chan struct{})
		closeGoRoutineDone = make(chan struct{})
	)

	st := NewState()

	go func() {
		<-st.Running()
		close(startGoRoutineDone)
	}()

	go func() {
		<-st.Stopped()
		close(stopGoRoutineDone)
	}()

	go func() {
		<-st.Closed()
		close(closeGoRoutineDone)
	}()

	// Not started, so stop and close does nothing
	require.False(t, st.Stop())
	require.False(t, st.Close())

	// Should not run the given function
	runBeforeStart := false
	require.False(t, st.IfRunning(func() {
		runBeforeStart = true
	}))
	require.False(t, runBeforeStart)

	// Start
	require.True(t, st.Start())
	require.False(t, st.Start()) // Second start does nothing
	require.False(t, st.Close()) // Not stopped, so can't close

	// Should run function now that we've started
	runAfterStart := false
	require.True(t, st.IfRunning(func() {
		runAfterStart = true
	}))
	require.True(t, runAfterStart)

	select {
	case <-time.NewTimer(time.Second * 5).C:
		require.Fail(t, "start go-routine did not complete within 5s")
	case <-startGoRoutineDone:
	}

	// Stop
	require.True(t, st.Stop())
	require.False(t, st.Stop())  // Second stop does nothing
	require.False(t, st.Start()) // Can't start

	select {
	case <-time.NewTimer(time.Second * 5).C:
		require.Fail(t, "stop go-routine did not complete within 5s")
	case <-stopGoRoutineDone:
	}

	// Should not run function when stopped
	runAfterStop := false
	require.False(t, st.IfRunning(func() {
		runAfterStop = true
	}))
	require.False(t, runAfterStop)

	// Close
	require.True(t, st.Close())
	require.False(t, st.Close()) // Close does nothing when closed
	require.False(t, st.Stop())  // Stop does nothing when closed
	require.False(t, st.Start()) // Start does nothing when closed

	select {
	case <-time.NewTimer(time.Second * 5).C:
		require.Fail(t, "stop go-routine did not complete within 5s")
	case <-closeGoRoutineDone:
	}

	// Should not run function when stopped
	runAfterClose := false
	require.False(t, st.IfRunning(func() {
		runAfterClose = true
	}))
	require.False(t, runAfterClose)
}

func TestState_IfRunningReentrant(t *testing.T) {
	st := NewState()
	st.Start()

	reentrantRunAllowed := false
	st.IfRunning(func() {
		st.IfRunning(func() {
			reentrantRunAllowed = true
		})
	})
	require.True(t, reentrantRunAllowed)
}
