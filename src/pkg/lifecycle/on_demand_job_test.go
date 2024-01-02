package lifecycle

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOnDemandJob(t *testing.T) {
	var counter atomic.Int32
	processed := make(chan struct{})
	job := StartOnDemandJob(func() {
		counter.Add(1)
		processed <- struct{}{}
	})

	for i := int32(0); i < 150; i++ {
		assert.Equal(t, counter.Load(), i)
		job.Signal(context.Background())
		select {
		case <-processed:
		case <-time.After(time.Second * 5):
			t.Fatalf("timed out iteration %d waiting for loop to fire", i)
		}
		assert.Equal(t, counter.Load(), i+1)
	}

	job.Stop()
	<-job.Stopped()
	assert.Equal(t, counter.Load(), int32(150))
}
