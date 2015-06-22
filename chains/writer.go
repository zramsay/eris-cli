package chains

import (
	"os"
	"path/filepath"

	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/BurntSushi/toml"
)

// if given empty string for fileName will use Service
// Definition Name
func WriteChainDefinitionFile(chainDef *def.Chain, fileName string) error {
	// writer := os.Stdout

	writer, err := os.Create(fileName)
	defer writer.Close()
	if err != nil {
		return err
	}

	if fileName == "" {
		fileName = chainDef.Name + ".toml"
	}

	switch filepath.Ext(fileName) {
	case ".toml":
		enc := toml.NewEncoder(writer)
		enc.Indent = ""
		writer.Write([]byte("name = \"" + chainDef.Name + "\"\n"))
		writer.Write([]byte("type = \"" + chainDef.Type + "\"\n"))
		writer.Write([]byte("\n[service]\n"))
		enc.Encode(chainDef.Service)
	}
	return nil
}
