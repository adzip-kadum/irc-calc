package errs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecover(t *testing.T) {
	err := recovered()
	require.NotNil(t, err)
	require.Equal(t, "runtime error: invalid memory address or nil pointer dereference", err.Error())
}

func recovered() (rerr error) {
	defer Recover(&rerr)
	s := struct {
		f func()
	}{}
	s.f()
	return nil
}
