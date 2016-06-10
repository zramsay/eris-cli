// +build !arm

package version

import (
	"fmt"
)

var (
	ERIS_REG_DEF = "quay.io"
	ERIS_REG_BAK = "" //dockerhub

	//ERIS_IMG_BASE = "eris/base" // only needed for [eris update]
	ERIS_IMG_DATA = fmt.Sprintf("eris/data:%s", VERSION)
	ERIS_IMG_KEYS = fmt.Sprintf("eris/keys:%s", VERSION)
	ERIS_IMG_DB   = fmt.Sprintf("eris/erisdb:%s", VERSION)
	ERIS_IMG_PM   = fmt.Sprintf("eris/epm:%s", VERSION)
	ERIS_IMG_CM   = fmt.Sprintf("eris/eris-cm:%s", VERSION)
	ERIS_IMG_IPFS = "eris/ipfs"
)
