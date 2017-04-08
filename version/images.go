// +build !arm

package version

import (
	"fmt"
)

const (
	DB_VERSION        = "0.16.2"
	KEYS_VERSION      = "0.16.0"
	COMPILERS_VERSION = "0.16.0"
)

var (
	DefaultRegistry = "quay.io"
	BackupRegistry  = ""

	ImageData      = fmt.Sprintf("monax/data:%s", VERSION_MAJOR)
	ImageKeys      = fmt.Sprintf("monax/keys:%s", KEYS_VERSION)
	ImageDB        = fmt.Sprintf("monax/db:%s", DB_VERSION)
	ImageIPFS      = "ipfs/go-ipfs"
	ImageCompilers = fmt.Sprintf("monax/compilers:%s", COMPILERS_VERSION)
)
