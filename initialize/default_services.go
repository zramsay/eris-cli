package initialize

import (
	"fmt"
	"os"
	"path"

	"github.com/eris-ltd/eris/definitions"
	"github.com/eris-ltd/eris/version"
)

var SERVICE_DEFINITIONS = []string{
	"compilers",
	"ipfs",
	"keys",
	// used by [eris chains start myChain --logrotate]
	// but its docker image is not pulled on [eris init]
	"logrotate",
}

func defaultServices(service string) *definitions.ServiceDefinition {
	serviceDefinition := definitions.BlankServiceDefinition()

	serviceDefinition.Maintainer.Name = "Monax Industries"
	serviceDefinition.Maintainer.Email = "support@monax.io"

	switch service {
	case "keys":

		serviceDefinition.Name = "keys"
		serviceDefinition.Description = `Eris keys is meant for quick prototyping. You must replace it with a hardened key signing daemon to use in production. Eris does not intend to harden this for production, but rather will keep it as a rapid prototyping server.

This service is usually linked to a chain and/or an application. Its functionality is wrapped by [eris keys].`
		serviceDefinition.Status = "unfit for production"

		serviceDefinition.Service.Image = path.Join(version.DefaultRegistry, version.ImageKeys)

		serviceDefinition.Service.AutoData = true
		// XXX these exposed ports are dangerous
		// and a gaping security flaw
		serviceDefinition.Service.Ports = []string{`"4767:4767"`}

	case "compilers":
		serviceDefinition.Name = "compilers"
		serviceDefinition.Description = `
Monax's Solidity Compiler Server.

This eris service compiles smart contract languages.
`
		serviceDefinition.Status = "beta"

		serviceDefinition.Service.Image = path.Join(version.DefaultRegistry, version.ImageCompilers)

		serviceDefinition.Service.AutoData = true
		//serviceDefinition.Service.Ports = []string{`"9090:9090"`}
	case "ipfs":
		port_to_use := os.Getenv("ERIS_CLI_TESTS_PORT")
		if port_to_use == "" {
			port_to_use = "8080"
		}
		serviceDefinition.Name = "ipfs"
		serviceDefinition.Description = `
IPFS is The Permanent Web: A new peer-to-peer hypermedia protocol. IPFS uses content-based addressing versus http's location-based addressing.

This eris service is all but essential as part of the eris tool. The [eris files] relies upon this running service.
`
		serviceDefinition.Status = "alpha"

		serviceDefinition.Service.Image = path.Join(version.DefaultRegistry, version.ImageIPFS)

		serviceDefinition.Service.AutoData = true
		serviceDefinition.Service.Ports = []string{`"4001:4001"`, `"5001:5001"`, `"` + port_to_use + `:` + port_to_use + `"`}
	case "logrotate":
		serviceDefinition.Name = "logrotate"
		serviceDefinition.Description = `
Truncates docker container logs when the grow in size.

This eris service can also be run by adding the [--logrotate] flag on [eris chains start]

It is essential for long-running chain nodes.

Alternatively, use logspout to pipe logs to a service of you choosing"
`
		serviceDefinition.Status = "ready"

		serviceDefinition.Service.Image = "tutum/logrotate"

		serviceDefinition.Service.AutoData = false
		serviceDefinition.Service.Volumes = []string{`"/var/lib/docker/containers:/var/lib/docker/containers:rw"`}
	default:
		panic(fmt.Errorf("not allowed"))
	}
	//serviceDefinition.Location.Repository = "https://github.com/eris-ltd/eris-keys"
	//serviceDefinition.Location.Website = "https://monax.io/docs/documentation/keys"
	return serviceDefinition

}
