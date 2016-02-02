package keys

import (
	"path"
	"path/filepath"

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
	do.Operations.ContainerNumber = 1
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}
	//destination on host -> given as flag default
	//if do.Destination == "" {
	//	do.Destination = filepath.Join(KeysPath, "data")
	//}

	//src in container
	do.Source = path.Join(ErisContainerRoot, "keys", "data", do.Address)
	if err := data.ExportData(do); err != nil {
		return err
	}
	return nil
}

func ImportKey(do *definitions.Do) error {

	do.Name = "keys"
	do.Operations.ContainerNumber = 1
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}

	do.Operations.Interactive = false
	dir := path.Join(ErisContainerRoot, "keys", "data", do.Address)
	do.Operations.Args = []string{"mkdir", dir} //need to mkdir for import
	if err := srv.ExecService(do); err != nil {
		return err
	}
	//src on host
	//if default given (from flag), join addrs
	if do.Source == filepath.Join(KeysPath, "data") {
		do.Source = filepath.Join(KeysPath, "data", do.Address, do.Address)
	}
	//dest in container
	do.Destination = dir

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
