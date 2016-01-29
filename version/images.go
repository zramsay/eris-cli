package version

import (
	"fmt"
)

var (
	ERIS_REG_DEF = "quay.io"
	ERIS_REG_BAK = "" //dockerhub

	ERIS_IMG_BASE = "eris/base"
	ERIS_IMG_DATA = "eris/data"
	ERIS_IMG_KEYS = "eris/keys"
	ERIS_IMG_DB   = fmt.Sprintf("eris/erisdb:%s", VERSION)
	ERIS_IMG_PM   = fmt.Sprintf("eris/epm:%s", VERSION)
	ERIS_IMG_IPFS = "eris/ipfs"
)
