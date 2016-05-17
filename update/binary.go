package update

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/services"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/Sirupsen/logrus"
	"github.com/eris-ltd/common/go/common"
)

// branch to update in build container
// binaryPath to replace with new binary
func BuildErisBinContainer(branch, binaryPath string) error {

	dockerfile := MakeDockerfile(branch)
	imageName := "eris-binary-update:temporary-image"
	serviceName := "eris-binary-update-temporary-service"
	if err := perform.DockerBuild(imageName, dockerfile); err != nil {
		return err
	}

	// new the service for which the image has just been built
	doNew := definitions.NowDo()
	doNew.Name = serviceName
	doNew.Operations.Args = []string{imageName}
	if err := services.MakeService(doNew); err != nil {
		return err
	}

	// start the service up: binary has already been built
	doUpdate := definitions.NowDo()
	doUpdate.Operations.Args = []string{serviceName}

	if err := services.StartService(doUpdate); err != nil {
		return nil
	}

	// copy (export) the binary from serviceName's data container
	// into the scratch path to be used later
	doCp := definitions.NowDo()
	doCp.Name = serviceName
	// where the bin will go; see below
	newPath := filepath.Join(common.ScratchPath, "bin")
	//$INSTALL_BASE/eris as set by
	if runtime.GOOS == "windows" {
		doCp.Source = "/usr/local/bin/eris.exe"
	} else {
		doCp.Source = "/usr/local/bin/eris"
	}
	doCp.Destination = newPath
	doCp.Operations.SkipCheck = true
	if err := data.ExportData(doCp); err != nil {
		return err
	}

	// remove all trace of the service and its image
	doRm := definitions.NowDo()
	doRm.Operations.Args = []string{serviceName}
	doRm.RmD = true     // remove data container
	doRm.Volumes = true // remove volumes
	doRm.Force = true   // remove by force (no pesky warnings)
	doRm.File = true    // remove the service defintion file
	doRm.RmImage = true // remove the temporary image

	if err := services.RmService(doRm); err != nil {
		return err
	}

	// binaryPath comes in from function
	if err := ReplaceOldBinaryWithNew(binaryPath, filepath.Join(newPath, "eris")); err != nil {
		return err
	}

	return nil
}

// takes a new binary and replaces the old one
// prompts windows users to do manually
func ReplaceOldBinaryWithNew(oldPath, newPath string) error {

	if runtime.GOOS != "windows" {
		if err := os.Remove(oldPath); err != nil {
			return err
		}

		if err := os.Rename(newPath, oldPath); err != nil {
			return err
		}

		chmodArgs := []string{"+x", oldPath}
		stdOut, err := exec.Command("chmod", chmodArgs...).CombinedOutput()
		if err != nil {
			return fmt.Errorf(string(stdOut))
		}

	} else {
		cpString := fmt.Sprintf("%s %s", newPath, oldPath)
		log.Warn(`
To complete the update on Windows, run:

del /f ` + oldPath + `
move ` + cpString + `
`)
	}

	return nil
}

func MakeDockerfile(branch string) string {
	var dockerfile string

	// baseImage is `quay.io/eris/base`
	baseImage := path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_BASE)
	// todo: clean up Dockerfile as much as possible

	baseDockerfile := fmt.Sprintf(`
ENV NAME         eris-cli
ENV REPO 	 eris-ltd/$NAME
ENV BRANCH       %s
ENV CLONE_PATH   $GOPATH/src/github.com/$REPO
ENV GO15VENDOREXPERIMENT 1

RUN mkdir --parents $CLONE_PATH

RUN git clone --quiet https://github.com/$REPO $CLONE_PATH
`, branch)

	if branch == "master" {
		one := fmt.Sprintf("RUN cd $CLONE_PATH/cmd/eris && GOOS=%s go build -o $INSTALL_BASE/eris", runtime.GOOS)
		two := `CMD ["/bin/bash"]`
		dockerfile = fmt.Sprintf("FROM %s\n%s\n%s\n%s", baseImage, baseDockerfile, one, two)
		return dockerfile
	} else {
		one := "RUN cd $CLONE_PATH && git checkout --quiet -b $BRANCH && git pull --quiet origin $BRANCH"
		two := fmt.Sprintf("RUN cd $CLONE_PATH/cmd/eris && GOOS=%s go build -o $INSTALL_BASE/eris", runtime.GOOS)
		three := `CMD ["/bin/bash"]`
		dockerfile = fmt.Sprintf("FROM %s\n%s\n%s\n%s\n%s", baseImage, baseDockerfile, one, two, three)
		return dockerfile
	}
}
