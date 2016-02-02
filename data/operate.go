package data

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/loaders"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// ImportData does what it says. It imports from a host's Source to a Dest
// in a data container. It returns an error.
//
//  do.Name                       - name of the data container to use (required)
//  do.Operations.ContainerNumber - container number (optional)
//  do.Source                     - directory which should be imported (required)
//  do.Destination                - directory to _unload_ the payload into (required)
//
func ImportData(do *definitions.Do) error {
	log.WithFields(log.Fields{
		"from": do.Source,
		"to":   do.Destination,
	}).Debug("Importing")
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {

		srv := PretendToBeAService(do.Name, do.Operations.ContainerNumber)
		service, exists := perform.ContainerExists(srv.Operations)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}
		if err := checkErisContainerRoot(do, "import"); err != nil {
			return err
		}

		containerName := util.DataContainersName(do.Name, do.Operations.ContainerNumber)
		os.Chdir(do.Source)

		reader, err := util.Tar(do.Source, 0)
		if err != nil {
			return err
		}
		defer reader.Close()

		opts := docker.UploadToContainerOptions{
			InputStream:          reader,
			Path:                 do.Destination,
			NoOverwriteDirNonDir: true,
		}

		log.WithField("=>", containerName).Info("Copying into container")
		log.WithField("path", do.Source).Debug()
		if err := util.DockerClient.UploadToContainer(service.ID, opts); err != nil {
			return err
		}

		doChown := definitions.NowDo()
		doChown.Operations.DataContainerName = containerName
		doChown.Operations.ContainerType = "data"
		doChown.Operations.ContainerNumber = do.Operations.ContainerNumber
		//required b/c `docker cp` (UploadToContainer) goes in as root
		doChown.Operations.Args = []string{"chown", "--recursive", "eris", do.Destination}
		_, err = perform.DockerRunData(doChown.Operations, nil)
		if err != nil {
			return fmt.Errorf("Error changing owner: %v\n", err)
		}
	} else {
		log.WithField("name", do.Name).Info("Data container does not exist.")
		ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
		if err := perform.DockerCreateData(ops); err != nil {
			return fmt.Errorf("Error creating data container %v.", err)
		}
		return ImportData(do)
	}
	do.Result = "success"
	return nil
}

func ExecData(do *definitions.Do) (buf *bytes.Buffer, err error) {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		log.WithField("=>", do.Operations.DataContainerName).Info("Executing data container")

		ops := loaders.LoadDataDefinition(do.Name, do.Operations.ContainerNumber)
		util.Merge(ops, do.Operations)
		buf, err = perform.DockerExecData(ops, nil)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("The marmots cannot find that data container.\nPlease check the name of the data container with [eris data ls].")
	}
	do.Result = "success"
	return buf, nil
}

//export from: do.Source(in container), to: do.Destination(on host)
func ExportData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		log.WithField("=>", do.Name).Info("Exporting data container")

		// we want to export to a temp directory.
		exportPath, err := ioutil.TempDir(os.TempDir(), do.Name) // TODO: do.Operations.ContainerNumber ?
		defer os.Remove(exportPath)
		if err != nil {
			return err
		}

		containerName := util.DataContainersName(do.Name, do.Operations.ContainerNumber)
		srv := PretendToBeAService(do.Name, do.Operations.ContainerNumber)
		service, exists := perform.ContainerExists(srv.Operations)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}

		reader, writer := io.Pipe()
		defer reader.Close()

		if err := checkErisContainerRoot(do, "export"); err != nil {
			return err
		}
		opts := docker.DownloadFromContainerOptions{
			OutputStream: writer,
			Path:         do.Source,
		}

		go func() {
			log.WithField("=>", containerName).Info("Copying out of container")
			log.WithField("path", do.Source).Debug()
			IfExit(util.DockerClient.DownloadFromContainer(service.ID, opts)) // TODO: be smarter about catching this error
			writer.Close()
		}()

		log.WithField("=>", exportPath).Debug("Untarring package from container")
		if err = util.Untar(reader, do.Name, exportPath); err != nil {
			return err
		}

		// now if docker dumps to exportPath/.eris we should remove
		// move everything from .eris to exportPath
		if err := moveOutOfDirAndRmDir(filepath.Join(exportPath, ".eris"), exportPath); err != nil {
			return err
		}

		// finally remove everything in the data directory and move
		// the temp contents there
		if _, err := os.Stat(do.Destination); os.IsNotExist(err) {
			if e2 := os.MkdirAll(do.Destination, 0755); e2 != nil {
				return fmt.Errorf("Error:\tThe marmots could neither find, nor had access to make the directory: (%s)\n", do.Destination)
			}
		}
		if err := moveOutOfDirAndRmDir(exportPath, do.Destination); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	do.Result = "success"
	return nil
}

//TODO test that this doesn't fmt things up, see note in #400
func moveOutOfDirAndRmDir(src, dest string) error {
	log.WithFields(log.Fields{
		"from": src,
		"to":   dest,
	}).Info("Move all files/dirs out of a dir and `rm -rf` that dir")
	toMove, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	if len(toMove) == 0 {
		log.Debug("No files to move")
	}

	for _, f := range toMove {
		// using a copy (read+write) strategy to get around swap partitions and other
		//   problems that cause a simple rename strategy to fail. it is more io overhead
		//   to do this, but for now that is preferable to alternative solutions.
		Copy(f, filepath.Join(dest, filepath.Base(f)))
	}

	log.WithField("=>", src).Info("Removing directory")
	err = os.RemoveAll(src)
	if err != nil {
		return err
	}

	return nil
}

// check path for ErisContainerRoot
// XXX this is opiniated & we may want to change in future
// for more flexibility with filesystem of data conts
func checkErisContainerRoot(do *definitions.Do, typ string) error {

	r, err := regexp.Compile(ErisContainerRoot)
	if err != nil {
		return err
	}

	switch typ {
	case "import":
		if r.MatchString(do.Destination) != true { //if not there join it
			do.Destination = path.Join(ErisContainerRoot, do.Destination)
			return nil
		} else { // matches: do nothing
			return nil
		}
	case "export":
		if r.MatchString(do.Source) != true {
			do.Source = path.Join(ErisContainerRoot, do.Source)
			return nil
		} else {
			return nil
		}
	}
	return nil
}
