package initialize

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
)

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

func WriteServiceDefinitionFile(name string, serviceDefinition *definitions.ServiceDefinition) error {

	var buffer bytes.Buffer

	// Write toml header.
	buffer.WriteString(tomlHeader)

	// Write service header.
	buffer.WriteString(serviceHeader)

	// Write main section.
	if err := serviceDefinitionTemplate.Execute(&buffer, serviceDefinition); err != nil {
		return fmt.Errorf("Failed to write service definition template %s", err)
	}

	// Name the file.
	file := filepath.Join(config.ServicesPath, fmt.Sprintf("%s.toml", name))

	// Write the file.
	return ioutil.WriteFile(file, buffer.Bytes(), 0644)
}

func writeAccountTypeDefinitionFile(name string, accountDefinition *definitions.MonaxDBAccountType) error {

	var buffer bytes.Buffer

	buffer.WriteString(tomlHeader)

	if err := accountTypeDefinitionTemplate.Execute(&buffer, accountDefinition); err != nil {
		return fmt.Errorf("Failed to write account type definition template %s", err)
	}

	file := filepath.Join(config.AccountsTypePath, fmt.Sprintf("%s.toml", name))

	return ioutil.WriteFile(file, buffer.Bytes(), 0644)
}

func writeChainTypeDefinitionFile(name string, chainDefinition *definitions.ChainType) error {

	var buffer bytes.Buffer

	buffer.WriteString(tomlHeader)

	if err := chainTypeDefinitionTemplate.Execute(&buffer, chainDefinition); err != nil {
		return fmt.Errorf("Failed to write chain type definition template %s", err)
	}

	file := filepath.Join(config.ChainTypePath, fmt.Sprintf("%s.toml", name))

	return ioutil.WriteFile(file, buffer.Bytes(), 0644)
}
