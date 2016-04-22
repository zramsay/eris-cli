package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/Sirupsen/logrus"
	"github.com/eris-ltd/common/go/common"
)

func UpdateEris(do *definitions.Do) error {

	whichEris, binPath, err := GoOrBinary()
	if err != nil {
		return err
	}

	if whichEris == "go" {
		// ensure git and go are installed
		hasGit, hasGo := CheckGitAndGo(true, true)
		if !hasGit || !hasGo {
			return fmt.Errorf("either git or go is not installed. both are required for non-binary update")
		}

		log.WithField("branch", do.Branch).Warn("Building eris binary via go with:")
		if err := UpdateErisGo(do.Branch); err != nil {
			return err
		}
	} else if whichEris == "binary" {
		if err := UpdateErisViaBinary(do.Branch, binPath); err != nil {
			return err
		}

	} else {
		return fmt.Errorf("The marmots could not figure out how eris was installed. Exiting.")
	}

	//checks for deprecated dir names and renames them
	// false = no prompt
	if err := util.MigrateDeprecatedDirs(common.DirsToMigrate, false); err != nil {
		log.Warn(fmt.Sprintf("Directory migration error: %v\nCheck your eris directory.", err))
	}
	log.Warn("Eris update successful. Please re-run [eris init]")
	return nil
}

func UpdateErisViaBinary(branch, binPath string) error {

	log.WithField("branch", branch).Warn("Building Eris binary in container with:")
	if err := BuildErisBinContainer(branch, binPath); err != nil {
		return err
	}

	platform := runtime.GOOS
	if platform != "windows" {
		ver, err := version() //because version.Version will be in RAM.
		if err != nil {
			return err
		}

		log.WithField("=>", ver).Warn("The marmots have updated Eris successfully")
	} // else: { windows instructions already sent displayed }

	return nil
}

func UpdateErisGo(branch string) error {
	// all the following functions are in gitandgo.go

	// change pwd to eris-cli for the next functions
	if err := ChangeDirectoryToCLI(); err != nil {
		return err
	}

	if err := CheckoutBranch(branch); err != nil {
		return err
	}

	if err := PullBranch(branch); err != nil {
		return err
	}

	if err := InstallErisGo(); err != nil {
		return err
	}

	ver, err := version() //because version.Version will be in RAM.
	if err != nil {
		return err
	}

	log.WithField("=>", ver).Warn("The marmots have updated Eris successfully")
	return nil
}

// returns a string of each the type of installation
// and path to the binary; the latter is only used for
// non-go binary installation
func GoOrBinary() (string, string, error) {

	erisLook, err := exec.LookPath("eris")
	if err != nil {
		return "", "", err
	}

	toCheck := strings.Split(string(erisLook), "/")
	length := len(toCheck)
	bin := util.TrimString(toCheck[length-2])
	eris := util.TrimString(toCheck[length-1]) //sometimes ya just gotta trim

	gopath := filepath.Join(os.Getenv("GOPATH"), bin, eris)

	trimEris := util.TrimString(string(erisLook))

	if eris == "eris" {
		// check if eris is installed via go
		if util.TrimString(gopath) == trimEris {
			goWarn := fmt.Sprintf(`The marmots have detected a source installation located at: (%s)
Continuing will update eris to either the latest version, or the branch you've specified.
Do you wish to continue?`, trimEris)
			if util.QueryYesOrNo(goWarn) == util.Yes {
				return "go", erisLook, nil
			} else {
				return "", "", fmt.Errorf("Permission to update denied; exiting.")
			}
		} else { // eris is installed via binary
			binWarn := fmt.Sprintf(`The marmots have detected a binary installation located at: (%s)
Continuing will update eris to either the latest version, or the branch you've specified.
Do you wish to continue?`, trimEris)
			if util.QueryYesOrNo(binWarn) == util.Yes {
				return "binary", erisLook, nil
			} else {
				return "", "", fmt.Errorf("Permission to update denied; exiting.")
			}
		}
	}
	return "", "", fmt.Errorf("could not determine how eris is installed")
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
