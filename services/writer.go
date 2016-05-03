package services

import (
	"encoding/json"
	"os"
	"path/filepath"

	def "github.com/eris-ltd/eris-cli/definitions"

	. "github.com/eris-ltd/common/go/common"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

// if given empty string for fileName will use Service
// Definition Name
// TODO return ErisError? maybe not
func WriteServiceDefinitionFile(serviceDef *def.ServiceDefinition, fileName string) error {

	if filepath.Ext(fileName) == "" {
		fileName = serviceDef.Service.Name + ".toml"
		fileName = filepath.Join(ServicesPath, fileName)
	}

	writer, err := os.Create(fileName)
	defer writer.Close()
	if err != nil {
		return err
	}

	switch filepath.Ext(fileName) {
	case ".json":
		mar, err := json.MarshalIndent(serviceDef, "", "  ")
		if err != nil {
			return err
		}
		mar = append(mar, '\n')
		writer.Write(mar)
	case ".yaml":
		mar, err := yaml.Marshal(serviceDef)
		if err != nil {
			return err
		}
		mar = append(mar, '\n')
		writer.Write(mar)
	default:
		WriteDefaultServiceTOML(writer, serviceDef)
	}
	return nil
}

func WriteDefaultServiceTOML(writer *os.File, serviceDef *def.ServiceDefinition) {

	writer.Write([]byte("# This is a TOML config file.\n# For more information, see https://github.com/toml-lang/toml\n\n"))
	enc := toml.NewEncoder(writer)
	enc.Indent = ""
	writer.Write([]byte("name = \"" + serviceDef.Name + "\"\n\n"))
	if serviceDef.ServiceID != "" {
		writer.Write([]byte("service_id = \"" + serviceDef.ServiceID + "\"\n"))
	}
	if serviceDef.Chain != "" {
		writer.Write([]byte("chain = \"" + serviceDef.Chain + "\"\n\n"))
	}

	writer.Write([]byte("description = \"\"\"\n" + "# describe your service" + "\n\"\"\"\n\n"))
	writer.Write([]byte("status = \"\"" + " # alpha, beta, ready" + "\n\n"))
	writer.Write([]byte("[service]\n"))
	enc.Encode(serviceDef.Service)
	writer.Write([]byte("\n"))
	writer.Write([]byte("[dependencies]\n"))
	if serviceDef.Dependencies != nil {
		if len(serviceDef.Dependencies.Services) != 0 || len(serviceDef.Dependencies.Chains) != 0 {
			enc.Encode(serviceDef.Dependencies)
		}
	}
	writer.Write([]byte("\n[maintainer]\n"))
	enc.Encode(serviceDef.Maintainer)
	writer.Write([]byte("\n[location]\n"))
	enc.Encode(serviceDef.Location)
	writer.Write([]byte("dockerfile = \"\"\n"))
	writer.Write([]byte("repository = \"\"\n"))
	writer.Write([]byte("website = \"\"\n"))
}
