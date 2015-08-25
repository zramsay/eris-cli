package initialize

func defAct() string {
	return `name = "do not use"
services = [ "ipfs" ]
chain = ""
steps = [
  "printenv",
  "echo hello",
  "echo goodbye"
]

[environment]
HELLO = "WORLD"

[maintainer]
name = "Eris Industries"
email = "support@erisindustries.com"

[location]
repository = "github.com/eris-ltd/eris-cli"

[machine]
include = ["docker"]
requires = [""]
`
}
