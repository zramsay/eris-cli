package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/eris-ltd/eris-logger"
	"github.com/eris-ltd/common/go/common"
)

func ChangeDirectoryToCLI() error {
	dir := filepath.Join(common.ErisGo, "eris-cli")
	if err := os.Chdir(dir); err != nil {
		return err
	}
	log.WithField("dir", dir).Debug("Directory changed to")

	return nil
}

func CheckoutBranch(branch string) error {
	checkoutArgs := []string{"checkout", branch}

	stdOut, err := exec.Command("git", checkoutArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(stdOut))
	}
	log.WithField("branch", branch).Debug("Branch checked-out")

	return nil
}

func PullBranch(branch string) error {
	pullArgs := []string{"pull", "origin", branch}

	stdOut, err := exec.Command("git", pullArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(stdOut))
	}
	log.WithField("branch", branch).Debug("Branch pulled successfully")

	return nil
}

func InstallErisGo() error {
	goArgs := []string{"install", "./cmd/eris"}

	stdOut, err := exec.Command("go", goArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(stdOut))
	}
	log.Debug("go install worked correctly")

	return nil
}

func version() (string, error) {
	verArgs := []string{"version"}

	stdOut, err := exec.Command("eris", verArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(string(stdOut))
	}

	return string(stdOut), nil
}
