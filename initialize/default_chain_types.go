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

[messenger]

[manager]

[consensus]
`
}

func defaultAdminChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "adminchain"

definition = """
An adminchain type has settings for prototyping a larger chain from a sysadmin point of view. With four Validator and three Full account_types, at minimum of five nodes must be up for consensus to happen.
"""

[account_types]
Full = 3
Developer = 1
Participant = 1
Root = 1
Validator = 4

[messenger]

[manager]

[consensus]
`
}

func defaultDemoChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "demochain"

definition = """
A demochain type has one node for setting up a simplechain but comes with additional Developer and Participant accounts for demonstrating a typical application. Use for prototyping only.
"""

[account_types]
Full = 1
Developer = 5
Participant = 5
Root = 0
Validator = 0

[messenger]

[manager]

[consensus]
`
}

func defaultGoChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "gochain"

definition = """
A gochain type will build a three node chain. It is a quick andeasy way to get started with a multi-validator chain. The Full account_type includes validator and deploy permissions, allowing for experimentation with setting up a chain and halting it by taking down a single node. Use for prototyping only.
"""

[account_types]
Full = 3
Developer = 0
Participant = 0
Root = 0
Validator = 0

[messenger]

[manager]

[consensus]
`
}

func defaultSprawlChainType() string {
	return `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "sprawlchain"

definition = """
A sprawlchain type has a little bit of everything. Modify as necessary for your ecosystem application. Will tolerate three nodes down.
"""

[account_types]
Full = 5
Developer = 10
Participant = 20 
Root = 3
Validator = 5

[messenger]

[manager]

[consensus]
`
}
