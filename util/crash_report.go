package util

import (
	"github.com/monax/cli/config"
	"github.com/monax/cli/log"
	"github.com/monax/cli/version"
)

var crashReport CrashReport

// CrashReport interface represents operations for sending out panics
// remotely and hooking to a logging library to collect debug messages.
type CrashReport interface {
	SendReport(interface{}, bool) error
	Hook() log.Hook
}

// CrashReportHook sets up a remote logging implementation (depending on
// the 'CrashReport' value in the `monax.toml` configuration file) and returns
// a hook for the Monax logging library.
func CrashReportHook(dockerVersion string) log.Hook {
	switch config.Global.CrashReport {
	case "bugsnag":
		crashReport = log.NewBugsnagReporter(ConfigureCrashReport(dockerVersion))
	default:
		crashReport = log.NewStubReporter(nil)
	}

	return crashReport.Hook()
}

// SendPanic sends a panic message to Bugsnag.
func SendPanic(message interface{}) error {
	return crashReport.SendReport(message, true)
}

// SendReport sends a message to Bugsnag (without
// a stack trace).
func SendReport(message interface{}) error {
	return crashReport.SendReport(message, false)
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
