package lifecycle

import (
	"sync"
)

type state int

const (
	stateInit state = iota
	stateStarted
	stateStopped
	stateClosed
)

// State is a very simple goroutine-safe Init -> Run -> Stop -> Close state machine,
// allowing goroutines to listen on state transitions.
type State struct {
	mut     sync.RWMutex
	state   state
	started chan struct{}
	stopped chan struct{}
	closed  chan struct{}
}

func NewState() *State {
	return &State{
		started: make(chan struct{}),
		stopped: make(chan struct{}),
		closed:  make(chan struct{}),
	}
}

// Start starts the lifecycle
func (s *State) Start() bool {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.state != stateInit {
		return false
	}

	s.state = stateStarted
	close(s.started)
	return true
}

// Stop stops the lifecycle, unblocking
// any goroutines waiting on the Stopped
// channel.
func (s *State) Stop() bool {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.state != stateStarted {
		// Either it's never started, or it's already been stopped
		return false
	}

	s.state = stateStopped
	close(s.stopped)
	return true
}

// Close marks the lifecycle as fully closed,
// unblocking any goroutines waiting on the
// closed channel.
func (s *State) Close() bool {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.state != stateStopped {
		return false
	}

	s.state = stateClosed
	close(s.closed)
	return true
}

// Started returns a channel that goroutines can block on to
// wait until Start() is called.
func (s *State) Started() <-chan struct{} { return s.started }

// Stopped returns a channel that goroutines can block on to
// wait until Stop() is called.
func (s *State) Stopped() <-chan struct{} { return s.stopped }

// Closed returns a channel that goroutines can block on to
// wait until Close() is called.
func (s *State) Closed() <-chan struct{} { return s.closed }

// IfRunning runs the given block only if the state
// is in the running state.
func (s *State) IfRunning(fn func()) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if s.state != stateStarted {
		return false
	}

	fn()
	return true
}
