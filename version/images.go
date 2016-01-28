package version

import (
	"fmt"
)

var (
	QUAY = "quay.io"
	HUB  = "" //dockerhub

	ERIS_IMG_BASE = "eris/base"
	ERIS_IMG_DATA = "eris/data"
	ERIS_IMG_KEYS = "eris/keys"
	ERIS_IMG_DB   = fmt.Sprintf("eris/erisdb:%s", VERSION)
	ERIS_IMG_PM   = fmt.Sprintf("eris/epm:%s", VERSION)
	ERIS_IMG_IPFS = "eris/ipfs"
)

func getReg(imgRaw string) string {
	return ""
}
