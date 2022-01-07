package log

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestSentry(t *testing.T) {
	t.SkipNow()
	conf := Config{
		Level: "debug",
		Sentry: SentryConfig{
			DSN: "https://fd6ac004f0724bf3b9831cd41e6230c3@sentry.improvity.ru/5",
			Tags: map[string]string{
				"component": "log-test",
			},
		},
	}
	err := Init(conf)
	require.NoError(t, err)
	Error(errors.New("test"))
	Sync()
}

type myerr struct{}

func (myerr) Error() string { return "myerr" }

func TestError(t *testing.T) {
	Error(errors.Errorf("err: %d", 1))
	Error(myerr{}, String("key", "value"))
}

func TestDebug(t *testing.T) {
	Debug("test1", Stringer("str1", mystring("666")))
	require.Equal(t, 1, getIntCalled)
	err := Init(Config{Level: "debug"})
	require.NoError(t, err)
	Debug("test2", Stringer("str2", mystring("777")))
	require.Equal(t, 1, getIntCalled)
}

var getIntCalled = 0

type mystring string

func (s mystring) String() string {
	getIntCalled++
	return string(s)
}
