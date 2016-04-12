package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"
	"github.com/eris-ltd/eris-cli/perform"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func UpdateEris(do *definitions.Do) error {
	
	log.Warn("building eris bin container with branch:")
	log.Warn(do.Branch)
	binPath := "" //get from stuff below
	if err := BuildErisBinContainer(do.Branch, binPath); err != nil {
		return err
	}

/*	whichEris, err := GoOrBinary()
	if err != nil {
		return err
	}
	// TODO check flags!

	if whichEris == "go" {
		hasGit, hasGo := CheckGitAndGo(true, true)
		if !hasGit || !hasGo {
			return fmt.Errorf("either git or go is not installed. both are required for non-binary update")
		}
		if err := UpdateErisGo(do); err != nil {
			return err
		}
	} else if whichEris == "binary" {
		if err := UpdateErisBinary(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("The marmots could not figure out how eris was installed")
	}*/

	//checks for deprecated dir names and renames them
	// false = no prompt
	if err := util.MigrateDeprecatedDirs(common.DirsToMigrate, false); err != nil {
		log.Warn(fmt.Sprintf("Directory migration error: %v\nContinuing update without migration", err))
	}
	log.Warn("Eris update successful. Please re-run `eris init`.")
	return nil
}

func BuildErisBinContainer(branch, binaryPath string) error {
	// quay.io does not parse!
	//dTest := fmt.Sprintf("FROM base\nMAINTAINER Eris Industries <support@erisindustries.com>\n")
	dockerfile := `FROM base
MAINTAINER Eris Industries <support@erisindustries.com>

ENV NAME         eris-cli
ENV REPO 	 eris-ltd/$NAME
ENV BRANCH       ` + branch + `
ENV CLONE_PATH   $GOPATH/src/github.com/eris-ltd/eris-cli

RUN mkdir --parents $CLONE_PATH

RUN git clone -q https://github.com/$REPO $CLONE_PATH
RUN cd $CLONE_PATH && git checkout -q $BRANCH
RUN cd $CLONE_PATH/cmd/eris && go build -o $INSTALL_BASE/eris

CMD ["/bin/bash"]`

	//log.Warn(dockerfile)
	if err := perform.DockerBuild(dockerfile); err != nil {
		return err
	}
	
	doUpdate := definitions.NowDo()
	doUpdate.Operations.Args = []string{"update"}
	
	if err := services.StartService(doUpdate); err != nil {
		return nil
	}

	doCp := definitions.NowDo()
	doCp.Name = "update"

	//$INSTALL_BASE/eris
	doCp.Source = "/usr/local/bin/eris"
	doCp.Destination = binaryPath
	if err := data.ExportData(doCp); err != nil {
		return err
	}

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
