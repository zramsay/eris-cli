package log

import (
	"bytes"
	"fmt"
	"os"
	"runtime/debug"

	log "github.com/Sirupsen/logrus"
	bugsnag "github.com/bugsnag/bugsnag-go"
)

// Default API Key. Can be overridden with the ERIS_BUGSNAG_TOKEN
// environment variable.
var APIKey = "1b9565bb7a4f8fd6dc446f2efd238fa3"

// Bugsnag implements the CrashReporter and the logrus.Hook interfaces.
type Bugsnag struct {
	config map[string]string

	remoteLogger *log.Logger
}

// NewBugsnagReporter configures the Bugsnag library and sets up a logger
// for collecting logging messages.
func NewBugsnagReporter(config map[string]string) Bugsnag {
	if os.Getenv("ERIS_BUGSNAG_TOKEN") != "" {
		APIKey = os.Getenv("ERIS_BUGSNAG_TOKEN")
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey:       APIKey,
		Synchronous:  true,
		AppVersion:   config["version"],
		ReleaseStage: config["branch"],
		// Bugsnag tries to say something itself occasionally.
		Logger: &log.Logger{
			Out:       os.Stdout,
			Formatter: ConsoleFormatter(log.DebugLevel),
			Level:     log.DebugLevel,
		},
		// Using our own panic recover.
		PanicHandler: func() {},
	})

	return Bugsnag{
		// Logger for silently collecting logging messages on all levels.
		remoteLogger: &log.Logger{
			Out:       new(bytes.Buffer),
			Formatter: RemoteFormatter(log.DebugLevel),
			Level:     log.DebugLevel,
		},
		config: config,
	}
}

func (b Bugsnag) Hook() log.Hook {
	return b
}

func (b Bugsnag) Levels() []log.Level {
	// Collecting messages on all levels.
	return []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
		log.InfoLevel,
		log.DebugLevel,
	}
}

func (b Bugsnag) Fire(e *log.Entry) error {
	out, err := b.remoteLogger.Formatter.Format(e)
	if err != nil {
		// Not important.
		return nil
	}

	b.remoteLogger.Out.Write(out)

	return nil
}

func (b Bugsnag) SendReport(message interface{}) error {
	debug.PrintStack()

	// Sending out a panic along with some useful bits of information.
	return bugsnag.Notify(
		fmt.Errorf("%v", message),
		bugsnag.ErrorClass{"panic"},
		bugsnag.SeverityError,
		bugsnag.User{Id: b.config["user"], Email: b.config["email"]},
		bugsnag.MetaData{
			"Log": {
				"Debug Output": b.remoteLogger.Out.(*bytes.Buffer).String(),
			},
			"Docker": {
				"Client": b.config["docker client"],
			},
		},
	)
}
