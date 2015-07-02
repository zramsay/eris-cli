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
		fmt.Println(cmd)

		s, err := pipes.RunString(cmd)
		if err != nil {
			return err
		}

		logger.Infoln(s)
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
		logger.Infoln("Running exec on container with volumes from data container for " + name)
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

		var dest string
		vol := cont.Volumes
		for k, v := range vol {
			if k == "/home/eris/.eris" {
				dest = filepath.Base(v)
			}
		}
		dest = filepath.Join(exportPath, dest)

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

		var toMove []string
		os.Chdir(exportPath)
		toMove, err = filepath.Glob(filepath.Join(dest, "*"))
		if err != nil {
			return err
		}

		for _, f := range toMove {
			err = os.Rename(f, filepath.Join(exportPath, filepath.Base(f)))
			if err != nil {
				return err
			}
		}
		err = os.RemoveAll(dest)
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	return nil
}
