// +build !arm

package version

import (
	"fmt"
)

const (
	DB_VERSION   = "0.17.0"
	KEYS_VERSION = "0.17.0"
	SOLC_VERSION = "stable"
)

var (
	DefaultRegistry = "quay.io"
	BackupRegistry  = ""

	ImageData = fmt.Sprintf("monax/data:%s", VERSION_MAJOR)
	ImageKeys = fmt.Sprintf("monax/keys:%s", KEYS_VERSION)
	ImageDB   = fmt.Sprintf("monax/db:%s", DB_VERSION)
	ImageSolc = fmt.Sprintf("ethereum/solc:%s", SOLC_VERSION)
)
