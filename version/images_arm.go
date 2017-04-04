package version

import (
	"fmt"
)

const ARCH = "arm"

var (
	DefaultRegistry = "quay.io"
	BackupRegistry  = ""

	ImageData      = fmt.Sprintf("monax/data:%s-%s", ARCH, VERSION)
	ImageKeys      = fmt.Sprintf("monax/keys:%s-%s", ARCH, VERSION)
	ImageDB        = fmt.Sprintf("monax/db:%s-%s", ARCH, VERSION)
	ImageCompilers = fmt.Sprintf("monax/compilers:%s-%s", ARCH, VERSION)
	ImageIPFS      = fmt.Sprintf("monax/ipfs:%s", ARCH)
)
