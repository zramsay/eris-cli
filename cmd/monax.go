package commands

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/initialize"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"
	"github.com/monax/monax/version"

	"github.com/spf13/cobra"
)

const VERSION = version.VERSION
const dVerMin = version.DOCKER_VER_MIN

// Defining the root command
var MonaxCmd = &cobra.Command{
	Use:   "monax COMMAND [FLAG ...]",
	Short: "The Ecosystem Application Platform",
	Long: `Monax is an application platform for building, testing, maintaining, and operating applications built to run on an ecosystem level.

Made with <3 by Monax Industries.

Complete documentation is available at https://monax.io/docs
` + "\nVersion:\n  " + VERSION,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.WarnLevel)
		if do.Verbose {
			log.SetLevel(log.InfoLevel)
		} else if do.Debug {
			log.SetLevel(log.DebugLevel)
		}

		// Don't try to connect to Docker for informational
		// or bug fixing commands.
		switch cmd.Use {
		case "version", "update", "man":
			return
		}

		util.DockerConnect(do.Verbose, "default") // see note at the function

		if os.Getenv("TEST_ON_WINDOWS") == "true" || os.Getenv("TEST_ON_MACOSX") == "true" {
			return
		}

		if !util.DoesDirExist(config.MonaxRoot) && cmd.Use != "init" {
			log.Warn("Monax root directory doesn't exist. The marmots will initialize it for you")
			do := definitions.NowDo()
			do.Yes = true
			do.Pull = false
			do.Quiet = true
			if err := initialize.Initialize(do); err != nil {
				log.Errorf("Error: couldn't initialize the Monax root directory: %v", err)
			}

			log.Warn()
		}

		// Compare Docker client API versions.
		dockerVersion, err := util.DockerClientVersion()
		if err != nil {
			util.IfExit(fmt.Errorf("There was an error connecting to your Docker daemon.\nCome back after you have resolved the issue and the marmots will be happy to service your blockchain management needs: %v", util.DockerError(err)))
		}
		marmot := "Come back after you have upgraded and the marmots will be happy to service your blockchain management needs"
		if !util.CompareVersions(dockerVersion, dVerMin) {
			util.IfExit(fmt.Errorf("Monax requires [docker] version >= %v\nThe marmots have detected [docker] version: %v\n%s", dVerMin, dockerVersion, marmot))
		}
	},

	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	InitializeConfig()
	AddGlobalFlags()
	AddCommands()
	util.IfExit(MonaxCmd.Execute())
}

// Define the commands
func AddCommands() {
	buildServicesCommand()
	MonaxCmd.AddCommand(Services)
	buildChainsCommand()
	MonaxCmd.AddCommand(Chains)
	buildPackagesCommand()
	MonaxCmd.AddCommand(Packages)
	buildKeysCommand()
	MonaxCmd.AddCommand(Keys)
	buildListCommand()
	MonaxCmd.AddCommand(List)
	buildCleanCommand()
	MonaxCmd.AddCommand(Clean)
	buildInitCommand()
	MonaxCmd.AddCommand(Init)
	buildVerSionCommand()
	MonaxCmd.AddCommand(VerSion)

	if runtime.GOOS != "windows" {
		buildManCommand()
		MonaxCmd.AddCommand(ManPage)
	}

	MonaxCmd.SetHelpCommand(Help)
	MonaxCmd.SetHelpTemplate(helpTemplate)
}

// Global Do struct
var do *definitions.Do

// Flags that are to be used by commands are handled by the Do struct
// Define the persistent commands (globals)
func AddGlobalFlags() {
	MonaxCmd.PersistentFlags().BoolVarP(&do.Verbose, "verbose", "v", false, "verbose output")
	MonaxCmd.PersistentFlags().BoolVarP(&do.Debug, "debug", "d", false, "debug level output")
}

func InitializeConfig() {
	var (
		err    error
		stdout io.Writer
		stderr io.Writer
	)

	do = definitions.NowDo()

	if os.Getenv("MONAX_CLI_WRITER") != "" {
		stdout, err = os.Open(os.Getenv("MONAX_CLI_WRITER"))
		if err != nil {
			log.Errorf("Could not open: %v", err)
			return
		}
	} else {
		stdout = os.Stdout
	}

	if os.Getenv("MONAX_CLI_ERROR_WRITER") != "" {
		stderr, err = os.Open(os.Getenv("MONAX_CLI_ERROR_WRITER"))
		if err != nil {
			log.Errorf("Could not open: %v", err)
			return
		}
	} else {
		stderr = os.Stderr
	}

	config.Global, err = config.New(stdout, stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ArgCheck(num int, comp string, cmd *cobra.Command, args []string) error {
	switch comp {
	case "eq":
		if len(args) != num {
			cmd.Help()
			return fmt.Errorf("\n**Note** you sent our marmots the wrong number of arguments.\nPlease send the marmots %d arguments only.", num)
		}
	case "ge":
		if len(args) < num {
			cmd.Help()
			return fmt.Errorf("\n**Note** you sent our marmots the wrong number of arguments.\nPlease send the marmots at least %d argument(s).", num)
		}
	}
	return nil
}

//restrict flag behaviour when needed (rare but used sometimes)
func FlagCheck(num int, comp string, cmd *cobra.Command, flags []string) error {
	switch comp {
	case "eq":
		if len(flags) != num {
			cmd.Help()
			return fmt.Errorf("\n**Note** you sent our marmots the wrong number of flags.\nPlease send the marmots %d flags only.", num)
		}
	case "ge":
		if len(flags) < num {
			cmd.Help()
			return fmt.Errorf("\n**Note** you sent our marmots the wrong number of flags.\nPlease send the marmots at least %d flag(s).", num)
		}
	}
	return nil
}
