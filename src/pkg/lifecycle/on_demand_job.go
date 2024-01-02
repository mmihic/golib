package lifecycle

import (
	"context"
)

// An OnDemandJob runs in the background, executing a background function
// when signalled.
type OnDemandJob struct {
	loopFn    func()
	state     *State
	signalled chan struct{}
}

// StartOnDemandJob starts an OnDemandJob with a loop function.
func StartOnDemandJob(loopFn func()) *OnDemandJob {
	j := &OnDemandJob{
		state:     NewState(),
		loopFn:    loopFn,
		signalled: make(chan struct{}, 1),
	}

	j.state.Start()

	go func() {
	done:
		for {
			select {
			case <-j.state.Stopped():
				break done
			case <-j.signalled:
				j.loopFn()
			}
		}

		close(j.signalled)
		j.state.Close()
	}()

	return j
}

// Stop stops the background job.
func (j *OnDemandJob) Stop() {
	j.state.Stop()
}

// Stopped returns a channel that callers can block on to wait until
// the job has completely stopped.
func (j *OnDemandJob) Stopped() <-chan struct{} {
	return j.state.Closed()
}

// Signal signals the background job to process a loop., without blocking.
// If there is already a signal pending, the second signal is not sent.
func (j *OnDemandJob) Signal(_ context.Context) {
	j.state.IfRunning(func() {
		select {
		case j.signalled <- struct{}{}:
		default:
		}
	})
}
