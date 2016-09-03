// +build !arm

package version

import (
	"fmt"
)

var (
	DefaultRegistry = "quay.io"
	BackupRegistry  = ""

	ImageData      = fmt.Sprintf("eris/data:%s", VERSION)
	ImageKeys      = fmt.Sprintf("eris/keys:%s", VERSION)
	ImageDB        = fmt.Sprintf("eris/erisdb:%s", VERSION)
	ImagePM        = fmt.Sprintf("eris/epm:%s", VERSION)
	ImageCM        = fmt.Sprintf("eris/eris-cm:%s", VERSION)
	ImageIPFS      = "eris/ipfs"
	ImageCompilers = fmt.Sprintf("eris/compilers:%s", VERSION)
)
