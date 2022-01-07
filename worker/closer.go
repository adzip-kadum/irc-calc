package worker

import (
	"context"
	"sync"
)

type Closer struct {
	Context   context.Context
	WaitGroup *sync.WaitGroup

	once   sync.Once
	cancel func()

	// for tests
	Closed bool
}

func NewCloser(ctx context.Context, workersNum int) *Closer {
	c := &Closer{}
	c.Context, c.cancel = context.WithCancel(context.Background())
	c.WaitGroup = &sync.WaitGroup{}
	c.WaitGroup.Add(workersNum)
	return c
}

func (c *Closer) Close() {
	c.once.Do(func() {
		c.cancel()
		c.WaitGroup.Wait()
		c.Closed = true
	})
}
