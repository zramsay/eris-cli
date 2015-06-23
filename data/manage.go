package data

import (
	"io"
	"fmt"
	"os"
	"regexp"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func ListKnown(cmd *cobra.Command, args []string) {
	dataCont, err := ListKnownRaw()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, s := range dataCont {
		fmt.Println(s)
	}
}

func Rename(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	if len(args) != 2 {
		fmt.Println("Please give me: eris data rename [oldName] [newName]")
		return
	}
	IfExit(RenameDataRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func Inspect(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	if len(args) == 1 {
		args = append(args, "all")
	}
	IfExit(InspectDataRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func Rm(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(RmDataRaw(args[0], cmd.Flags().Lookup("verbose").Changed, os.Stdout))
}

func ListKnownRaw() ([]string, error) {
	dataCont := []string{}
	r := regexp.MustCompile(`\/eris_data_(.+)_\d`)

	contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	for _, con := range contns {
		for _, c := range con.Names {
			match := r.FindAllStringSubmatch(c, 1)
			if len(match) != 0 {
				dataCont = append(dataCont, r.FindAllStringSubmatch(c, 1)[0][1])
			}
		}
	}

	return dataCont, nil
}

func RenameDataRaw(oldName, newName string, verbose bool, w io.Writer) error {
	if parseKnown(oldName) {
		if verbose {
			w.Write([]byte("Renaming data container" + oldName + "to" + newName))
		}

		srv, ops := mockService(oldName)
		err := perform.DockerRename(srv, ops, oldName, newName, verbose, w)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func InspectDataRaw(name, field string, verbose bool, w io.Writer) error {
	if parseKnown(name) {
		if verbose {
			w.Write([]byte("Inspecting data container" + name))
		}

		srv, ops := mockService(name)
		err := perform.DockerInspect(srv, ops, field, verbose, w)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}
	return nil
}

func RmDataRaw(name string, verbose bool, w io.Writer) error {
	if parseKnown(name) {
		if verbose {
			w.Write([]byte("Removing data container" + name))
		}

		srv, ops := mockService(name)
		err := perform.DockerRemove(srv, ops, verbose, w)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("I cannot find that data container. Please check the data container name you sent me.")
	}

	return nil
}
