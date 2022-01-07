package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func Wait() error {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-ch
		cancel()
	}()

	<-ctx.Done()

	if ctx.Err() != nil && ctx.Err() != context.Canceled {
		return ctx.Err()
	}

	return nil
}
