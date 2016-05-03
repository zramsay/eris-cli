package errno

import (
	"errors"
	"fmt"
	"runtime"
)

// the error structure is as follows:
// top-level functions (any functioned called directly by a cmd)
// will, if err != nil { return &ErisError{code, BaseError*(ErrAny, err), "fix"} }
// note: BaseError*(ErrAny, err) can be supplemented for any Error() as required

// to prevent double ErisError returns from nested function calls,
// any second-level function must return, if err != nil {
//	- err 		=> the unformatted error
//	- ErrAny 	=> a simple formatted/hardcoded error
//	- BaseError*()	=> a custom formatted error

// ------------------ base error framework ------------------------
var (
	ErrGo = 400
	ErrDependency = 401
	ErrDocker = 402
	ErrDockerMachine = 403

	ErrErisPreRun = 500
	ErrEris = 501
)


type ErisError struct {
	Code int
	ErrMsg error
	FixMsg string
}

// Take a string error defined in this file & concates with thrown error
// TODO make BaseErrorEE
func BaseError(errMsg string, thrownError error) error {
	return fmt.Errorf(errMsg, thrownError)
}

// takes two strings & returns an error
func BaseErrorES(errMsg, thing string) error {
	return fmt.Errorf(errMsg, thing)
}

func BaseErrorESS(errMsg, thing1, thing2 string) error {
	return fmt.Errorf(errMsg, thing1, thing2)
}

func BaseErrorEE(errMsg string, err1, err2 error) error {
	return fmt.Errorf(errMsg, err1, err2)
}

func BaseErrorEI(errMsg, id string, code int) error {
	return fmt.Errorf(errMsg, id, code)
}

// takes an error and two strings & returns an error
func BaseErrorESE(errMsg, thing string, err error) error {
	return errors.New(fmt.Sprintf(errMsg, thing, err))
}

func (e *ErisError) Error() string {
	return fmt.Sprintf("error code %d/nerror: %v/Try fixing it with: %s/n", e.Code, e.ErrMsg, e.FixMsg)
}

