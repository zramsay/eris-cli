package data

import (
	"fmt"
	"regexp"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func ListKnown(cmd *cobra.Command, args []string) {
	dataCont := ListKnownRaw()
	for _, s := range dataCont {
		fmt.Println(s)
	}
}

func Rename(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	if len(args) != 2 {
		fmt.Println("Please give me: eris data rename [oldName] [newName]")
		return
	}
	RenameDataRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
}

func Inspect(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	if len(args) == 1 {
		args = append(args, "all")
	}
	InspectDataRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
}

func Rm(cmd *cobra.Command, args []string) {
	checkServiceGiven(args)
	RmDataRaw(args[0], cmd.Flags().Lookup("verbose").Changed)
}

func ListKnownRaw() []string {
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

	return dataCont
}

func RenameDataRaw(oldName, newName string, verbose bool) {
	if parseKnown(oldName) {
		if verbose {
			fmt.Println("Renaming data container", oldName, "to", newName)
		}

		srv, ops := mockService(oldName)
		perform.DockerRename(srv, ops, oldName, newName, verbose)
	} else {
		if verbose {
			fmt.Println("I cannot find that data container. Please check the data container name you sent me.")
		}
	}
}

func InspectDataRaw(name, field string, verbose bool) {
	if parseKnown(name) {
		if verbose {
			fmt.Println("Inspecting data container", name)
		}

		srv, ops := mockService(name)
		perform.DockerInspect(srv, ops, field, verbose)
	} else {
		if verbose {
			fmt.Println("I cannot find that data container. Please check the data container name you sent me.")
		}
	}
}

func RmDataRaw(name string, verbose bool) {
	if parseKnown(name) {
		if verbose {
			fmt.Println("Removing data container", name)
		}

		srv, ops := mockService(name)
		perform.DockerRemove(srv, ops, verbose)
	} else {
		if verbose {
			fmt.Println("I cannot find that data container. Please check the data container name you sent me.")
		}
	}
}
