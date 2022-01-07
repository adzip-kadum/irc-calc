package probe

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/textproto"
	"runtime"
	"time"

	"github.com/pkg/errors"
)

func TCPDialProbe(addr string, timeout time.Duration) Probe {
	return func() error {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			return err
		}
		return conn.Close()
	}
}

func HTTPGetProbe(url string, timeout time.Duration) Probe {
	client := http.Client{
		Timeout: timeout,
		// never follow redirects
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return func() error {
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("returned status %d", resp.StatusCode)
		}
		return nil
	}
}

func DatabasePingProbe(db *sql.DB, timeout time.Duration) Probe {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if db == nil {
			return fmt.Errorf("database is nil")
		}
		return db.PingContext(ctx)
	}
}

func DNSResolveProbe(host string, timeout time.Duration) Probe {
	resolver := net.Resolver{}
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		addrs, err := resolver.LookupHost(ctx, host)
		if err != nil {
			return err
		}
		if len(addrs) < 1 {
			return fmt.Errorf("could not resolve host")
		}
		return nil
	}
}

func GoroutineCountProbe(threshold int) Probe {
	return func() error {
		count := runtime.NumGoroutine()
		if count > threshold {
			return fmt.Errorf("too many goroutines (%d > %d)", count, threshold)
		}
		return nil
	}
}

func RedisProbe(addr string, timeout time.Duration) Probe {
	return func() error {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			return err
		}

		redis := textproto.NewConn(conn)

		_, err = redis.Cmd("PING")
		if err != nil {
			redis.Close()
			return err
		}

		resp, err := redis.ReadLine()
		if err != nil {
			redis.Close()
			return err
		}

		if resp != "+PONG" {
			redis.Close()
			return errors.New(resp)
		}

		return redis.Close()
	}
}
