package initialize

import (
	"fmt"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

func DefaultKeys() string {
	return fmt.Sprintf(`[service]
name = "keys"
image = "eris/keys"
data_container = false
volumes = [ "%s:/home/eris/.eris/keys" ]
`, common.KeysPath)
}

func DefaultIpfs() string {
	return `name = "ipfs"

[service]
name = "ipfs"
image = "eris/ipfs"
data_container = true
ports = ["4001:4001", "5001:5001", "8080:8080"]
user = "root"

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"

[location]
repository = "github.com/eris-ltd/eris-services"

[machine]
include = ["docker"]
requires = [""]
`
}

func DefaultIpfs2() string {
	return `name = "ipfs"

[service]
name = "ipfs"
image = "eris/ipfs"
data_container = true
ports = ["4001:4001", "5001:5001", "8080:8080"]
user = "root"

[dependencies]
services = ["keys"]

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"

[location]
repository = "github.com/eris-ltd/eris-services"

[machine]
include = ["docker"]
requires = [""]
`
}
