package initialize

import (
	"fmt"
	"os"
	"path"

	"github.com/eris-ltd/eris/definitions"
	"github.com/eris-ltd/eris/version"
)

var ServiceDefinitions = []string{
	"compilers",
	"ipfs",
	"keys",
	// used by [eris chains start myChain --logrotate]
	// but its docker image is not pulled on [eris init]
	"logrotate",
}

var AccountTypeDefinitions = []string{
	"participant",
	"developer",
	"validator",
	"full",
	"root",
}

var ChainTypeDefinitions = []string{
	"simplechain",
	"sprawlchain",
	"adminchain",
	"demochain",
	"gochain",
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
		serviceDefinition.Service.Ports = []string{`"4767:4767"`} // XXX these exposed ports are a gaping security flaw

	case "compilers":

		serviceDefinition.Name = "compilers"
		serviceDefinition.Description = `Monax's Solidity Compiler Server.

This eris service compiles smart contract languages.`
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
		serviceDefinition.Description = `IPFS is The Permanent Web: A new peer-to-peer hypermedia protocol. IPFS uses content-based addressing versus http's location-based addressing.

This eris service is all but essential as part of the eris tool. The [eris files] relies upon this running service.`
		serviceDefinition.Status = "alpha"
		serviceDefinition.Service.Image = path.Join(version.DefaultRegistry, version.ImageIPFS)
		serviceDefinition.Service.AutoData = true
		serviceDefinition.Service.Ports = []string{`"4001:4001"`, `"5001:5001"`, `"` + port_to_use + `:` + port_to_use + `"`}

	case "logrotate":

		serviceDefinition.Name = "logrotate"
		serviceDefinition.Description = `Truncates docker container logs when the grow in size.

This eris service can also be run by adding the [--logrotate] flag on [eris chains start]

It is essential for long-running chain nodes.

Alternatively, use logspout to pipe logs to a service of you choosing"`
		serviceDefinition.Status = "ready"
		serviceDefinition.Service.Image = "tutum/logrotate"
		serviceDefinition.Service.AutoData = false
		serviceDefinition.Service.Volumes = []string{`"/var/lib/docker/containers:/var/lib/docker/containers:rw"`}

	default:
		panic(fmt.Errorf("not allowed"))
	}

	return serviceDefinition

}

