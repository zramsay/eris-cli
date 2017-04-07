package log

import (
	"runtime/debug"
)

// Stub is a void implementation of the CrashReporter and Monax logger Hook interfaces.
type Stub struct{}

// NewStubReporter returns a new Stub implementation.
func NewStubReporter(c map[string]string) Stub {
	return Stub{}
}

// Levels is an implementation of the logger Levels method.
func (s Stub) Levels() []Level {
	return []Level{}
}

// Fire is an implementation of the logger Fire method.
func (s Stub) Fire(e *Entry) error {
	return nil
}

// Hook is an implementation of the logger Hook method.
func (s Stub) Hook() Hook {
	return s
}

// SendReport optionally prints a stack trace to the console
// to make sure stub actually is used as an implementation.
func (s Stub) SendReport(message interface{}, stack bool) error {
	if stack == true {
		debug.PrintStack()
	}

	return nil
}
