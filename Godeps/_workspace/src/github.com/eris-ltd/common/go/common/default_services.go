package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func InitDefaultServices(toPull, verbose bool) error {
	if toPull {
		if err := pullRepo("eris-services", ServicesPath, verbose); err != nil {
			if verbose {
				fmt.Println("Using default defs.")
			}
			if err2 := dropServiceDefaults(); err2 != nil {
				return fmt.Errorf("Cannot pull services: %s. %s.\n", err, err2)
			}
		} else {
			if err2 := pullRepo("eris-actions", ActionsPath, verbose); err2 != nil {
				return fmt.Errorf("Cannot pull actions: %s.\n", err2)
			}
		}
	} else {
		if err := dropServiceDefaults(); err != nil {
			return err
		}
	}
	return nil
}

func pullRepo(name, location string, verbose bool) error {
	src := ErisGH + name
	c := exec.Command("git", "clone", src, location)
	if verbose {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func dropServiceDefaults() error {
	if err := WriteFile(defKeys(), filepath.Join(ServicesPath, "keys.toml")); err != nil {
		return err
	}
	if err := WriteFile(defIpfs(), filepath.Join(ServicesPath, "ipfs.toml")); err != nil {
		return err
	}
	if err := WriteFile(defAct(), filepath.Join(ActionsPath, "do_not_use.toml")); err != nil {
		return err
	}
	return nil
}

func defKeys() string {
	return `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

[service]
name = "keys"

image = "eris/keys"
data_container = true
`
}

func defIpfs() string {
	return `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "ipfs"

[service]
name = "ipfs"
image = "eris/ipfs"
data_container = true
ports = ["4001:4001", "5001", "8080:8080"]
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

func defAct() string {
	return `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "do not use"
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
