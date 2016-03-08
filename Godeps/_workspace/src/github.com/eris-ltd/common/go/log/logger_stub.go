package log

import (
	"runtime/debug"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

// Stub is a void implementation of the CrashReporter and logrus Hook interfaces.
type Stub struct{}

func NewStubReporter(c map[string]string) Stub {
	return Stub{}
}

func (s Stub) Levels() []log.Level {
	return []log.Level{}
}

func (s Stub) Fire(e *log.Entry) error {
	return nil
}

func (s Stub) Hook() log.Hook {
	return s
}

func (s Stub) SendReport(message interface{}) error {
	debug.PrintStack()

	return nil
}
