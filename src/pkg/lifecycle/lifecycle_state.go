// Package lifecycle contains utilities making it easier to manage common
// application lifecycle.
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

// State is a very simple goroutine-safe Init -> Running -> Stopped -> Closed state machine,
// allowing goroutines to listen on state transitions. States start in the Init state,
// then transition to the Running state in response to a call to the Start method. Applications
// can use the IfRunning method to perform a block if and only if the state is Running.
//
// States have a two part termination - first being Stopped (through a call to Stop) and then
// being Closed (through a call to Close).
//
// Goroutines can block on the Running(), Stopped(), or Closed() channels to wait for
// the state to transition as desired.
type State struct {
	mut     sync.RWMutex
	state   state
	started chan struct{}
	stopped chan struct{}
	closed  chan struct{}
}

// NewState creates a new State in the Initialized state,
func NewState() *State {
	return &State{
		started: make(chan struct{}),
		stopped: make(chan struct{}),
		closed:  make(chan struct{}),
	}
}

// Start transitions from the "Initialized" state to the "Running" state.
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

// Stop transitions from the "Running" state to the "Stopped" state.
// Goroutines blocked on the Stopped() channel will wake up once this
// transition is complete.
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

// Close transitions from the "Stopped" state to the "Closed" state.
// Goroutines blocked on the Closed() channel will wake up once this
// transition is complete.
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

// Running returns a channel that goroutines can block on to
// wait until the state is "Running"
func (s *State) Running() <-chan struct{} { return s.started }

// Stopped returns a channel that goroutines can block on to
// wait until the state is "Stopped"
func (s *State) Stopped() <-chan struct{} { return s.stopped }

// Closed returns a channel that goroutines can block on to
// wait until the stte is "Closed"
func (s *State) Closed() <-chan struct{} { return s.closed }

// IfRunning runs the given block only if the state is in the "Running" state.
func (s *State) IfRunning(fn func()) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	if s.state != stateStarted {
		return false
	}

	fn()
	return true
}
