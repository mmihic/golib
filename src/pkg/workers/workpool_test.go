package workers

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	const MaxConcurrency = 5

	var (
		unblock = make(chan struct{})
		start   = make(chan struct{})
	)

	var scheduledConcurrency int
	var maxConcurrency int
	var mut sync.Mutex

	Do(MaxConcurrency, func(p *Pool) {
		for i := 0; i < 20; i++ {
			p.AddTask(func() {
				<-start

				mut.Lock()
				scheduledConcurrency++
				if scheduledConcurrency > maxConcurrency {
					maxConcurrency++
				}
				scheduledConcurrency--
				mut.Unlock()

				<-unblock
			})
		}

		// Allow all scheduled goroutines to start
		close(start)

		// Wait a second so that we get max concurrency
		time.Sleep(time.Second)

		// Allow all active goroutines to start
		close(unblock)
	})

	assert.Less(t, maxConcurrency, MaxConcurrency)
}
