package commands

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/version"

	"github.com/spf13/cobra"
)

const VERSION = version.VERSION
const dVerMin = version.DOCKER_VER_MIN
const dmVerMin = version.DM_VER_MIN

// Defining the root command
var ErisCmd = &cobra.Command{
	Use:   "eris COMMAND [FLAG ...]",
	Short: "The Ecosystem Application Platform",
	Long: `Eris is an application platform for building, testing, maintaining, and operating applications built to run on an ecosystem level.

Made with <3 by Monax Industries.

Complete documentation is available at https://monax.io/docs/documentation
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

		util.DockerConnect(do.Verbose, do.MachineName)
		util.IpfsHost = config.Global.IpfsHost
		util.IpfsPort = config.Global.IpfsPort

		if os.Getenv("TEST_ON_WINDOWS") == "true" || os.Getenv("TEST_ON_MACOSX") == "true" {
			return
		}

		if !util.DoesDirExist(config.ErisRoot) && cmd.Use != "init" {
			log.Warn("Eris root directory doesn't exist. The marmots will initialize it for you")
			do := definitions.NowDo()
			do.Yes = true
			do.Pull = false
			do.Quiet = true
			if err := initialize.Initialize(do); err != nil {
				log.Errorf("Error: couldn't initialize the Eris root directory: %v", err)
			}

			if err := config.Save(&config.Global.Settings); err != nil {
				log.Error(err)
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
			util.IfExit(fmt.Errorf("Eris requires [docker] version >= %v\nThe marmots have detected [docker] version: %v\n%s", dVerMin, dockerVersion, marmot))
		}
		log.AddHook(util.CrashReportHook(dockerVersion))

		// Compare `docker-machine` versions but don't fail if not installed.
		dmVersion, err := util.DockerMachineVersion()
		if err != nil {
			log.Info("The marmots could not find [docker-machine] installed. While it is not required to be used with Eris, we strongly recommend it be installed for maximum blockchain awesomeness")
		} else if !util.CompareVersions(dmVersion, dmVerMin) {
			util.IfExit(fmt.Errorf("Eris requires [docker-machine] version >= %v\nThe marmots have detected version: %v\n%s", dmVerMin, dmVersion, marmot))
		}
	},

	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
}

func Execute() {
	// Handle panics within Execute().
	defer func() {
		if err := recover(); err != nil {
			util.SendPanic(err)
		}
	}()

	InitializeConfig()
	AddGlobalFlags()
	AddCommands()
	util.IfExit(ErisCmd.Execute())
}

// Define the commands
func AddCommands() {
	buildServicesCommand()
	ErisCmd.AddCommand(Services)
	buildChainsCommand()
	ErisCmd.AddCommand(Chains)
	buildPackagesCommand()
	ErisCmd.AddCommand(Packages)
	buildKeysCommand()
	ErisCmd.AddCommand(Keys)
	buildFilesCommand()
	ErisCmd.AddCommand(Files)
	buildDataCommand()
	ErisCmd.AddCommand(Data)
	buildListCommand()
	ErisCmd.AddCommand(List)
	//buildAgentsCommand()
	//ErisCmd.AddCommand(Agents)
	buildCleanCommand()
	ErisCmd.AddCommand(Clean)
	buildInitCommand()
	ErisCmd.AddCommand(Init)
	buildVerSionCommand()
	ErisCmd.AddCommand(VerSion)

	if runtime.GOOS != "windows" {
		buildManCommand()
		ErisCmd.AddCommand(ManPage)
	}

	ErisCmd.SetHelpCommand(Help)
	ErisCmd.SetHelpTemplate(helpTemplate)
}

// Global Do struct
var do *definitions.Do

// Flags that are to be used by commands are handled by the Do struct
// Define the persistent commands (globals)
func AddGlobalFlags() {
	ErisCmd.PersistentFlags().BoolVarP(&do.Verbose, "verbose", "v", false, "verbose output")
	ErisCmd.PersistentFlags().BoolVarP(&do.Debug, "debug", "d", false, "debug level output")
	ErisCmd.PersistentFlags().StringVarP(&do.MachineName, "machine", "m", "eris", "machine name for docker-machine that is running VM")
}

func InitializeConfig() {
	var (
		err    error
		stdout io.Writer
		stderr io.Writer
	)

	do = definitions.NowDo()

	if os.Getenv("ERIS_CLI_WRITER") != "" {
		stdout, err = os.Open(os.Getenv("ERIS_CLI_WRITER"))
		if err != nil {
			log.Errorf("Could not open: %v", err)
			return
		}
	} else {
		stdout = os.Stdout
	}

	if os.Getenv("ERIS_CLI_ERROR_WRITER") != "" {
		stderr, err = os.Open(os.Getenv("ERIS_CLI_ERROR_WRITER"))
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
