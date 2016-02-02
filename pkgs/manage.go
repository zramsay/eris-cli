package pkgs

import (
	"github.com/eris-ltd/eris-cli/definitions"
)

// TODO: finish when the PR which is blocking
//   eris files put --dir is integrated into
//   ipfs
func GetPackage(do *definitions.Do) error {
	// do.Name = args[0]
	// do.Path = args[1]

	return nil
}

// TODO: finish when the PR which is blocking
//   eris files put --dir is integrated into
//   ipfs
func PutPackage(do *definitions.Do) error {
	// do.Name = args[0]
	var hash string = ""

	do.Result = hash
	return nil
}
