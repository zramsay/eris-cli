package util

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func UpdateEris(branch string) {

	//check that git/go are installed
	CheckGitAndGo(true, true)

	//checks for deprecated dir names and renames them
	err := MigrateDeprecatedDirs(common.DirsToMigrate, false) // false = no prompt
	if err != nil {
		logger.Printf("directory migration error: %v\ncontinuing with update without migration\n", err)
	}

	//change pwd to eris/cli
	ChangeDirectory()

	if branch == "" {
		branch = "master"
	}

	CheckoutBranch(branch)
	PullBranch(branch)

	InstallEris()
	ver := version() //because version.Version will be in RAM.

	logger.Printf("The marmots have updated eris successfully.\n%s\n", ver)
}

func CheckGitAndGo(git, gO bool) {
	if git {
		stdOut1, err := exec.Command("git", "version").CombinedOutput()
		if err != nil {
			logger.Printf("ensure you have git installed:\n%v\n", string(stdOut1))
			logger.Println("or if running `eris init` use the --skip-pull flag")
			os.Exit(1)

		}
	}
	if gO {
		stdOut2, err := exec.Command("go", "version").CombinedOutput()
		if err != nil {
			logger.Printf("ensure you have go installed:\n%s\n", string(stdOut2))
			os.Exit(1)
		}
	}
}

func ChangeDirectory() {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		fmt.Printf("You do not have $GOPATH set. Please make sure this is set and rerun the command.\n")
		os.Exit(1)
	}

	dir := path.Join(goPath, "src/github.com/eris-ltd/eris-cli/")
	err := os.Chdir(dir)

	if err != nil {
		fmt.Printf("error changing directory\n%v\n", err)
		os.Exit(1)
	}

	logger.Debugf("directory changed to:\n%s\n", dir)
}

func CheckoutBranch(branch string) {
	checkoutArgs := []string{"checkout", branch}

	stdOut, err := exec.Command("git", checkoutArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("error checking out %s:\n%s\n", branch, string(stdOut))
		os.Exit(1)
	}

	logger.Debugf("%s checked-out\n", branch)
}

func PullBranch(branch string) {
	pullArgs := []string{"pull", "origin", branch}

	stdOut, err := exec.Command("git", pullArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("error pulling from github:\n%s\n", string(stdOut))
		os.Exit(1)
	}

	logger.Debugf("%s pulled successfully\n", branch)
}

func InstallEris() {
	goArgs := []string{"install", "./cmd/eris"}

	stdOut, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("error with go install ./cmd/eris:\n%s\n", string(stdOut))
		os.Exit(1)
	}

	logger.Debugf("Go install worked correctly.\n")
}

func version() string {
	verArgs := []string{"version"}

	stdOut, err := exec.Command("eris", verArgs...).CombinedOutput()
	if err != nil {
		fmt.Printf("error getting version:\n%s\n", string(stdOut))
		os.Exit(1)
	}
	return string(stdOut)

}
