package log

import (
	"sync"
)

type Hooker interface {
	Debug(string, ...Field)
	Info(string, ...Field)
	Error(string, ...Field)
}

func NewCollector() *Collector {
	return &Collector{}
}

type Collector struct {
	Debugs LogEntries
	Infos  LogEntries
	Errors LogEntries
	sync.Mutex
}

type LogEntry struct {
	Num    int
	Msg    string
	Fields Fields
}

type Fields []Field

func (f Fields) Get(key string) Field {
	for _, field := range f {
		if field.Key == key {
			return field
		}
	}
	return Field{}
}

type LogEntries []LogEntry

func (e LogEntries) Find(msg string) (entries LogEntries) {
	for _, entry := range e {
		if entry.Msg == msg {
			entries = append(entries, entry)
		}
	}
	return
}

func (c *Collector) Reset() {
	c.Lock()
	defer c.Unlock()
	c.Debugs = nil
	c.Infos = nil
	c.Errors = nil
}

func (c *Collector) Debug(msg string, args ...Field) {
	c.Lock()
	defer c.Unlock()
	c.Debugs = append(c.Debugs, LogEntry{len(c.Debugs), msg, args})
}

func (c *Collector) Info(msg string, args ...Field) {
	c.Lock()
	defer c.Unlock()
	c.Infos = append(c.Infos, LogEntry{len(c.Infos), msg, args})
}

func (c *Collector) Error(msg string, args ...Field) {
	c.Lock()
	defer c.Unlock()
	c.Errors = append(c.Errors, LogEntry{len(c.Errors), msg, args})
}

func (c *Collector) Equal(num int, msg string, fields ...Field) (count int) {
	c.Lock()
	defer c.Unlock()
	eq := newEquality(num, msg, fields...)
	count += c.debugEqual(eq)
	count += c.infoEqual(eq)
	count += c.errorEqual(eq)
	return
}

func (c *Collector) DebugEqual(num int, msg string, fields ...Field) int {
	c.Lock()
	defer c.Unlock()
	eq := newEquality(num, msg, fields...)
	return c.debugEqual(eq)
}

func (c *Collector) InfoEqual(num int, msg string, fields ...Field) int {
	c.Lock()
	defer c.Unlock()
	eq := newEquality(num, msg, fields...)
	return c.infoEqual(eq)
}

func (c *Collector) ErrorEqual(num int, msg string, fields ...Field) int {
	c.Lock()
	defer c.Unlock()
	eq := newEquality(num, msg, fields...)
	return c.errorEqual(eq)
}

func (c *Collector) debugEqual(eq equality) (count int) {
	for _, entry := range c.Debugs {
		if eq.equal(entry) {
			count++
		}
	}
	return
}

func (c *Collector) infoEqual(eq equality) (count int) {
	for _, entry := range c.Infos {
		if eq.equal(entry) {
			count++
		}
	}
	return
}

func (c *Collector) errorEqual(eq equality) (count int) {
	for _, entry := range c.Errors {
		if eq.equal(entry) {
			count++
		}
	}
	return
}

type equality struct {
	num    int
	msg    string
	fields map[string]Field
}

func newEquality(num int, msg string, fields ...Field) equality {
	mfields := make(map[string]Field, len(fields))
	for _, f := range fields {
		mfields[f.Key] = f
	}
	return equality{
		num:    num,
		msg:    msg,
		fields: mfields,
	}
}

func (eq equality) equal(e LogEntry) bool {
	if eq.num != -1 && eq.num != e.Num {
		return false
	}
	if eq.msg != e.Msg {
		return false
	}
	for _, field := range e.Fields {
		if f, ok := eq.fields[field.Key]; ok {
			if !field.Equals(f) {
				return false
			}
		}
	}
	return true
}
