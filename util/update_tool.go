package util

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/eris-ltd/common/go/common"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/definitions"
)

func UpdateErisGo(do *definitions.Do) error {
	// TODO handle errors!
	// cleaner ch dir functionality
	//change pwd to eris/cli
	ChangeDirectory("src")

	if do.Branch == "" {
		do.Branch = "master"
	}

	//TODO fix proper: add commit & version!
	// do it all in these funcs
	CheckoutBranch(do.Branch)
	PullBranch(do.Branch)

	InstallErisGo()
	ver := version() //because version.Version will be in RAM.

	log.WithField("=>", ver).Warn("The marmots have updated Eris successfully")
	return nil
}

func UpdateErisBinary() error {
	ChangeDirectory("bin")
	_, err := DownloadLatestBinaryRelease()
	return err
}

func CheckGitAndGo(git, gO bool) (bool, bool) {
	hasGit := false
	hasGo := false
	if git {
		stdOut1, err := exec.Command("git", "version").CombinedOutput()
		if err != nil {
			log.WithField("version", string(stdOut1)).Warn("Ensure you have git installed.")
		} else {
			hasGit = true
		}
	}
	if gO {
		stdOut2, err := exec.Command("go", "version").CombinedOutput()
		if err != nil {
			log.WithField("version", string(stdOut2)).Warn("Ensure you have Go installed.")
		} else {
			hasGo = true
		}
	}
	return hasGit, hasGo
}

func ChangeDirectory(to string) {
	if to == "bin" {
		erisLoc, err := exec.LookPath("eris")
		if err != nil {
			log.Fatalf("Error finding eris binary: %v", err)
		}
		err = os.Chdir(filepath.Dir(erisLoc))
		if err != nil {
			log.Fatalf("Error changing directory: %v", err)
		}
		log.WithField("dir", erisLoc).Debug("Directory changed to")
	} else if to == "src" {
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			log.Fatal("You do not have $GOPATH set. Please make sure this is set and rerun the command.")
		}
		// TODO use from common!
		dir := filepath.Join(common.ErisGo, "eris-cli")
		err := os.Chdir(dir)

		if err != nil {
			log.Fatalf("Error changing directory: %v", err)
		}
		log.WithField("dir", dir).Debug("Directory changed to")
	}
}
