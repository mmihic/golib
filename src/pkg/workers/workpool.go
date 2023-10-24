// Package workers contains a helper for parallelizing work
// in a finite number of goroutines.
package workers

import (
	"sync"
)

// A Pool is a pool of workers, allowing for concurrency restrictions.
type Pool struct {
	leases chan struct{}
	wg     sync.WaitGroup
}

// AddTask adds a task to the worker pool.
func (p *Pool) AddTask(fn func()) {
	p.wg.Add(1)
	go func() {
		<-p.leases

		defer func() {
			p.leases <- struct{}{}
			p.wg.Done()
		}()
		fn()
	}()
}

// Do runs a function with a set of workers
func Do(numWorkers int, f func(p *Pool)) {
	p := &Pool{
		leases: make(chan struct{}, numWorkers),
	}

	for i := 0; i < numWorkers; i++ {
		p.leases <- struct{}{}
	}

	f(p)
	p.wg.Wait()
}