// ------------------ somewhat general errors ---------------------
// TODO split errors by their BaseError* format
var (
	ErrNonInteractiveExec    = errors.New("non-interactive exec sessions must provide arguments to execute.")
	ErrRenaming              = errors.New("cannot rename to same name.")
	ErrCannotFindService     = errors.New("cannot find that service.") // "retry with a known service: [eris services ls --known]"
	ErrCantFindChain         = errors.New("cannot find that chain.") // "retry with a known chain: [eris chains ls --known]"
	ErrCantFindData          = errors.New("cannot find that data container.") // "retry with a known data container: [eris data ls]"
	ErrCantFindAction        = errors.New("cannot find that action.") // "retry with a known action: [eris actions ls --known]"
	ErrContainerNameNotFound = errors.New("cannot find that container name.") //
	ErrNoChainName		 = errors.New("cannot start a chain without a name") // TODO generalize this error message
	ErrStoppingContainer = "cannot stop container: %v"

	ErrInvalidPkgJSON	 = "cannot figure that package.json out." // "check that your package.json is properly formatted"
	ErrNoChainSpecified	 = "cannot start service which has a `$chain` variable without a chain specified." //fix "rerun the command either after [eris chains checkout CHAINNAME] *or* with a --chain flag."
	//duplicate with above!
	ErrNeedChainCheckedOut = errors.New("A checked out chain is needed to continue. Please check out the appropriate chain or rerun with a chain flag")
	ErrServiceNotRunning = "the requested service is not running." // "start it with [eris services start <serviceName>]"

	ErrContainerExists     = errors.New("container exists.") // "context needed"
	ErrImageNotExist       = errors.New("cannot find image.") // "check your images with [docker images]
	ErrNoFile              = errors.New("cannot find that file.") //
	ErrNoPermGiven         = errors.New("cannot continue without permission.")
	ErrNeedGitAndGo        = errors.New("cannot find git or go installed. both are required for non-binary update.")
	ErrMarmotWTF           = errors.New("cannot determine how eris was installed.")
	ErrNotLetMePull        = errors.New("cannot start a container based on an image you will not let me pull.")
	ErrMergeParameters     = errors.New("parameters are not pointers to struct")

	ErrCreatingDataCont = "cannot create data container:%v"
	ErrPermissionNotGiven = "permission to %s denied."
	ErrPathIsNotDirectory = "path (%s) is not a directory" // "please provide a path to a directory"
	ErrPathDoesNotExist = "path (%s) does not exist; please rerun command with a proper path" // "please provide a path to a directory"
	ErrBadConfigOptions = "config options should be <key>=<value> pairs. Got %s"
	ErrUnknownCatCmd = "unknown cat subcommand: %s"

	//has a duplicate some where below!
	ErrRunningCommand = "cannot run command (%s): %v"
	ErrMakingDirectory = "cannot find or permission denied to directory (%s): %v"
	ErrBadTemplate = "%stemplate error: %v"
	ErrWritingFile = "cannot add default %s: %v"

	ErrMigratingDirs = "cannot migrate directories: %v"
	ErrInitErisRoot ="cannot initialize the Eris root directory: %v"

	// these two need to be deduped
	ErrInitDefaults = "cannot instantiate default files: %v"
	ErrDropDefaults = "cannot drop default files: %v\ntoadserver may be down: re-run [eris init] with [--source=rawgit]"

	BadGatewayURL = "invalid gateway URL provided: %v"
	ErrEnsureRunningIPFS = "failed to ensure IPFS is running: %v"
	ErrCantGetFromIPFS = "cannot get from IPFS: %v"
	FixGetFromIPFS = "ensure it is running and that you can connect to it"
	ErrNoFileToExport = errors.New("cannot find file to export")
	WarnAllOrNothing = errors.New("either remove a file by hash or all of them.")

	//these have duplicates above
	ErrStartingService = "cannot start service: %v"
	ErrNoServiceGiven = errors.New("no service given")
	ErrWritingDefinitionFile = "cannot write definition file: %v"
	ErrNoImage = errors.New(`an "image" field is required in the service definition file.`)

	ErrReadingGenesisFile = "cannot read genesis file: %v"
	ErrStartingChain = "cannot start chain: %v"
	ErrReadingFromGenesisFile = "cannot read %s genesis file: %v"
	ErrExecChain = "cannot exec chain %s: %v"

	ErrListingContainers = "cannot list containers: %v"
	ErrTypeListContainers = "cannot determine the type %s to list containers for"
	ErrRemovingContainer = "cannot remove container: %v"

	ErrBadReport = "cannot print a nice report: %v"

	ParseIPFShost = "parse the URL"
	SplitHP       = "split the host and port"
	ErrConnectDockerTLS = "cannot connect to Docker Backend via TLS: %v"
	ErrStartingDockerMachine = "cannot start the newly created docker-machine: %v"
	ErrDockerWindows = "cannot add ssh.exe to PATH: %v"
	ErrConnectDockerMachine = "cannot evaluate the env vars for the %s docker-machine: %v"
	ErrConnectDockerDaemon = "cannot connect to your Docker daemon: %v\nCome back after you have resolved the issue and the marmots will be happy to service your blockchain management needs"
	ErrParseIPFS = "cannot %s for the DockerHost to populate the IPFS Host: %v" // "check that your docker-machine VM is running with [docker-machine ls]"
	ErrRmDataContainer = "cannot remove data container after executing (%v): %v"

	ErrLoadViperConfig = "cannot load viper config from that definition file. there may be an issue with the formatting of the .toml file: %v"
	ErrLoadingDefFile = "cannot load definition file: %v"
	ErrContainerExit = "container %s exited with status %d"
	ErrRunningArguments = "cannot run arguments (%s): %v"
	ErrWrongLength = "%s length !=%d"
	ErrChainMissing = "chain %s depends on chain %s but the latter is not running"
	ErrCleaningUpChain = "Tragic! Our marmots encountered an error during setupChain.\nThey also failed to cleanup after themselves (remove containers).\nFirst error: %v\nCleanup error: %v"
	ErrBadCommandLength = "**Note** you sent our marmots the wrong number of %s.\nPlease send the marmots %s"
	ErrNoDirectories = "neither deprecated (%s) or new (%s) exists." // "run [eris init] prior to [eris update]"
	ErrGitConfigUser = errors.New(`cannot find either or username and e-mail in git config settings; using "" if empty`)

)
// ----------------------------------------------------------------

func MustInstallDockerError() error {
	install := `The marmots cannot connect to Docker. Do you have Docker installed?
If not, please visit here: https://docs.docker.com/installation/`
	// for linux only
	run := `Do you have Docker running? If not, please type [sudo service docker start].
Also check that your user is in the "docker" group. If not, you can add it
using the [sudo usermod -a -G docker $USER] command or rerun as [sudo eris]`

	switch runtime.GOOS {
	case "linux":
		return fmt.Errorf("%slinux/\n\n%s", install, run)
	case "darwin":
		return fmt.Errorf("%smac/", install)
	case "windows":
		return fmt.Errorf("%swindows/", install)
	}
	return fmt.Errorf(install)
}

func ErrBadWhaleVersions(thing, verMin, verDetected string) error {
	return fmt.Errorf("Eris requires %s version >= %s\nThe marmots have detected docker version: %s\nCome back after you have upgraded and the marmots will be happy to service your blockchain management needs", thing, verMin, verDetected)
}

func ErrCheckKeysAndCerts(thing, file string, err error) error {
	return fmt.Errorf("The marmots could not find a file that was required to connect to Docker. %s\n%s\nFile needed: %s\nerror:", thing, file, err)
}

// only a warning
func ErrSettingUpChain(err error) string {
	return fmt.Sprintf("error setting up chain: %v\nCleaning up...", err)
}
