package log

import (
	"fmt"
	"io"
	"os"
	"sync"
)

//--------------------------------------------------------------------------------
// thread safe logger that fires messages from multiple packages one at a time

func init() {
	go readLoop()
}

var (
	// control access to loggers
	mtx     sync.Mutex
	loggers = make(map[string]*Logger)

	// access to writers managed by channels
	writer    io.Writer = os.Stdout
	errWriter io.Writer = os.Stderr

	chanBuffer = 100
	writeCh    = make(chan []byte, chanBuffer)
	errorCh    = make(chan []byte, chanBuffer)

	quitCh = make(chan struct{})
)

type Logger struct {
	Level int
	Pkg   string

	// these are here for easy access
	// for functions that want the writer
	Writer    *SafeWriter
	ErrWriter *SafeWriter
}

// add a default logger with pkg name
func AddLogger(pkg string) *Logger {
	l := &Logger{
		Level:     0,
		Pkg:       pkg,
		Writer:    NewSafeWriter(writeCh),
		ErrWriter: NewSafeWriter(errorCh),
	}
	mtx.Lock()
	loggers[pkg] = l
	mtx.Unlock()
	return l
}

// set levels for individual packages
func SetLogLevel(pkg string, level int) {
	mtx.Lock()
	defer mtx.Unlock()

	if l, ok := loggers[pkg]; ok {
		l.Level = level
		if level > 1 {
			// TODO: wrap the writers to print [<pkg>]
		}
	}
}

// set level and writer for all loggers
func SetLoggers(level int, w io.Writer, ew io.Writer) {
	mtx.Lock()
	defer mtx.Unlock()
	for _, l := range loggers {
		l.Level = level
		if l.Level > 1 {
			// TODO: wrap the writers to print [<pkg>]
		}
	}
	writer = w
	errWriter = ew
}

//--------------------------------------------------------------------------------
// concurrency

func readLoop() {
LOOP:
	for {
		select {
		case b := <-writeCh:
			writer.Write(b)
		case b := <-errorCh:
			errWriter.Write(b)
		case <-quitCh:
			break LOOP

		}
	}
}

func Flush() {
	quitCh <- struct{}{}
LOOP:
	for {
		select {
		case b := <-writeCh:
			writer.Write(b)
		case b := <-errorCh:
			errWriter.Write(b)
		default:
			break LOOP
		}
	}
}

// a SafeWriter implements Writer and fires its bytes over the channel
// to be written to the writer or errWriter
type SafeWriter struct {
	ch chan []byte
}

func (sw *SafeWriter) Write(b []byte) (int, error) {
	sw.ch <- b
	return len(b), nil
}

func NewSafeWriter(ch chan []byte) *SafeWriter {
	return &SafeWriter{ch}
}

// thread safe writes
func writef(s string, args ...interface{}) {
	writeCh <- []byte(fmt.Sprintf(s, args...))
}

func writeln(s ...interface{}) {
	writeCh <- []byte(fmt.Sprintln(s...))
}

func errorf(s string, args ...interface{}) {
	errorCh <- []byte(fmt.Sprintf(s, args...))
}

func errorln(s ...interface{}) {
	errorCh <- []byte(fmt.Sprintln(s...))
}

//--------------------------------------------------------------------------------
// public logger functions

// Printf and Println write to the Writer no matter what
func (l *Logger) Printf(s string, args ...interface{}) {
	writef(s, args...)
}

func (l *Logger) Println(s ...interface{}) {
	writeln(s...)
}

// Errorf and Errorln write to the ErrWriter no matter what
func (l *Logger) Errorf(s string, args ...interface{}) {
	errorf(s, args...)
}

func (l *Logger) Errorln(s ...interface{}) {
	errorln(s...)
}

// Infof and Infoln write to the Writer if log level >= 1
func (l *Logger) Infof(s string, args ...interface{}) {
	if l.Level > 0 {
		writef(s, args...)
	}
}

func (l *Logger) Infoln(s ...interface{}) {
	if l.Level > 0 {
		writeln(s...)
	}
}

// Debugf and Debugln write to the Writer if log level >= 2
func (l *Logger) Debugf(s string, args ...interface{}) {
	if l.Level > 1 {
		writef(s, args...)
	}
}

func (l *Logger) Debugln(s ...interface{}) {
	if l.Level > 1 {
		writeln(s...)
	}
}