func defaultAccountTypes(accountType string) *definitions.ErisDBAccountType {
	accountTypeDefinition := definitions.BlankAccountType()

	switch accountType {
	case "participant":

		accountTypeDefinition.Name = "Participant"
		accountTypeDefinition.Description = `Users who have a key which is registered with participant privileges can send
tokens; call contracts; and use the name registry`
		accountTypeDefinition.TypicalUser = `Generally the number of participants in your chain who do not need elevated
privileges should be given these keys.

Usually this group will have the most number of keys of all of the groups.`
		accountTypeDefinition.DefaultNumber = 25
		accountTypeDefinition.DefaultTokens = 9999999999
		accountTypeDefinition.DefaultBond = 0
		accountTypeDefinition.Perms = map[string]int{
			"root":            0,
			"send":            1,
			"call":            1,
			"create_contract": 0,
			"create_account":  0,
			"bond":            0,
			"name":            1,
			"has_base":        0,
			"set_base":        0,
			"unset_base":      0,
			"set_global":      0,
			"has_role":        1,
			"add_role":        0,
			"rm_role":         0,
		}

	case "developer":

		accountTypeDefinition.Name = "Developer"
		accountTypeDefinition.Description = `Users who have a key which is registered with developer privileges can send
tokens; call contracts; create contracts; create accounts; use the name registry;
and modify account's roles.`
		accountTypeDefinition.TypicalUser = `Generally the development team seeking to build the application on top of the
given chain would be within the group. If this is a multi organizational
chain then developers from each of the stakeholders should generally be registered
within this group, although that design is up to you.`
		accountTypeDefinition.DefaultNumber = 6
		accountTypeDefinition.DefaultTokens = 9999999999
		accountTypeDefinition.DefaultBond = 0
		accountTypeDefinition.Perms = map[string]int{
			"root":            0,
			"send":            1,
			"call":            1,
			"create_contract": 1,
			"create_account":  1,
			"bond":            0,
			"name":            1,
			"has_base":        0,
			"set_base":        0,
			"unset_base":      0,
			"set_global":      0,
			"has_role":        1,
			"add_role":        1,
			"rm_role":         1,
		}

	case "validator":

		accountTypeDefinition.Name = "Validator"
		accountTypeDefinition.Description = `Users who have a key which is registered with validator privileges can
only post a bond and begin validating the chain. This is the only privilege
this account group gets.`
		accountTypeDefinition.TypicalUser = `Generally the marmots recommend that you put your validator nodes onto a cloud
(IaaS) provider so that they will be always running.

We also recommend that if you are in a multi organizational chain then you would
have some separation of the validators to be ran by the different organizations
in the system.`
		accountTypeDefinition.DefaultNumber = 7
		accountTypeDefinition.DefaultTokens = 9999999999
		accountTypeDefinition.DefaultBond = 9999999998
		accountTypeDefinition.Perms = map[string]int{
			"root":            0,
			"send":            0,
			"call":            0,
			"create_contract": 0,
			"create_account":  0,
			"bond":            1,
			"name":            0,
			"has_base":        0,
			"set_base":        0,
			"unset_base":      0,
			"set_global":      0,
			"has_role":        0,
			"add_role":        0,
			"rm_role":         0,
		}

	case "full":

		accountTypeDefinition.Name = "Full"
		accountTypeDefinition.Description = `Users who have a key which is registered with root privileges can do everything
on the chain. They have all of the permissions possible. These users are also
bonded at the genesis time, so these should be used only for simple chains with
a few nodes who will be on during the prototyping session.`
		accountTypeDefinition.TypicalUser = `If you are making a small chain just to play around then usually you would
give all of the accounts needed for your experiment full accounts.

If you are making a more complex chain, don't use this account type.`
		accountTypeDefinition.DefaultNumber = 1
		accountTypeDefinition.DefaultTokens = 99999999999999
		accountTypeDefinition.DefaultBond = 9999999999
		accountTypeDefinition.Perms = map[string]int{
			"root":            1,
			"send":            1,
			"call":            1,
			"create_contract": 1,
			"create_account":  1,
			"bond":            1,
			"name":            1,
			"has_base":        1,
			"set_base":        1,
			"unset_base":      1,
			"set_global":      1,
			"has_role":        1,
			"add_role":        1,
			"rm_role":         1,
		}

	case "root":

		accountTypeDefinition.Name = "Root"
		accountTypeDefinition.Description = `Users who have a key which is registered with root privileges can do everything
on the chain. They have all of the permissions possible.`
		accountTypeDefinition.TypicalUser = `If you are making a small chain just to play around then usually you would
give all of the accounts needed for your experiment root privileges (unless you
were testing different) privilege types.

If you are making a more complex chain, then you would usually have a few
keys which have registered root permissions and as such will act in a capacity
similar to a network administrator in other data management situations.`
		accountTypeDefinition.DefaultNumber = 3
		accountTypeDefinition.DefaultTokens = 9999999999
		accountTypeDefinition.DefaultBond = 0
		accountTypeDefinition.Perms = map[string]int{
			"root":            1,
			"send":            1,
			"call":            1,
			"create_contract": 1,
			"create_account":  1,
			"bond":            1,
			"name":            1,
			"has_base":        1,
			"set_base":        1,
			"unset_base":      1,
			"set_global":      1,
			"has_role":        1,
			"add_role":        1,
			"rm_role":         1,
		}

	default:
		panic(fmt.Errorf("not allowed"))
	}

	return accountTypeDefinition
}

func defaultChainTypes(chainType string) {

}
