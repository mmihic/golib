package workers

import (
	"errors"
	"go.uber.org/multierr"
	"sync/atomic"
)

var (
	// ErrChannelAlreadyStarted is returned when trying to start a channel that
	// was already started.
	ErrChannelAlreadyStarted = errors.New("error channel already started")

	// ErrChannelNotOpen is returned when trying to close a channel that is not open.
	ErrChannelNotOpen = errors.New("error channel not open")

	// ErrChannelNotStarted is returned when trying to pull errors from a channel
	// that has not been started. Doing so would otherwise result in a deadlock.
	ErrChannelNotStarted = errors.New("error channel not started")

	// ErrChannelStillOpen is returned when trying to pull errors from a channel
	// that is still open.
	ErrChannelStillOpen = errors.New("error channel is still open")
)

// An ErrorCh is a channel of errors from workers. Allows async workers to submit errors,
// and has a goroutine that collects errors and aggregates them into a multierr.
type ErrorCh struct {
	ch       chan error
	state    atomic.Int32
	started  atomic.Bool
	closed   atomic.Bool
	done     chan struct{}
	multierr error
}

// NewErrorCh creates a new channel of errors.
func NewErrorCh(size int) *ErrorCh {
	return &ErrorCh{
		ch:   make(chan error, size),
		done: make(chan struct{}),
	}
}

// Start starts a goroutine to collect errors.
func (ch *ErrorCh) Start() error {
	if !ch.state.CompareAndSwap(stateInitial, stateRunning) {
		return ErrChannelAlreadyStarted
	}

	go func() {
		for err := range ch.ch {
			ch.multierr = multierr.Append(ch.multierr, err)
		}

		close(ch.done)
	}()

	return nil
}

// Close closes the channel.
func (ch *ErrorCh) Close() error {
	if !ch.state.CompareAndSwap(stateRunning, stateClosed) {
		return ErrChannelNotOpen
	}

	close(ch.ch)
	return nil
}

// AddError adds an error to the channel.
func (ch *ErrorCh) AddError(err error) {
	ch.ch <- err
}

// Error pulls the errors off the channel and returns a multierr.
func (ch *ErrorCh) Error() error {
	if ch.state.Load() != stateClosed {
		return ErrChannelStillOpen
	}

	<-ch.done
	return ch.multierr
}

const (
	stateInitial int32 = 0
	stateRunning int32 = 1
	stateClosed  int32 = 2
)
