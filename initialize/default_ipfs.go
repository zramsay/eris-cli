package initialize

import (
	"fmt"
	"os"
	"path"

	"github.com/eris-ltd/eris-cli/version"
)

func defServiceIPFS() string {
	port_to_use := os.Getenv("ERIS_CLI_TESTS_PORT")
	if port_to_use == "" {
		port_to_use = "8080"
	}
	return `
# For more information on configurations, see the services specification:
# https://docs.erisindustries.com/documentation/eris-cli/latest/services_specification/

# These fields marshal roughly into the [docker run] command, see:
# https://docs.docker.com/engine/reference/run/

# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "ipfs"
description = """
IPFS is The Permanent Web: A new peer-to-peer hypermedia protocol. IPFS uses content-based addressing versus http's location-based addressing.

This eris service is all but essential as part of the eris tool. The [eris files] command (and several other eris subcommands) require IPFS to be running. The toadserver is an example of IPFS + the name registry feature of erisdb.
"""

status = "alpha"

[service]
image = "` + path.Join(version.ERIS_REG_DEF, version.ERIS_IMG_IPFS) + fmt.Sprintf(`"
data_container = true
ports = ["4001:4001", "5001:5001", "%s:%s"]
user = "root"
exec_host = "ERIS_IPFS_HOST"

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"

[location]
dockerfile = "https://github.com/eris-ltd/eris-cli/blob/master/docker/x86/ipfs/Dockerfile"
repository = "https://github.com/ipfs/go-ipfs"
website = "https://ipfs.io/"
`, port_to_use, port_to_use)
}
