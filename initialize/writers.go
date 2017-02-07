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

const serviceHeader = `
# For more information on configurations, see the services specification:
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

var serviceDefinitionTemplate *template.Template

func init() {
	var err error
	if serviceDefinitionTemplate, err = template.New("serviceDefinition").Parse(serviceDefinitionGeneral); err != nil {
		panic(err)
	}
}

func writeServiceDefinitionFile(name string) error {

	fileBytes, err := getServiceDefinitionFileBytes(name)
	if err != nil {
		return err
	}

	file := filepath.Join(config.ServicesPath, fmt.Sprintf("%s.toml", name))

	log.WithField("path", file).Debug("Saving File.")
	if err := config.WriteFile(string(fileBytes), file); err != nil {
		return err
	}
	return nil
}

func getServiceDefinitionFileBytes(name string) ([]byte, error) {

	serviceDefinition := defaultServices(name)

	var buffer bytes.Buffer

	// write copyright header
	buffer.WriteString(serviceHeader)

	// write section [service]
	if err := serviceDefinitionTemplate.Execute(&buffer, serviceDefinition); err != nil {
		return nil, fmt.Errorf("Failed to write service definition template %s", err)
	}
	return buffer.Bytes(), nil
}
