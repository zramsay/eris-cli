package initialize

import (
	"path"

	"github.com/eris-ltd/eris/version"
)

func defServiceCompilers() string {
	return `
# For more information on configurations, see the services specification:
# https://monax.io/docs/documentation/cli/latest/specifications/services_specification/

# These fields marshal roughly into the [docker run] command, see:
# https://docs.docker.com/engine/reference/run/

# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name           = "compilers"
description = """
Monax's Solidity Compiler Server.

This eris service compiles smart contract languages.
"""

status = "beta"

[service]
image          = "` + path.Join(version.DefaultRegistry, version.ImageCompilers) + `"
data_container = true
ports          = [  ]
volumes        = [  ]
environment    = [  ]

[maintainer]
name = "Monax Industries"
email = "support@monax.io"

[location]
repository = "https://github.com/eris-ltd/eris-compilers"
website = "https://monax.io/docs/documentation/compilers"
`
}
