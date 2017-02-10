package initialize

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/eris-ltd/eris/config"
	//"github.com/eris-ltd/eris/definitions"
	"github.com/eris-ltd/eris/log"
)

// ------------------ services ------------------

const serviceHeader = `# For more information on configurations, see the services specification:
# https://monax.io/docs/documentation/cli/latest/services_specification/

# These fields marshal roughly into the [docker run] command, see:
# https://docs.docker.com/engine/reference/run/

# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml
`

const serviceDefinitionGeneral = `
name = "{{ .Name}}"

description = """
{{ .Description}}
"""

status = "{{ .Status}}"

[service]
image = "{{ .Service.Image}}"
data_container = {{ .Service.AutoData}}
ports = {{ .Service.Ports}}
exec_host = "{{ .Service.ExecHost}}"
volumes = {{ .Service.Volumes}}

[maintainer]
name = "{{ .Maintainer.Name}}"
email = "{{ .Maintainer.Email}}"
`

// ------------------ account types ------------------

const accountTypeHeader = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml
`

// [description] used to be [definition]
const accountTypeDefinitionGeneral = `
name = "{{ .Name}}"

description = """
{{ .Description}}
"""

typical_user = """
{{ .TypicalUser}}
"""

default_number = {{ .DefaultNumber}}

default_tokens = {{ .DefaultTokens}}

default_bond = {{ .DefaultBond}}

[perms]
root = {{index .Perms "root"}}
send = {{index .Perms "send"}}
call = {{index .Perms "call"}}
create_contract = {{index .Perms "create_contract"}}
create_account = {{index .Perms "create_account"}}
bond = {{index .Perms "bond"}}
name = {{index .Perms "name"}}
has_base = {{index .Perms "has_base"}}
set_base = {{index .Perms "set_base"}}
unset_base = {{index .Perms "unset_base"}}
set_global = {{index .Perms "set_global"}}
has_role = {{index .Perms "has_role"}}
add_role = {{index .Perms "add_role"}}
rm_role = {{index .Perms "rm_role"}}
`

// ------------------ chain types ------------------

const chainTypeHeader = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml
`

// [description] used to be [definition]
const chainTypeDefinitionGeneral = `
name = "{{ .Name}}"

description = """
{{ .Description}}
"""

[account_types]
Full = {{index .AccountTypes "Full"}}
Developer = {{index .AccountTypes "Developer"}}
Participant = {{index .AccountTypes "Participant"}}
Root = {{index .AccountTypes "Root"}}
Validator = {{index .AccountTypes "Validator"}}

[servers]

[erismint]

[tendermint]
`

var serviceDefinitionTemplate *template.Template
var accountTypeDefinitionTemplate *template.Template
var chainTypeDefinitionTemplate *template.Template

func init() {
	var err error
	if serviceDefinitionTemplate, err = template.New("serviceDefinition").Parse(serviceDefinitionGeneral); err != nil {
		panic(err)
	}
	if accountTypeDefinitionTemplate, err = template.New("accountTypeDefinition").Parse(accountTypeDefinitionGeneral); err != nil {
		panic(err)
	}
	if chainTypeDefinitionTemplate, err = template.New("chainTypeDefinition").Parse(chainTypeDefinitionGeneral); err != nil {
		panic(err)
	}
}

func writeServiceDefinitionFile(name string) error {

	serviceDefinition := defaultServices(name)

	var buffer bytes.Buffer

	// write copyright header
	buffer.WriteString(serviceHeader)

	// write main section
	if err := serviceDefinitionTemplate.Execute(&buffer, serviceDefinition); err != nil {
		return fmt.Errorf("Failed to write service definition template %s", err)
	}

	// create file
	file := filepath.Join(config.ServicesPath, fmt.Sprintf("%s.toml", name))

	// write file
	log.WithField("path", file).Debug("Saving File.")
	if err := config.WriteFile(buffer.String(), file); err != nil {
		return err
	}
	return nil
}

func writeAccountTypeDefinitionFile(name string) error {
	accountDefinition := defaultAccountTypes(name)

	var buffer bytes.Buffer

	// write copyright header
	buffer.WriteString(accountTypeHeader)

	// write section [service]
	if err := accountTypeDefinitionTemplate.Execute(&buffer, accountDefinition); err != nil {
		return fmt.Errorf("Failed to write account type definition template %s", err)
	}
	// create file
	file := filepath.Join(config.AccountsTypePath, fmt.Sprintf("%s.toml", name))

	// write file
	log.WithField("path", file).Debug("Saving File.")
	if err := config.WriteFile(buffer.String(), file); err != nil {
		return err
	}
	return nil
}

func writeChainTypeDefinitionFile(name string) error {
	chainDefinition := defaultChainTypes(name)

	var buffer bytes.Buffer

	// write copyright header
	buffer.WriteString(chainTypeHeader)

	// write main section
	if err := chainTypeDefinitionTemplate.Execute(&buffer, chainDefinition); err != nil {
		return fmt.Errorf("Failed to write chain type definition template %s", err)
	}
	// create file
	file := filepath.Join(config.ChainTypePath, fmt.Sprintf("%s.toml", name))

	// write file
	log.WithField("path", file).Debug("Saving File.")
	if err := config.WriteFile(buffer.String(), file); err != nil {
		return err
	}
	return nil
}
