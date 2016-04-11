package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func UpdateEris(do *definitions.Do) error {
	whichEris, err := GoOrBinary()
	if err != nil {
		return err
	}
	// TODO check flags!

	if whichEris == "go" {
		hasGit, hasGo := util.CheckGitAndGo(true, true)
		if !hasGit || !hasGo {
			return fmt.Errorf("either git or go is not installed. both are required for non-binary update")
		}
		if err := util.UpdateErisGo(do); err != nil {
			return err
		}
	} else if whichEris == "binary" {
		if err := util.UpdateErisBinary(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("The marmots could not figure out how eris was installed")
	}

	//checks for deprecated dir names and renames them
	// false = no prompt
	if err := util.MigrateDeprecatedDirs(common.DirsToMigrate, false); err != nil {
		log.Warn(fmt.Sprintf("Directory migration error: %v\nContinuing update without migration", err))
	}
	log.Warn("Eris update successful. Please re-run `eris init`.")
	return nil
}

func GoOrBinary() (string, error) {
	which, err := exec.Command("which", "eris").CombinedOutput()
	if err != nil {
		return "", err
	}

	toCheck := strings.Split(string(which), "/")
	length := len(toCheck)
	usr := toCheck[length-3]
	bin := util.TrimString(toCheck[length-2])
	eris := util.TrimString(toCheck[length-1]) //sometimes ya just gotta trim

	gopath := filepath.Join(os.Getenv("GOPATH"), bin, eris)
	
	erisLook, err := exec.LookPath("eris")
	if err != nil {
		return "", err
	}

	if string(which) != erisLook {
		return "", fmt.Errorf("`which eris` returned (%s) while the exec.LookPath(`eris`) command returned (%s). these need to match", string(which), erisLook)
	}
	
	if bin == "bin" && eris == "eris" {
		if util.TrimString(gopath) == util.TrimString(string(which)) { // gotta trim those strings!
			log.Debug("looks like eris was instaled via go")
			return "go", nil
		} else if usr == "usr" { //binary check
			log.Debug("looks like eris was instaled via binary")
			// lookPath ...?
			// "usr/bin/eris"
			return "binary", nil
		}
	} else {
		return "", fmt.Errorf("could not determine how eris is installed")
	}
	return "", err
}
