package data

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/ebuchman/go-shell-pipes"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func ImportDataRaw(name string, containerNumber int) error {
	if parseKnown(name, containerNumber) {

		containerName := nameToContainerName(name, containerNumber)
		importPath := filepath.Join(DataContainersPath, name)

		// temp until docker cp works both ways.
		os.Chdir(importPath)
		cmd := "tar chf - . | docker run -i --rm --volumes-from " + containerName + " eris/data tar xf - -C /home/eris/.eris"
		s, err := pipes.RunString(cmd)
		if err != nil {
			cmd := "tar chf - . | docker run -i --volumes-from " + containerName + " eris/data tar xf - -C /home/eris/.eris"
			s2, e2 := pipes.RunString(cmd)
			if e2 != nil {
				return fmt.Errorf("Could not import the data container.\n\nTried with docker --rm: %v.\n Tried without docker --rm: %v.\n", err, e2)
			}
			logger.Infoln(s2)
		} else {
			logger.Infoln(s)
		}
	} else {
		if err := perform.DockerCreateDataContainer(name, containerNumber); err != nil {
			return fmt.Errorf("Error creating data container %v.", err)
		}
		return ImportDataRaw(name, containerNumber)
	}

	return nil
}

func ExecDataRaw(name string, containerNumber int, interactive bool, args []string) error {
	if parseKnown(name, containerNumber) {
		name = nameToContainerName(name, containerNumber)
		logger.Infoln("Running exec on container with volumes from data container " + name)
		if err := perform.DockerRunVolumesFromContainer(name, interactive, args); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func ExportDataRaw(name string, containerNumber int) error {
	if parseKnown(name, containerNumber) {
		logger.Infoln("Exporting data container" + name)

		exportPath := filepath.Join(DataContainersPath, name) // TODO: containerNumber ?
		_, ops := MockService(name, containerNumber)
		service, exists := perform.ContainerExists(ops)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}
		logger.Infoln("Service ID: " + service.ID)

		cont, err := util.DockerClient.InspectContainer(service.ID)
		if err != nil {
			return err
		}

		reader, writer := io.Pipe()
		opts := docker.CopyFromContainerOptions{
			OutputStream: writer,
			Container:    service.ID,
			Resource:     "/home/eris/.eris/",
		}

		go func() {
			IfExit(util.DockerClient.CopyFromContainer(opts))
			writer.Close()
		}()

		err = util.Untar(reader, name, exportPath)
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
