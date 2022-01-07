package errs

import (
	"github.com/pkg/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type Error struct {
	err    error
	causer error
	fields []Field
}

var (
	Bool     = zap.Bool
	Int      = zap.Int
	Int32    = zap.Int32
	Int64    = zap.Int64
	Float32  = zap.Float32
	Float64  = zap.Float64
	String   = zap.String
	Strings  = zap.Strings
	Stringer = zap.Stringer
	Duration = zap.Duration
	Any      = zap.Any
	Err      = zap.Error
	Object   = zap.Object
)

type Field = zapcore.Field

func New(message string, fields ...Field) error {
	return &Error{err: errors.New(message), fields: fields}
}

func Wrap(cause error, message string, fields ...Field) error {
	if cause == nil {
		return nil
	}
	return Error{causer: cause, err: errors.New(message), fields: fields}
}

func WrapErr(cause, err error, fields ...Field) error {
	if cause == nil {
		return nil
	}
	return Error{causer: cause, err: err, fields: fields}
}

func (s Error) JSONBuffer() *buffer.Buffer {
	// NOTE: ignoring the error here is safe with the current version of zap's JSON encoder, as it
	// is always nil
	buf, _ := zapcore.NewJSONEncoder(jsonEncConf).EncodeEntry(s.entry())
	return buf
}

func (s Error) MarshalJSON() ([]byte, error) {
	buf := s.JSONBuffer()
	bufBs := buf.Bytes()
	bs := make([]byte, len(bufBs))
	copy(bs, bufBs)
	buf.Free()

	return bs, nil
}

func (s Error) JSON() string {
	bs, _ := s.MarshalJSON()
	return string(bs)
}

func (s Error) Fields() []Field {
	fs := make([]Field, 0, len(s.fields)+1)

	fs = append(fs, s.fields...)

	if stre, ok := AsStructured(s.Unwrap()); ok {
		fs = append(fs, Object("cause", stre))
		return fs
	}

	return fs
}

func (s Error) MarshalLogObject(oe zapcore.ObjectEncoder) error {
	oe.AddString("msg", s.errorOrCause())
	for _, field := range s.Fields() {
		field.AddTo(oe)
	}
	return nil
}

func (s Error) Unwrap() error {
	return s.causer
}

func (s Error) Is(target error) bool {
	return target == s.err
}

func (s Error) Cause() error {
	return s.causer
}

func (s Error) errorOrCause() string {
	if s.err != nil {
		return s.err.Error()
	}
	return s.causer.Error()
}

func (s Error) Error() string {
	if s.err != nil {
		if s.causer == nil { // only an error but no cause, so return that
			return s.err.Error()
		} else { // have an error and a cause for it, return both
			return s.err.Error() + ": " + s.causer.Error()
		}
	}

	return s.causer.Error()
}

func (s Error) entry() (zapcore.Entry, []zapcore.Field) {
	return zapcore.Entry{Message: s.errorOrCause()}, s.Fields()
}

func AsStructured(e error) (err Error, ok bool) {
	ok = errors.As(e, &err)
	return
}

func IsStructured(e error) bool {
	return errors.Is(e, Error{})
}

// Field returns a zap field for err under the key "error". If err is nil, returns a no-op field. If
// err is a structured error or has one in its error chain, returns a zap.Object field, if err is a
// plain 'ol error, returns zap.Error
//func Field(err error) zapcore.Field {
//	if err == nil {
//		return zap.Skip()
//	}
//	stre, ok := AsStructured(err)
//	if ok {
//		return zap.Object("error", stre)
//	}
//	return zap.Error(err)
//}

var jsonEncConf zapcore.EncoderConfig

func init() {
	jsonEncConf = zap.NewProductionEncoderConfig()
	jsonEncConf.CallerKey = ""
	jsonEncConf.StacktraceKey = ""
	jsonEncConf.LevelKey = ""
	jsonEncConf.TimeKey = ""
	jsonEncConf.NameKey = ""
	jsonEncConf.EncodeCaller = nil
}

/*
https://forum.golangbridge.org/t/go-wrapped-errors/24898

type wrappedError struct {
	innerErr error
	outerErr error
}

func (e *wrappedError) Error() string { return e.outerErr.Error() }

func (e *wrappedError) Is(v error) bool { return v == e.outerErr }

func (e *wrappedError) Unwrap() error { return e.innerErr }

func (e *wrappedError) As(target interface{}) bool {
	return errors.As(e.outerErr, target)
}

func Wrap(outerErr error, innerErr error) error {
	return &wrappedError{outerErr: outerErr, innerErr: innerErr}
}
*/
