package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime/debug"

	bugsnag "github.com/bugsnag/bugsnag-go"
)

// APIKey can be overridden with the MONAX_BUGSNAG_TOKEN
// environment variable.
var APIKey = "1b9565bb7a4f8fd6dc446f2efd238fa3"

// Bugsnag implements the CrashReporter and the Monax logger Hook interfaces.
type Bugsnag struct {
	config map[string]string

	remoteLogger *Logger
}

// NewBugsnagReporter configures the Bugsnag library and sets up a logger
// for collecting logging messages.
func NewBugsnagReporter(config map[string]string) Bugsnag {
	if os.Getenv("MONAX_BUGSNAG_TOKEN") != "" {
		APIKey = os.Getenv("MONAX_BUGSNAG_TOKEN")
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       APIKey,
		Synchronous:  true,
		AppVersion:   config["version"],
		ReleaseStage: config["branch"],
		// Bugsnag tries to say something itself occasionally.
		Logger: &Logger{
			Out:       os.Stdout,
			Formatter: MonaxFormatter{},
			Level:     DebugLevel,
		},
		// Using our own panic recover.
		PanicHandler: func() {},
	})

	return Bugsnag{
		remoteLogger: &Logger{
			Out:       new(bytes.Buffer),
			Formatter: MonaxFormatter{IgnoreLevel: true},
			Level:     DebugLevel,
		},
		config: config,
	}
}

// Hook is an implementation of the logger Hook method.
func (b Bugsnag) Hook() Hook {
	return b
}

// Levels is an implementation of the logger Levels method.
func (b Bugsnag) Levels() []Level {
	// Collecting messages on all levels.
	return []Level{
		PanicLevel,
		FatalLevel,
		ErrorLevel,
		WarnLevel,
		InfoLevel,
		DebugLevel,
	}
}

// Fire is an implementation of the logger Fire method.
func (b Bugsnag) Fire(e *Entry) error {
	out, err := b.remoteLogger.Formatter.Format(e)
	if err != nil {
		// Not important.
		return nil
	}

	b.remoteLogger.Out.Write(out)

	return nil
}

// SendReport method connects to the Bugsnag server and
// sends out collected debugging and optional stack trace info.
func (b Bugsnag) SendReport(message interface{}, stack bool) error {
	if stack == true {
		debug.PrintStack()
	}

	// Sending out a panic along with some useful bits of information.
	return bugsnag.Notify(
		fmt.Errorf("%v", message),
		bugsnag.ErrorClass{"panic"},
		bugsnag.SeverityError,
		bugsnag.User{Id: b.config["user"], Email: b.config["email"]},
		bugsnag.MetaData{
			"log": {
				"debug": b.remoteLogger.Out.(*bytes.Buffer).String(),
			},
			"docker": {
				"client": b.config["docker client"],
			},
		},
	)
}
