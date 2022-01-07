package errs

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestWrap(t *testing.T) {
	e := Wrap(myerr{}, "message")
	e = WrapErr(e, ErrNotExists)
	//	require.Equal(t, "message: my err: not exists", e.Error())
	require.Equal(t, "my err", errors.Cause(e).Error())
	require.True(t, errors.Is(e, myerr{}))
	require.True(t, errors.Is(e, ErrNotExists))
	require.False(t, errors.Is(e, ErrAlreadyExists))
	my := myerr{}
	require.True(t, errors.As(e, &my))
	require.True(t, HasKind(e))
}

type myerr struct {
}

func (e myerr) Error() string {
	return "my err"
}
