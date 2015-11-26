package keys

import (
	"path"

	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func GenerateKey(do *definitions.Do) error {

	do.Name = "keys"
	do.Operations.ContainerNumber = 1

	if err := srv.EnsureRunning(do); err != nil {
		return err
	}
	do.Operations.Interactive = false
	do.Operations.Args = []string{"eris-keys", "gen", "--no-pass"}

	if err := srv.ExecService(do); err != nil {
		return err
	}

	return nil
}

func GetPubKey(do *definitions.Do) error {

	do.Name = "keys"
	do.Operations.ContainerNumber = 1
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}
	do.Operations.Interactive = false
	do.Operations.Args = []string{"eris-keys", "pub", "--addr", do.Address}

	if err := srv.ExecService(do); err != nil {
		return err
	}
	return nil
}

//from /home/eris/.eris/keys/data/ to /home/user/.eris/keys/data/
func ExportKey(do *definitions.Do) error {

	do.Name = "keys" //for cont as well as path-joined for final dir
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}
	//destination on host
	if do.Destination == "" {
		do.Destination = KeysPath
	}
	//src in container (hardcoded)
	do.Source = path.Join(ErisContainerRoot, "keys", "data")

	if err := data.ExportData(do); err != nil {
		return err
	}
	return nil
}

func ImportKey(do *definitions.Do) error {

	do.Name = "keys" //for cont as well as path-joined for final dir
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}
	//destination in container (harcode; same as do.Source from Export)
	do.Destination = path.Join(ErisContainerRoot, "keys", "data")
	//src on host
	if do.Source == "" {
		do.Source = KeysPath
	}

	if err := data.ImportData(do); err != nil {
		return err
	}
	return nil
}

func ConvertKey(do *definitions.Do) error {

	do.Name = "keys"
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}

	do.Operations.Args = []string{"mintkey", "mint", do.Address}
	if err := srv.ExecService(do); err != nil {
		return err
	}
	return nil
}
