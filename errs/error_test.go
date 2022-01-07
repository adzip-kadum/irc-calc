package errs

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adzip-kadum/irc-calc/log"
)

func TestError(t *testing.T) {
	// TODO: test sentry, test stack
	err := New("message", String("field1", "value1"), Int("field2", 666), Float64("field3", 0.666))
	err = WrapErr(err, ErrNotExists, String("id", "ORDER-ID"))
	err = Wrap(err, "message1", String("field11", "value1"), Int("field22", 777), Float64("field33", 0.777))
	err = Wrap(err, "message2", String("field111", "value1"), Int("field222", 888), Float64("field333", 0.888))
	if stre, ok := AsStructured(err); ok {
		log.Error(stre, stre.Fields()...)
	}
	data, _ := json.Marshal(err)
	fmt.Printf("JSON: %s\n", string(data))
}
