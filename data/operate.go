package data

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/ebuchman/go-shell-pipes"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Import(cmd *cobra.Command, args []string) {
	common.IfExit(checkServiceGiven(args))
	common.IfExit(ImportDataRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func Export(cmd *cobra.Command, args []string) {
	common.IfExit(checkServiceGiven(args))
	common.IfExit(ExportDataRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func ImportDataRaw(name string, verbose bool, w io.Writer) error {
	if parseKnown(name) {

		containerName := nameToContainerName(name)
		importPath := filepath.Join(common.DataContainersPath, name)

		// temp until docker cp works both ways.
		os.Chdir(importPath)
		cmd := "tar cf - . | docker run -i --rm --volumes-from " + containerName + " eris/data tar xf - -C /home/eris/.eris"

		s, err := pipes.RunString(cmd)
		if err != nil {
			return err
		}

		if verbose {
			w.Write([]byte(s))
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	return nil
}

func ExportDataRaw(name string, verbose bool, w io.Writer) error {
	if parseKnown(name) {
		if verbose {
			w.Write([]byte("Exporting data container" + name))
		}

		exportPath := filepath.Join(common.DataContainersPath, name)
		_, ops := mockService(name)
		service, exists := perform.ContainerExists(ops)

		if !exists {
			return fmt.Errorf("There is no data container for that service.")
		}
		if verbose {
			w.Write([]byte("Service ID: " + service.ID))
		}

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
			common.IfExit(util.DockerClient.CopyFromContainer(opts))
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
