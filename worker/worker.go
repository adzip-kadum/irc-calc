package worker

import (
	"context"
	"sync"
	"time"

	"github.com/adzip-kadum/irc-calc/log"
)

func Worker(
	ctx context.Context,
	name string,
	interval time.Duration,
	periodic func(),
	finished func(),
	wg *sync.WaitGroup,
) {

	log.Info("worker started", log.String("name", name), log.Duration("interval", interval))
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
		if finished != nil {
			finished()
		}
		log.Info("worker finished", log.String("name", name))
		wg.Done()
	}()

	for {
		select {
		case <-timer.C:
			periodic()
			timer.Reset(interval)
		case <-ctx.Done():
			return
		}
	}
}
