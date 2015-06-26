package data

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/ebuchman/go-shell-pipes"
	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------

func Import(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(ImportDataRaw(args[0]))
}

func Export(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(ExportDataRaw(args[0]))
}

func Exec(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	srv := args[0]

	// if interactive, we ignore args. if not, run args as command
	interactive := cmd.Flags().Lookup("interactive").Changed
	if !interactive {
		if len(args) < 2 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
		args = args[1:]
		if len(args) == 1 {
			args = strings.Split(args[0], " ")
		}
	}

	IfExit(ExecDataRaw(srv, interactive, args))
}

//----------------------------------------------------

func ImportDataRaw(name string) error {
	if parseKnown(name) {

		containerName := nameToContainerName(name)
		importPath := filepath.Join(DataContainersPath, name)

		// temp until docker cp works both ways.
		os.Chdir(importPath)
		cmd := "tar chf - . | docker run -i --rm --volumes-from " + containerName + " eris/data tar xf - -C /home/eris/.eris"

		s, err := pipes.RunString(cmd)
		if err != nil {
			return err
		}

		logger.Infoln(s)
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	return nil
}

func ExecDataRaw(name string, interactive bool, args []string) error {
	if parseKnown(name) {
		containerNumber := 1
		name = "eris_data_" + name + "_" + strconv.Itoa(containerNumber)
		logger.Infoln("Running exec on container with volumes from data container for " + name)
		if err := perform.DockerRunVolumesFromContainer(name, interactive, args); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func ExportDataRaw(name string) error {
	if parseKnown(name) {
		logger.Infoln("Exporting data container" + name)

		exportPath := filepath.Join(DataContainersPath, name)
		_, ops := mockService(name)
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
