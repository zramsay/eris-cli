package data

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/ebuchman/go-shell-pipes"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func ImportData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {

		containerName := util.DataContainersName(do.Name, do.Operations.ContainerNumber)
		importPath := filepath.Join(DataContainersPath, do.Name)

		// temp until docker cp works both ways.
		logger.Debugf("Importing FROM =>\t\t%s\n", importPath)
		os.Chdir(importPath)
		// TODO [eb]: deal with hardcoded user
		// TODO [csk]: drop the whole damn cmd call
		//         use go's tar lib to make a tarball of the directory
		//         read the tar file into an io.Reader
		//         start a container with its Stdin open, connect to an io.Writer
		//         connect them up with io.Pipe
		//         this will free us from any quirks that the cli has
		if do.Path != "" {
			do.Path = do.Path
		} else {
			do.Path = "/home/eris/.eris"
		}

		logger.Debugf("Importing TO =>\t\t\t%s\n", do.Path)
		cmd := "tar chf - . | docker run -i --rm --volumes-from " + containerName + " --user eris eris/data tar xf - -C " + do.Path
		_, err := pipes.RunString(cmd)
		if err != nil {
			cmd := "tar chf - . | docker run -i --volumes-from " + containerName + " --user eris eris/data tar xf - -C " + do.Path
			_, e2 := pipes.RunString(cmd)
			if e2 != nil {
				return fmt.Errorf("Could not import the data container.\nTried with docker --rm =>\t%v\nTried without docker --rm =>\t%v", err, e2)
			}
		}
	} else {
		if err := perform.DockerCreateDataContainer(do.Name, do.Operations.ContainerNumber); err != nil {
			return fmt.Errorf("Error creating data container %v.", err)
		}
		return ImportData(do)
	}
	do.Result = "success"
	return nil
}

func ExecData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		do.Name = util.DataContainersName(do.Name, do.Operations.ContainerNumber)
		logger.Infoln("Running exec on container with volumes from data container " + do.Name)
		if err := perform.DockerRunVolumesFromContainer(do.Name, do.Interactive, do.Args); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	do.Result = "success"
	return nil
}

func ExportData(do *definitions.Do) error {
	if util.IsDataContainer(do.Name, do.Operations.ContainerNumber) {
		logger.Infoln("Exporting data container", do.Name)

		exportPath := filepath.Join(DataContainersPath, do.Name) // TODO: do.Operations.ContainerNumber ?
		srv := PretendToBeAService(do.Name, do.Operations.ContainerNumber)

		service, exists := perform.ContainerExists(srv.Operations)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}
		logger.Infoln("Service ID: " + service.ID)

		cont, err := util.DockerClient.InspectContainer(service.ID)
		if err != nil {
			return err
		}

		reader, writer := io.Pipe()

		if do.Path != "" {
			do.Path = do.Path
		} else {
			do.Path = "/home/eris/.eris"
		}
		opts := docker.CopyFromContainerOptions{
			OutputStream: writer,
			Container:    service.ID,
			Resource:     do.Path,
		}

		go func() {
			IfExit(util.DockerClient.CopyFromContainer(opts))
			writer.Close()
		}()

		err = util.Untar(reader, do.Name, exportPath)
		if err != nil {
			return err
		}

		// docker actually exports to a `_data` folder for volumes
		//   this section of the function moves whatever docker dumps
		//   into exportPath/_data into export. ranging through the
		//   volumes is probably overkill as we could just assume
		//   that it *was* `_data` but in case docker changes later
		//   we'll just keep it for now.
		os.Chdir(exportPath)
		var unTarDestination string
		for k, v := range cont.Volumes {
			if k == "/home/eris/.eris" {
				unTarDestination = filepath.Base(v)
			}
		}
		if err := moveOutOfDirAndRmDir(filepath.Join(exportPath, unTarDestination), exportPath); err != nil {
			return err
		}

		// now if docker dumps to exportPath/.eris we should remove
		//   move everything from .eris to exportPath
		if err := moveOutOfDirAndRmDir(filepath.Join(exportPath, ".eris"), exportPath); err != nil {
			return err
		}

	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	do.Result = "success"
	return nil
}

func moveOutOfDirAndRmDir(src, dest string) error {
	logger.Debugln("\nMove all files/dirs out of a dir and rm -rf that dir.")
	logger.Debugf("Source of the move:\t%s.\n", src)
	logger.Debugf("Destin of the move:\t%s.\n\n", dest)
	toMove, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	if len(toMove) == 0 {
		logger.Debugln("No files to move.")
	}

	for _, f := range toMove {
		logger.Debugf("Moving file [%s] to [%s].\n", f, filepath.Join(dest, filepath.Base(f)))
		err = os.Rename(f, filepath.Join(dest, filepath.Base(f)))
		if err != nil {
			return err
		}
	}

	logger.Debugf("\nRemoving directory:\t%s.\n", src)
	err = os.RemoveAll(src)
	if err != nil {
		return err
	}

	return nil
}
