package version

import (
	"fmt"
)

const ARCH = "arm"

var (
	DefaultRegistry = "quay.io"
	BackupRegistry  = ""

	ImageData      = fmt.Sprintf("eris/data:%s-%s", ARCH, VERSION)
	ImageKeys      = fmt.Sprintf("eris/keys:%s-%s", ARCH, VERSION)
	ImageDB        = fmt.Sprintf("eris/db:%s-%s", ARCH, VERSION)
	ImagePM        = fmt.Sprintf("eris/pm:%s-%s", ARCH, VERSION)
	ImageCompilers = fmt.Sprintf("eris/compilers:%s-%s", ARCH, VERSION)
	ImageIPFS      = fmt.Sprintf("eris/ipfs:%s", ARCH)
)
