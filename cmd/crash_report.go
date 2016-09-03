package commands

import (
	"github.com/eris-ltd/eris-cli/config"
	log "github.com/eris-ltd/eris-logger"

	"github.com/eris-ltd/eris-cli/version"
)

var crashReport CrashReport

// CrashReport interface represents operations for sending out panics
// remotely and hooking to a logging library to collect debug messages.
type CrashReport interface {
	SendReport(interface{}) error
	Hook() log.Hook
}

// CrashReportHook sets up a remote logging implementation (depending on
// the 'CrashReport' value in the `eris.toml` configuration file) and returns
// a hook for the Eris logging library.
func CrashReportHook(dockerVersion string) log.Hook {
	switch config.Global.CrashReport {
	case "bugsnag":
		crashReport = log.NewBugsnagReporter(ConfigureCrashReport(dockerVersion))
	default:
		crashReport = log.NewStubReporter(nil)
	}

	return crashReport.Hook()
}

// SendReport executes the actual transmission.
func SendReport(message interface{}) error {
	return crashReport.SendReport(message)
}

// ConfigureCrashReport collects variables from various places
// to send them along with a crash report.
func ConfigureCrashReport(dockerVersion string) map[string]string {
	user, email, err := config.GitConfigUser()
	if err != nil {
		user, email = "n/a", "n/a"
	}

	return map[string]string{
		"version":       version.VERSION,
		"branch":        "*",
		"user":          user,
		"email":         email,
		"docker client": dockerVersion,
	}
}
