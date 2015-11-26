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

func ExportKey(do *definitions.Do) error {

	do.Name = "keys"
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}
	//destination on host
	if do.Destination == "" {
		do.Destination = path.Join(KeysPath, "data")
	}
	//src in container
	if do.Address != "" {
		do.Source = path.Join(ErisContainerRoot, "keys", "data", do.Address, do.Address)
	} else {
		do.Source = path.Join(ErisContainerRoot, "keys", "data")
	}

	if err := data.ExportData(do); err != nil {
		return err
	}
	return nil
}

func ImportKey(do *definitions.Do) error {

	do.Name = "keys"
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}
	if do.Address != "" {
		do.Operations.Interactive = false
		dir := path.Join(ErisContainerRoot, "keys", "data", do.Address)
		do.Operations.Args = []string{"mkdir", dir} //need to mkdir for import
		if err := srv.ExecService(do); err != nil {
			return err
		}
		//src on host
		//dest in container
		do.Source = path.Join(KeysPath, "data", do.Address, do.Address)
		do.Destination = dir

	} else {
		do.Source = path.Join(KeysPath, "data")
		do.Destination = path.Join(ErisContainerRoot, "keys", "data")
	}
	//TODO [zr] src flag

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
