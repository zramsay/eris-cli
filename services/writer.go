package services

import (
	"encoding/json"
	"os"
	"path/filepath"

	def "github.com/eris-ltd/eris-cli/definitions"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

// if given empty string for fileName will use Service
// Definition Name
func WriteServiceDefinitionFile(serviceDef *def.ServiceDefinition, fileName string) error {
	// writer := os.Stdout

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
			enc := toml.NewEncoder(writer)
			enc.Indent = ""
			writer.Write([]byte("[service]\n"))
			enc.Encode(serviceDef.Service)
			writer.Write([]byte("\n[maintainer]\n"))
			enc.Encode(serviceDef.Maintainer)
			writer.Write([]byte("\n[location]\n"))
			enc.Encode(serviceDef.Location)
			writer.Write([]byte("\n[machine]\n"))
			enc.Encode(serviceDef.Machine)
	}
	return nil
}
