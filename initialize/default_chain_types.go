package initialize

func defaultSimpleChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "simplechain"

definition = """
A simple chain type will build a single node chain. This chain type is useful
for quick and easy prototyping. It should not be used for anything more than
the most simple prototyping as it only has one node and the keys to that node
could get lost or compromised and the chain would thereafter become useless.
"""

[account_types]
Full = 1
Developer = 0
Participant = 0
Root = 0
Validator = 0

[servers]

[erismint]

[tendermint]
`
}

func defaultAdminChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "adminchain"

definition = """
An adminchain type has settings for prototyping a larger chain from a sysadmin point of view. With four Validator and three Full account_types, at minimum of five nodes must be up for consensus to happen. This account combination is what we use to test long running chains on our CI system.
"""

[account_types]
Full = 3
Developer = 1
Participant = 1
Root = 1
Validator = 4

[servers]

[erismint]

[tendermint]
`
}

func defaultDemoChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "demochain"

definition = """
A demo chain is useful for setting up proof of concept demonstration chains. It is a quick and easy way to have multi-validator, multi-developer, multi-participant chains set up for your application. This chain will tolerate 2 validators becoming byzantine or going off-line while still moving forward. You should utilize 7 different cloud instances and deploy one of the validator chain directories to each.
"""

[account_types]
Full = 0
Developer = 5
Participant = 20
Root = 3
Validator = 7

[servers]

[erismint]

[tendermint]
`
}

func defaultGoChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "gochain"

definition = """
A gochain type will build a three node chain. It is a quick and easy way to get started with a multi-validator chain. The Full account_type includes validator and deploy permissions, allowing for experimentation with setting up a chain and halting it by taking down a single node. This Full account should be deployed on your local machine and cloud nodes should have only Validator accounts. Use for prototyping only.
"""

[account_types]
Full = 1
Developer = 0
Participant = 0
Root = 0
Validator = 2

[servers]

[erismint]

[tendermint]
`
}

func defaultSprawlChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "sprawlchain"

definition = """
A sprawlchain type has a little bit of everything. Modify as necessary for your ecosystem application. Will tolerate three nodes down. As with other chains, Validator accounts ought to go on cloud. No Full accounts are provided since these are prefered for quick development only.
"""

[account_types]
Full = 0
Developer = 10
Participant = 20 
Root = 3
Validator = 10

[servers]

[erismint]

[tendermint]
`
}
