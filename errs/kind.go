package errs

import "github.com/pkg/errors"

type Kind string

const (
	ErrNotExists        Kind = "not exists"
	ErrAlreadyExists    Kind = "already exists"
	ErrPermissionDenied Kind = "permission denied"
)

func (k Kind) Error() string {
	return string(k)
}

func (k Kind) Kind() Kind {
	return k
}

func AsKind(e error) (err Kind, ok bool) {
	ok = errors.As(e, &err)
	return
}

func HasKind(e error) bool {
	return errors.Is(e, new(Kind))
}
