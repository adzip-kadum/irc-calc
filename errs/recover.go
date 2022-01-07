package errs

import (
	"github.com/pkg/errors"
)

func Recover(rerr *error) {
	if r := recover(); r != nil {
		switch err := (r).(type) {
		case error:
			*rerr = errors.WithStack(err)
		default:
			*rerr = errors.Errorf("%s", r)
		}
	}
}
