package commands

import (
	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	logger "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/log"
	"github.com/eris-ltd/eris-cli/config"

	"github.com/eris-ltd/eris-cli/util"
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
// a hook for the logrus logging library.
func CrashReportHook() log.Hook {
	switch config.GlobalConfig.Config.CrashReport {
	case "bugsnag":
		crashReport = logger.NewBugsnagReporter(ConfigureCrashReport())
	default:
		crashReport = logger.NewStubReporter(ConfigureCrashReport())
	}

	return crashReport.Hook()
}

// SendReport executes the actual transmission.
func SendReport(message interface{}) error {
	return crashReport.SendReport(message)
}

// ConfigureCrashReport collects variables from various places
// to send them along with a crash report.
func ConfigureCrashReport() map[string]string {
	user, email, err := config.GitConfigUser()
	if err != nil {
		user, email = "n/a", "n/a"
	}

	dockerClient, err := util.DockerClientVersion()
	if err != nil {
		dockerClient = "n/a"
	}

	return map[string]string{
		"version":       version.VERSION,
		"branch":        "*",
		"user":          user,
		"email":         email,
		"docker client": dockerClient,
	}
}
