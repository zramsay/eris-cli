package actions

import (
	"fmt"
	"path/filepath"
	"strings"

	dir "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

func Get(args []string) {

}

func New(cmd *cobra.Command, args []string) {

}

func Add(args []string) {

}

func ListGlobal() {

}

func ListProject() {

}

func ListKnown() {
	actions := ListKnownRaw()
	for _, s := range actions {
		fmt.Println(strings.Replace(s, "_", " ", -1))
	}
}

func Edit(args []string) {

}

func Rename(args []string) {

}

func Remove(args []string) {

}

func ListKnownRaw() []string {
	acts := []string{}
	fileTypes := []string{}
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(dir.ActionsPath, t))
	}
	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			s1 = strings.Split(filepath.Base(s1), ".")[0]
			acts = append(acts, s1)
		}
	}
	return acts
}
