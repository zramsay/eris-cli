package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func UpdateEris(branch string) {

	//check that git/go are installed
	CheckGitAndGo(true, true)

	//checks for deprecated dir names and renames them
	err := MigrateDeprecatedDirs(common.DirsToMigrate, false) // false = no prompt
	if err != nil {
		log.Warnf("Directory migration error: %v", err)
		log.Warn("Continuing with update without migration")
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

	log.WithField("=>", ver).Warn("The marmots have updated Eris successfully")
}

func CheckGitAndGo(git, gO bool) {
	if git {
		stdOut1, err := exec.Command("git", "version").CombinedOutput()
		if err != nil {
			log.WithField("version", string(stdOut1)).Fatal("Ensure you have git installed or if running `eris init` use the `--skip-pull` flag")
		}
	}
	if gO {
		stdOut2, err := exec.Command("go", "version").CombinedOutput()
		if err != nil {
			log.WithField("version", string(stdOut2)).Fatal("Ensure you have Go installed")
		}
	}
}

func ChangeDirectory() {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatal("You do not have $GOPATH set. Please make sure this is set and rerun the command")
	}

	dir := filepath.Join(goPath, "src", "github.com", "eris-ltd", "eris-cli")
	err := os.Chdir(dir)

	if err != nil {
		log.Fatalf("Error changing directory: %v")
	}

	log.WithField("dir", dir).Debug("Directory changed to")
}

func CheckoutBranch(branch string) {
	checkoutArgs := []string{"checkout", branch}

	stdOut, err := exec.Command("git", checkoutArgs...).CombinedOutput()
	if err != nil {
		log.WithField("branch", branch).Fatalf("Error checking out branch: %v", string(stdOut))
	}

	log.WithField("branch", branch).Debug("Branch checked-out")
}

func PullBranch(branch string) {
	pullArgs := []string{"pull", "origin", branch}

	stdOut, err := exec.Command("git", pullArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error pulling from GitHub: %v", string(stdOut))
	}

	log.WithField("branch", branch).Debug("Branch pulled successfully")
}

func InstallEris() {
	goArgs := []string{"install", "./cmd/eris"}

	stdOut, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		log.Fatalf("Error with go install ./cmd/eris: %v", string(stdOut))
	}

	log.Debug("Go install worked correctly")
}

func version() string {
	verArgs := []string{"version"}

	stdOut, err := exec.Command("eris", verArgs...).CombinedOutput()
	if err != nil {
		common.IfExit(fmt.Errorf("error getting version:\n%s\n", string(stdOut)))
	}
	return string(stdOut)

}
