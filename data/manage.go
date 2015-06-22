package data

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/eris-ltd/eris-cli/perform"
	"github.com/eris-ltd/eris-cli/util"

	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Import(args []string) {

}

func ListKnown(args []string) {
	dataCont := ListKnownRaw()
	for _, s := range dataCont {
		fmt.Println(s)
	}
}

func Rename(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		fmt.Println("Please give me: eris data rename [oldName] [newName]")
		return
	}
	RenameDataRaw(args[0], args[1], cmd.Flags().Lookup("verbose").Changed)
}

func Export(args []string) {

}

func Rm(cmd *cobra.Command, args []string) {
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

		srv := &def.Service{}
		ops := &def.ServiceOperation{}
		containerNumber := 1 // tmp
		ops.SrvContainerName = "eris_data_" + oldName + "_" + strconv.Itoa(containerNumber)
		perform.DockerRename(srv, ops, oldName, newName, verbose)
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

		srv := &def.Service{}
		ops := &def.ServiceOperation{}
		containerNumber := 1 // tmp
		ops.SrvContainerName = "eris_data_" + name + "_" + strconv.Itoa(containerNumber)
		perform.DockerRemove(srv, ops, verbose)
	} else {
		if verbose {
			fmt.Println("I cannot find that data container. Please check the data container name you sent me.")
		}
	}
}

func parseKnown(name string) bool {
	known := ListKnownRaw()
	if len(known) != 0 {
		for _, srv := range known {
			if srv == name {
				return true
			}
		}
	}
	return false
}
