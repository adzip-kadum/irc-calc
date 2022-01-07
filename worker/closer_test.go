package worker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloser(t *testing.T) {
	closer := NewCloser(context.Background(), 2)
	var done1, done2 bool

	go func() {
		defer func() {
			done1 = true
			closer.WaitGroup.Done()
		}()
		<-closer.Context.Done()
	}()

	go func() {
		defer func() {
			done2 = true
			closer.WaitGroup.Done()
		}()
		<-closer.Context.Done()
	}()

	closer.Close()

	assert.Equal(t, true, closer.Closed)
	assert.Equal(t, true, done1)
	assert.Equal(t, true, done2)
}
