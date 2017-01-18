package initialize

import (
	"path"

	"github.com/eris-ltd/eris-cli/version"
)

func defServiceKeys() string {
	return `
# For more information on configurations, see the services specification:
# https://monax.io/docs/documentation/cli/latest/services_specification/

# These fields marshal roughly into the [docker run] command, see:
# https://docs.docker.com/engine/reference/run/

# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "keys"
description = """
Eris keys is meant for quick prototyping. You must replace it with a hardened key signing daemon to use in production. Eris does not intend to harden this for production, but rather will keep it as a rapid prototyping server.

This service is usually linked to a chain and/or an application. Its functionality is wrapped by [eris keys].
"""

status = "unfit for production"

[service]
image = "` + path.Join(version.DefaultRegistry, version.ImageKeys) + `"
data_container = true
ports = ["4767:4767"]
exec_host = "ERIS_KEYS_HOST"

[maintainer]
name = "Monax Industries"
email = "support@monax.io"

[location]
repository = "https://github.com/eris-ltd/eris-keys"
website = "https://monax.io/docs/documentation/keys"
`
}
