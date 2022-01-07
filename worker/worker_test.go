package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	periodic := 0
	finished := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go Worker(ctx, "test", 300*time.Millisecond,
		func() { periodic++ },
		func() { finished++ },
		wg,
	)

	time.Sleep(time.Second)
	cancel()
	wg.Wait()

	assert.Equal(t, 3, periodic)
	assert.Equal(t, 1, finished)
}
