package services

import (
	"os"
	"path/filepath"

	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
)

// if given empty string for fileName will use Service
// Definition Name
func WriteServiceDefinitionFile(serviceDef *def.ServiceDefinition, fileName string) error {
	// writer := os.Stdout

	writer, err := os.Create(fileName)
	defer writer.Close()
	if err != nil {
		return err
	}

	if fileName == "" {
		fileName = serviceDef.Service.Name + ".toml"
	}

	switch filepath.Ext(fileName) {
	case ".toml":
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
