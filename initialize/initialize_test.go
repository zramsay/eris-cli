package initialize

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
)

const serviceToNeverUseToml = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

# These fields marshal roughly into the [docker run] command, see:
# https://docs.docker.com/engine/reference/run/

# For more information on configurations, see the services specification:
# https://monax.io/docs/documentation/cli/latest/services_specification/

name = "tester"

description = """
This is for testing only.
"""

status = "do not use"

[service]
image = "cats"
data_container = false
ports = ["12:12"]
user = ""
exec_host = ""
volumes = []

[maintainer]
name = "Marmotoshi"
email = "burrow@rollachain.now"`

const accountTypeToNeverUseToml = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "tester"

description = """
Use this account for all the people
"""

typical_user = """
Everyone.
"""

default_number = 4

default_tokens = 1234

default_bond = 0

[perms]
root = false
send = false
call = false
createContract = false
createAccount = false
bond = false
name = false
hasBase = true
setBase = true
unsetBase = true
setGlobal = true
hasRole = true
addRole = true
rmRole = true`

const chainTypeToNeverUseToml = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "tester"

description = """
This is a description.
"""

[account_types]
Full = 7
Developer = 8
Participant = 9
Root = 101
Validator = 2000

[servers]

[monaxmint]

[tendermint]`

func TestWriteServiceDefinitionFile(t *testing.T) {

	const serviceName = "tester"
	testFile := filepath.Join(config.ServicesPath, "tester.toml")

	serviceDefinition := definitions.BlankServiceDefinition()
	serviceDefinition.Name = "tester"
	serviceDefinition.Description = "This is for testing only."
	serviceDefinition.Status = "do not use"
	serviceDefinition.Service.Image = "cats"
	serviceDefinition.Service.AutoData = false
	serviceDefinition.Service.Ports = []string{`"12:12"`}
	serviceDefinition.Maintainer.Name = "Marmotoshi"
	serviceDefinition.Maintainer.Email = "burrow@rollachain.now"

	if err := WriteServiceDefinitionFile(serviceName, serviceDefinition); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile)

	fileBytes, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(fileBytes) != serviceToNeverUseToml {
		t.Fatalf("got %s, expected %s", string(fileBytes), serviceToNeverUseToml)
	}
}

func TestWriteAccountTypeDefinitionFile(t *testing.T) {

	const accountTypeName = "tester"
	testFile := filepath.Join(config.AccountsTypePath, "tester.toml")

	accountDefinition := definitions.BlankAccountType()
	accountDefinition.Name = "tester"
	accountDefinition.Description = "Use this account for all the people"
	accountDefinition.TypicalUser = "Everyone."

	accountDefinition.DefaultNumber = 4
	accountDefinition.DefaultTokens = 1234
	accountDefinition.DefaultBond = 0
	accountDefinition.Perms = map[string]bool{
		"root":           false,
		"send":           false,
		"call":           false,
		"createContract": false,
		"createAccount":  false,
		"bond":           false,
		"name":           false,
		"hasBase":        true,
		"setBase":        true,
		"unsetBase":      true,
		"setGlobal":      true,
		"hasRole":        true,
		"addRole":        true,
		"rmRole":         true,
	}

	if err := writeAccountTypeDefinitionFile(accountTypeName, accountDefinition); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile)

	fileBytes, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(fileBytes) != accountTypeToNeverUseToml {
		t.Fatalf("got %s, expected %s", string(fileBytes), accountTypeToNeverUseToml)
	}
}

func TestWriteChainTypeDefinitionFile(t *testing.T) {

	const chainTypeName = "tester"
	testFile := filepath.Join(config.ChainTypePath, "tester.toml")

	chainTypeDefinition := definitions.BlankChainType()
	chainTypeDefinition.Name = "tester"
	chainTypeDefinition.Description = "This is a description."
	chainTypeDefinition.AccountTypes = map[string]int64{
		"Full":        7,
		"Developer":   8,
		"Participant": 9,
		"Root":        101,
		"Validator":   2000,
	}

	if err := writeChainTypeDefinitionFile(chainTypeName, chainTypeDefinition); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile)

	fileBytes, err := ioutil.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(fileBytes) != chainTypeToNeverUseToml {
		t.Fatalf("got %s, expected %s", string(fileBytes), chainTypeToNeverUseToml)
	}
}
