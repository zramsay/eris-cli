package initialize

func DefaultKeys() string {
  return `[service]
name = "keys"

image = "eris/keys"
data_container = true
`
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

[services]
dependencies = ["keys"]

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
