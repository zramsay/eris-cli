package commands

import (
	"os"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cm/definitions"
	"github.com/eris-ltd/eris-cm/version"

	log "github.com/eris-ltd/eris-logger"
	"github.com/spf13/cobra"
)

const VERSION = version.VERSION

// Global Do struct
var do *definitions.Do
var keysAddr string

// Defining the root command
var ErisCMCmd = &cobra.Command{
	Use:   "eris-cm",
	Short: "The Eris Chain Manager is a utility for performing complex operations on eris chains",
	Long: `The Eris Chain Manager is a utility for performing complex operations on eris chains.

Made with <3 by Monax Industries.

Complete documentation is available at https://monax.io/docs/documentation/cm/
` + "\nVersion:\n  " + VERSION,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.WarnLevel)
		if do.Verbose {
			log.SetLevel(log.InfoLevel)
		} else if do.Debug {
			log.SetLevel(log.DebugLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func Execute() {
	InitErisChainManager()
	AddGlobalFlags()
	AddCommands()
	ErisCMCmd.Execute()
}

func InitErisChainManager() {
	do = definitions.NowDo()
}

func AddCommands() {
	buildMakerCommand()
	ErisCMCmd.AddCommand(MakerCmd)
}

func AddGlobalFlags() {
	ErisCMCmd.PersistentFlags().BoolVarP(&do.Verbose, "verbose", "v", defaultVerbose(), "verbose output; more output than no output flags; less output than debug level; default respects $ERIS_CHAINMANAGER_VERBOSE")
	ErisCMCmd.PersistentFlags().BoolVarP(&do.Debug, "debug", "d", defaultDebug(), "debug level output; the most output available for eris-cm; if it is too chatty use verbose flag; default respects $ERIS_CHAINMANAGER_DEBUG")
	ErisCMCmd.PersistentFlags().BoolVarP(&do.Output, "output", "o", defaultOutput(), "should eris-cm provide an output of its job; default respects $ERIS_CHAINMANAGER_OUTPUT")
}

// ---------------------------------------------------
// Defaults

func defaultVerbose() bool {
	return setDefaultBool("ERIS_CHAINMANAGER_VERBOSE", false)
}

func defaultDebug() bool {
	return setDefaultBool("ERIS_CHAINMANAGER_DEBUG", false)
}

func defaultOutput() bool {
	return setDefaultBool("ERIS_CHAINMANAGER_OUTPUT", true)
}

func setDefaultBool(envVar string, def bool) bool {
	env := os.Getenv(envVar)
	if env != "" {
		i, _ := strconv.ParseBool(env)
		return i
	}
	return def
}

func setDefaultString(envVar, def string) string {
	env := os.Getenv(envVar)
	if env != "" {
		return env
	}
	return def
}

func setDefaultStringSlice(envVar string, def []string) []string {
	env := os.Getenv(envVar)
	if env != "" {
		return strings.Split(env, ",")
	}
	return def
}
