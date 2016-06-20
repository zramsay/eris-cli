package logger

import (
	"runtime/debug"
)

// Stub is a void implementation of the CrashReporter and Eris logger Hook interfaces.
type Stub struct{}

func NewStubReporter(c map[string]string) Stub {
	return Stub{}
}

func (s Stub) Levels() []Level {
	return []Level{}
}

func (s Stub) Fire(e *Entry) error {
	return nil
}

func (s Stub) Hook() Hook {
	return s
}

func (s Stub) SendReport(message interface{}) error {
	debug.PrintStack()

	return nil
}
