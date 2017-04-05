// +build !arm

package version

import (
	"fmt"
)

var (
	DefaultRegistry = "quay.io"
	BackupRegistry  = ""

	ImageData      = fmt.Sprintf("monax/data:%s", VERSION_MAJOR)
	ImageKeys      = fmt.Sprintf("monax/keys:%s", VERSION)
	ImageDB        = fmt.Sprintf("monax/db:%s", VERSION)
	ImageIPFS      = "ipfs/go-ipfs"
	ImageCompilers = fmt.Sprintf("monax/compilers:%s", VERSION)
)
